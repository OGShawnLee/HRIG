# 011 — Verification Spec

Every check is binary (`PASS` / `FAIL`). All checks must pass before ticket `011` is accepted.

## 1. Go Module & Dependencies (`app/`)

- [x] `app/go.mod` declares `hrig/shared v0.0.0` with `replace hrig/shared => ../shared` and direct `github.com/coder/websocket`, and `GOWORK=off go -C app vet .` exits with `0`.
- [x] `app/go.mod` does **not** depend on `github.com/gorilla/websocket` or `server/`.

## 2. REST Client, WebSocket Bridge & Bindings (`app/client.go`, `app/app.go`)

- [x] `app/client.go` reads `HRIG_API_URL` (default `http://127.0.0.1:80`), derives `<wsBase>/ws/realtime`, and decodes `model.Astronaut`, `model.Metric`, `model.Snapshot`, `model.TrendPoint`, `model.HistoryEntry`, `model.AlertEvent`, and `model.HealthStatus`.
- [x] `RunRealtimeLoop` connects via `github.com/coder/websocket`, emits `"telemetry:connection"` (`ConnectionState`) and `"telemetry:envelope"` (`model.RealtimeEnvelope`) events without blocking startup, and shuts down cleanly when context is canceled.
- [x] `app/app.go` preserves `OnMinimize`, `OnMaximize`, `OnClose`, `Ping`, implements `ServiceStartup` and `ServiceShutdown`, and exposes `GetConnectionStatus`, `GetHealth`, `GetAstronauts`, `GetMetrics`, `GetSnapshot`, `GetTrends`, `GetHistory`, and `AckAlert`.

## 3. Generated TypeScript Bindings & Tests (`app/frontend/bindings/`, `app/client_test.go`)

- [x] `app/frontend/bindings/hrig/app/app.ts` exports `GetConnectionStatus`, `GetHealth`, `GetAstronauts`, `GetMetrics`, `GetSnapshot`, `GetTrends`, `GetHistory`, `AckAlert`, alongside `OnClose`, `OnMaximize`, `OnMinimize`, `Ping`.
- [x] `app/frontend/bindings/hrig/shared/model/models.ts` (and `index.ts`) exists and exports `Astronaut`, `Metric`, `Snapshot`, `LatestReading`, `AlertEvent`, `RealtimeEnvelope`, `TrendPoint`, `HistoryEntry`, and `HealthStatus`.
- [x] `go -C app test -v -count=1 ./...` exits with `0` (all unit & `httptest.Server` REST/WS bridge tests pass).
- [x] `wails3 task client:check` exits with `0` (bindings regeneration + `vue-tsc --noEmit` + `vite build` + `go build .`).

## 4. Scope & Ground Rules

- [x] `git diff --name-only` and untracked files show changes only inside `May create or edit` paths listed in `task.md`.
- [x] `server/**`, `bot/**`, `shared/**`, `db/**`, `app/main.go`, `app/app_window.go`, and `app/frontend/src/**` remain completely unmodified.
- [x] Code, `.tickets/`, and `changelog.md` are written in English; `.tickets/011-wails-backend-api-bindings/changelog.md` records all changed files and exact command outputs.
