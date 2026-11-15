# 007 — Implement `GUIDE.md` sensor ingestion happy flow & `bot/` integration

| Field | Value |
| :--- | :--- |
| **ID** | `007` |
| **Status** | `completed` |
| **Worker** | `antigravity` / `gemini-4-hydrogen` (`thinking_level: medium`) |
| **Role** | Go Backend & PostgreSQL Ingestion Engineer |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `server/internal/config/config.go`, `server/.env.example` (default `HRIG_API_ADDR=:80`; read `DATABASE_URL` with fallback `HRIG_DATABASE_URL`)
  - `server/cmd/server/main.go` (initialize `pgxpool.Pool`, ping on startup, wire ingestion service + router, close pool on shutdown)
  - `server/internal/ingestion/*.go` (DTOs, `ClassifyReading`, transactional ingestion service + unit & `HRIG_TEST_DATABASE_URL` integration tests)
  - `server/internal/http/router.go`, `server/internal/http/router_test.go` (preserve `GET /`, add `POST /` ingestion handler)
  - `bot/go.mod` (initialize `module hrig/bot`, `go 1.26.0` — `bot/main.go` must remain unmodified unless a compile error is proven)
  - `go.work` (optional: add `./bot` if workspace check passes cleanly)
  - `Taskfile.yml` (add `bot:dev`, `db:migrate`, and `db:seed` while preserving all existing tasks)
  - `.tickets/007-guide-sensor-ingestion/changelog.md`
- **Must not touch:**
  - `db/001_schema.sql`, `db/002_seed.sql`, `db/003_demo_seed.sql`, `db/README.md`, `db/queries/**`
  - `sqlc.yaml`, `server/internal/db/*.go`
  - `bot/main.go`
  - `app/**`, `shared/**`, `local/**`
  - `.gitignore`, `.routing/**`, `.claude/**`, `docs/**`, `README.md`, `PROJECT-REVIEW.md`, `GUIDE.md`, `SENSOR-READINGS.md`
  - `.tickets/001-*`, `.tickets/002-*`, `.tickets/003-*`, `.tickets/004-*`, `.tickets/005-*`, `.tickets/006-*`
  - `.tickets/007-guide-sensor-ingestion/task.md`, `.tickets/007-guide-sensor-ingestion/spec.md`

## Inputs

Read these files before acting:
1. [`GUIDE.md`](../../GUIDE.md)
2. [`bot/main.go`](../../bot/main.go)
3. [`db/001_schema.sql`](../../db/001_schema.sql)
4. [`db/002_seed.sql`](../../db/002_seed.sql)
5. [`db/003_demo_seed.sql`](../../db/003_demo_seed.sql)
6. [`server/internal/db/querier.go`](../../server/internal/db/querier.go)
7. [`server/internal/db/models.go`](../../server/internal/db/models.go)
8. [`server/internal/db/telemetry.sql.go`](../../server/internal/db/telemetry.sql.go)
9. [`server/internal/config/config.go`](../../server/internal/config/config.go)
10. [`server/internal/http/router.go`](../../server/internal/http/router.go)
11. [`server/cmd/server/main.go`](../../server/cmd/server/main.go)
12. [`Taskfile.yml`](../../Taskfile.yml)

## Grounding & Anti-Hallucination Constraints

- **SQLC & Schema Fidelity:**
  - Use the generated `server/internal/db` package (`pgx/v5`) and `pgxpool`—do **not** write raw ad-hoc SQL queries outside `server/internal/db/` or modify `db/001_schema.sql`–`003_demo_seed.sql`.
  - Let PostgreSQL generate UUIDs (`measurement.id`, `measurement_value.id`).
- **Threshold Classification Rule (`db/001_schema.sql` + `db/002_seed.sql`):**
  - Evaluate critical bounds before warning bounds (`critical_min`/`critical_max` inclusive: `value <= critical_min || value >= critical_max` → `critical`; `warning_min`/`warning_max` exclusive: `value < warning_min || value > warning_max` → `abnormal`; otherwise `normal`).
  - For `classification_mode == "manual"` (e.g., `ultrasound`), always return `normal`.
- **HTTP & Validation Contract (`GUIDE.md` + `bot/main.go`):**
  - Default `HRIG_API_ADDR` to `:80` so `bot/main.go` (`http://localhost:80/`) works with zero env overrides.
  - Only return documented status codes for `POST /`: `201 Created`, `400 Bad Request`, `404 Not Found`, `409 Conflict`, `500 Internal Server Error` (never leak SQL text, DSNs, or credentials in HTTP responses).
  - Read `DATABASE_URL` (falling back to `HRIG_DATABASE_URL`) and fail fast on startup if missing or `pool.Ping` fails.
- **Do NOT commit or push** any changes to Git. **Do NOT leave background processes (`hrig-server`, `bot`, or `hrig_test` DB) running** at handback.

## Task

1. **Initialize `bot/go.mod` & Taskfile helpers:**
   - Write `bot/go.mod` with `module hrig/bot` and `go 1.26.0`. Verify `go vet ./...` and `go build -o ../bin/hrig-bot.exe .` inside `bot/` pass cleanly (if adding `./bot` to `go.work` keeps `go test` and `wails3 task server:check` clean, add it; otherwise leave `go.work` untouched and explain in `changelog.md`).
   - Add `bot:dev` (`dir: bot`, `cmds: ["go run ."]`) plus Windows `db:migrate` and `db:seed` tasks to `Taskfile.yml` executing `db/001_schema.sql`, `db/002_seed.sql`, and `db/003_demo_seed.sql` via `./local/pgsql/bin/psql.exe` with `-v ON_ERROR_STOP=1`.
2. **Update `server/internal/config/config.go` & `server/.env.example`:**
   - Default `HRIG_API_ADDR` to `:80`.
   - Read `DATABASE_URL` (falling back to `HRIG_DATABASE_URL`) into `Config.DatabaseURL`.
   - Document `HRIG_API_ADDR=:80`, `HRIG_ALLOWED_ORIGINS=...`, and illustrative `DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable` in `server/.env.example`.
3. **Create `server/internal/ingestion/` service & threshold classifier:**
   - Define `IngestRequest` (`device_code`, `environment`, `recorded_at_utc`, `astronaut_name *string`, `readings []ReadingInput`) and `IngestResult` (`measurement_id`, `recorded_at_utc`, `values_count`).
   - Implement `ClassifyReading(value float64, mode string, warnMin, warnMax, critMin, critMax *float64) db.ReadingStatus`.
   - Implement transactional `Service.Ingest(ctx, req)`:
     - Validate non-empty `device_code`, `environment` in `{"iva", "eva"}`, RFC3339/RFC3339Nano `recorded_at_utc` with timezone, non-empty `readings` slice, trimmed non-empty `astronaut_name` when present, and no duplicate `metric_code` within `readings` (`409` on duplicate).
     - Begin `pgx` transaction (`defer tx.Rollback(ctx)`) using `db.New(tx)` (or `q.WithTx(tx)`).
     - Lookup `device_code` via `GetActiveDeviceByCode` (`pgx.ErrNoRows` → `404`).
     - If `astronaut_name != nil`, lookup via `ListActiveAstronautsByFullName` (`0` rows → `404`; `>1` rows → `409`; `1` row → set `astronaut_id`).
     - Insert parent row via `CreateMeasurement`.
     - For each reading: lookup `metric_code` via `GetActiveMetricByCode` (`404` if missing), enforce `metric_type == 2` requires `astronaut_id == NULL` and `metric_type == 1` requires `astronaut_id != NULL` (`400`), enforce exact single value matching `value_type` (`numeric`: finite & fits `numeric(12,4)`; `scale`: within `[scale_min, scale_max]`; `text`/`image`: non-empty `value_text`), enforce `unit_code` via `GetActiveUnitByCode` (`404` if missing) + matching `dimension` for `numeric`/`scale` and absent `unit_code` for `text`/`image` (`400`), compute `status` via `ClassifyReading`, and insert via `CreateMeasurementValue`.
     - Map PostgreSQL `pgconn.PgError` `23505` → `409`, `23514`/`23503` → `400`, others → `500`.
     - Commit transaction; on any error roll back atomically.
4. **Wire HTTP `POST /` & startup DB pool:**
   - Update `server/internal/http/router.go` so `NewRouter(cfg config.Config, ingester Ingester)` preserves `GET /` (`{"message":"hello world"}`) and mounts `POST /`:
     - Enforce `Content-Type` (`application/json` media type → `400` otherwise), `1 MiB` body limit (`http.MaxBytesReader`), strict JSON decoding (reject malformed JSON or trailing tokens → `400`).
     - Return `201 Created` JSON on success and `400`/`404`/`409`/`500` JSON `{"error":"..."}` on failure.
   - Update `server/cmd/server/main.go` to require `cfg.DatabaseURL`, connect `pgxpool.New` + `pool.Ping`, defer `pool.Close()`, and pass `ingestion.NewService(pool)` to `NewRouter`.
5. **Write Unit & PostgreSQL Integration Tests:**
   - Unit tests for `ClassifyReading` boundary values (`heart_rate` `40/41/50/60/100/101/199/200`, `manual` mode), `POST /` request validation (`GET /` preserved, malformed JSON, wrong `Content-Type`, empty `readings`, `400`/`404`/`409`/`500` mapping via mock `Ingester`).
   - Integration tests in `server/internal/ingestion/` gated on `HRIG_TEST_DATABASE_URL`:
     - Load `db/001_schema.sql`, `db/002_seed.sql`, `db/003_demo_seed.sql` into disposable `hrig_test`.
     - Verify `201 Created` + persisted DB counts & `status` for all 7 `bot/main.go` device shapes (including `cabin_sensor_01` without `astronaut_name`, `suit_sensor_01` on `eva`, `mood` `{score}`, `ultrasound` `value_text` without `unit_code`, plus abnormal/critical threshold classification).
     - Verify atomic rollback (0 new `measurement` / `measurement_value` rows) on unknown `metric_code` (`404`), unknown `unit_code` (`404`), unit dimension mismatch (`400`), and `metric_type=2` with `astronaut_name` (`400`).
6. **End-to-End Smoke Verification (`server` + `bot` + PostgreSQL):**
   - Start PostgreSQL (`wails3 task db:start`), run `wails3 task db:migrate` & `wails3 task db:seed` against `hrig`.
   - Create disposable `hrig_test`, run `HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable go test -v ./...` inside `server/`, then drop `hrig_test`.
   - Run live end-to-end smoke test: start built `hrig-server.exe` on `:80` (or verify port `:80` availability; if port `:80` is occupied/restricted in test harness, document exact env override test and verify `:80` default in config/test) with `DATABASE_URL` pointing to `hrig`, run `bot` until at least 1 complete 17-request batch succeeds (`lote de lecturas enviado correctamente`), verify new `measurement` / `measurement_value` rows in `hrig`, then stop both `bot` and `hrig-server.exe` cleanly (and stop PostgreSQL via `wails3 task db:stop` if started solely for verification).
   - Run `wails3 task server:check` and `wails3 task client:check` (restore `app/frontend/bindings/` if dirtied) and record all outputs in `.tickets/007-guide-sensor-ingestion/changelog.md`.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/007-guide-sensor-ingestion/changelog.md` listing:
- Files created and modified
- Unit, integration (`hrig_test`), and live `bot` smoke test outputs
- Verification command outputs (`server:check`, `client:check`, `bot` build/vet)
- Any blockers or assumptions avoided
