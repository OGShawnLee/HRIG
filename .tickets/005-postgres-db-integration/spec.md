# 005 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Dependencies, Config & Taskfile
- [ ] `server/go.mod` includes `github.com/jackc/pgx/v5` and no ORM (`gorm`) dependencies.
- [ ] `server/.env.example` defines active `DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable` and documents `HRIG_TEST_DATABASE_URL`.
- [ ] `Taskfile.yml` defines `db:migrate` (`db/001_schema.sql`) and `db:seed` (`db/002_seed.sql` + `db/003_demo_seed.sql`) while keeping `client:*`, `db:start/stop/status/psql`, `shared:*`, and `server:*` tasks intact.

## Schema Load & Automated Test Suite (`TEMP-DB-TESTING.md` + Enhancements)
- [ ] `wails3 task db:migrate` and `wails3 task db:seed` exit `0` against `hrig`, and re-running `wails3 task db:seed` exits `0` without duplicating rows.
- [ ] `HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable go test -v ./internal/store/...` inside `server/` exits `0` with zero skipped integration tests (`astronaut=3`, `metric=23`, `metric_threshold=23`, `measurement=15`, `measurement_value=117`, all triggers/constraints including separate nonexistent UUID FK checks on `astronaut_id`/`device_id`/`metric_id`/`unit_id`, `media_asset_mime`, `v_astronaut_age`, `trg_astronaut_updated_at`, `ON DELETE RESTRICT`, `GetThresholdByMetricCode("heart_rate")` + boundary thresholds, and `PostgresStore` round-trips all passing).
- [ ] Disposable `hrig_test` database was dropped cleanly after verification (`psql -l` does not list `hrig_test`).

## Go API Integration & Offline Safety
- [ ] `wails3 task server:check` exits `0` without `HRIG_TEST_DATABASE_URL` or `DATABASE_URL` set (tests skip/fallback cleanly).
- [ ] `wails3 task client:check` exits `0`.
- [ ] With PostgreSQL 17 and `DATABASE_URL` active, `GET /health` returns `200` with `"store":"postgres"`.
- [ ] `GET /api/astronauts` returns 3 astronauts with derived `age`, and `GET /api/metrics` returns 23 metrics with units and thresholds.
- [ ] Sending a telemetry frame over `WS /ws/ingest` persists a row in PostgreSQL and is returned by `GET /api/trends` and `GET /api/history`.
- [ ] `POST /api/checkins` persists `measurement` + `measurement_value` (+ `media_asset` when included) in PostgreSQL and returns `201`.

## Scope & Ground Rules
- [ ] `git status --short` shows no changes in `app/**`, `shared/**`, `db/**`, `.routing/**`, `docs/**`, `README.md`, `PROJECT-REVIEW.md`, or `.claude/**`.
- [ ] `wails3 task db:stop` leaves PostgreSQL cleanly stopped if started only for verification (or status noted in `changelog.md`).
- [ ] No git commits or pushes were executed.
- [ ] `.tickets/005-postgres-db-integration/changelog.md` exists and records files touched, PG version, and `PASS`/`FAIL` per `TEMP-DB-TESTING.md` check.
