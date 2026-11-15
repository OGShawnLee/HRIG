# 010 — Verification Spec

Every check is binary (`PASS` / `FAIL`). All checks must pass before ticket `010` is accepted.

## 1. Shared DTOs (`shared/model/`)

- [x] `shared/model/telemetry.go` defines `BaseUnitDimension` and `BaseUnitDecimals` on `Metric`, `Offset` on `HistoryQuery`, and `TrendPoint` and `HistoryEntry` structs with camelCase JSON tags matching `Astronaut` / `Metric` / `TelemetryFrame`.
- [x] `GOWORK=off go -C shared test -v ./...` exits with `0` and verifies JSON marshaling/unmarshaling for `TrendPoint`, `HistoryEntry`, and `Metric`.

## 2. Query Service (`server/internal/query/`)

- [x] `server/internal/query/service.go` compiles cleanly and maps `ListActiveAstronauts`, `ListActiveMetrics`, `ListTrendsByMetric`, and `ListHistory` without altering `db/queries/**` or `server/internal/db/**`.
- [x] `ListTrends` rejects empty `metric_code`, malformed `astronaut_id` UUIDs, negative limits, or limits above `1000` with `ErrBadRequest`, and computes `MET` using `realtime.FormatMET`.
- [x] `ListHistory` rejects malformed `astronaut_id` UUIDs, negative `offset`, negative limits, or limits above `500` with `ErrBadRequest`, and maps optional `media_asset` columns when present.

## 3. HTTP Router & Server Wiring (`server/internal/http/`, `server/cmd/server/`)

- [x] `server/internal/http/router.go` preserves `NewRouter(cfg, ingester)`, `GET /`, `POST /`, `GET /health`, `GET /ws/realtime`, `GET /api/snapshot`, `POST /api/alerts/{id}/ack`, and mounts:
  - `GET /api/astronauts`
  - `GET /api/metrics`
  - `GET /api/trends`
  - `GET /api/history`
- [x] `GET /api/trends` validates `metric_code` (required), `astronaut_id`, `since` (RFC3339/RFC3339Nano), and `limit`, returning `400` JSON `{"error":"..."}` on invalid inputs and `503` when `Reader == nil`.
- [x] `GET /api/history` validates `astronaut_id`, `limit`, and `offset`, returning `400` JSON `{"error":"..."}` on invalid inputs and `503` when `Reader == nil`.
- [x] `server/cmd/server/main.go` instantiates `query.NewService(pool, cfg.MissionEpochUTC)` and passes `Reader` in `RouterDeps`.

## 4. Build, Unit & PostgreSQL Integration Verification

- [x] `GOWORK=off go -C server vet ./...` exits with `0`.
- [x] `wails3 task server:check` exits with `0`.
- [x] With PostgreSQL running and disposable `hrig_test` created, `HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable go -C server test -v -count=1 ./...` exits with `0` (including `server/internal/query` integration tests verifying 3 seeded astronauts, 23 seeded metrics, trends, and history) and `hrig_test` is dropped afterward.
- [x] No `hrig-server.exe`, `hrig-bot.exe`, or `postgres.exe` process remains running after verification.

## 5. Scope & Ground Rules

- [x] `git diff --name-only` and untracked files show changes only inside `May create or edit` paths listed in `task.md`.
- [x] `bot/**`, `app/**`, `db/**`, `server/internal/db/**`, `server/internal/ingestion/**`, and `server/internal/realtime/**` remain completely unmodified.
- [x] Code, `.tickets/`, and `changelog.md` are written in English; `.tickets/010-server-read-endpoints/changelog.md` records all changed files and exact verification outputs.
