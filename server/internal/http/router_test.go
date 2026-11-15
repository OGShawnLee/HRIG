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

	"hrig/server/internal/config"
	"hrig/server/internal/ingestion"
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
