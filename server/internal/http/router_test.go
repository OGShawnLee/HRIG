package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"hrig/server/internal/config"
	"hrig/server/internal/ingestion"
	"hrig/server/internal/query"
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
	unsetEnvForTest := func(t *testing.T, keys ...string) {
		t.Helper()
		for _, key := range keys {
			t.Setenv(key, "")
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("unset %s: %v", key, err)
			}
		}
	}

	t.Run("default_database_url_when_unset_or_empty", func(t *testing.T) {
		t.Chdir(t.TempDir())
		unsetEnvForTest(t, "HRIG_API_ADDR", "DATABASE_URL", "HRIG_DATABASE_URL")

		cfg := config.Load()
		if cfg.APIAddr != ":80" {
			t.Fatalf("expected default APIAddr :80, got %q", cfg.APIAddr)
		}
		if cfg.DatabaseURL != "postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable" {
			t.Fatalf("expected default DatabaseURL, got %q", cfg.DatabaseURL)
		}

		t.Setenv("DATABASE_URL", "")
		t.Setenv("HRIG_DATABASE_URL", "")
		cfg = config.Load()
		if cfg.DatabaseURL != "postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable" {
			t.Fatalf("expected default DatabaseURL when env vars are empty, got %q", cfg.DatabaseURL)
		}
	})

	t.Run("hrig_database_url_and_primary_precedence", func(t *testing.T) {
		t.Chdir(t.TempDir())
		t.Setenv("HRIG_API_ADDR", "")
		t.Setenv("DATABASE_URL", "")
		t.Setenv("HRIG_DATABASE_URL", "postgres://fallback@127.0.0.1:5432/hrig?sslmode=disable")

		cfg := config.Load()
		if cfg.DatabaseURL != "postgres://fallback@127.0.0.1:5432/hrig?sslmode=disable" {
			t.Fatalf("expected HRIG_DATABASE_URL fallback, got %q", cfg.DatabaseURL)
		}

		t.Setenv("DATABASE_URL", "postgres://primary@127.0.0.1:5432/hrig?sslmode=disable")
		cfg = config.Load()
		if cfg.DatabaseURL != "postgres://primary@127.0.0.1:5432/hrig?sslmode=disable" {
			t.Fatalf("expected DATABASE_URL to take precedence, got %q", cfg.DatabaseURL)
		}
	})

	t.Run("dotenv_loaded_without_overriding_existing_env", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)
		unsetEnvForTest(t,
			"HRIG_API_ADDR",
			"HRIG_ALLOWED_ORIGINS",
			"DATABASE_URL",
			"HRIG_DATABASE_URL",
			"HRIG_STALE_AFTER",
			"HRIG_MISSION_EPOCH_UTC",
		)

		dotenvContent := strings.Join([]string{
			"# Local dev configuration",
			"",
			"export HRIG_API_ADDR=\":8088\"",
			"HRIG_ALLOWED_ORIGINS='http://localhost:5173, http://127.0.0.1:5173'",
			"DATABASE_URL=postgres://dotenv@127.0.0.1:5432/hrig_dev?sslmode=disable",
			"HRIG_STALE_AFTER=45s",
			"HRIG_MISSION_EPOCH_UTC=2026-11-15T12:00:00Z",
		}, "\n")
		if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte(dotenvContent), 0o600); err != nil {
			t.Fatalf("write .env: %v", err)
		}

		cfg := config.Load()
		if cfg.APIAddr != ":8088" {
			t.Fatalf("expected APIAddr from .env :8088, got %q", cfg.APIAddr)
		}
		if len(cfg.AllowedOrigins) != 2 || cfg.AllowedOrigins[0] != "http://localhost:5173" || cfg.AllowedOrigins[1] != "http://127.0.0.1:5173" {
			t.Fatalf("expected AllowedOrigins from .env, got %#v", cfg.AllowedOrigins)
		}
		if cfg.DatabaseURL != "postgres://dotenv@127.0.0.1:5432/hrig_dev?sslmode=disable" {
			t.Fatalf("expected DatabaseURL from .env, got %q", cfg.DatabaseURL)
		}
		if cfg.StaleAfter != 45*time.Second {
			t.Fatalf("expected StaleAfter 45s from .env, got %s", cfg.StaleAfter)
		}
		wantEpoch := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
		if !cfg.MissionEpochUTC.Equal(wantEpoch) {
			t.Fatalf("expected MissionEpochUTC %s from .env, got %s", wantEpoch, cfg.MissionEpochUTC)
		}

		// Explicit env vars (including empty ones) must take precedence over .env
		t.Setenv("HRIG_API_ADDR", ":9090")
		t.Setenv("DATABASE_URL", "")
		cfg = config.Load()
		if cfg.APIAddr != ":9090" {
			t.Fatalf("expected explicit HRIG_API_ADDR :9090 to override .env, got %q", cfg.APIAddr)
		}
		if cfg.DatabaseURL != "postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable" {
			t.Fatalf("expected empty explicit DATABASE_URL not to fall back to .env, got %q", cfg.DatabaseURL)
		}
	})

	t.Run("server_dotenv_loaded_from_repo_root", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)
		unsetEnvForTest(t, "DATABASE_URL", "HRIG_DATABASE_URL", "HRIG_API_ADDR")

		serverDir := filepath.Join(tmpDir, "server")
		if err := os.MkdirAll(serverDir, 0o755); err != nil {
			t.Fatalf("mkdir server: %v", err)
		}
		if err := os.WriteFile(filepath.Join(serverDir, ".env"), []byte("HRIG_DATABASE_URL=postgres://serverenv@127.0.0.1:5432/hrig?sslmode=disable\n"), 0o600); err != nil {
			t.Fatalf("write server/.env: %v", err)
		}

		cfg := config.Load()
		if cfg.DatabaseURL != "postgres://serverenv@127.0.0.1:5432/hrig?sslmode=disable" {
			t.Fatalf("expected DatabaseURL from server/.env, got %q", cfg.DatabaseURL)
		}
	})
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

type mockReader struct {
	listAstronautsFn func(ctx context.Context) ([]model.Astronaut, error)
	listMetricsFn    func(ctx context.Context) ([]model.Metric, error)
	listTrendsFn     func(ctx context.Context, filter query.TrendsFilter) ([]model.TrendPoint, error)
	listHistoryFn    func(ctx context.Context, filter query.HistoryFilter) ([]model.HistoryEntry, error)
}

func (m mockReader) ListAstronauts(ctx context.Context) ([]model.Astronaut, error) {
	if m.listAstronautsFn != nil {
		return m.listAstronautsFn(ctx)
	}
	return []model.Astronaut{}, nil
}

func (m mockReader) ListMetrics(ctx context.Context) ([]model.Metric, error) {
	if m.listMetricsFn != nil {
		return m.listMetricsFn(ctx)
	}
	return []model.Metric{}, nil
}

func (m mockReader) ListTrends(ctx context.Context, filter query.TrendsFilter) ([]model.TrendPoint, error) {
	if m.listTrendsFn != nil {
		return m.listTrendsFn(ctx, filter)
	}
	return []model.TrendPoint{}, nil
}

func (m mockReader) ListHistory(ctx context.Context, filter query.HistoryFilter) ([]model.HistoryEntry, error) {
	if m.listHistoryFn != nil {
		return m.listHistoryFn(ctx, filter)
	}
	return []model.HistoryEntry{}, nil
}

func TestReadEndpoints(t *testing.T) {
	cfg := config.Load()

	t.Run("returns_503_when_reader_nil", func(t *testing.T) {
		router := NewRouterWithDeps(cfg, RouterDeps{})
		for _, path := range []string{
			"/api/astronauts",
			"/api/metrics",
			"/api/trends?metric_code=spo2",
			"/api/history",
		} {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != http.StatusServiceUnavailable {
				t.Fatalf("%s expected 503 when Reader nil, got %d (%s)", path, rec.Code, rec.Body.String())
			}
			var errPayload struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &errPayload); err != nil || errPayload.Error == "" {
				t.Fatalf("%s expected JSON error payload, got %s", path, rec.Body.String())
			}
		}
	})

	t.Run("astronauts_and_metrics_200_and_500_sanitized", func(t *testing.T) {
		now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
		decimals := int16(1)
		dim := "temperature"
		router := NewRouterWithDeps(cfg, RouterDeps{
			Reader: mockReader{
				listAstronautsFn: func(ctx context.Context) ([]model.Astronaut, error) {
					return []model.Astronaut{
						{
							ID:              "11111111-1111-4111-8111-111111111111",
							FullName:        "Ana Torres",
							BirthDate:       time.Date(1988, 4, 12, 0, 0, 0, 0, time.UTC),
							Age:             38,
							NationalityCode: "MX",
							IsActive:        true,
							CreatedAt:       now,
							UpdatedAt:       now,
						},
					}, nil
				},
				listMetricsFn: func(ctx context.Context) ([]model.Metric, error) {
					return []model.Metric{
						{
							ID:                "22222222-2222-4222-8222-222222222222",
							Code:              "core_temp",
							Name:              "Core body temperature",
							Category:          "temperature",
							Tier:              model.TierPrimary,
							MetricType:        1,
							ValueType:         model.ValueNumeric,
							BaseUnitDimension: &dim,
							BaseUnitDecimals:  &decimals,
							DisplayOrder:      13,
							IsActive:          true,
						},
					}, nil
				},
			},
		})

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/astronauts", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /api/astronauts expected 200, got %d (%s)", rec.Code, rec.Body.String())
		}
		var astronauts []model.Astronaut
		if err := json.Unmarshal(rec.Body.Bytes(), &astronauts); err != nil {
			t.Fatalf("decode /api/astronauts: %v", err)
		}
		if len(astronauts) != 1 || astronauts[0].FullName != "Ana Torres" {
			t.Fatalf("unexpected astronauts payload: %+v", astronauts)
		}

		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/metrics", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /api/metrics expected 200, got %d (%s)", rec.Code, rec.Body.String())
		}
		var metrics []model.Metric
		if err := json.Unmarshal(rec.Body.Bytes(), &metrics); err != nil {
			t.Fatalf("decode /api/metrics: %v", err)
		}
		if len(metrics) != 1 || metrics[0].BaseUnitDecimals == nil || *metrics[0].BaseUnitDecimals != 1 {
			t.Fatalf("unexpected metrics payload: %+v", metrics)
		}

		errRouter := NewRouterWithDeps(cfg, RouterDeps{
			Reader: mockReader{
				listAstronautsFn: func(ctx context.Context) ([]model.Astronaut, error) {
					return nil, errors.New("postgres://secret@db/hrig connection refused")
				},
			},
		})
		rec = httptest.NewRecorder()
		errRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/astronauts", nil))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 on unexpected db error, got %d", rec.Code)
		}
		if strings.Contains(rec.Body.String(), "secret") || !strings.Contains(rec.Body.String(), "internal server error") {
			t.Fatalf("expected sanitized 500 response, got %s", rec.Body.String())
		}
	})

	t.Run("trends_200_and_400_cases", func(t *testing.T) {
		val := 98.0
		var gotFilter query.TrendsFilter
		router := NewRouterWithDeps(cfg, RouterDeps{
			Reader: mockReader{
				listTrendsFn: func(ctx context.Context, filter query.TrendsFilter) ([]model.TrendPoint, error) {
					gotFilter = filter
					if filter.Limit > 1000 {
						return nil, errors.Join(query.ErrBadRequest, errors.New("limit out of bounds"))
					}
					return []model.TrendPoint{
						{
							ID:            "11111111-1111-4111-8111-111111111111",
							MeasurementID: "22222222-2222-4222-8222-222222222222",
							MetricID:      "33333333-3333-4333-8333-333333333333",
							MetricCode:    filter.MetricCode,
							DeviceID:      "44444444-4444-4444-8444-444444444444",
							Environment:   model.EnvIVA,
							ValueNumeric:  &val,
							Status:        model.StatusNormal,
							RecordedAtUTC: time.Date(2026, 11, 15, 8, 0, 0, 0, time.UTC),
							MET:           "T+002d 08:00:00",
						},
					}, nil
				},
			},
		})

		// Valid request with all query params
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(
			http.MethodGet,
			"/api/trends?metric_code=spo2&astronaut_id=11111111-1111-4111-8111-111111111111&since=2026-11-14T00:00:00Z&limit=25",
			nil,
		))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /api/trends expected 200, got %d (%s)", rec.Code, rec.Body.String())
		}
		if gotFilter.MetricCode != "spo2" || gotFilter.AstronautID == nil || *gotFilter.AstronautID != "11111111-1111-4111-8111-111111111111" || gotFilter.Limit != 25 || gotFilter.SinceUTC == nil {
			t.Fatalf("unexpected TrendsFilter parsed: %+v", gotFilter)
		}
		var points []model.TrendPoint
		if err := json.Unmarshal(rec.Body.Bytes(), &points); err != nil {
			t.Fatalf("decode /api/trends: %v", err)
		}
		if len(points) != 1 || points[0].MetricCode != "spo2" {
			t.Fatalf("unexpected trends payload: %+v", points)
		}

		// 400 validation cases
		for _, badURL := range []string{
			"/api/trends",
			"/api/trends?metric_code=spo2&since=not-rfc3339",
			"/api/trends?metric_code=spo2&limit=abc",
			"/api/trends?metric_code=spo2&limit=5000",
		} {
			rec = httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, badURL, nil))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("%s expected 400, got %d (%s)", badURL, rec.Code, rec.Body.String())
			}
		}
	})

	t.Run("history_200_and_400_cases", func(t *testing.T) {
		val := 72.0
		var gotFilter query.HistoryFilter
		router := NewRouterWithDeps(cfg, RouterDeps{
			Reader: mockReader{
				listHistoryFn: func(ctx context.Context, filter query.HistoryFilter) ([]model.HistoryEntry, error) {
					gotFilter = filter
					if filter.Offset < 0 {
						return nil, errors.Join(query.ErrBadRequest, errors.New("offset must be non-negative"))
					}
					return []model.HistoryEntry{
						{
							MeasurementID: "11111111-1111-4111-8111-111111111111",
							ValueID:       "22222222-2222-4222-8222-222222222222",
							DeviceCode:    "biomonitor_01",
							DeviceName:    "IVA biomonitor",
							Environment:   model.EnvIVA,
							MetricCode:    "heart_rate",
							MetricName:    "Heart rate",
							ValueType:     model.ValueNumeric,
							ValueNumeric:  &val,
							Status:        model.StatusNormal,
							RecordedAtUTC: time.Date(2026, 11, 15, 8, 0, 0, 0, time.UTC),
							MET:           "T+002d 08:00:00",
						},
					}, nil
				},
			},
		})

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(
			http.MethodGet,
			"/api/history?astronaut_id=11111111-1111-4111-8111-111111111111&limit=10&offset=20",
			nil,
		))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /api/history expected 200, got %d (%s)", rec.Code, rec.Body.String())
		}
		if gotFilter.AstronautID == nil || *gotFilter.AstronautID != "11111111-1111-4111-8111-111111111111" || gotFilter.Limit != 10 || gotFilter.Offset != 20 {
			t.Fatalf("unexpected HistoryFilter parsed: %+v", gotFilter)
		}
		var entries []model.HistoryEntry
		if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
			t.Fatalf("decode /api/history: %v", err)
		}
		if len(entries) != 1 || entries[0].MetricCode != "heart_rate" {
			t.Fatalf("unexpected history payload: %+v", entries)
		}

		for _, badURL := range []string{
			"/api/history?limit=notint",
			"/api/history?offset=notint",
			"/api/history?offset=-5",
		} {
			rec = httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, badURL, nil))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("%s expected 400, got %d (%s)", badURL, rec.Code, rec.Body.String())
			}
		}
	})
}

