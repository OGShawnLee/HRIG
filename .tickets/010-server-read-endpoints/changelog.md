# 010 — Change Log

## Files Created & Modified

### Created
- `server/internal/query/service.go` — `query.Service` (`NewService`, `NewServiceWithQuerier`) implementing `ListAstronauts`, `ListMetrics`, `ListTrends`, and `ListHistory` over `db.Querier` (`ListActiveAstronauts`, `ListActiveMetrics`, `ListTrendsByMetric`, `ListHistory`), with `ErrBadRequest` validation, `realtime.FormatMET` calculation, `pgtype.Numeric`/`pgtype.UUID`/`pgtype.Timestamptz` mapping, and optional `media_asset` mapping.
- `server/internal/query/service_test.go` — Mock `db.Querier` unit tests covering astronaut mapping, metric + threshold + `BaseUnitDecimals` mapping, trend validation/trailing-limit slicing in ASC order, and history validation/pagination/media mapping.
- `server/internal/query/service_integration_test.go` — PostgreSQL integration test (`TestIntegration_QueryService`) executing `001_schema.sql`, `002_seed.sql`, and `003_demo_seed.sql` in an isolated `query_test` schema against `HRIG_TEST_DATABASE_URL`, verifying 3 active seeded astronauts (`Ana Torres`, `Liam Okafor`, `Sofia Petrov`) with valid UUIDs and derived `Age`, 23 active metrics with thresholds and unit decimals, chronological `spo2` and `cabin_o2` trends, `limit`/`offset` and `astronaut_id` history filtering, and `media_asset` mapping.
- `.tickets/010-server-read-endpoints/changelog.md` — This change log.

### Modified
- `shared/model/telemetry.go` — Added `BaseUnitDimension` and `BaseUnitDecimals` to `Metric`, added `Offset` to `HistoryQuery`, and added `TrendPoint` and `HistoryEntry` DTOs with camelCase JSON tags.
- `shared/model/telemetry_test.go` — Added JSON round-trip assertions for `Metric` base unit fields, `TrendPoint`, `HistoryEntry` (numeric and media variants), and `HistoryQuery`.
- `server/internal/http/router.go` — Added `TelemetryReader` interface, `Reader` field on `RouterDeps`, mounted `GET /api/astronauts`, `GET /api/metrics`, `GET /api/trends`, and `GET /api/history`, added query-parameter validation (`400` JSON `{"error":"..."}`), `503` handling when `Reader == nil`, and sanitized `500` error responses while preserving all existing routes and `NewRouter(cfg, ingester)`.
- `server/internal/http/router_test.go` — Added `TestReadEndpoints` verifying `503` when `Reader == nil`, `200` and sanitized `500` responses for `/api/astronauts` and `/api/metrics`, and `200` / `400` query validation cases for `/api/trends` and `/api/history`.
- `server/cmd/server/main.go` — Constructed `reader := query.NewService(pool, cfg.MissionEpochUTC)` and wired `Reader: reader` into `serverhttp.RouterDeps`.

## Verification Command Outputs

### 1. Shared Model Tests
Command:
```powershell
$env:GOWORK='off'; go -C shared test -v ./...
```
Output (exit code `0`):
```text
=== RUN   TestEnumValidation
--- PASS: TestEnumValidation (0.00s)
=== RUN   TestJSONRoundTrip
--- PASS: TestJSONRoundTrip (0.00s)
=== RUN   TestIngestDTOJSONContract
--- PASS: TestIngestDTOJSONContract (0.00s)
PASS
ok  	hrig/shared/model	1.497s
```

### 2. Server Vet
Command:
```powershell
$env:GOWORK='off'; go -C server vet ./...
```
Output (exit code `0`):
```text
(no warnings or errors)
```

### 3. Server Check (`wails3 task server:check`)
Command:
```powershell
$env:GOWORK='off'; wails3 task server:check
```
Output (exit code `0`):
```text
task: [server:check] go vet ./...
task: [server:check] go test ./...
?   	hrig/server/cmd/server	[no test files]
?   	hrig/server/internal/config	[no test files]
?   	hrig/server/internal/db	[no test files]
ok  	hrig/server/internal/http	3.689s
ok  	hrig/server/internal/ingestion	0.344s
ok  	hrig/server/internal/query	0.348s
ok  	hrig/server/internal/realtime	3.340s
task: [server:check] go build -o ../bin/hrig-server.exe ./cmd/server
```

### 4. Disposable PostgreSQL (`hrig_test`) Integration Tests
Commands:
```powershell
wails3 task db:start
./local/pgsql/bin/psql.exe -h 127.0.0.1 -p 5432 -U postgres -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS hrig_test;" -c "CREATE DATABASE hrig_test;"
$env:HRIG_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable'; go -C server test -v -count=1 ./...
./local/pgsql/bin/psql.exe -h 127.0.0.1 -p 5432 -U postgres -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS hrig_test;"
wails3 task db:stop
```
Output (exit code `0`):
```text
?   	hrig/server/cmd/server	[no test files]
?   	hrig/server/internal/config	[no test files]
?   	hrig/server/internal/db	[no test files]
=== RUN   TestConfigDefaultsAndDatabaseURLFallback
--- PASS: TestConfigDefaultsAndDatabaseURLFallback (0.00s)
=== RUN   TestRootHelloWorld
--- PASS: TestRootHelloWorld (0.00s)
=== RUN   TestPostIngest_HTTPValidation
--- PASS: TestPostIngest_HTTPValidation (0.01s)
=== RUN   TestConfigStaleAfterAndMissionEpoch
--- PASS: TestConfigStaleAfterAndMissionEpoch (0.00s)
=== RUN   TestRealtimeAndHealthEndpoints
--- PASS: TestRealtimeAndHealthEndpoints (0.01s)
=== RUN   TestReadEndpoints
    --- PASS: TestReadEndpoints/returns_503_when_reader_nil (0.00s)
    --- PASS: TestReadEndpoints/astronauts_and_metrics_200_and_500_sanitized (0.00s)
    --- PASS: TestReadEndpoints/trends_200_and_400_cases (0.00s)
    --- PASS: TestReadEndpoints/history_200_and_400_cases (0.00s)
--- PASS: TestReadEndpoints (0.00s)
PASS
ok  	hrig/server/internal/http	3.751s
=== RUN   TestClassifyReading_HeartRateBoundaries
--- PASS: TestClassifyReading_HeartRateBoundaries (0.00s)
=== RUN   TestClassifyReading_ManualMode
--- PASS: TestClassifyReading_ManualMode (0.00s)
=== RUN   TestClassifyReading_OneSidedAndScaleBounds
--- PASS: TestClassifyReading_OneSidedAndScaleBounds (0.00s)
=== RUN   TestIntegration_IngestAllBotSensorShapesAndClassification
--- PASS: TestIntegration_IngestAllBotSensorShapesAndClassification (0.74s)
=== RUN   TestIntegration_AtomicRollbackAndErrorCodes
--- PASS: TestIntegration_AtomicRollbackAndErrorCodes (0.63s)
=== RUN   TestIntegration_PostCommitPublisher_PublishesOnlyOnCommit
--- PASS: TestIntegration_PostCommitPublisher_PublishesOnlyOnCommit (0.54s)
PASS
ok  	hrig/server/internal/ingestion	2.207s
=== RUN   TestServiceListAstronauts
--- PASS: TestServiceListAstronauts (0.00s)
=== RUN   TestServiceListMetrics
--- PASS: TestServiceListMetrics (0.00s)
=== RUN   TestServiceListTrends_ValidationAndSlicing
--- PASS: TestServiceListTrends_ValidationAndSlicing (0.00s)
=== RUN   TestServiceListHistory_ValidationAndMediaMapping
--- PASS: TestServiceListHistory_ValidationAndMediaMapping (0.00s)
=== RUN   TestIntegration_QueryService
--- PASS: TestIntegration_QueryService (0.83s)
PASS
ok  	hrig/server/internal/query	1.127s
=== RUN   TestFormatMET
--- PASS: TestFormatMET (0.00s)
=== RUN   TestHubPublishFrames_ReadingAndCriticalAlert
--- PASS: TestHubPublishFrames_ReadingAndCriticalAlert (0.00s)
=== RUN   TestHubStaleMonitorAndAckAlert
--- PASS: TestHubStaleMonitorAndAckAlert (0.01s)
=== RUN   TestHydrateFromDB_MarksOldReadingsStaleInMemoryOnly
--- PASS: TestHydrateFromDB_MarksOldReadingsStaleInMemoryOnly (0.00s)
=== RUN   TestWebSocketEndToEnd_SnapshotBroadcastAckAndOriginCheck
--- PASS: TestWebSocketEndToEnd_SnapshotBroadcastAckAndOriginCheck (0.02s)
=== RUN   TestWebSocketSlowClient_DoesNotBlockPublishFrames
--- PASS: TestWebSocketSlowClient_DoesNotBlockPublishFrames (0.01s)
PASS
ok  	hrig/server/internal/realtime	3.377s
DROP DATABASE
task: [db:stop] ./local/pgsql/bin/pg_ctl.exe -D ./local/pgsql/data stop -m fast
esperando que el servidor se detenga.... listo
servidor detenido
```

## Notes & Assumptions Avoided
- **Parallel Integration Isolation:** Because `go -C server test -v -count=1 ./...` executes `server/internal/ingestion` and `server/internal/query` packages concurrently against the same `HRIG_TEST_DATABASE_URL`, `query_test`'s `setupTestDB` creates and drops a dedicated `query_test` schema (via `search_path=query_test`) rather than resetting `public`, preventing cross-package schema reset races without modifying `server/internal/ingestion/**`.
- **SQLC Read-Only Compliance:** Reused the existing `db.Querier` methods (`ListActiveAstronauts`, `ListActiveMetrics`, `ListTrendsByMetric`, `ListHistory`) without modifying `db/queries/**` or `server/internal/db/*.go`.
- **Process Cleanup:** Confirmed `hrig_test` was dropped and no `hrig-server.exe`, `hrig-bot.exe`, or `postgres.exe` process remains running.
