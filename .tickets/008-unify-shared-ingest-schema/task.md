# 008 — Unify server and bot under `hrig/shared/model` ingestion schema

| Field | Value |
| :--- | :--- |
| **ID** | `008` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Go Shared Contracts & Refactoring Engineer |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `shared/model/ingest.go` (canonical `POST /` wire DTOs: `IngestRequest`, `IngestReading`, `IngestResult`)
  - `shared/model/telemetry_test.go` (JSON round-trip and wire-tag verification tests for the ingestion DTOs)
  - `bot/go.mod`, `server/go.mod` (add `require hrig/shared v0.0.0` and `replace hrig/shared => ../shared`)
  - `bot/main.go` (remove duplicate wire structs `payload` and `reading`; import `hrig/shared/model` and use `model.IngestRequest`, `model.IngestReading`, `model.EnvironmentMode` constants `EnvIVA`/`EnvEVA`, and `model.ValueType` constants `ValueNumeric`/`ValueScale`/`ValueImage`)
  - `server/internal/ingestion/types.go` (alias `IngestRequest = model.IngestRequest`, `ReadingInput = model.IngestReading`, `IngestResult = model.IngestResult`)
  - `server/internal/ingestion/service.go` (validate `req.Environment` via `model.EnvironmentMode` / `IsValid()`)
  - `server/internal/ingestion/service_integration_test.go` (import `hrig/shared/model` if using shared constants/assertions without altering test assertions)
  - `.tickets/008-unify-shared-ingest-schema/changelog.md`
- **Must not touch:**
  - `shared/go.mod`, `shared/model/telemetry.go`
  - `db/001_schema.sql`, `db/002_seed.sql`, `db/003_demo_seed.sql`, `db/README.md`, `db/queries/**`
  - `sqlc.yaml`, `server/internal/db/*.go`
  - `server/internal/config/config.go`, `server/internal/http/router.go`, `server/cmd/server/main.go`
  - `app/**`, `local/**`, `docs/**`, `Taskfile.yml`, `go.work`
  - `.gitignore`, `.routing/**`, `.claude/**`, `README.md`, `PROJECT-REVIEW.md`, `GUIDE.md`, `SENSOR-READINGS.md`
  - `.tickets/001-*`, `.tickets/002-*`, `.tickets/003-*`, `.tickets/004-*`, `.tickets/005-*`, `.tickets/006-*`, `.tickets/007-*`
  - `.tickets/008-unify-shared-ingest-schema/task.md`, `.tickets/008-unify-shared-ingest-schema/spec.md`

## Inputs

Read these files before acting:
1. [`GUIDE.md`](../../GUIDE.md)
2. [`go.work`](../../go.work)
3. [`shared/go.mod`](../../shared/go.mod)
4. [`shared/model/telemetry.go`](../../shared/model/telemetry.go)
5. [`shared/model/telemetry_test.go`](../../shared/model/telemetry_test.go)
6. [`bot/go.mod`](../../bot/go.mod)
7. [`bot/main.go`](../../bot/main.go)
8. [`server/go.mod`](../../server/go.mod)
9. [`server/internal/ingestion/types.go`](../../server/internal/ingestion/types.go)
10. [`server/internal/ingestion/service.go`](../../server/internal/ingestion/service.go)
11. [`server/internal/http/router.go`](../../server/internal/http/router.go)
12. [`server/internal/http/router_test.go`](../../server/internal/http/router_test.go)
13. [`server/internal/ingestion/service_integration_test.go`](../../server/internal/ingestion/service_integration_test.go)

## Grounding & Anti-Hallucination Constraints

- **Exact `GUIDE.md` Wire Schema:**
  - `model.IngestRequest`:
    - `DeviceCode string` (`json:"device_code"`)
    - `Environment EnvironmentMode` (`json:"environment"`)
    - `RecordedAtUTC string` (`json:"recorded_at_utc"`)
    - `AstronautName *string` (`json:"astronaut_name,omitempty"`)
    - `Readings []IngestReading` (`json:"readings"`)
  - `model.IngestReading`:
    - `MetricCode string` (`json:"metric_code"`)
    - `ValueNumeric *float64` (`json:"value_numeric,omitempty"`)
    - `ValueScale *int` (`json:"value_scale,omitempty"`)
    - `ValueText *string` (`json:"value_text,omitempty"`)
    - `UnitCode *string` (`json:"unit_code,omitempty"`)
  - `model.IngestResult`:
    - `MeasurementID string` (`json:"measurement_id"`)
    - `RecordedAtUTC time.Time` (`json:"recorded_at_utc"`)
    - `ValuesCount int` (`json:"values_count"`)
- **Timestamp & Unit Pointer Semantics:**
  - Keep `RecordedAtUTC` as `string` in `model.IngestRequest` so `server` preserves its exact `400` validation messages (`"recorded_at_utc is required"` and `"recorded_at_utc must be RFC3339 with timezone"`). `bot/main.go` must serialize `time.Now().UTC().Format(time.RFC3339Nano)`.
  - Keep `UnitCode` as `*string` in `model.IngestReading` so `server` continues distinguishing omitted/`null` from `""` on `text`/`image` metrics. In `bot/main.go`, leave `UnitCode` as `nil` when `m.Unit == ""` and pass a pointer copy `&unit` when `m.Unit != ""`.
- **Zero Behavioral Drift in `bot/main.go`:**
  - Keep the exact same 7 devices (`biomonitor_01`, `baumanometro_01`, `dosimeter_01`, `cabin_sensor_01`, `suit_sensor_01`, `lab_assay_01`, `ultrasound_01`), 3 crew names (`Ana Torres`, `Liam Okafor`, `Sofia Petrov`), 22 metric codes, numeric/scale ranges, CLI env vars (`SENSOR_ENDPOINT`, `SENSOR_INTERVAL`, `SENSOR_CREW_COUNT`), and Spanish log output.
  - Keep simulator-only configuration structs (`sensor`, `metric`, `generator`) local to `bot/main.go`, while typing `sensor.Environment` as `model.EnvironmentMode` (`model.EnvIVA` / `model.EnvEVA`) and `metric.Kind` as `model.ValueType` (`model.ValueNumeric` / `model.ValueScale` / `model.ValueImage`).
- **Stdlib-Only Shared Module:**
  - Do **not** add external dependencies to `shared/go.mod`.
- **No Unsolicited Git / Daemon Side Effects:**
  - Do **not** commit or push changes.
  - Do **not** leave `hrig-server.exe`, `hrig-bot.exe`, PostgreSQL, or `hrig_test` running after verification.

## Task

1. **Create `shared/model/ingest.go`:**
   - Define `IngestReading`, `IngestRequest`, and `IngestResult` in package `model` with the exact field names, Go types, and `json` struct tags listed in Grounding & Anti-Hallucination Constraints.
2. **Add shared contract tests in `shared/model/telemetry_test.go`:**
   - Add a test (`TestIngestDTOJSONContract`) verifying JSON round-trip and exact `snake_case` wire keys (`device_code`, `environment`, `recorded_at_utc`, `astronaut_name`, `readings`, `metric_code`, `value_numeric`, `value_scale`, `value_text`, `unit_code`, `measurement_id`, `values_count`), plus verification that `unit_code` and `astronaut_name` are omitted when `nil`.
3. **Wire `hrig/shared` into `server/go.mod` and `bot/go.mod`:**
   - Add `require hrig/shared v0.0.0` and `replace hrig/shared => ../shared` to both `server/go.mod` and `bot/go.mod` so both modules resolve `hrig/shared/model` in `go.work` and standalone `GOWORK=off` mode.
4. **Switch `server/internal/ingestion/` to consume `hrig/shared/model`:**
   - Update `server/internal/ingestion/types.go` to import `"hrig/shared/model"` and declare type aliases:
     - `type IngestRequest = model.IngestRequest`
     - `type ReadingInput = model.IngestReading`
     - `type IngestResult = model.IngestResult`
   - Keep `Error`, `badRequest`, `notFound`, `conflict`, `internalError`, and `StatusCodeFromError` unchanged in `server/internal/ingestion/types.go`.
   - Update `server/internal/ingestion/service.go` to trim and normalize `req.Environment` as `model.EnvironmentMode` and validate using `env.IsValid()` (returning `badRequest("environment must be 'iva' or 'eva'")` when invalid), mapping `model.EnvIVA` -> `db.EnvironmentModeIva` and `model.EnvEVA` -> `db.EnvironmentModeEva`.
5. **Switch `bot/main.go` to consume `hrig/shared/model`:**
   - Import `"hrig/shared/model"` in `bot/main.go`.
   - Delete the duplicate local `reading` and `payload` structs from `bot/main.go`.
   - Update `sensor.Environment` to `model.EnvironmentMode` and `metric.Kind` to `model.ValueType`, replacing raw literals `"iva"`, `"eva"`, `"numeric"`, `"scale"`, `"image"` with `model.EnvIVA`, `model.EnvEVA`, `model.ValueNumeric`, `model.ValueScale`, `model.ValueImage`.
   - Update `sendBatch` to build `model.IngestRequest` (`RecordedAtUTC: time.Now().UTC().Format(time.RFC3339Nano)`) and `makeReadings` to return `[]model.IngestReading` setting `UnitCode` to `&unit` only when `m.Unit != ""`.
6. **Verify across `shared`, `server`, `bot`, and live PostgreSQL 17:**
   - Run `go -C shared test -v ./...`.
   - Run `wails3 task server:check` and `GOWORK=off go -C server vet ./...`.
   - Run `go -C bot vet .`, `GOWORK=off go -C bot vet .`, and `go -C bot build -o ../bin/hrig-bot.exe .`.
   - Start PostgreSQL (`wails3 task db:start`), create disposable `hrig_test`, run uncached `HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable go -C server test -v -count=1 ./...`, then drop `hrig_test`.
   - Run a live `:80` end-to-end smoke test (`hrig-server.exe` + `hrig-bot.exe` against `hrig` DB for at least 1 complete batch), confirm `measurement` and `measurement_value` counts increase by `+17` and `+57` (or exact multiple per batch), then stop `bot`, `hrig-server.exe`, and `wails3 task db:stop`.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/008-unify-shared-ingest-schema/changelog.md` listing:
- Files created and modified
- Exact test & build command outputs (`shared`, `server`, `bot`, `GOWORK=off` checks, `hrig_test` integration suite, and live `bot` smoke test)
- Any blockers or assumptions avoided
