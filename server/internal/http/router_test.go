package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hrig/server/internal/config"
)

func TestRootHelloWorld(t *testing.T) {
	cfg := config.Load()
	router := NewRouter(cfg)

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
