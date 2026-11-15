package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"hrig/server/internal/config"
	"hrig/server/internal/ingestion"
	"hrig/server/internal/realtime"
	"hrig/shared/model"
)

const (
	maxIngestBodyBytes = 1 << 20 // 1 MiB
	maxAckBodyBytes    = 64 << 10
)

type Ingester interface {
	Ingest(ctx context.Context, req ingestion.IngestRequest) (ingestion.IngestResult, error)
}

type RealtimeController interface {
	ServeWS(allowedOrigins []string) http.HandlerFunc
	Snapshot() model.Snapshot
	AckAlert(id string, acknowledgedBy string) (model.AlertEvent, error)
	ClientCount() int
	ActiveAlertsCount() int
}

type DBPinger interface {
	Ping(ctx context.Context) error
}

type RouterDeps struct {
	Ingester Ingester
	Realtime RealtimeController
	DBPinger DBPinger
}

func NewRouter(cfg config.Config, ingester Ingester) http.Handler {
	return NewRouterWithDeps(cfg, RouterDeps{Ingester: ingester})
}

func NewRouterWithDeps(cfg config.Config, deps RouterDeps) http.Handler {
	r := chi.NewRouter()
	var activeIngests atomic.Int64

	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, struct {
			Message string `json:"message"`
		}{
			Message: "hello world",
		})
	})

	r.Post("/", handleIngest(deps.Ingester, &activeIngests))
	r.Get("/health", handleHealth(deps.Realtime, deps.DBPinger, &activeIngests))
	r.Get("/ws/realtime", handleRealtimeWS(cfg.AllowedOrigins, deps.Realtime))
	r.Get("/api/snapshot", handleSnapshot(deps.Realtime))
	r.Post("/api/alerts/{id}/ack", handleAckAlert(deps.Realtime))

	return r
}

func handleIngest(ingester Ingester, activeIngests *atomic.Int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeIngests.Add(1)
		defer activeIngests.Add(-1)

		ct := r.Header.Get("Content-Type")
		mediaType, _, err := mime.ParseMediaType(ct)
		if err != nil || mediaType != "application/json" {
			writeError(w, http.StatusBadRequest, "Content-Type must be application/json")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxIngestBodyBytes)
		dec := json.NewDecoder(r.Body)

		var req ingestion.IngestRequest
		if err := dec.Decode(&req); err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				writeError(w, http.StatusBadRequest, "request body exceeds 1 MiB limit")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}

		var trailing any
		if err := dec.Decode(&trailing); err != io.EOF {
			writeError(w, http.StatusBadRequest, "invalid JSON payload: unexpected trailing data")
			return
		}

		if ingester == nil {
			writeError(w, http.StatusInternalServerError, "ingestion service unavailable")
			return
		}

		res, err := ingester.Ingest(r.Context(), req)
		if err != nil {
			status, msg := ingestion.StatusCodeFromError(err)
			writeError(w, status, msg)
			return
		}

		writeJSON(w, http.StatusCreated, res)
	}
}

func handleRealtimeWS(allowedOrigins []string, rt RealtimeController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rt == nil {
			writeError(w, http.StatusServiceUnavailable, "realtime service unavailable")
			return
		}
		rt.ServeWS(allowedOrigins)(w, r)
	}
}

func handleSnapshot(rt RealtimeController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rt == nil {
			writeError(w, http.StatusServiceUnavailable, "realtime service unavailable")
			return
		}
		writeJSON(w, http.StatusOK, rt.Snapshot())
	}
}

func handleAckAlert(rt RealtimeController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rt == nil {
			writeError(w, http.StatusServiceUnavailable, "realtime service unavailable")
			return
		}

		id := strings.TrimSpace(chi.URLParam(r, "id"))
		if id == "" {
			writeError(w, http.StatusBadRequest, "alert id is required")
			return
		}

		var ackReq model.AckAlertRequest
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxAckBodyBytes)
			dec := json.NewDecoder(r.Body)
			dec.DisallowUnknownFields()

			err := dec.Decode(&ackReq)
			if err != nil && !errors.Is(err, io.EOF) {
				var maxBytesErr *http.MaxBytesError
				if errors.As(err, &maxBytesErr) {
					writeError(w, http.StatusBadRequest, "request body exceeds limit")
					return
				}
				writeError(w, http.StatusBadRequest, "invalid JSON payload")
				return
			}
			if err == nil {
				var trailing any
				if err := dec.Decode(&trailing); err != io.EOF {
					writeError(w, http.StatusBadRequest, "invalid JSON payload: unexpected trailing data")
					return
				}
			}
		}

		alert, err := rt.AckAlert(id, ackReq.AcknowledgedBy)
		if err != nil {
			if errors.Is(err, realtime.ErrAlertNotFound) {
				writeError(w, http.StatusNotFound, "alert not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusOK, alert)
	}
}

func handleHealth(rt RealtimeController, pinger DBPinger, activeIngests *atomic.Int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := "degraded"
		if pinger != nil {
			pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			err := pinger.Ping(pingCtx)
			cancel()
			if err == nil {
				store = "postgres"
			}
		}

		var activeAlerts, realtimeClients int
		if rt != nil {
			activeAlerts = rt.ActiveAlertsCount()
			realtimeClients = rt.ClientCount()
		}

		writeJSON(w, http.StatusOK, model.HealthStatus{
			Status:          "ok",
			Store:           store,
			TimeUTC:         time.Now().UTC(),
			ActiveAlerts:    activeAlerts,
			RealtimeClients: realtimeClients,
			IngestClients:   int(activeIngests.Load()),
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, struct {
		Error string `json:"error"`
	}{
		Error: message,
	})
}
