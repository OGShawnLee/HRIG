# 014E — Medical Knowledge DB Schema, Seed Data, and Search API Changelog

## Files Created or Modified

### Database Schema, Seed & SQLc
- `db/001_schema.sql`: Added `escalation_level` enum (`'self_care'`, `'cmo'`, `'flight_surgeon'`), `space_hazard` table, `medical_procedure` table (with `estimated_minutes > 0` and `array_length(steps, 1) >= 1` checks), and `idx_medical_procedure_active` index while preserving all existing tables, triggers, and views.
- `db/004_medical_manual.sql`: Created idempotent DDL upgrade (`DO $$ ... EXCEPTION WHEN duplicate_object`, `CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`) and `ON CONFLICT (code) DO UPDATE` seed for the 5 NASA HRP hazards (`space_radiation`, `isolation_confinement`, `distance_from_earth`, `gravity_fields`, `hostile_closed_env`) and 20 operational medical procedures covering all 23 cataloged metrics in `db/002_seed.sql`, strictly using non-physician operational checklists without fabricated medication dosages or unverified clinical cutoffs.
- `db/queries/medical.sql`: Added `ListSpaceHazards`, `SearchSpaceHazards`, `ListMedicalProcedures`, and `SearchMedicalProcedures`.
- `db/README.md`: Documented `db/004_medical_manual.sql`.
- `Taskfile.yml`: Appended `./db/004_medical_manual.sql` execution to `db:seed`.
- `server/internal/db/medical.sql.go`, `server/internal/db/models.go`, `server/internal/db/querier.go`: Regenerated via `wails3 task db:sqlc`.

### Shared Contracts
- `shared/model/telemetry.go`: Added `EscalationLevel` (`EscalationSelfCare`, `EscalationCMO`, `EscalationFlightSurgeon`, `IsValid()`), `SpaceHazard`, `MedicalProcedure`, and `SearchResults` with camelCase JSON tags and `.Normalize()` slice-normalization helpers.
- `shared/model/telemetry_test.go`: Added `EscalationLevel.IsValid()` checks, JSON round-trip assertions for `SpaceHazard`, `MedicalProcedure`, and `SearchResults`, and `TestMedicalKnowledgeNormalization`.

### Query Service & HTTP Router
- `server/internal/query/service.go`: Implemented `ListHazards`, `ListProcedures`, and `Search` (validating `tab` against `""`, `"all"`, `"crew"`, `"procedures"`, `"hazards_metrics"`, returning `ErrBadRequest` for invalid `tab`, normalizing slices, and filtering astronauts/metrics case-insensitively).
- `server/internal/query/service_test.go`: Extended `mockQuerier` with the 4 medical SQLc methods and added `TestServiceMedicalKnowledgeAndSearch`.
- `server/internal/query/service_integration_test.go`: Loaded `db/004_medical_manual.sql` twice in `setupTestDB` to verify idempotency and added integration checks for the 5 hazards, 20 procedures, full 23-metric coverage, category/metric/keyword filtering, and `GET /api/search` HTTP behavior.
- `server/internal/http/router.go`: Extended `TelemetryReader` with `ListHazards`, `ListProcedures`, and `Search`, and mounted `GET /api/hazards`, `GET /api/procedures`, and `GET /api/search` returning `503` when `Reader == nil` and `400` on `query.ErrBadRequest`.
- `server/internal/http/router_test.go`: Updated `mockReader` and 503 checks, and added `hazards_procedures_and_search_cases`.

## Verification Results

1. `wails3 task db:sqlc` — **PASS** (exit code `0`, generated `server/internal/db/medical.sql.go`, updated `models.go` and `querier.go`).
2. `GOWORK=off go -C shared test -v -count=1 ./...` — **PASS** (`TestEnumValidation`, `TestJSONRoundTrip`, `TestMedicalKnowledgeNormalization`, `TestIngestDTOJSONContract` passed).
3. `wails3 task server:check` — **PASS** (`go vet ./...`, `go test ./...`, and `go build -o ../bin/hrig-server.exe ./cmd/server` exited `0`).
4. `wails3 task db:seed` on local `hrig` (executed twice) — **PASS** (first run created `escalation_level`, `space_hazard`, `medical_procedure`, `idx_medical_procedure_active`, inserted `5` hazards and `20` procedures; second run skipped existing DDL objects cleanly and upserted `5` hazards and `20` procedures with verified counts `hazards = 5`, `procedures = 20`).
5. `HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable go -C server test -v -count=1 ./...` — **PASS** (`hrig/server/internal/http`, `hrig/server/internal/ingestion`, `hrig/server/internal/query` including `TestIntegration_QueryService`, and `hrig/server/internal/realtime` all passed).
6. Cleanup: Dropped disposable `hrig_test` database (`DROP DATABASE IF EXISTS hrig_test;`) and stopped PostgreSQL (`wails3 task db:stop`) — **PASS**.

## Notes & Edge Cases

- `psql -c "DROP DATABASE IF EXISTS hrig_test; CREATE DATABASE hrig_test;"` fails when combined into one `-c` string because PostgreSQL wraps multi-statement `-c` arguments in a transaction block; issuing separate `-c` arguments resolved this.
- When `Search` receives an empty `tab` query parameter, it normalizes `SearchResults.Tab` to `"all"` and initializes excluded category slices as non-nil empty slices (`[]`) so JSON serialization never produces `null`.
- PostgreSQL was already running when verification began; per instructions, `hrig_test` was dropped and PostgreSQL was stopped via `wails3 task db:stop` after all database checks completed.
