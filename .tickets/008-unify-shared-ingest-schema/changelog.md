# 008 — Change Log

## Summary

Unified `server` and `bot` ingestion wire contracts under `hrig/shared/model` by introducing canonical `POST /` DTOs (`IngestRequest`, `IngestReading`, `IngestResult`) in `shared/model/ingest.go`, wiring `hrig/shared` into both `server/go.mod` and `bot/go.mod`, aliasing `server/internal/ingestion` DTOs to `model`, validating `environment` via `model.EnvironmentMode.IsValid()`, and removing duplicate wire structs and raw enum literals from `bot/main.go`.

## Files Created & Modified

- **Created:**
  - `shared/model/ingest.go`: Exported canonical `IngestReading`, `IngestRequest` (`Environment EnvironmentMode`, `RecordedAtUTC string`), and `IngestResult` wire DTOs with `GUIDE.md` snake_case JSON tags.
  - `.tickets/008-unify-shared-ingest-schema/changelog.md`: This change log.
- **Modified:**
  - `shared/model/telemetry_test.go`: Added `TestIngestDTOJSONContract` covering JSON round-trip, snake_case wire keys, and `nil` omission for `astronaut_name` and `unit_code`.
  - `bot/go.mod`: Added `require hrig/shared v0.0.0` and `replace hrig/shared => ../shared`.
  - `server/go.mod`: Added `hrig/shared v0.0.0`, `replace hrig/shared => ../shared`, and indirect dependencies (`github.com/jackc/puddle/v2 v2.2.3`, `golang.org/x/sync v0.22.0`) required by `GOWORK=off go -C server mod tidy`.
  - `server/internal/ingestion/types.go`: Imported `hrig/shared/model` and replaced local `IngestRequest`, `ReadingInput`, and `IngestResult` structs with type aliases to `model.IngestRequest`, `model.IngestReading`, and `model.IngestResult`.
  - `server/internal/ingestion/service.go`: Normalized and validated `req.Environment` using `model.EnvironmentMode` and `env.IsValid()`, mapping `model.EnvIVA` -> `db.EnvironmentModeIva` and `model.EnvEVA` -> `db.EnvironmentModeEva`.
  - `bot/main.go`: Removed duplicate local `payload` and `reading` structs, imported `hrig/shared/model`, typed `sensor.Environment` as `model.EnvironmentMode` and `metric.Kind` as `model.ValueType` using `model.EnvIVA`/`model.EnvEVA` and `model.ValueNumeric`/`model.ValueScale`/`model.ValueImage`, formatted `RecordedAtUTC` with `time.RFC3339Nano`, and set `UnitCode` only when `m.Unit != ""` (leaving `ultrasound` `UnitCode` as `nil`).

## Verification Outputs

### 1. Shared Module Tests (`go -C shared test -v -count=1 ./...`)

Exit code: `0`

```text
=== RUN   TestEnumValidation
--- PASS: TestEnumValidation (0.00s)
=== RUN   TestJSONRoundTrip
--- PASS: TestJSONRoundTrip (0.00s)
=== RUN   TestIngestDTOJSONContract
--- PASS: TestIngestDTOJSONContract (0.00s)
PASS
ok  	hrig/shared/model	1.312s
```

### 2. Standalone Module Checks (`GOWORK=off`)

Initial `GOWORK=off go -C server vet ./...` reported `go: updates to go.mod needed; to update it: go mod tidy`. Running `GOWORK=off go -C server mod tidy` added the two indirect dependencies (`github.com/jackc/puddle/v2 v2.2.3` and `golang.org/x/sync v0.22.0`) to `server/go.mod`, after which both standalone vet checks passed:

```text
$env:GOWORK='off'; go -C server vet ./...
server_vet_exit=0

$env:GOWORK='off'; go -C bot vet .
bot_vet_exit=0
```

### 3. Workspace Server Check (`wails3 task server:check`)

Exit code: `0`

```text
task: [server:check] go vet ./...
task: [server:check] go test ./...
?   	hrig/server/cmd/server	[no test files]
?   	hrig/server/internal/config	[no test files]
?   	hrig/server/internal/db	[no test files]
ok  	hrig/server/internal/http	(cached)
ok  	hrig/server/internal/ingestion	(cached)
task: [server:check] go build -o ../bin/hrig-server.exe ./cmd/server
```

### 4. Bot Vet & Build (`go -C bot vet .` and `go -C bot build -o ../bin/hrig-bot.exe .`)

```text
bot_vet_exit=0
bot_build_exit=0
```

### 5. PostgreSQL 17 Integration Tests (`hrig_test`)

Started PostgreSQL via `wails3 task db:start`, created disposable `hrig_test`, ran `HRIG_TEST_DATABASE_URL="postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable" go -C server test -v -count=1 ./...`, dropped `hrig_test`, and stopped PostgreSQL via `wails3 task db:stop`.

Exit code: `0`

```text
task: [db:start] ./local/pgsql/bin/pg_ctl.exe -D ./local/pgsql/data -l ./local/pgsql/logfile -o "-h 127.0.0.1 -p 5432" start
esperando que el servidor se inicie.... listo
servidor iniciado
DROP DATABASE
CREATE DATABASE
?   	hrig/server/cmd/server	[no test files]
?   	hrig/server/internal/config	[no test files]
?   	hrig/server/internal/db	[no test files]
=== RUN   TestConfigDefaultsAndDatabaseURLFallback
--- PASS: TestConfigDefaultsAndDatabaseURLFallback (0.00s)
=== RUN   TestRootHelloWorld
--- PASS: TestRootHelloWorld (0.00s)
=== RUN   TestPostIngest_HTTPValidation
=== RUN   TestPostIngest_HTTPValidation/rejects_wrong_content_type
=== RUN   TestPostIngest_HTTPValidation/rejects_malformed_json
=== RUN   TestPostIngest_HTTPValidation/rejects_trailing_json_token
=== RUN   TestPostIngest_HTTPValidation/rejects_body_exceeding_1mib
=== RUN   TestPostIngest_HTTPValidation/returns_201_on_success
=== RUN   TestPostIngest_HTTPValidation/maps_ingestion_errors_and_hides_internal_errors
=== RUN   TestPostIngest_HTTPValidation/maps_ingestion_errors_and_hides_internal_errors/bad_request_400
=== RUN   TestPostIngest_HTTPValidation/maps_ingestion_errors_and_hides_internal_errors/not_found_404
=== RUN   TestPostIngest_HTTPValidation/maps_ingestion_errors_and_hides_internal_errors/conflict_409
=== RUN   TestPostIngest_HTTPValidation/maps_ingestion_errors_and_hides_internal_errors/unexpected_error_sanitized_500
--- PASS: TestPostIngest_HTTPValidation (0.01s)
PASS
ok  	hrig/server/internal/http	2.124s
=== RUN   TestClassifyReading_HeartRateBoundaries
--- PASS: TestClassifyReading_HeartRateBoundaries (0.00s)
=== RUN   TestClassifyReading_ManualMode
--- PASS: TestClassifyReading_ManualMode (0.00s)
=== RUN   TestClassifyReading_OneSidedAndScaleBounds
--- PASS: TestClassifyReading_OneSidedAndScaleBounds (0.00s)
=== RUN   TestIntegration_IngestAllBotSensorShapesAndClassification
--- PASS: TestIntegration_IngestAllBotSensorShapesAndClassification (0.28s)
=== RUN   TestIntegration_AtomicRollbackAndErrorCodes
--- PASS: TestIntegration_AtomicRollbackAndErrorCodes (0.58s)
PASS
ok  	hrig/server/internal/ingestion	1.081s
DROP DATABASE
task: [db:stop] ./local/pgsql/bin/pg_ctl.exe -D ./local/pgsql/data stop -m fast
esperando que el servidor se detenga.... listo
servidor detenido
```

### 6. Live `:80` End-to-End Smoke Test (`bin/hrig-server.exe` + `bin/hrig-bot.exe` against `hrig` DB)

Exit code: `0`

```text
BEFORE_COUNTS=66|288
AFTER_COUNTS=83|345
DELTA_MEASUREMENT=17
DELTA_MEASUREMENT_VALUE=57
BATCH_OK=True
--- BOT STDERR ---
2026/11/15 14:56:58 simulando 7 sensores lógicos para 3 astronautas; POST http://localhost:80/ cada 2s
2026/11/15 14:56:58 lote de lecturas enviado correctamente
--- SERVER LOGS ---
2026/11/15 14:56:56 hrig server listening on :80
task: [db:stop] ./local/pgsql/bin/pg_ctl.exe -D ./local/pgsql/data stop -m fast
esperando que el servidor se detenga.... listo
servidor detenido
```

### 7. Cleanup & Scope Verification

- `Get-Process -Name hrig-server,hrig-bot,postgres -ErrorAction SilentlyContinue` returned no running processes.
- `git diff --name-only` and `git status -s` confirmed only files in `task.md`'s `May create or edit` list were touched:
  - `M bot/go.mod`
  - `M bot/main.go`
  - `M server/go.mod`
  - `M server/internal/ingestion/service.go`
  - `M server/internal/ingestion/types.go`
  - `M shared/model/telemetry_test.go`
  - `?? .tickets/008-unify-shared-ingest-schema/changelog.md`
  - `?? shared/model/ingest.go`

## Notes & Assumptions Avoided

- Because the execution environment terminates child processes attached to a command's Windows job object when the command returns, PostgreSQL (`wails3 task db:start`), `psql`/`go test`/smoke-test binaries, and `wails3 task db:stop` were executed within self-contained `try`/`finally` PowerShell blocks so no stray daemons remained after exit.
- `shared/go.mod` and `shared/model/telemetry.go` remained completely unmodified.
