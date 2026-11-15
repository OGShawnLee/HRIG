# 009 — Set up Gorilla WebSocket realtime server & alert/stale core

| Field | Value |
| :--- | :--- |
| **ID** | `009` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Go Realtime WebSocket & Backend Engineer |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `server/go.mod`, `server/go.sum` (add `github.com/gorilla/websocket`)
  - `server/internal/config/config.go`, `server/.env.example` (add `StaleAfter` from `HRIG_STALE_AFTER` defaulting to `15s`, and `MissionEpochUTC` from `HRIG_MISSION_EPOCH_UTC`)
  - `server/internal/realtime/*.go` (`hub.go`, `client.go`, `met.go`, `hub_test.go`, `ws_test.go` — Hub, Gorilla WebSocket upgrader/pumps, post-commit frame & critical alert fan-out, stale monitor, snapshot, alert ack, MET helper)
  - `server/internal/ingestion/types.go`, `server/internal/ingestion/service.go`, `server/internal/ingestion/service_integration_test.go` (optional post-commit `FramePublisher` hook publishing persisted frames strictly after `tx.Commit` succeeds)
  - `server/internal/http/router.go`, `server/internal/http/router_test.go` (`NewRouterWithDeps`, `GET /ws/realtime`, `GET /api/snapshot`, `POST /api/alerts/{id}/ack`, `GET /health` while preserving `NewRouter`, `GET /`, and `POST /`)
  - `server/cmd/server/main.go` (hydrate Hub from DB, start `RunStaleMonitor`, wire Hub to ingestion service and router, shut down gracefully)
  - `.tickets/009-realtime-gorilla-ws-server/changelog.md`
- **Must not touch:**
  - `shared/**`, `bot/**`, `app/**`, `local/**`, `docs/**`
  - `db/001_schema.sql`, `db/002_seed.sql`, `db/003_demo_seed.sql`, `db/README.md`, `db/queries/**`
  - `sqlc.yaml`, `server/internal/db/*.go`
  - `Taskfile.yml`, `go.work`
  - `.gitignore`, `.routing/**`, `.claude/**`, `README.md`, `PROJECT-REVIEW.md`, `GUIDE.md`, `SENSOR-READINGS.md`
  - `.tickets/001-*` .. `.tickets/008-*`
  - `.tickets/009-realtime-gorilla-ws-server/task.md`, `.tickets/009-realtime-gorilla-ws-server/spec.md`

## Inputs

Read these files before acting:
1. [`shared/model/telemetry.go`](../../shared/model/telemetry.go)
2. [`shared/model/ingest.go`](../../shared/model/ingest.go)
3. [`server/go.mod`](../../server/go.mod)
4. [`server/internal/config/config.go`](../../server/internal/config/config.go)
5. [`server/internal/ingestion/types.go`](../../server/internal/ingestion/types.go)
6. [`server/internal/ingestion/service.go`](../../server/internal/ingestion/service.go)
7. [`server/internal/http/router.go`](../../server/internal/http/router.go)
8. [`server/internal/http/router_test.go`](../../server/internal/http/router_test.go)
9. [`server/cmd/server/main.go`](../../server/cmd/server/main.go)
10. [`server/internal/db/querier.go`](../../server/internal/db/querier.go)
11. [`server/internal/db/telemetry.sql.go`](../../server/internal/db/telemetry.sql.go)

## Grounding & Anti-Hallucination Constraints

- **Use Canonical `hrig/shared/model` Types:**
  - Wire WebSocket payloads must use `model.RealtimeEnvelope` with `Type` in `{"snapshot", "reading", "stale", "alert", "alert_ack"}`.
  - Snapshot payload must use `model.Snapshot`; health payload must use `model.HealthStatus`; alert ack request must use `model.AckAlertRequest`; frames must use `model.TelemetryFrame`, `model.LatestReading`, and `model.AlertEvent`.
- **PostgreSQL Fidelity (`RNF10`, `.routing/BD.md` D7):**
  - `model.StatusStale` (`"stale"`) is **runtime-only** in memory and over WebSockets — **never** persist `"stale"` to PostgreSQL.
  - `model.AlertEvent` alerts are **in-memory only** (triggered whenever a committed reading has `Status == model.StatusCritical`); **never** attempt to insert alerts into PostgreSQL.
  - Publish frames to the realtime hub **strictly after** `tx.Commit(ctx)` succeeds in `ingestion.Service.Ingest`; rolled-back or failed `POST /` requests must publish zero frames and zero alerts.
- **Gorilla WebSocket Concurrency Rules:**
  - Use `github.com/gorilla/websocket` (`websocket.Upgrader`) with `CheckOrigin` validating non-empty `Origin` headers against `cfg.AllowedOrigins` (exact case-insensitive match, allowing empty `Origin` for native/CLI WebSocket clients).
  - Enforce Gorilla's single-writer rule: every `*websocket.Conn` has one dedicated `writePump` reading from a buffered `send chan model.RealtimeEnvelope` (`64`), plus `readPump` with `SetReadLimit`, `SetReadDeadline`, and `SetPongHandler`.
  - Never block HTTP `POST /` callers on slow WebSocket clients: drop/close slow clients whose `send` buffer is full using non-blocking `select`.
- **Preserve Existing Behavior:**
  - Preserve `GET /` (`{"message":"hello world"}`) and `POST /` (`201 Created`, `400`, `404`, `409`, `500`) behavior verbatim.
  - Keep `NewRouter(cfg config.Config, ingester Ingester) http.Handler` and `NewService(pool *pgxpool.Pool)` signatures working for existing tests; add `NewRouterWithDeps(cfg config.Config, deps RouterDeps)` and `Service.SetPublisher(pub FramePublisher)` (or `NewServiceWithPublisher`).
- **No Unsolicited Git / Daemon Side Effects:**
  - Do **not** commit or push changes.
  - Do **not** leave `hrig-server.exe`, `hrig-bot.exe`, PostgreSQL, or `hrig_test` running after verification.

## Task

1. **Add `github.com/gorilla/websocket` to `server/`:**
   - Inside `server/`, run `go get github.com/gorilla/websocket@v1.5.3` (or latest compatible `v1.5.x`) and verify `server/go.mod` and `server/go.sum` update cleanly (`GOWORK=off go -C server mod tidy` if needed).
2. **Extend `server/internal/config/config.go` & `server/.env.example`:**
   - Add `StaleAfter time.Duration` (read `HRIG_STALE_AFTER`, parse with `time.ParseDuration`, default `15 * time.Second` if empty, invalid, or `<= 0`).
   - Add `MissionEpochUTC time.Time` (read `HRIG_MISSION_EPOCH_UTC` RFC3339/RFC3339Nano, defaulting to `time.Now().UTC()` at `Load()` if empty or invalid).
   - Document `HRIG_STALE_AFTER=15s` and optional `HRIG_MISSION_EPOCH_UTC` in `server/.env.example`.
3. **Create `server/internal/realtime/` Hub, MET Helper & Gorilla WebSocket Transport:**
   - `met.go`: implement `FormatMET(recordedAtUTC, epochUTC time.Time) string` returning `T+000d HH:MM:SS` (or `T-000d HH:MM:SS` if `recordedAtUTC.Before(epochUTC)`).
   - `hub.go`:
     - Store `latest` readings (`map[readingKey]*model.LatestReading`), `alerts` (`[]*model.AlertEvent` or map+order, bounded e.g. last 200 alerts), `clients` (`map[*Client]struct{}`), `staleAfter`, and `missionEpoch`.
     - Implement `HydrateFromDB(ctx context.Context, q *db.Queries) error` calling `q.ListLatestReadingsByMetric(ctx, pgtype.UUID{})` to seed `latest` readings on server startup (marking any reading whose age exceeds `staleAfter` as `IsStale: true`, `Status: model.StatusStale` in memory only).
     - Implement `PublishFrames(frames []model.TelemetryFrame)`: compute `LastSeenUTC = now`, `IsStale = false`, populate `MET` via `FormatMET`, update `latest`, broadcast `model.RealtimeEnvelope{Type: "reading", Reading: &copy, SentAtUTC: now}`, and for any `frame.Status == model.StatusCritical`, create a `model.AlertEvent` with unique ID, append/store in `alerts`, and broadcast `model.RealtimeEnvelope{Type: "alert", Alert: &alertCopy, SentAtUTC: now}`.
     - Implement `RunStaleMonitor(ctx context.Context, checkInterval time.Duration)`: periodically check all `latest` entries; when `!r.IsStale && now.Sub(r.LastSeenUTC) > h.staleAfter`, set `r.IsStale = true`, `r.Status = model.StatusStale`, and broadcast `model.RealtimeEnvelope{Type: "stale", Reading: &copy, SentAtUTC: now}`.
     - Implement `AckAlert(id string, acknowledgedBy string) (model.AlertEvent, error)` (return sentinel not-found error if ID does not exist; default `acknowledgedBy` to `"crew"` if empty; mark acknowledged and broadcast `model.RealtimeEnvelope{Type: "alert_ack", Alert: &alertCopy, SentAtUTC: now}`).
     - Implement `Snapshot() model.Snapshot`, `ClientCount() int`, and `ActiveAlertsCount() int`.
   - `client.go`:
     - Implement `ServeWS(hub *Hub, allowedOrigins []string) http.HandlerFunc` using `gorilla/websocket.Upgrader`.
     - Enqueue initial `model.RealtimeEnvelope{Type: "snapshot", Readings: snap.Readings, Alerts: snap.Alerts, SentAtUTC: now}` immediately upon registration.
     - Run single-writer `writePump` (`SetWriteDeadline`, `WriteJSON`, 30s `PingMessage` keepalive) and `readPump` (`SetReadLimit(64 << 10)`, `SetReadDeadline`, `SetPongHandler`).
     - When `readPump` receives a JSON `model.RealtimeEnvelope` with `Type == "alert_ack"` and `Alert != nil && Alert.ID != ""`, call `hub.AckAlert(env.Alert.ID, ackBy)` so WS bidirectional `alert_ack` works cleanly alongside REST.
     - On shutdown (`Close()` or canceled server context), cleanly close registered client channels/connections.
4. **Emit Post-Commit Frames from `server/internal/ingestion/`:**
   - Define `FramePublisher` interface (`PublishFrames(frames []model.TelemetryFrame)`) in `server/internal/ingestion/types.go` and add `SetPublisher(pub FramePublisher)` on `ingestion.Service`.
   - Collect `model.TelemetryFrame` entries for each value inserted inside `Service.Ingest`, and **only after** `tx.Commit(ctx)` succeeds, if `s.publisher != nil`, call `s.publisher.PublishFrames(frames)`.
5. **Mount Realtime & Health Endpoints in `server/internal/http/router.go` and Wire `main.go`:**
   - Keep `NewRouter(cfg, ingester)` calling `NewRouterWithDeps(cfg, RouterDeps{Ingester: ingester})`.
   - Define `RealtimeController` interface in `server/internal/http` (covering WS handler / ServeWS, `Snapshot() model.Snapshot`, `AckAlert(id, by string) (model.AlertEvent, error)`, `ClientCount() int`, `ActiveAlertsCount() int`) and optional `DBPinger` (`Ping(ctx) error`) for health reporting.
   - Mount:
     - `GET /ws/realtime` → Gorilla WebSocket endpoint (returns `503 Service Unavailable` or `500` JSON error if realtime controller is `nil`).
     - `GET /api/snapshot` → `200 OK` JSON `model.Snapshot`.
     - `POST /api/alerts/{id}/ack` → strict JSON decode (allowing optional empty body or `model.AckAlertRequest`), calls `AckAlert`, returns `200 OK` JSON `model.AlertEvent` or `404` / `400`.
     - `GET /health` → `200 OK` JSON `model.HealthStatus` (`Status: "ok"`, `Store: "postgres"` if ping succeeds else `"degraded"`, `TimeUTC`, `ActiveAlerts`, `RealtimeClients`, `IngestClients: 0` or tracked active ingest requests).
   - In `server/cmd/server/main.go`:
     - Construct `hub := realtime.NewHub(cfg.StaleAfter, cfg.MissionEpochUTC)`.
     - Call `hub.HydrateFromDB(pingCtx, db.New(pool))` on startup.
     - Attach `ingester.SetPublisher(hub)` and start `go hub.RunStaleMonitor(ctx, time.Second)`.
     - Pass `RouterDeps{Ingester: ingester, Realtime: hub, DBPinger: pool}` to `serverhttp.NewRouterWithDeps`.
     - Close `hub` cleanly during shutdown.
6. **Write Tests & Run Full Verification:**
   - Unit tests in `server/internal/realtime/` (`hub_test.go`, `ws_test.go`):
     - `PublishFrames` triggers `"reading"` envelope + `"alert"` envelope on `StatusCritical` and increments `ActiveAlertsCount`.
     - `RunStaleMonitor` / stale sweep flips old reading to `IsStale: true` / `StatusStale` and broadcasts `"stale"` envelope.
     - `AckAlert` marks alert acknowledged and broadcasts `"alert_ack"`; unknown ID returns not-found error.
     - Real `gorilla/websocket` client over `httptest.Server` connects to `/ws/realtime`: receives initial `"snapshot"`, receives live `"reading"` + `"alert"` broadcast, sends inbound `"alert_ack"` frame and receives `"alert_ack"` confirmation, and slow-client test proves full buffer does not block `PublishFrames`.
   - HTTP tests in `server/internal/http/router_test.go`:
     - `GET /health`, `GET /api/snapshot`, `POST /api/alerts/{id}/ack` (`200` and `404`), and `GET /ws/realtime` end-to-end WebSocket handshake.
   - PostgreSQL integration test in `server/internal/ingestion/service_integration_test.go`:
     - Verify `Service.Ingest` publishes frames on `201 Created` commit AND publishes **0 frames** on rolled-back errors (e.g., unknown `metric_code`, constraint error).
   - Run `wails3 task server:check`, `GOWORK=off go -C server vet ./...`, PostgreSQL `hrig_test` integration tests, and a live `:80` end-to-end smoke test connecting a Go Gorilla WebSocket test/client (or `httptest`/scripted WS consumer) + `hrig-server.exe` + `hrig-bot.exe` proving live bot batches stream out over `/ws/realtime` as `"reading"`/`"alert"` frames!

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/009-realtime-gorilla-ws-server/changelog.md` listing:
- Files created and modified
- Unit, Gorilla WebSocket, PostgreSQL integration (`hrig_test`), and live `:80` + `bot` + `/ws/realtime` smoke test outputs
- Any blockers or assumptions avoided
