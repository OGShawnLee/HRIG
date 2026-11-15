package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/wailsapp/wails/v3/pkg/application"

	"hrig/shared/model"
)

const (
	defaultAPIBaseURL = "http://127.0.0.1:80"
	defaultHTTPTimeout = 8 * time.Second
	maxHTTPBodyBytes   = 4 << 20
	maxWSMessageBytes  = 4 << 20
)

var wsBackoffSteps = []time.Duration{
	1 * time.Second,
	2 * time.Second,
	5 * time.Second,
}

func init() {
	application.RegisterEvent[model.RealtimeEnvelope]("telemetry:envelope")
	application.RegisterEvent[ConnectionState]("telemetry:connection")
}

// ConnectionState describes the current WebSocket bridge connection status.
type ConnectionState struct {
	Connected    bool      `json:"connected"`
	URL          string    `json:"url"`
	LastError    string    `json:"lastError"`
	UpdatedAtUTC time.Time `json:"updatedAtUtc"`
}

type apiClient struct {
	baseURL    string
	wsURL      string
	httpClient *http.Client

	mu        sync.RWMutex
	connState ConnectionState
	emitFn    func(name string, data any)
}

func defaultEmit(name string, data any) {
	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit(name, data)
	}
}

// resolveAPIBaseURL normalizes HRIG_API_URL (defaulting to http://127.0.0.1:80)
// and derives the corresponding WebSocket endpoint URL (<wsBase>/ws/realtime).
func resolveAPIBaseURL(raw string) (httpBase string, wsURL string, err error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		trimmed = defaultAPIBaseURL
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", "", fmt.Errorf("invalid HRIG_API_URL %q: %w", raw, err)
	}

	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "http" && scheme != "https" {
		return "", "", fmt.Errorf("invalid HRIG_API_URL scheme %q: must be http or https", parsed.Scheme)
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return "", "", fmt.Errorf("invalid HRIG_API_URL %q: host is required", raw)
	}

	parsed.Scheme = scheme
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""

	httpBase = parsed.String()

	wsParsed := *parsed
	if scheme == "https" {
		wsParsed.Scheme = "wss"
	} else {
		wsParsed.Scheme = "ws"
	}
	wsParsed.Path = parsed.Path + "/ws/realtime"

	return httpBase, wsParsed.String(), nil
}

func newAPIClient(rawURL string, emitFn func(name string, data any)) (*apiClient, error) {
	baseURL, wsURL, err := resolveAPIBaseURL(rawURL)
	if err != nil {
		return nil, err
	}
	if emitFn == nil {
		emitFn = defaultEmit
	}
	return &apiClient{
		baseURL: baseURL,
		wsURL:   wsURL,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		connState: ConnectionState{
			Connected:    false,
			URL:          wsURL,
			UpdatedAtUTC: time.Now().UTC(),
		},
		emitFn: emitFn,
	}, nil
}

func (c *apiClient) SetEmitFn(fn func(name string, data any)) {
	if fn == nil {
		fn = defaultEmit
	}
	c.mu.Lock()
	c.emitFn = fn
	c.mu.Unlock()
}

func (c *apiClient) ConnectionStatus() ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connState
}

func (c *apiClient) emit(name string, data any) {
	c.mu.RLock()
	fn := c.emitFn
	c.mu.RUnlock()
	if fn != nil {
		fn(name, data)
	}
}

func (c *apiClient) updateConnectionState(connected bool, lastErr string) {
	c.mu.Lock()
	changed := c.connState.Connected != connected ||
		c.connState.LastError != lastErr ||
		c.connState.URL != c.wsURL
	state := ConnectionState{
		Connected:    connected,
		URL:          c.wsURL,
		LastError:    lastErr,
		UpdatedAtUTC: time.Now().UTC(),
	}
	c.connState = state
	fn := c.emitFn
	c.mu.Unlock()

	if changed && fn != nil {
		fn("telemetry:connection", state)
	}
}

func (c *apiClient) GetHealth(ctx context.Context) (model.HealthStatus, error) {
	var out model.HealthStatus
	if err := c.doJSON(ctx, http.MethodGet, "/health", nil, &out); err != nil {
		return model.HealthStatus{}, err
	}
	return out, nil
}

func (c *apiClient) GetAstronauts(ctx context.Context) ([]model.Astronaut, error) {
	var out []model.Astronaut
	if err := c.doJSON(ctx, http.MethodGet, "/api/astronauts", nil, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Astronaut{}
	}
	return out, nil
}

func (c *apiClient) GetMetrics(ctx context.Context) ([]model.Metric, error) {
	var out []model.Metric
	if err := c.doJSON(ctx, http.MethodGet, "/api/metrics", nil, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Metric{}
	}
	return out, nil
}

func (c *apiClient) GetSnapshot(ctx context.Context) (model.Snapshot, error) {
	var out model.Snapshot
	if err := c.doJSON(ctx, http.MethodGet, "/api/snapshot", nil, &out); err != nil {
		return model.Snapshot{}, err
	}
	if out.Readings == nil {
		out.Readings = []model.LatestReading{}
	}
	if out.Alerts == nil {
		out.Alerts = []model.AlertEvent{}
	}
	return out, nil
}

func (c *apiClient) GetTrends(ctx context.Context, query model.TrendQuery) ([]model.TrendPoint, error) {
	params := url.Values{}
	if metricCode := strings.TrimSpace(query.MetricCode); metricCode != "" {
		params.Set("metric_code", metricCode)
	}
	if astroID := strings.TrimSpace(query.AstronautID); astroID != "" {
		params.Set("astronaut_id", astroID)
	}
	if query.Limit > 0 {
		params.Set("limit", strconv.Itoa(query.Limit))
	}

	path := "/api/trends"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out []model.TrendPoint
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.TrendPoint{}
	}
	return out, nil
}

func (c *apiClient) GetHistory(ctx context.Context, query model.HistoryQuery) ([]model.HistoryEntry, error) {
	params := url.Values{}
	if astroID := strings.TrimSpace(query.AstronautID); astroID != "" {
		params.Set("astronaut_id", astroID)
	}
	if query.Limit > 0 {
		params.Set("limit", strconv.Itoa(query.Limit))
	}
	if query.Offset > 0 {
		params.Set("offset", strconv.Itoa(query.Offset))
	}

	path := "/api/history"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out []model.HistoryEntry
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.HistoryEntry{}
	}
	return out, nil
}

func (c *apiClient) AckAlert(ctx context.Context, alertID, acknowledgedBy string) (model.AlertEvent, error) {
	id := strings.TrimSpace(alertID)
	if id == "" {
		return model.AlertEvent{}, errors.New("alertID is required")
	}

	reqBody := model.AckAlertRequest{
		AcknowledgedBy: strings.TrimSpace(acknowledgedBy),
	}
	path := "/api/alerts/" + url.PathEscape(id) + "/ack"

	var out model.AlertEvent
	if err := c.doJSON(ctx, http.MethodPost, path, reqBody, &out); err != nil {
		return model.AlertEvent{}, err
	}
	return out, nil
}

func (c *apiClient) GetHazards(ctx context.Context, q string) ([]model.SpaceHazard, error) {
	params := url.Values{}
	if trimmedQ := strings.TrimSpace(q); trimmedQ != "" {
		params.Set("q", trimmedQ)
	}

	path := "/api/hazards"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out []model.SpaceHazard
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.SpaceHazard{}
	}
	for i := range out {
		out[i].Normalize()
	}
	return out, nil
}

func (c *apiClient) GetProcedures(ctx context.Context, q, category, metricCode string) ([]model.MedicalProcedure, error) {
	params := url.Values{}
	if trimmedQ := strings.TrimSpace(q); trimmedQ != "" {
		params.Set("q", trimmedQ)
	}
	if trimmedCategory := strings.TrimSpace(category); trimmedCategory != "" {
		params.Set("category", trimmedCategory)
	}
	if trimmedMetric := strings.TrimSpace(metricCode); trimmedMetric != "" {
		params.Set("metric_code", trimmedMetric)
	}

	path := "/api/procedures"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out []model.MedicalProcedure
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.MedicalProcedure{}
	}
	for i := range out {
		out[i].Normalize()
	}
	return out, nil
}

func (c *apiClient) Search(ctx context.Context, q, tab string) (model.SearchResults, error) {
	params := url.Values{}
	if trimmedQ := strings.TrimSpace(q); trimmedQ != "" {
		params.Set("q", trimmedQ)
	}
	if trimmedTab := strings.TrimSpace(tab); trimmedTab != "" {
		params.Set("tab", trimmedTab)
	}

	path := "/api/search"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var out model.SearchResults
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return model.SearchResults{}, err
	}
	out.Normalize()
	return out, nil
}

func (c *apiClient) doJSON(ctx context.Context, method, path string, reqBody any, respOut any) error {
	if ctx == nil {
		ctx = context.Background()
	}

	var bodyReader io.Reader
	if reqBody != nil {
		payload, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build request %s %s: %w", method, path, err)
	}
	req.Header.Set("Accept", "application/json")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(io.LimitReader(resp.Body, maxHTTPBodyBytes))
	if err != nil {
		return fmt.Errorf("read response %s %s: %w", method, path, err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var apiErr struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(rawBody, &apiErr) == nil && strings.TrimSpace(apiErr.Error) != "" {
			return fmt.Errorf("api %s %s returned %d: %s", method, path, resp.StatusCode, strings.TrimSpace(apiErr.Error))
		}
		msg := strings.TrimSpace(string(rawBody))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return fmt.Errorf("api %s %s returned %d: %s", method, path, resp.StatusCode, msg)
	}

	if respOut != nil {
		if err := json.Unmarshal(rawBody, respOut); err != nil {
			return fmt.Errorf("decode response %s %s: %w", method, path, err)
		}
	}
	return nil
}

// RunRealtimeLoop maintains a persistent WebSocket connection to /ws/realtime,
// emitting "telemetry:connection" and "telemetry:envelope" events until ctx is canceled.
func (c *apiClient) RunRealtimeLoop(ctx context.Context) {
	if ctx == nil {
		return
	}

	attempt := 0
	for {
		if ctx.Err() != nil {
			c.updateConnectionState(false, "")
			return
		}

		conn, _, err := websocket.Dial(ctx, c.wsURL, nil)
		if err != nil {
			if ctx.Err() != nil {
				c.updateConnectionState(false, "")
				return
			}
			c.updateConnectionState(false, err.Error())
			if !waitBackoff(ctx, attempt) {
				c.updateConnectionState(false, "")
				return
			}
			attempt++
			continue
		}

		conn.SetReadLimit(maxWSMessageBytes)
		attempt = 0
		c.updateConnectionState(true, "")

		for {
			var env model.RealtimeEnvelope
			if err := wsjson.Read(ctx, conn, &env); err != nil {
				if ctx.Err() != nil {
					_ = conn.Close(websocket.StatusNormalClosure, "shutdown")
					c.updateConnectionState(false, "")
					return
				}
				_ = conn.CloseNow()
				c.updateConnectionState(false, err.Error())
				break
			}
			c.emit("telemetry:envelope", env)
		}

		if !waitBackoff(ctx, attempt) {
			c.updateConnectionState(false, "")
			return
		}
		attempt++
	}
}

func backoffDuration(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if attempt >= len(wsBackoffSteps) {
		return wsBackoffSteps[len(wsBackoffSteps)-1]
	}
	return wsBackoffSteps[attempt]
}

func waitBackoff(ctx context.Context, attempt int) bool {
	timer := time.NewTimer(backoffDuration(attempt))
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
