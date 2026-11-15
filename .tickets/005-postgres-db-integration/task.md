# 005 — Wire PostgreSQL DB schema, seeds & `PostgresStore` persistence

| Field | Value |
| :--- | :--- |
| **ID** | `005` |
| **Status** | `draft` *(ready to dispatch after `002` and `004` complete)* |
| **Worker** | `antigravity` / `gemini-4-hydrogen` (`thinking_level: medium`) |
| **Role** | Backend & PostgreSQL Integration Engineer |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `server/go.mod`, `server/go.sum` (add `github.com/jackc/pgx/v5`)
  - `server/.env.example` (set active local `DATABASE_URL` default and document `HRIG_TEST_DATABASE_URL`)
  - `server/internal/config/config.go` (read `DATABASE_URL`)
  - `server/internal/store/postgres.go` (`PostgresStore` implementing `TelemetryStore`)
  - `server/internal/store/postgres_test.go` (covering all checks from `.routing/TEMP-DB-TESTING.md` plus `media_asset`, `v_astronaut_age`, `trg_astronaut_updated_at`, `ON DELETE RESTRICT`, boundary thresholds, and `PostgresStore` round-trips)
  - `server/cmd/server/main.go` (select `PostgresStore` via `pgxpool` when `DATABASE_URL` is set, failing fast if `Ping` fails; keep `MemoryStore` fallback when `DATABASE_URL` is empty)
  - `Taskfile.yml` (add `db:migrate` and `db:seed` tasks pointing to `db/001_schema.sql`, `db/002_seed.sql`, `db/003_demo_seed.sql`; keep existing `client:*`, `db:start/stop/status/psql`, `shared:*`, and `server:*` tasks intact)
  - `.tickets/005-postgres-db-integration/changelog.md`
- **Must not touch:**
  - `app/**`, `shared/**`, `db/**`, `local/**`
  - `.gitignore`, `go.work`
  - `.routing/**`, `.claude/**`
  - `docs/**`, `README.md`, `PROJECT-REVIEW.md`
  - `.tickets/001-scaffold-wails3-app/**`, `.tickets/002-setup-portable-postgres-17/**`, `.tickets/003-full-stack-install-guide/**`, `.tickets/004-scaffold-go-api/**`
  - `.tickets/005-postgres-db-integration/task.md`, `.tickets/005-postgres-db-integration/spec.md`

## Inputs

Read these files before acting:
1. [`db/001_schema.sql`](../../db/001_schema.sql)
2. [`db/002_seed.sql`](../../db/002_seed.sql)
3. [`db/003_demo_seed.sql`](../../db/003_demo_seed.sql)
4. [`db/README.md`](../../db/README.md)
5. [`.routing/BD.md`](../../.routing/BD.md)
6. [`.routing/BDML.md`](../../.routing/BDML.md)
7. [`.routing/TEMP-DB-TESTING.md`](../../.routing/TEMP-DB-TESTING.md)
8. [`shared/model/telemetry.go`](../../shared/model/telemetry.go)
9. [`server/internal/store/store.go`](../../server/internal/store/store.go)
10. [`server/internal/store/writer.go`](../../server/internal/store/writer.go)
11. [`server/internal/threshold/evaluator.go`](../../server/internal/threshold/evaluator.go)
12. [`Taskfile.yml`](../../Taskfile.yml)

## Grounding & Anti-Hallucination Constraints

- **1:1 Schema Fidelity (`db/001_schema.sql` & `.routing/BD.md`):**
  - Never alter `db/*.sql` or invent new columns/tables.
  - Resolve `metric.id`, `unit.id`, and `device.id` by `code` (never hardcode catalog UUIDs).
  - Read astronauts from `astronaut` joined with `v_astronaut_age` (`is_active = true`).
  - Read metrics from `metric` joined with `unit` and `metric_threshold` (`is_active = true`, ordered by `metric.display_order ASC`).
  - **Trigger Compliance:**
    - `trg_measurement_metric_type`: when `metric.metric_type = 2` (environmental), always insert `measurement.astronaut_id = NULL`.
    - `trg_measurement_value_dimension`: require incoming readings to match `metric.base_unit_id` (if `UnitCode` is omitted on a numeric/scale payload, default to the metric's base unit; if provided and its `unit.dimension` differs, surface a validation error cleanly rather than building a unit converter).
    - `trg_measurement_value_timestamp`: let the trigger populate `measurement_value.recorded_at_utc` from `measurement.recorded_at_utc` (still pass `recorded_at_utc` on `measurement` or omit/mirror safely).
  - **Threshold Evaluation (`.routing/BD.md` §5):** Evaluate `critical` limits first (`<= critical_min` or `>= critical_max` -> `critical`), then `warning` limits (`< warning_min` or `> warning_max` -> `abnormal`), else `normal` (`manual` mode for `ultrasound`). Persist `status` in `measurement_value.status`.
  - **Alerts Out of DB (`.routing/BD.md` D7):** Never store alerts in PostgreSQL; active alerts and `/api/alerts/{id}/ack` remain in-memory on `ws.Hub`.
- **Store Selection Policy:**
  - When `DATABASE_URL` is non-empty, `server/cmd/server/main.go` initializes `PostgresStore` (`pgxpool.New`) and calls `Ping(ctx)` at startup — failing fast with non-zero exit if PostgreSQL is unreachable.
  - When `DATABASE_URL` is empty, keep `MemoryStore` as the in-memory fallback (`RNF09` offline/standalone resilience).
- **Test Database Safety (`.routing/TEMP-DB-TESTING.md`):**
  - `db/001_schema.sql` is one-shot DDL (not an idempotent migration). Never drop or overwrite an existing `hrig` database automatically.
  - `postgres_test.go` must gate on `HRIG_TEST_DATABASE_URL` (skipping cleanly with `t.Skip` when unset so offline `wails3 task server:check` passes without a running DB) and run destructive/trigger tests inside rolled-back transactions (`tx.Rollback`).
- **Do NOT commit or push** any changes to Git.

## Task

1. Add `github.com/jackc/pgx/v5` to `server/go.mod` (`go get github.com/jackc/pgx/v5`).
2. Update `server/internal/config/config.go` (`DatabaseURL string` from `DATABASE_URL`) and `server/.env.example` (set `DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable` and document `# HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable`).
3. Update `Taskfile.yml` to add Windows-safe `db:migrate` and `db:seed` tasks (matching `db:start` / `db:psql` conventions; note `db:migrate` runs once on a fresh DB):
   - `db:migrate`: runs `.\local\pgsql\bin\psql.exe -h 127.0.0.1 -p 5432 -U postgres -d hrig -v ON_ERROR_STOP=1 -f db/001_schema.sql` (`platforms: [windows]`).
   - `db:seed`: runs `.\local\pgsql\bin\psql.exe -h 127.0.0.1 -p 5432 -U postgres -d hrig -v ON_ERROR_STOP=1 -f db/002_seed.sql` followed by `-f db/003_demo_seed.sql` (`platforms: [windows]`).
4. Implement `PostgresStore` in `server/internal/store/postgres.go` using `pgxpool.Pool` (`BackendName() == "postgres"`):
   - `Ping(ctx)`: executes `pool.Ping(ctx)`.
   - `ListAstronauts(ctx)`: selects `id`, `full_name`, `birth_date`, `age` (from `v_astronaut_age`), `nationality_code`, `nationality_label`, `is_active` where `is_active = true`.
   - `ListMetrics(ctx)`: selects active `metric` rows joined with `unit` and `metric_threshold` ordered by `metric.display_order ASC`.
   - `GetThresholdByMetricCode(ctx, metricCode)`: fetches `metric_threshold` joined by `metric.code`.
   - `SaveTelemetry(ctx, frame)`: in a transaction, resolves `metric` (including `metric_type`) and `unit`/`device`, inserts `measurement` (`astronaut_id = NULL` when `metric_type = 2`) and `measurement_value` (`value_numeric`, `value_scale`, `value_text`, `status`, `note`).
   - `LatestReadings(ctx)`: returns latest reading per `(metric_id, astronaut_id)` using `idx_measurement_value_metric_time` joined to `measurement`, `metric`, `unit`, and `device`.
   - `QueryTrends(ctx, q)` and `ListHistory(ctx, q)`: query `measurement_value` joined to `measurement`, `metric`, `unit`, and `device`, filtered by optional `astronautId` / `metricCode` and ordered by `recorded_at_utc DESC` with `LIMIT`.
   - `SaveCheckIn(ctx, req, status)`: inserts `measurement` (`device_id` resolved if provided, else `NULL`) + `measurement_value` + optional `media_asset` rows (`kind`, `file_path`, `mime_type`, `findings`, `captured_at_utc`) inside a single transaction.
5. Update `server/cmd/server/main.go` to construct `PostgresStore` when `cfg.DatabaseURL != ""` (calling `Ping` and exiting non-zero on error) or `MemoryStore` when `cfg.DatabaseURL == ""`.
6. Create `server/internal/store/postgres_test.go` (gated on `HRIG_TEST_DATABASE_URL`) testing:
   - Catalog & demo seed counts (`3` astronauts, `23` metrics, `23` `metric_threshold` rows: `22` `numeric` + `1` `manual` for `ultrasound`, `15` demo measurements, `117` demo values) and zero inverted/empty bounds.
   - All constraint & trigger failures from `.routing/TEMP-DB-TESTING.md` inside rolled-back transactions: `measurement_not_orphan`, `trg_measurement_metric_type`, `trg_measurement_value_dimension`, `(measurement_id, metric_id)` uniqueness, `measurement_value_numeric_needs_unit`, `measurement_value_has_value`, explicit non-existent UUID FK rejections separately on `astronaut_id`, `device_id`, `metric_id`, and `unit_id`, `trg_measurement_value_timestamp`, and `trg_measurement_timestamp_cascade`.
   - Additional schema checks: `media_asset_mime` constraint (reject `'video/mp4'`; accept `'image/jpeg'`, `'image/png'`, `'application/pdf'`), `v_astronaut_age` derivation, `trg_astronaut_updated_at`, and `ON DELETE RESTRICT` blocking deletion of a referenced `measurement`.
   - `PostgresStore` round-trip tests: `ListAstronauts`, `ListMetrics`, `GetThresholdByMetricCode("heart_rate")` (`unit_symbol="/min"`, `warning_min=50`, `warning_max=100`, `critical_min=40`, `critical_max=200`, `classification_mode="numeric"`), `LatestReadings`, `SaveTelemetry` + boundary classification (`heart_rate`: `40`, `41`, `50`, `60`, `100`, `101`, `110`, `199`, `200`), `SaveCheckIn` (with `media_asset`), `QueryTrends`, and `ListHistory`.
7. Execute end-to-end verification against local PostgreSQL 17:
   - Start PG via `wails3 task db:start`.
   - Run `wails3 task db:migrate` and `wails3 task db:seed` against `hrig` (and verify re-running `wails3 task db:seed` succeeds idempotently).
   - Create disposable `hrig_test`, load `001_schema.sql` + `002_seed.sql` + `003_demo_seed.sql` (plus re-run `002_seed.sql` + `003_demo_seed.sql`), run `HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable go test -v ./internal/store/...` inside `server/`, and drop `hrig_test` via `dropdb.exe` afterward.
   - Start `server` with `DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable`, verify `GET /health` returns `"store":"postgres"`, `GET /api/astronauts` returns 3 astronauts, `GET /api/snapshot` hydrates from demo data, and `WS /ws/ingest` persists a reading queryable via `GET /api/trends`; then stop `server` and run `wails3 task db:stop`.
   - Run `wails3 task server:check` (with `HRIG_TEST_DATABASE_URL` unset) and `wails3 task client:check`, confirming both exit `0`.
8. Write `.tickets/005-postgres-db-integration/changelog.md` documenting all files touched, exact test outputs (`PASS`/`FAIL` per case from `TEMP-DB-TESTING.md`), and any discrepancies found.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/005-postgres-db-integration/changelog.md` listing:
- Files created and modified
- PostgreSQL version and exact SQL scripts applied
- Verification commands and `TEMP-DB-TESTING.md` case results
- Any blockers or assumptions avoided
