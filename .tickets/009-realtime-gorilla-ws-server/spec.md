# 009 — Verification Spec

Every check is binary (`PASS` / `FAIL`). All checks must pass before ticket `009` is accepted.

## 1. Dependencies & Config

- [x] `server/go.mod` includes `github.com/gorilla/websocket` and `GOWORK=off go -C server vet ./...` exits with `0`.
- [x] `server/internal/config/config.go` loads `HRIG_STALE_AFTER` (default `15s`) into `Config.StaleAfter` and `HRIG_MISSION_EPOCH_UTC` into `Config.MissionEpochUTC`.
- [x] `server/.env.example` documents `HRIG_STALE_AFTER=15s` and `HRIG_MISSION_EPOCH_UTC`.

## 2. Gorilla WebSocket Hub & Stale/Alert Logic (`server/internal/realtime/`)

- [x] `server/internal/realtime/` uses `github.com/gorilla/websocket` (`websocket.Upgrader`, single-writer `writePump` + `readPump`, non-blocking broadcast to buffered `send` channels, origin validation via `cfg.AllowedOrigins`).
- [x] `PublishFrames` broadcasts `"reading"` `model.RealtimeEnvelope` messages and automatically creates/broadcasts in-memory `model.AlertEvent` (`"alert"` envelope) when `Status == model.StatusCritical` without writing alerts to PostgreSQL.
- [x] Stale monitor marks quiet readings `IsStale = true` and `Status = model.StatusStale` in memory (`"stale"` envelope) after `StaleAfter` without writing `"stale"` to PostgreSQL.
- [x] `AckAlert` marks alerts acknowledged and broadcasts `"alert_ack"` `model.RealtimeEnvelope`; inbound `"alert_ack"` frames on `/ws/realtime` also trigger acknowledgement cleanly.
- [x] New WebSocket connections receive an initial `"snapshot"` `model.RealtimeEnvelope` containing current readings and alerts.

## 3. Post-Commit Ingestion Hook & Endpoints (`server/internal/ingestion/`, `server/internal/http/`, `server/cmd/server/`)

- [x] `Service.Ingest` collects committed `model.TelemetryFrame` values and calls `publisher.PublishFrames` **only after** `tx.Commit(ctx)` succeeds (verified by integration test confirming 0 published frames on rollback).
- [x] `GET /` still returns `200 OK` `{"message":"hello world"}`, and `POST /` preserves `GUIDE.md` validation and `201`/`400`/`404`/`409`/`500` responses.
- [x] Router mounts `GET /ws/realtime`, `GET /api/snapshot`, `POST /api/alerts/{id}/ack`, and `GET /health` using `hrig/shared/model` DTOs (`RealtimeEnvelope`, `Snapshot`, `AckAlertRequest`, `AlertEvent`, `HealthStatus`).
- [x] `server/cmd/server/main.go` hydrates the hub via `ListLatestReadingsByMetric`, starts `RunStaleMonitor`, wires `ingester.SetPublisher(hub)`, and shuts down cleanly.

## 4. Build, Unit, Integration & Live End-to-End Smoke Verification

- [x] `wails3 task server:check` exits with `0` (all unit tests in `realtime`, `http`, `ingestion` pass + binary builds).
- [x] `go -C server test -race ./internal/realtime/... ./internal/http/...` (or if CGO disabled on Windows, verify `-race` availability note cleanly and run `-count=1 -v ./...`) passes with zero data races or failures.
- [x] With PostgreSQL running and disposable `hrig_test` created, `HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable go -C server test -v -count=1 ./...` exits with `0` and `hrig_test` is dropped afterward.
- [x] Live `:80` end-to-end smoke test (`bin/hrig-server.exe` + `bin/hrig-bot.exe` + live Gorilla WS consumer on `ws://127.0.0.1:80/ws/realtime`) verifies:
  - WS receives initial `"snapshot"` frame,
  - After bot sends batch (`lote de lecturas enviado correctamente`), WS receives live `"reading"` frames matching bot metrics,
  - `GET http://127.0.0.1:80/health` reports `"status":"ok"` and `realtimeClients >= 1`,
  - `GET http://127.0.0.1:80/api/snapshot` returns populated readings.
- [x] No `hrig-server.exe`, `hrig-bot.exe`, or `postgres.exe` process remains running after verification.

## 5. Scope & Ground Rules

- [x] `git diff --name-only` and untracked files show changes only inside `May create or edit` paths listed in `task.md`.
- [x] `shared/**`, `bot/**`, `app/**`, `db/**`, `server/internal/db/**` remain completely unmodified.
- [x] Code, `.tickets/`, and `changelog.md` are written in English; `.tickets/009-realtime-gorilla-ws-server/changelog.md` records all changed files and exact outputs.
