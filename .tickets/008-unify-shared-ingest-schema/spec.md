# 008 — Verification Spec

Every check is binary (`PASS` / `FAIL`). All checks must pass before ticket `008` is accepted.

## 1. Shared Ingestion Contract (`shared/`)

- [x] `shared/go.mod` contains only `module hrig/shared` and `go 1.26.0` (no external `require` block).
- [x] `shared/model/telemetry.go` is unmodified compared to `HEAD`.
- [x] `shared/model/ingest.go` exists and exports `IngestReading`, `IngestRequest`, and `IngestResult` with snake_case `json` tags matching `GUIDE.md` (`device_code`, `environment`, `recorded_at_utc`, `astronaut_name`, `readings`, `metric_code`, `value_numeric`, `value_scale`, `value_text`, `unit_code`, `measurement_id`, `values_count`).
- [x] `IngestRequest.Environment` uses `EnvironmentMode`; `IngestRequest.RecordedAtUTC` is `string`; `IngestReading.UnitCode` is `*string`.
- [x] `go -C shared test -v -count=1 ./...` exits with `0` and includes `TestIngestDTOJSONContract`.

## 2. Module Wiring (`server/` & `bot/`)

- [x] `server/go.mod` includes `require hrig/shared v0.0.0` and `replace hrig/shared => ../shared`.
- [x] `bot/go.mod` includes `require hrig/shared v0.0.0` and `replace hrig/shared => ../shared`.
- [x] `GOWORK=off go -C server vet ./...` exits with `0`.
- [x] `GOWORK=off go -C bot vet .` exits with `0`.

## 3. Single Source of Truth Adoption (`server/` & `bot/`)

- [x] `server/internal/ingestion/types.go` imports `hrig/shared/model` and defines type aliases `IngestRequest = model.IngestRequest`, `ReadingInput = model.IngestReading`, and `IngestResult = model.IngestResult` (no duplicate struct field definitions).
- [x] `server/internal/ingestion/service.go` validates `environment` using `model.EnvironmentMode` and `IsValid()`.
- [x] `bot/main.go` imports `hrig/shared/model`, has zero `type payload struct` or `type reading struct` definitions, uses `model.IngestRequest` and `model.IngestReading`, and uses `model.EnvIVA`, `model.EnvEVA`, `model.ValueNumeric`, `model.ValueScale`, and `model.ValueImage`.
- [x] `bot/main.go` sets `RecordedAtUTC` using `time.RFC3339Nano` and leaves `UnitCode` `nil` for unitless metrics (`ultrasound`).

## 4. Build, Unit, Integration & Live Smoke Verification

- [x] `wails3 task server:check` exits with `0`.
- [x] `go -C bot vet .` and `go -C bot build -o ../bin/hrig-bot.exe .` exit with `0`.
- [x] With PostgreSQL running and disposable `hrig_test` created, `HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable go -C server test -v -count=1 ./...` exits with `0` and `hrig_test` is dropped afterward.
- [x] Live `:80` smoke test (`bin/hrig-server.exe` + `bin/hrig-bot.exe` against `hrig` DB) completes at least 1 batch (`lote de lecturas enviado correctamente`) and increases `measurement` and `measurement_value` counts by an exact multiple of `+17` and `+57`.
- [x] No `hrig-server.exe`, `hrig-bot.exe`, or stray PostgreSQL process remains running after verification.

## 5. Scope & Ground Rules

- [x] `git diff --name-only` and untracked files show changes only inside `May create or edit` paths listed in `task.md`.
- [x] Code, `.tickets/`, and `changelog.md` are written in English (preserving existing Spanish user-facing log messages in `bot/main.go`).
- [x] `.tickets/008-unify-shared-ingest-schema/changelog.md` exists and records all changed files and exact command outputs.
