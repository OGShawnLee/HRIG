package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"hrig/server/internal/config"
	"hrig/server/internal/ingestion"
	"hrig/server/internal/realtime"
	"hrig/shared/model"
)

type mockIngester struct {
	fn func(ctx context.Context, req ingestion.IngestRequest) (ingestion.IngestResult, error)
}

func (m mockIngester) Ingest(ctx context.Context, req ingestion.IngestRequest) (ingestion.IngestResult, error) {
	return m.fn(ctx, req)
}

func TestConfigDefaultsAndDatabaseURLFallback(t *testing.T) {
	t.Setenv("HRIG_API_ADDR", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("HRIG_DATABASE_URL", "postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable")

	cfg := config.Load()
	if cfg.APIAddr != ":80" {
		t.Fatalf("expected default APIAddr :80, got %q", cfg.APIAddr)
	}
	if cfg.DatabaseURL != "postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable" {
		t.Fatalf("expected HRIG_DATABASE_URL fallback, got %q", cfg.DatabaseURL)
	}

	t.Setenv("DATABASE_URL", "postgres://primary@127.0.0.1:5432/hrig?sslmode=disable")
	cfg = config.Load()
	if cfg.DatabaseURL != "postgres://primary@127.0.0.1:5432/hrig?sslmode=disable" {
		t.Fatalf("expected DATABASE_URL to take precedence, got %q", cfg.DatabaseURL)
	}
}

func TestRootHelloWorld(t *testing.T) {
	cfg := config.Load()
	router := NewRouter(cfg, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("expected Content-Type application/json, got %q", contentType)
	}

	body := strings.TrimSpace(rec.Body.String())
	if body != `{"message":"hello world"}` {
		t.Fatalf("expected body %q, got %q", `{"message":"hello world"}`, body)
	}

	var payload struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %v", err)
	}
	if payload.Message != "hello world" {
		t.Fatalf("expected message %q, got %q", "hello world", payload.Message)
	}
}

func TestPostIngest_HTTPValidation(t *testing.T) {
	validBody := `{"device_code":"cabin_sensor_01","environment":"iva","recorded_at_utc":"2026-11-15T08:20:00Z","readings":[{"metric_code":"cabin_pressure","value_numeric":101.3,"unit_code":"kPa"}]}`

	t.Run("rejects_wrong_content_type", func(t *testing.T) {
		router := NewRouter(config.Load(), mockIngester{
			fn: func(ctx context.Context, req ingestion.IngestRequest) (ingestion.IngestResult, error) {
				t.Fatal("ingester should not be called")
				return ingestion.IngestResult{}, nil
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(validBody))
		req.Header.Set("Content-Type", "text/plain")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d (%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("rejects_malformed_json", func(t *testing.T) {
		router := NewRouter(config.Load(), mockIngester{
			fn: func(ctx context.Context, req ingestion.IngestRequest) (ingestion.IngestResult, error) {
				t.Fatal("ingester should not be called")
				return ingestion.IngestResult{}, nil
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"device_code":`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("rejects_trailing_json_token", func(t *testing.T) {
		router := NewRouter(config.Load(), mockIngester{
			fn: func(ctx context.Context, req ingestion.IngestRequest) (ingestion.IngestResult, error) {
				t.Fatal("ingester should not be called")
				return ingestion.IngestResult{}, nil
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(validBody+`{"extra":true}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("rejects_body_exceeding_1mib", func(t *testing.T) {
		router := NewRouter(config.Load(), mockIngester{
			fn: func(ctx context.Context, req ingestion.IngestRequest) (ingestion.IngestResult, error) {
				t.Fatal("ingester should not be called")
				return ingestion.IngestResult{}, nil
			},
		})
		oversized := `{"device_code":"` + strings.Repeat("a", (1<<20)+100) + `"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(oversized))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("returns_201_on_success", func(t *testing.T) {
		ts := time.Date(2026, 11, 15, 8, 20, 0, 0, time.UTC)
		router := NewRouter(config.Load(), mockIngester{
			fn: func(ctx context.Context, req ingestion.IngestRequest) (ingestion.IngestResult, error) {
				if req.DeviceCode != "cabin_sensor_01" {
					t.Fatalf("unexpected device_code %q", req.DeviceCode)
				}
				return ingestion.IngestResult{
					MeasurementID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
					RecordedAtUTC: ts,
					ValuesCount:   1,
				}, nil
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(validBody))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d (%s)", rec.Code, rec.Body.String())
		}
		var out ingestion.IngestResult
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if out.MeasurementID != "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" || out.ValuesCount != 1 {
			t.Fatalf("unexpected response: %+v", out)
		}
	})

	t.Run("maps_ingestion_errors_and_hides_internal_errors", func(t *testing.T) {
		statusCases := []struct {
			name       string
			err        error
			wantStatus int
			wantBody   string
		}{
			{
				name:       "bad_request_400",
				err:        &ingestion.Error{StatusCode: http.StatusBadRequest, Message: "readings must contain at least one reading"},
				wantStatus: http.StatusBadRequest,
			},
			{
				name:       "not_found_404",
				err:        &ingestion.Error{StatusCode: http.StatusNotFound, Message: "device not found or inactive"},
				wantStatus: http.StatusNotFound,
			},
			{
				name:       "conflict_409",
				err:        &ingestion.Error{StatusCode: http.StatusConflict, Message: "duplicate metric_code in readings"},
				wantStatus: http.StatusConflict,
			},
			{
				name:       "unexpected_error_sanitized_500",
				err:        errors.New("postgres://secret:pass@db:5432/hrig SQL syntax error"),
				wantStatus: http.StatusInternalServerError,
				wantBody:   "internal server error",
			},
		}

		for _, tc := range statusCases {
			t.Run(tc.name, func(t *testing.T) {
				router := NewRouter(config.Load(), mockIngester{
					fn: func(ctx context.Context, req ingestion.IngestRequest) (ingestion.IngestResult, error) {
						return ingestion.IngestResult{}, tc.err
					},
				})
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(validBody))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()

				router.ServeHTTP(rec, req)
				if rec.Code != tc.wantStatus {
					t.Fatalf("expected %d, got %d", tc.wantStatus, rec.Code)
				}
				if tc.wantBody != "" {
					if strings.Contains(rec.Body.String(), "secret") || strings.Contains(rec.Body.String(), "SQL") {
						t.Fatalf("internal error leaked sensitive details: %s", rec.Body.String())
					}
					if !strings.Contains(rec.Body.String(), tc.wantBody) {
						t.Fatalf("expected body to contain %q, got %s", tc.wantBody, rec.Body.String())
					}
				}
			})
		}
	})
}

type mockPinger struct {
	err error
}

func (m mockPinger) Ping(ctx context.Context) error {
	return m.err
}

func TestConfigStaleAfterAndMissionEpoch(t *testing.T) {
	t.Setenv("HRIG_STALE_AFTER", "")
	t.Setenv("HRIG_MISSION_EPOCH_UTC", "")
	cfg := config.Load()
	if cfg.StaleAfter != 15*time.Second {
		t.Fatalf("expected default StaleAfter 15s, got %s", cfg.StaleAfter)
	}
	if cfg.MissionEpochUTC.IsZero() {
		t.Fatal("expected default MissionEpochUTC to be non-zero")
	}

	t.Setenv("HRIG_STALE_AFTER", "30s")
	t.Setenv("HRIG_MISSION_EPOCH_UTC", "2026-11-15T01:02:03Z")
	cfg = config.Load()
	if cfg.StaleAfter != 30*time.Second {
		t.Fatalf("expected StaleAfter 30s, got %s", cfg.StaleAfter)
	}
	wantEpoch := time.Date(2026, 11, 15, 1, 2, 3, 0, time.UTC)
	if !cfg.MissionEpochUTC.Equal(wantEpoch) {
		t.Fatalf("expected MissionEpochUTC %s, got %s", wantEpoch, cfg.MissionEpochUTC)
	}

	t.Setenv("HRIG_STALE_AFTER", "-5s")
	t.Setenv("HRIG_MISSION_EPOCH_UTC", "not-a-timestamp")
	cfg = config.Load()
	if cfg.StaleAfter != 15*time.Second {
		t.Fatalf("expected invalid StaleAfter to fall back to 15s, got %s", cfg.StaleAfter)
	}
	if cfg.MissionEpochUTC.IsZero() {
		t.Fatal("expected invalid MissionEpochUTC to fall back to current UTC time")
	}
}

func TestRealtimeAndHealthEndpoints(t *testing.T) {
	cfg := config.Load()

	// Without realtime controller wired, /ws/realtime and /api/snapshot return 503 and /health reports degraded store.
	noRT := NewRouterWithDeps(cfg, RouterDeps{})
	{
		rec := httptest.NewRecorder()
		noRT.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ws/realtime", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503 when realtime nil, got %d", rec.Code)
		}

		rec = httptest.NewRecorder()
		noRT.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/snapshot", nil))
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503 on snapshot when realtime nil, got %d", rec.Code)
		}

		rec = httptest.NewRecorder()
		noRT.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 on health, got %d", rec.Code)
		}
		var h model.HealthStatus
		if err := json.Unmarshal(rec.Body.Bytes(), &h); err != nil {
			t.Fatalf("decode health: %v", err)
		}
		if h.Status != "ok" || h.Store != "degraded" {
			t.Fatalf("unexpected health when no pinger: %+v", h)
		}
	}

	hub := realtime.NewHub(15*time.Second, time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC))
	defer hub.Close()

	val := 190.0
	unit := "/min"
	hub.PublishFrames([]model.TelemetryFrame{
		{
			Source:        "wearable",
			DeviceCode:    "biomonitor_01",
			Environment:   model.EnvIVA,
			MetricType:    1,
			MetricCode:    "heart_rate",
			UnitCode:      &unit,
			ValueNumeric:  &val,
			Status:        model.StatusCritical,
			RecordedAtUTC: time.Now().UTC(),
		},
	})

	router := NewRouterWithDeps(cfg, RouterDeps{
		Realtime: hub,
		DBPinger: mockPinger{err: nil},
	})

	// WebSocket handshake via httptest.Server
	srv := httptest.NewServer(router)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/realtime"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial /ws/realtime: %v", err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var wsSnap model.RealtimeEnvelope
	if err := conn.ReadJSON(&wsSnap); err != nil {
		t.Fatalf("read initial ws snapshot: %v", err)
	}
	if wsSnap.Type != "snapshot" || len(wsSnap.Readings) != 1 || len(wsSnap.Alerts) != 1 {
		t.Fatalf("unexpected ws snapshot: %+v", wsSnap)
	}
	alertID := wsSnap.Alerts[0].ID

	// GET /health with active WS client, active alert, and healthy DB pinger
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on /health, got %d", rec.Code)
	}
	var health model.HealthStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &health); err != nil {
		t.Fatalf("decode /health: %v", err)
	}
	if health.Status != "ok" || health.Store != "postgres" || health.ActiveAlerts != 1 || health.RealtimeClients != 1 {
		t.Fatalf("unexpected /health payload: %+v", health)
	}

	// GET /api/snapshot
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/snapshot", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on /api/snapshot, got %d", rec.Code)
	}
	var snap model.Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatalf("decode /api/snapshot: %v", err)
	}
	if len(snap.Readings) != 1 || len(snap.Alerts) != 1 {
		t.Fatalf("unexpected /api/snapshot payload: %+v", snap)
	}

	// POST /api/alerts/{id}/ack -> 404 on unknown ID
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/alerts/does-not-exist/ack", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown alert id, got %d", rec.Code)
	}

	// POST /api/alerts/{id}/ack -> 400 on malformed JSON
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/alerts/"+alertID+"/ack", strings.NewReader(`{"bad":`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d", rec.Code)
	}

	// POST /api/alerts/{id}/ack -> 200 with AckAlertRequest
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/alerts/"+alertID+"/ack", strings.NewReader(`{"acknowledgedBy":"CMO-1"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on ack, got %d (%s)", rec.Code, rec.Body.String())
	}
	var acked model.AlertEvent
	if err := json.Unmarshal(rec.Body.Bytes(), &acked); err != nil {
		t.Fatalf("decode ack response: %v", err)
	}
	if !acked.Acknowledged || acked.AcknowledgedBy == nil || *acked.AcknowledgedBy != "CMO-1" {
		t.Fatalf("unexpected ack response: %+v", acked)
	}

	// WS client should also receive alert_ack broadcast
	var ackEnv model.RealtimeEnvelope
	if err := conn.ReadJSON(&ackEnv); err != nil {
		t.Fatalf("read ws alert_ack: %v", err)
	}
	if ackEnv.Type != "alert_ack" || ackEnv.Alert == nil || ackEnv.Alert.ID != alertID {
		t.Fatalf("unexpected ws alert_ack envelope: %+v", ackEnv)
	}
}
