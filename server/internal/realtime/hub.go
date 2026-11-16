package realtime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"hrig/server/internal/db"
	"hrig/shared/model"
)

const (
	defaultStaleAfter      = 15 * time.Second
	dailyCheckinStaleAfter = 24 * time.Hour
	maxStoredAlerts        = 200
)

func isDailyCheckinReading(metricCode, deviceCode, source string) bool {
	if deviceCode == "manual_checkin" || source == "manual" {
		return true
	}
	switch metricCode {
	case "sleep_hours", "mood", "hydration_intake", "body_mass":
		return true
	default:
		return false
	}
}

var ErrAlertNotFound = errors.New("alert not found")

type readingKey struct {
	metricCode  string
	astronautID string
}

// Hub maintains in-memory latest readings, active/recent alerts, and connected
// Gorilla WebSocket clients.
type Hub struct {
	mu           sync.RWMutex
	latest       map[readingKey]*model.LatestReading
	alerts       []*model.AlertEvent
	alertByID    map[string]*model.AlertEvent
	clients      map[*Client]struct{}
	staleAfter   time.Duration
	missionEpoch time.Time
	nowFn        func() time.Time
	alertSeq     atomic.Uint64
	closed       bool
}

func NewHub(staleAfter time.Duration, missionEpochUTC time.Time) *Hub {
	if staleAfter <= 0 {
		staleAfter = defaultStaleAfter
	}
	if missionEpochUTC.IsZero() {
		missionEpochUTC = time.Now().UTC()
	} else {
		missionEpochUTC = missionEpochUTC.UTC()
	}

	return &Hub{
		latest:       make(map[readingKey]*model.LatestReading),
		alerts:       make([]*model.AlertEvent, 0),
		alertByID:    make(map[string]*model.AlertEvent),
		clients:      make(map[*Client]struct{}),
		staleAfter:   staleAfter,
		missionEpoch: missionEpochUTC,
		nowFn: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// HydrateFromDB seeds latest readings from PostgreSQL at startup without writing
// runtime-only stale state back to the database.
func (h *Hub) HydrateFromDB(ctx context.Context, q db.Querier) error {
	if q == nil {
		return nil
	}

	rows, err := q.ListLatestReadingsByMetric(ctx, pgtype.UUID{})
	if err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	now := h.nowFn().UTC()
	for _, row := range rows {
		var astroIDPtr *string
		astroKey := ""
		if row.AstronautID.Valid {
			s := formatUUID(row.AstronautID)
			astroIDPtr = &s
			astroKey = s
		}

		deviceCode := ""
		if row.DeviceCode != nil {
			deviceCode = *row.DeviceCode
		}

		var unitCodePtr *string
		if row.UnitCode != nil {
			u := *row.UnitCode
			unitCodePtr = &u
		}

		var valNumPtr *float64
		if row.ValueNumeric.Valid {
			if f, err := row.ValueNumeric.Float64Value(); err == nil && f.Valid {
				v := f.Float64
				valNumPtr = &v
			}
		}

		var valScalePtr *int16
		if row.ValueScale != nil {
			v := *row.ValueScale
			valScalePtr = &v
		}

		var valTextPtr *string
		if row.ValueText != nil {
			v := *row.ValueText
			valTextPtr = &v
		}

		var notePtr *string
		if row.Note != nil {
			v := *row.Note
			notePtr = &v
		}

		recordedAt := now
		if row.RecordedAtUtc.Valid {
			recordedAt = row.RecordedAtUtc.Time.UTC()
		}

		staleThreshold := h.staleAfter
		if isDailyCheckinReading(row.MetricCode, deviceCode, "postgres") {
			staleThreshold = dailyCheckinStaleAfter
		}
		isStale := now.Sub(recordedAt) > staleThreshold
		status := model.ReadingStatus(row.Status)
		if isStale {
			status = model.StatusStale
		}

		key := readingKey{
			metricCode:  row.MetricCode,
			astronautID: astroKey,
		}

		h.latest[key] = &model.LatestReading{
			TelemetryFrame: model.TelemetryFrame{
				Source:        "postgres",
				DeviceCode:    deviceCode,
				AstronautID:   astroIDPtr,
				Environment:   model.EnvironmentMode(row.Environment),
				MetricType:    row.MetricType,
				MetricCode:    row.MetricCode,
				UnitCode:      unitCodePtr,
				ValueNumeric:  valNumPtr,
				ValueScale:    valScalePtr,
				ValueText:     valTextPtr,
				Status:        status,
				RecordedAtUTC: recordedAt,
				MET:           FormatMET(recordedAt, h.missionEpoch),
				Note:          notePtr,
			},
			LastSeenUTC: recordedAt,
			IsStale:     isStale,
		}
	}

	return nil
}

// PublishFrames updates the in-memory latest map, broadcasts "reading" envelopes,
// and emits in-memory "alert" envelopes for critical readings.
func (h *Hub) PublishFrames(frames []model.TelemetryFrame) {
	if len(frames) == 0 {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}

	now := h.nowFn().UTC()
	for _, rawFrame := range frames {
		frame := cloneFrame(rawFrame)
		if frame.RecordedAtUTC.IsZero() {
			frame.RecordedAtUTC = now
		} else {
			frame.RecordedAtUTC = frame.RecordedAtUTC.UTC()
		}
		frame.MET = FormatMET(frame.RecordedAtUTC, h.missionEpoch)

		astroKey := ""
		if frame.AstronautID != nil {
			astroKey = *frame.AstronautID
		}
		key := readingKey{
			metricCode:  frame.MetricCode,
			astronautID: astroKey,
		}

		lr := &model.LatestReading{
			TelemetryFrame: frame,
			LastSeenUTC:    now,
			IsStale:        false,
		}
		h.latest[key] = lr

		h.broadcastLocked(model.RealtimeEnvelope{
			Type:      "reading",
			Reading:   cloneLatestReading(lr),
			SentAtUTC: now,
		})

		if frame.Status == model.StatusCritical {
			alert := h.createAlertLocked(frame, now)
			h.broadcastLocked(model.RealtimeEnvelope{
				Type:      "alert",
				Alert:     cloneAlert(alert),
				SentAtUTC: now,
			})
		}
	}
}

func (h *Hub) createAlertLocked(frame model.TelemetryFrame, now time.Time) *model.AlertEvent {
	id := fmt.Sprintf("alert-%d-%d", now.UnixNano(), h.alertSeq.Add(1))
	alert := &model.AlertEvent{
		ID:             id,
		AstronautID:    cloneStringPtr(frame.AstronautID),
		Source:         frame.Source,
		DeviceCode:     frame.DeviceCode,
		Environment:    frame.Environment,
		MetricCode:     frame.MetricCode,
		ValueNumeric:   cloneFloat64Ptr(frame.ValueNumeric),
		ValueScale:     cloneInt16Ptr(frame.ValueScale),
		ValueText:      cloneStringPtr(frame.ValueText),
		UnitCode:       cloneStringPtr(frame.UnitCode),
		Status:         frame.Status,
		Message:        formatAlertMessage(frame),
		TriggeredAtUTC: now,
		MET:            frame.MET,
		Acknowledged:   false,
	}

	if len(h.alerts) >= maxStoredAlerts {
		evicted := h.alerts[0]
		copy(h.alerts, h.alerts[1:])
		h.alerts[len(h.alerts)-1] = alert
		if evicted != nil {
			delete(h.alertByID, evicted.ID)
		}
	} else {
		h.alerts = append(h.alerts, alert)
	}
	h.alertByID[alert.ID] = alert

	return alert
}

// RunStaleMonitor periodically sweeps latest readings and marks quiet readings stale in memory.
func (h *Hub) RunStaleMonitor(ctx context.Context, checkInterval time.Duration) {
	if checkInterval <= 0 {
		checkInterval = time.Second
	}

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.SweepStale()
		}
	}
}

// SweepStale checks all latest readings and transitions quiet ones to StatusStale in memory.
func (h *Hub) SweepStale() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return 0
	}

	now := h.nowFn().UTC()
	keys := make([]readingKey, 0, len(h.latest))
	for k := range h.latest {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].metricCode != keys[j].metricCode {
			return keys[i].metricCode < keys[j].metricCode
		}
		return keys[i].astronautID < keys[j].astronautID
	})

	marked := 0
	for _, k := range keys {
		lr := h.latest[k]
		staleThreshold := h.staleAfter
		if isDailyCheckinReading(lr.MetricCode, lr.DeviceCode, lr.Source) {
			staleThreshold = dailyCheckinStaleAfter
		}
		if !lr.IsStale && now.Sub(lr.LastSeenUTC) > staleThreshold {
			lr.IsStale = true
			lr.Status = model.StatusStale
			marked++
			h.broadcastLocked(model.RealtimeEnvelope{
				Type:      "stale",
				Reading:   cloneLatestReading(lr),
				SentAtUTC: now,
			})
		}
	}
	return marked
}

// AckAlert marks an in-memory alert as acknowledged and broadcasts an "alert_ack" envelope.
func (h *Hub) AckAlert(id string, acknowledgedBy string) (model.AlertEvent, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	alert, ok := h.alertByID[id]
	if !ok {
		return model.AlertEvent{}, ErrAlertNotFound
	}

	by := strings.TrimSpace(acknowledgedBy)
	if by == "" {
		by = "crew"
	}
	now := h.nowFn().UTC()

	alert.Acknowledged = true
	alert.AcknowledgedBy = &by
	alert.AcknowledgedAtUTC = &now

	out := *cloneAlert(alert)
	if !h.closed {
		h.broadcastLocked(model.RealtimeEnvelope{
			Type:      "alert_ack",
			Alert:     cloneAlert(alert),
			SentAtUTC: now,
		})
	}
	return out, nil
}

// Snapshot returns a point-in-time copy of all latest readings and stored alerts.
func (h *Hub) Snapshot() model.Snapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.snapshotLocked(h.nowFn().UTC())
}

func (h *Hub) snapshotLocked(now time.Time) model.Snapshot {
	keys := make([]readingKey, 0, len(h.latest))
	for k := range h.latest {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].metricCode != keys[j].metricCode {
			return keys[i].metricCode < keys[j].metricCode
		}
		return keys[i].astronautID < keys[j].astronautID
	})

	readings := make([]model.LatestReading, 0, len(keys))
	for _, k := range keys {
		readings = append(readings, *cloneLatestReading(h.latest[k]))
	}

	alerts := make([]model.AlertEvent, 0, len(h.alerts))
	for _, a := range h.alerts {
		alerts = append(alerts, *cloneAlert(a))
	}

	return model.Snapshot{
		Readings:       readings,
		Alerts:         alerts,
		GeneratedAtUTC: now,
	}
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) ActiveAlertsCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, a := range h.alerts {
		if !a.Acknowledged {
			count++
		}
	}
	return count
}

// Close disconnects all active WebSocket clients and stops accepting new ones.
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}
	h.closed = true
	for c := range h.clients {
		delete(h.clients, c)
		close(c.send)
	}
}

// ServeWS returns an http.HandlerFunc upgrading connections to Gorilla WebSocket clients.
func (h *Hub) ServeWS(allowedOrigins []string) http.HandlerFunc {
	return ServeWS(h, allowedOrigins)
}

func (h *Hub) isClosed() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.closed
}

func (h *Hub) register(c *Client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return false
	}

	now := h.nowFn().UTC()
	snap := h.snapshotLocked(now)
	initial := model.RealtimeEnvelope{
		Type:      "snapshot",
		Readings:  snap.Readings,
		Alerts:    snap.Alerts,
		SentAtUTC: now,
	}

	select {
	case c.send <- initial:
		h.clients[c] = struct{}{}
		return true
	default:
		return false
	}
}

func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
}

func (h *Hub) broadcastLocked(env model.RealtimeEnvelope) {
	for c := range h.clients {
		select {
		case c.send <- env:
		default:
			delete(h.clients, c)
			close(c.send)
		}
	}
}

func formatAlertMessage(frame model.TelemetryFrame) string {
	unit := ""
	if frame.UnitCode != nil && *frame.UnitCode != "" {
		unit = " " + *frame.UnitCode
	}
	switch {
	case frame.ValueNumeric != nil:
		return fmt.Sprintf("Critical %s reading: %g%s", frame.MetricCode, *frame.ValueNumeric, unit)
	case frame.ValueScale != nil:
		return fmt.Sprintf("Critical %s reading: %d%s", frame.MetricCode, *frame.ValueScale, unit)
	case frame.ValueText != nil:
		return fmt.Sprintf("Critical %s reading: %s", frame.MetricCode, *frame.ValueText)
	default:
		return fmt.Sprintf("Critical %s reading", frame.MetricCode)
	}
}

func formatUUID(id pgtype.UUID) string {
	b := id.Bytes
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func cloneStringPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func cloneFloat64Ptr(p *float64) *float64 {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func cloneInt16Ptr(p *int16) *int16 {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func cloneTimePtr(p *time.Time) *time.Time {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

func cloneFrame(f model.TelemetryFrame) model.TelemetryFrame {
	return model.TelemetryFrame{
		Source:        f.Source,
		DeviceCode:    f.DeviceCode,
		AstronautID:   cloneStringPtr(f.AstronautID),
		Environment:   f.Environment,
		MetricType:    f.MetricType,
		MetricCode:    f.MetricCode,
		UnitCode:      cloneStringPtr(f.UnitCode),
		ValueNumeric:  cloneFloat64Ptr(f.ValueNumeric),
		ValueScale:    cloneInt16Ptr(f.ValueScale),
		ValueText:     cloneStringPtr(f.ValueText),
		Status:        f.Status,
		RecordedAtUTC: f.RecordedAtUTC,
		MET:           f.MET,
		Note:          cloneStringPtr(f.Note),
	}
}

func cloneLatestReading(lr *model.LatestReading) *model.LatestReading {
	if lr == nil {
		return nil
	}
	return &model.LatestReading{
		TelemetryFrame: cloneFrame(lr.TelemetryFrame),
		LastSeenUTC:    lr.LastSeenUTC,
		IsStale:        lr.IsStale,
	}
}

func cloneAlert(a *model.AlertEvent) *model.AlertEvent {
	if a == nil {
		return nil
	}
	return &model.AlertEvent{
		ID:                a.ID,
		AstronautID:       cloneStringPtr(a.AstronautID),
		Source:            a.Source,
		DeviceCode:        a.DeviceCode,
		Environment:       a.Environment,
		MetricCode:        a.MetricCode,
		ValueNumeric:      cloneFloat64Ptr(a.ValueNumeric),
		ValueScale:        cloneInt16Ptr(a.ValueScale),
		ValueText:         cloneStringPtr(a.ValueText),
		UnitCode:          cloneStringPtr(a.UnitCode),
		Status:            a.Status,
		Message:           a.Message,
		TriggeredAtUTC:    a.TriggeredAtUTC,
		MET:               a.MET,
		Acknowledged:      a.Acknowledged,
		AcknowledgedBy:    cloneStringPtr(a.AcknowledgedBy),
		AcknowledgedAtUTC: cloneTimePtr(a.AcknowledgedAtUTC),
	}
}
