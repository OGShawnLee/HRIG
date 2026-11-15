# 014E — Verification Spec

Every check is binary (`PASS` / `FAIL`). All checks must pass before ticket `014E` is accepted.

## 1. Database Schema, Idempotent Upgrade & SQLc

- [x] `db/001_schema.sql` defines `escalation_level` enum (`'self_care'`, `'cmo'`, `'flight_surgeon'`), `space_hazard`, and `medical_procedure` tables while preserving all existing tables, triggers, and views.
- [x] `db/004_medical_manual.sql` is idempotent and seeds exactly `5` `space_hazard` rows (`space_radiation`, `isolation_confinement`, `distance_from_earth`, `gravity_fields`, `hostile_closed_env`) and `20` `medical_procedure` rows whose `related_metric_codes` all exist in `db/002_seed.sql`.
- [x] `db/queries/medical.sql` defines `ListSpaceHazards`, `ListMedicalProcedures`, `SearchMedicalProcedures`, and `SearchSpaceHazards`, and `wails3 task db:sqlc` exits with `0`.
- [x] `Taskfile.yml` `db:seed` runs `./db/004_medical_manual.sql`.

## 2. Shared Models & Server Read/Search API

- [x] `shared/model/telemetry.go` defines `EscalationLevel`, `SpaceHazard`, `MedicalProcedure`, and `SearchResults` with camelCase JSON tags matching existing telemetry DTOs.
- [x] `GOWORK=off go -C shared test -v -count=1 ./...` exits with `0`.
- [x] `server/internal/query/service.go` implements `ListHazards`, `ListProcedures`, and `Search` (validating `tab` against `all|crew|procedures|hazards_metrics`), returning `ErrBadRequest` for invalid `tab`.
- [x] `server/internal/http/router.go` mounts `GET /api/hazards`, `GET /api/procedures`, and `GET /api/search` without breaking existing routes (`/`, `/health`, `/ws/realtime`, `/api/snapshot`, `/api/alerts/{id}/ack`, `/api/astronauts`, `/api/metrics`, `/api/trends`, `/api/history`).
- [x] `wails3 task server:check` exits with `0`.
- [x] With PostgreSQL 17 running and disposable `hrig_test` created, `HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable go -C server test -v -count=1 ./...` exits with `0` (all ingestion, query, and medical search integration tests pass), and `hrig_test` is dropped afterward.
- [x] Applying `db/004_medical_manual.sql` to local `hrig` database succeeds cleanly (`wails3 task db:seed`), verifying idempotency on an already-populated DB.

## 3. Scope & Ground Rules

- [x] No files inside `app/`, `bot/`, `bot-n/`, or `.routing/INPUT.md` are modified.
- [x] No fabricated Rx drug dosages or unverified clinical cutoffs appear in `db/004_medical_manual.sql`.
- [x] `.tickets/014E-medical-knowledge-db-and-search-api/changelog.md` is present and accurate.
