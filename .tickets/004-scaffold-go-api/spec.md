# 004 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Monorepo Workspace & Shared Models (`shared/`)
- [ ] `go.work` exists at repository root and lists `./app`, `./server`, and `./shared`.
- [ ] `shared/go.mod` declares `module hrig/shared` with zero third-party `require` dependencies.
- [ ] `shared/model/telemetry.go` defines `ReadingStatus` (`normal`, `abnormal`, `critical`, `stale`), `EnvironmentMode` (`iva`, `eva`), `MetricTier`, `ValueType`, `DeviceKind`, `MediaKind`, `Astronaut`, `Unit`, `Metric`, `MetricThreshold`, `TelemetryFrame`, `LatestReading`, `CheckInRequest`, `AlertEvent`, `RealtimeEnvelope`, and `HealthStatus` using concrete Wails-binding-safe types (no `any` / `interface{}`).
- [ ] `go vet ./... && go test ./...` inside `shared/` exits `0`.

## API Scaffold, Evaluation & Store Contract (`server/`)
- [ ] `server/go.mod` declares `module hrig/server` and requires `hrig/shared`, `github.com/go-chi/chi/v5`, `github.com/go-chi/cors`, and `github.com/gorilla/websocket` (no SQL/pgx driver added yet).
- [ ] `server/.env.example` exists and documents `HRIG_API_ADDR`, `HRIG_ALLOWED_ORIGINS`, `HRIG_STALE_AFTER`, `HRIG_MISSION_EPOCH_UTC`, and a commented `DATABASE_URL` placeholder referencing Ticket `005`.
- [ ] `server/internal/threshold/evaluator.go` evaluates `critical` limits before `warning` limits and handles `manual` mode; `go test ./internal/threshold/...` passes.
- [ ] `server/internal/store/store.go` defines `TelemetryStore` using `hrig/shared/model` with no SQL/ORM imports and no alert-persistence methods (`alerts` stay in-memory on `ws.Hub`).
- [ ] `server/internal/store/memory.go` implements `TelemetryStore` (`BackendName() == "memory"`) and enforces `AstronautID == nil` when `MetricType == 2`.
- [ ] `server/internal/store/writer.go` implements non-blocking buffered writes via a bounded channel.
- [ ] `server/internal/ws/hub.go` serves both `/ws/ingest` (bidirectional) and `/ws/realtime` (subscriber broadcast) and runs the `STALE / NO DATA` watchdog (`RNF10`).
- [ ] `server/internal/http/router.go` exposes `GET /health`, `GET /api/snapshot`, `GET /api/astronauts`, `GET /api/metrics`, `GET /api/trends`, `GET /api/history`, `POST /api/checkins`, `POST /api/alerts/{id}/ack`, `GET /ws/ingest`, and `GET /ws/realtime`.
- [ ] `wails3 task server:check` exits `0` (`go vet`, `go test`, and `go build` all pass).
- [ ] `wails3 task client:check` exits `0` (confirming `go.work` and `shared/` do not break `app/`).

## Scope & Ground Rules
- [ ] `git status --short` shows no changes in `app/**`, `db/**`, `.routing/**`, `docs/**`, `README.md`, `PROJECT-REVIEW.md`, or `.claude/**`.
- [ ] No unverified NASA clinical threshold numbers were hardcoded in Go source.
- [ ] No git commits or pushes were executed.
- [ ] `.tickets/004-scaffold-go-api/changelog.md` exists and records files touched and command outputs.
