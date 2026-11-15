# 004 — Scaffold Go API (`Chi` + `Gorilla WebSocket` + shared `hrig/shared/model`)

| Field | Value |
| :--- | :--- |
| **ID** | `004` |
| **Status** | `draft` |
| **Worker** | `antigravity` / `gemini-4-hydrogen` (`thinking_level: medium`) |
| **Role** | Go Backend & Monorepo Scaffolder |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `go.work` (linking `./app`, `./server`, and `./shared`)
  - `Taskfile.yml` (only add `server:check`, `server:test`, `server:build`, and `server:dev` tasks; keep `client:*` and `db:*` tasks intact)
  - `shared/go.mod`
  - `shared/model/telemetry.go`
  - `shared/model/telemetry_test.go`
  - `server/go.mod`, `server/go.sum`
  - `server/.env.example`
  - `server/cmd/server/main.go`
  - `server/internal/config/config.go`
  - `server/internal/threshold/evaluator.go`
  - `server/internal/threshold/evaluator_test.go`
  - `server/internal/store/store.go`
  - `server/internal/store/memory.go`
  - `server/internal/store/writer.go`
  - `server/internal/ws/hub.go`
  - `server/internal/ws/hub_test.go`
  - `server/internal/http/router.go`
  - `server/internal/http/router_test.go`
  - `.tickets/004-scaffold-go-api/changelog.md`
- **Must not touch:**
  - `app/**`
  - `db/**`, `local/**`
  - `.gitignore`
  - `.routing/**`
  - `.claude/**`
  - `docs/**`, `README.md`, `PROJECT-REVIEW.md`
  - `.tickets/001-scaffold-wails3-app/**`, `.tickets/002-setup-portable-postgres-17/**`, `.tickets/003-full-stack-install-guide/**`, `.tickets/005-postgres-db-integration/**`
  - `.tickets/004-scaffold-go-api/task.md`, `.tickets/004-scaffold-go-api/spec.md`

## Inputs

Read these files before acting:
1. [`app/go.mod`](../../app/go.mod)
2. [`Taskfile.yml`](../../Taskfile.yml)
3. [`.routing/CONTEXT.md`](../../.routing/CONTEXT.md)
4. [`.routing/BD.md`](../../.routing/BD.md)
5. [`.routing/BDML.md`](../../.routing/BDML.md)
6. [`db/001_schema.sql`](../../db/001_schema.sql)
7. [`db/002_seed.sql`](../../db/002_seed.sql)
8. [`.routing/SPECIFICATION.md`](../../.routing/SPECIFICATION.md) (`RF01`–`RF03`, `RF08`, `RF10`, `RNF06`, `RNF09`, `RNF10`, `RNF11`)
9. [`PROJECT-REVIEW.md`](../../PROJECT-REVIEW.md)
10. [`.tickets/005-postgres-db-integration/task.md`](../005-postgres-db-integration/task.md)

## Grounding & Anti-Hallucination Constraints

- **Shared Models Monorepo Package (`shared/`):**
  - `shared/go.mod` must declare `module hrig/shared` (`go 1.26.0`) with **zero external dependencies** (Go standard library only) so `app/` can later import `hrig/shared/model` and generate 100% type-safe TypeScript bindings via `wails3 generate bindings` without pulling `chi`, `gorilla/websocket`, or `pgx` into the Wails desktop binary.
  - All exported types in `shared/model/telemetry.go` must use concrete Go types (`string`, `int`, `int16`, `float64`, `bool`, `time.Time`, typed slices, pointers for nullable fields)—do **not** use `any` / `interface{}` or `map[string]any` on DTOs so Wails TypeScript binding generation stays 100% typed.
- **Root Workspace & Server Dependencies:**
  - Root `go.work`: `go 1.26.0`, `use ( ./app ./server ./shared )`.
  - `server/go.mod`: `module hrig/server`, `go 1.26.0`, requiring `hrig/shared v0.0.0` (with `replace hrig/shared => ../shared`), `github.com/go-chi/chi/v5`, `github.com/go-chi/cors`, and `github.com/gorilla/websocket`.
  - Do **not** add `pgx`, `lib/pq`, or SQL queries in this ticket—keep persistence behind the `TelemetryStore` interface (`server/internal/store/store.go`) with `MemoryStore` fallback; Ticket `005` adds `PostgresStore`.
- **Exact DB Schema Enum & Domain Alignment (`db/001_schema.sql` & `.routing/BD.md`):**
  - `ReadingStatus`: `"normal"`, `"abnormal"`, `"critical"`, plus runtime-only `"stale"` (`RNF10`; never persisted to DB). Do **not** use `"nominal"` or `"caution"`.
  - `EnvironmentMode`: `"iva"`, `"eva"`.
  - `MetricTier`: `"primary"`, `"secondary"`.
  - `ValueType`: `"numeric"`, `"scale"`, `"text"`, `"image"`.
  - `DeviceKind`: `"wearable"`, `"environment_sensor"`, `"manual"`, `"lab_assay"`, `"imaging"`.
  - `MediaKind`: `"ultrasound"`, `"ocular"`.
  - Environmental metrics (`metric_type = 2`, e.g., `cabin_o2`, `cabin_ppco2`, `cabin_pressure`, `cabin_temp`, `suit_pressure`, `suit_temp`) must enforce `AstronautID == nil` (matching `trg_measurement_metric_type` in `db/001_schema.sql`).
  - Alerts are **not** stored in PostgreSQL (`.routing/BD.md` D7); keep active alerts and acknowledgements in-memory on `ws.Hub` (`POST /api/alerts/{id}/ack`).
- **Threshold Classification & Stale Defaults:**
  - Implement threshold classification in `server/internal/threshold/evaluator.go` matching `.routing/BD.md` §5: for `classification_mode == "numeric"`, evaluate `critical_min` / `critical_max` first (at/beyond -> `"critical"`), then `warning_min` / `warning_max` (beyond -> `"abnormal"`), else `"normal"`; `nil` means unbounded on that side; for `classification_mode == "manual"`, return `"normal"` unless an explicit valid `ReadingStatus` is passed.
  - Do **not** hardcode unverified clinical threshold constants in Go code—`MemoryStore` returns empty/caller-seeded thresholds cleanly until Ticket `005` reads `metric_threshold` from PostgreSQL.
  - Mark `HRIG_STALE_AFTER` (`10s` default) explicitly as a configurable runtime timeout, not a NASA clinical rule.
  - Do **not** implement JWT, login, or external system auth (`PROJECT-REVIEW.md` cuts `RF09`/`RNF12`).
- **Do NOT commit or push** any changes to Git.

## Task

1. Create `shared/go.mod` (`module hrig/shared`, `go 1.26.0`, no third-party dependencies).
2. Create `shared/model/telemetry.go` defining Wails-binding-friendly structs and typed string constants aligned 1:1 with `db/001_schema.sql`:
   - Enums: `ReadingStatus` (`normal`, `abnormal`, `critical`, `stale`), `EnvironmentMode` (`iva`, `eva`), `MetricTier` (`primary`, `secondary`), `ValueType` (`numeric`, `scale`, `text`, `image`), `DeviceKind` (`wearable`, `environment_sensor`, `manual`, `lab_assay`, `imaging`), `MediaKind` (`ultrasound`, `ocular`).
   - Catalog & profile DTOs: `Astronaut` (including derived `Age int16`), `Unit`, `Metric`, `MetricThreshold`, `Device`.
   - Telemetry & session DTOs: `TelemetryFrame` (`Source`, `DeviceCode`, `AstronautID *string`, `Environment EnvironmentMode`, `MetricType int16`, `MetricCode string`, `UnitCode *string`, `ValueNumeric *float64`, `ValueScale *int16`, `ValueText *string`, `Status ReadingStatus`, `RecordedAtUTC time.Time`, `MET string`, `Note *string`), `LatestReading` (embeds/carries `TelemetryFrame` + `LastSeenUTC time.Time` + `IsStale bool`), `CheckInMedia`, `CheckInRequest`, `TrendQuery`, `HistoryQuery`, `AlertEvent`, `RealtimeEnvelope` (tagged payload for WS `/ws/realtime`: `Type string` `"snapshot"|"reading"|"stale"|"alert"|"alert_ack"`, plus typed optional fields), `HealthStatus`.
3. Create `shared/model/telemetry_test.go` verifying JSON round-trip serialization and `ReadingStatus` / `EnvironmentMode` validation helpers (`IsValid()`).
4. Create root `go.work` with `go 1.26.0` referencing `./app`, `./server`, and `./shared`.
5. Initialize `server/go.mod` (`module hrig/server`, `go 1.26.0`) requiring `hrig/shared v0.0.0` (`replace hrig/shared => ../shared`), `github.com/go-chi/chi/v5`, `github.com/go-chi/cors`, and `github.com/gorilla/websocket`.
6. Create `server/.env.example` and `server/internal/config/config.go` with defaults:
   - `HRIG_API_ADDR` = `:8080`
   - `HRIG_ALLOWED_ORIGINS` = `http://127.0.0.1:9245,http://localhost:9245,wails://wails`
   - `HRIG_STALE_AFTER` = `10s`
   - `HRIG_MISSION_EPOCH_UTC` = optional RFC3339 string (empty disables derived `MET`)
   - Commented placeholder `# DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable # Wired in Ticket 005`
7. Create `server/internal/threshold/evaluator.go` + `evaluator_test.go` implementing `.routing/BD.md` §5 evaluation (`Evaluate(th *model.MetricThreshold, value *float64, requested model.ReadingStatus) model.ReadingStatus`: critical boundary checked `<= critical_min` / `>= critical_max`, warning boundary checked `< warning_min` / `> warning_max`).
8. Create `server/internal/store/store.go` defining `TelemetryStore`:
   - `BackendName() string`
   - `Ping(ctx context.Context) error`
   - `ListAstronauts(ctx context.Context) ([]model.Astronaut, error)`
   - `ListMetrics(ctx context.Context) ([]model.Metric, error)`
   - `GetThresholdByMetricCode(ctx context.Context, metricCode string) (*model.MetricThreshold, error)`
   - `SaveTelemetry(ctx context.Context, frame model.TelemetryFrame) error`
   - `LatestReadings(ctx context.Context) ([]model.LatestReading, error)`
   - `QueryTrends(ctx context.Context, q model.TrendQuery) ([]model.TelemetryFrame, error)`
   - `ListHistory(ctx context.Context, q model.HistoryQuery) ([]model.TelemetryFrame, error)`
   - `SaveCheckIn(ctx context.Context, req model.CheckInRequest, status model.ReadingStatus) (model.TelemetryFrame, error)`
   - Include `// TODO(ticket-005): Implement PostgresStore satisfying TelemetryStore against db/001_schema.sql.`
9. Create `server/internal/store/memory.go` implementing `TelemetryStore` (`BackendName() == "memory"`, thread-safe with `sync.RWMutex`, enforcing `AstronautID == nil` when `MetricType == 2`) and `server/internal/store/writer.go` (`BufferedWriter` with bounded channel cap `256` and graceful `Close()` drain).
10. Create `server/internal/ws/hub.go` using `gorilla/websocket`:
    - `HandleIngest(w, r)`: bidirectional WS for wearables and environment sensors (`?role=wearable|ship&id=<codeOrUuid>`), parses `model.TelemetryFrame`, rejects `MetricType == 2` with non-nil `AstronautID`, evaluates threshold via `threshold.Evaluate`, updates latest snapshot + `lastSeen`, queues `BufferedWriter.Enqueue(frame)`, broadcasts `reading` envelope to `/ws/realtime`, and when `Status == "critical"`, creates an in-memory `AlertEvent` and broadcasts `alert` to both `/ws/realtime` and connected bidirectional wearable `/ws/ingest` sockets.
    - `HandleRealtime(w, r)`: sends initial `snapshot` envelope on connect, then streams `reading`, `stale`, `alert`, and `alert_ack` envelopes.
    - `RunStaleWatchdog(ctx)`: checks `lastSeen` every `1s`; transitions sources idle past `HRIG_STALE_AFTER` to `Status: "stale"` (`IsStale: true`, keeping last `RecordedAtUTC`) and broadcasts `stale` envelope (`RNF10`).
    - `AckAlert(alertID, ackBy string) (model.AlertEvent, bool)` and `ActiveAlerts() []model.AlertEvent`.
11. Create `server/internal/http/router.go` using `go-chi/chi/v5` and `go-chi/cors`:
    - `GET /health`
    - `GET /api/snapshot` (returns latest readings + active alerts)
    - `GET /api/astronauts`
    - `GET /api/metrics`
    - `GET /api/trends` (`astronautId`, `metricCode`, `limit`)
    - `GET /api/history` (`astronautId`, `limit`)
    - `POST /api/checkins` (validates `model.CheckInRequest`, classifies via `threshold.Evaluate`, calls `store.SaveCheckIn`, broadcasts `reading`/`alert` on `hub`, returns `201`)
    - `POST /api/alerts/{id}/ack` (calls `hub.AckAlert`, broadcasts `alert_ack`, returns `200` or `404`)
    - `GET /ws/ingest` and `GET /ws/realtime`
12. Create `server/cmd/server/main.go` wiring config, `MemoryStore`, `BufferedWriter`, `ws.Hub`, `http.NewRouter`, and graceful shutdown on `SIGINT`/`SIGTERM`.
13. Write tests in `server/internal/ws/hub_test.go` and `server/internal/http/router_test.go` covering `/health`, `/api/astronauts`, `/api/metrics`, `POST /api/checkins`, `GET /api/trends`, `WS /ws/ingest` -> `WS /ws/realtime` fan-out, bidirectional critical `alert` push + `/api/alerts/{id}/ack`, `MetricType == 2` non-null `AstronautID` rejection, and `stale` watchdog transition.
14. Update root `Taskfile.yml` to add `server:check` (`go vet ./... && go test ./... && go build -o ../bin/hrig-server ./cmd/server` / `.exe` or `go build ./cmd/server` with binary cleanup), `server:test`, `server:build`, and `server:dev` inside `dir: server` (and `shared:check` running `go vet ./... && go test ./...` inside `dir: shared`).
15. Run `wails3 task server:check`, `go test ./...` in `shared/`, and `wails3 task client:check` from repo root, confirming all exit `0`.
16. Write `.tickets/004-scaffold-go-api/changelog.md`.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/004-scaffold-go-api/changelog.md` listing:
- Files created and modified
- Verification checks executed and their outputs
- Any blockers or assumptions avoided
