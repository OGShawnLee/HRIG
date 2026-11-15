# 007 — Changelog (`GUIDE.md` Sensor Ingestion & `bot/` Integration)

## Files Created & Modified

### Created
- `bot/go.mod` — Initialized `module hrig/bot` (`go 1.26.0`) without modifying `bot/main.go`.
- `server/internal/ingestion/types.go` — Defined `IngestRequest`, `ReadingInput`, `IngestResult`, and HTTP-mapped `Error` helpers (`StatusCodeFromError`).
- `server/internal/ingestion/classify.go` — Implemented `ClassifyReading` evaluating inclusive critical bounds before exclusive warning bounds, and returning `normal` for `manual` mode.
- `server/internal/ingestion/service.go` — Implemented transactional `Service.Ingest(ctx, req)` using `pgxpool.Pool` and `server/internal/db` (`pgx/v5`), resolving active device/astronaut/metric/unit catalogs, validating `environment`, `recorded_at_utc`, `metric_type`, `value_type`, `numeric(12,4)` bounds, `scale` bounds, `unit_code` dimension matching, duplicate `metric_code` (`409`), threshold `status`, and PostgreSQL `pgconn.PgError` mapping.
- `server/internal/ingestion/classify_test.go` — Unit tests for `ClassifyReading` (`heart_rate` boundaries `40/41/50/60/100/101/199/200`, `manual` mode, one-sided & `mood` scale bounds).
- `server/internal/ingestion/service_integration_test.go` — PostgreSQL integration tests gated on `HRIG_TEST_DATABASE_URL`, loading `db/001_schema.sql`, `db/002_seed.sql`, `db/003_demo_seed.sql` into `hrig_test`, covering all 7 `bot/main.go` sensor shapes, abnormal/critical status persistence, and atomic rollback across 13 failure/conflict cases.
- `.tickets/007-guide-sensor-ingestion/changelog.md` — This changelog.

### Modified
- `go.work` (and `go.work.sum` synced by Go toolchain) — Added `./bot` so `go vet ./...` and `go build -o ../bin/hrig-bot.exe .` inside `bot/` resolve cleanly inside the Go workspace.
- `Taskfile.yml` — Added `bot:dev`, `db:migrate`, and `db:seed` while preserving all existing `client:*`, `server:*`, and `db:*` tasks.
- `server/internal/config/config.go` — Changed default `HRIG_API_ADDR` from `:8080` to `:80` and added `Config.DatabaseURL` reading `DATABASE_URL` with fallback `HRIG_DATABASE_URL`.
- `server/.env.example` — Documented `HRIG_API_ADDR=:80`, `HRIG_ALLOWED_ORIGINS`, `DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable`, and `HRIG_DATABASE_URL` fallback.
- `server/internal/http/router.go` — Updated `NewRouter(cfg config.Config, ingester Ingester)` preserving `GET /` (`{"message":"hello world"}`) and adding `POST /` with `application/json` `Content-Type` check, `1 MiB` `MaxBytesReader` limit, strict JSON decoding (rejecting malformed JSON and trailing tokens), `201 Created`, and sanitized `400`/`404`/`409`/`500` JSON responses.
- `server/internal/http/router_test.go` — Added tests for config defaults/`DATABASE_URL` fallback, `GET /`, and `POST /` validation & status mapping via mock `Ingester`.
- `server/cmd/server/main.go` — Enforced `cfg.DatabaseURL` presence, initialized `pgxpool.New` + `pool.Ping` on startup, wired `ingestion.NewService(pool)` to `serverhttp.NewRouter`, and closed pool cleanly on shutdown/failure.

### Workspace Hygiene Note
- `wails3 task client:check` regenerated `app/frontend/bindings/hrig/app/app.ts`; restored via `git checkout -- app/frontend/bindings/` so `app/**` remains untouched.

---

## Verification Outputs

### 1. `bot/` Vet & Build
Command (in `bot/`):
```powershell
go vet ./...
go build -o ../bin/hrig-bot.exe .
```
Result: exited `0` (no warnings or errors).

### 2. `db:migrate` & `db:seed` on Local `hrig` DB
Commands:
```powershell
wails3 task db:start
wails3 task db:migrate
wails3 task db:seed
```
Key output:
- `db:migrate`: `CREATE TYPE` / `CREATE TABLE` / `CREATE INDEX` / `CREATE FUNCTION` / `CREATE TRIGGER` / `CREATE VIEW` / `COMMIT` (exit `0`).
- `db:seed` (`002_seed.sql`): `INSERT 0 15`, `INSERT 0 23`, `INSERT 0 6`, `INSERT 0 22`, `INSERT 0 1`, `COMMIT` (exit `0`).
- `db:seed` (`003_demo_seed.sql`): `INSERT 0 3`, `INSERT 0 1`, `INSERT 0 15`, `INSERT 0 117`, `COMMIT` (exit `0`).

### 3. Unit & PostgreSQL Integration Tests (`hrig_test`)
Command (in `server/`, with disposable `hrig_test` created and dropped afterward):
```powershell
HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable go test -v ./...
```
Output summary:
```text
=== RUN   TestConfigDefaultsAndDatabaseURLFallback
--- PASS: TestConfigDefaultsAndDatabaseURLFallback (0.00s)
=== RUN   TestRootHelloWorld
--- PASS: TestRootHelloWorld (0.00s)
=== RUN   TestPostIngest_HTTPValidation
--- PASS: TestPostIngest_HTTPValidation (0.11s)
PASS
ok  	hrig/server/internal/http	6.022s
=== RUN   TestClassifyReading_HeartRateBoundaries
--- PASS: TestClassifyReading_HeartRateBoundaries (0.00s)
=== RUN   TestClassifyReading_ManualMode
--- PASS: TestClassifyReading_ManualMode (0.00s)
=== RUN   TestClassifyReading_OneSidedAndScaleBounds
--- PASS: TestClassifyReading_OneSidedAndScaleBounds (0.00s)
=== RUN   TestIntegration_IngestAllBotSensorShapesAndClassification
--- PASS: TestIntegration_IngestAllBotSensorShapesAndClassification (1.44s)
=== RUN   TestIntegration_AtomicRollbackAndErrorCodes
--- PASS: TestIntegration_AtomicRollbackAndErrorCodes (1.60s)
PASS
ok  	hrig/server/internal/ingestion	3.407s
```
`hrig_test` was dropped cleanly via `DROP DATABASE IF EXISTS hrig_test WITH (FORCE);`.

### 4. Live Server + Bot End-to-End Smoke Test (Port `:80`, Local `hrig` DB)
Ran `./bin/hrig-server.exe` with default `HRIG_API_ADDR` (`:80`) and `DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable`, plus `./bin/hrig-bot.exe` with default `SENSOR_ENDPOINT` (`http://localhost:80/`):
```text
BEFORE_COUNTS=15|117
GET_ROOT={"message":"hello world"}
BOT_STDERR:
2026/11/15 14:33:11 simulando 7 sensores lógicos para 3 astronautas; POST http://localhost:80/ cada 2s
2026/11/15 14:33:12 lote de lecturas enviado correctamente
SERVER_LOGS:
2026/11/15 14:33:08 hrig server listening on :80
AFTER_COUNTS=32|174
```
Confirmed +17 `measurement` rows (`15 -> 32`) and +57 `measurement_value` rows (`117 -> 174`) persisted in `hrig`, then stopped `hrig-bot.exe`, `hrig-server.exe`, and PostgreSQL (`wails3 task db:stop`).

### 5. `wails3 task server:check` & `wails3 task client:check`
- `wails3 task server:check`: `go vet ./...`, `go test ./...`, and `go build -o ../bin/hrig-server.exe ./cmd/server` exited `0` (also verified uncached `go test -count=1 ./...` in `server/` exited `0`).
- `wails3 task client:check`: `wails3 generate bindings -clean=true -ts`, `vue-tsc --noEmit`, `vite build --mode production`, and `go build .` exited `0`.

### 6. Process & Git State
- Verified zero running `hrig-server`, `hrig-bot`, or `postgres` processes remain.
- `git status --short` confirms no modifications to `db/**`, `sqlc.yaml`, `server/internal/db/**`, `bot/main.go`, `app/**`, `.routing/**`, `docs/**`, `README.md`, `GUIDE.md`, or `.claude/**`.
- No git commits or pushes were executed.
