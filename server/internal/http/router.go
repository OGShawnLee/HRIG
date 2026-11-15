package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"hrig/server/internal/config"
	"hrig/server/internal/ingestion"
)

const maxIngestBodyBytes = 1 << 20 // 1 MiB

type Ingester interface {
	Ingest(ctx context.Context, req ingestion.IngestRequest) (ingestion.IngestResult, error)
}

func NewRouter(cfg config.Config, ingester Ingester) http.Handler {
	r := chi.NewRouter()

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

	r.Post("/", handleIngest(ingester))

	return r
}

func handleIngest(ingester Ingester) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
