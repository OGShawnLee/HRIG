# 009 — Changelog: Gorilla WebSocket Realtime Server & Alert/Stale Core

## Files Created

- `server/internal/realtime/met.go` — `FormatMET(recordedAtUTC, epochUTC time.Time) string` formatting `T+000d HH:MM:SS` / `T-000d HH:MM:SS`.
- `server/internal/realtime/hub.go` — Thread-safe `Hub` managing `latest` readings, bounded in-memory `AlertEvent` store (`maxStoredAlerts = 200`), connected Gorilla WebSocket `Client`s, `HydrateFromDB`, `PublishFrames` (emits `"reading"` and critical `"alert"` envelopes without writing alerts to PostgreSQL), `RunStaleMonitor` / `SweepStale` (runtime-only `StatusStale` `"stale"` envelopes), `AckAlert`, `Snapshot`, `ClientCount`, `ActiveAlertsCount`, and `Close`.
- `server/internal/realtime/client.go` — `ServeWS(hub *Hub, allowedOrigins []string) http.HandlerFunc` using `github.com/gorilla/websocket` `Upgrader` with origin validation, initial `"snapshot"` enqueue, single-writer `writePump` (30s ping keepalive, write deadlines), `readPump` (64 KiB read limit, pong deadline handler, inbound `"alert_ack"` support), and non-blocking slow-client eviction.
- `server/internal/realtime/hub_test.go` — Unit tests for `FormatMET`, `PublishFrames` reading + critical alert creation, `RunStaleMonitor` + `AckAlert` (including `ErrAlertNotFound` and `"crew"` default), and `HydrateFromDB` stale marking in memory only.
- `server/internal/realtime/ws_test.go` — End-to-end Gorilla WebSocket tests against `httptest.Server` verifying `CheckOrigin` enforcement, initial `"snapshot"`, live `"reading"` + `"alert"` broadcast, bidirectional inbound `"alert_ack"`, and non-blocking slow-client eviction.
- `.tickets/009-realtime-gorilla-ws-server/changelog.md` — This changelog.

## Files Modified

- `server/go.mod`, `server/go.sum` — Added `github.com/gorilla/websocket v1.5.3`.
- `server/internal/config/config.go` — Added `StaleAfter` (`HRIG_STALE_AFTER`, default `15s`) and `MissionEpochUTC` (`HRIG_MISSION_EPOCH_UTC`, RFC3339/RFC3339Nano, defaulting to `time.Now().UTC()`).
- `server/.env.example` — Documented `HRIG_STALE_AFTER=15s` and optional `HRIG_MISSION_EPOCH_UTC`.
- `server/internal/ingestion/types.go` — Added `FramePublisher` interface (`PublishFrames([]model.TelemetryFrame)`).
- `server/internal/ingestion/service.go` — Added `SetPublisher(pub FramePublisher)` and post-commit frame collection/emission strictly after `tx.Commit(ctx)` succeeds.
- `server/internal/ingestion/service_integration_test.go` — Added `TestIntegration_PostCommitPublisher_PublishesOnlyOnCommit` proving 0 batches published on rolled-back errors and committed frames published after `201 Created`.
- `server/internal/http/router.go` — Added `RealtimeController`, `DBPinger`, `RouterDeps`, and `NewRouterWithDeps` mounting `GET /ws/realtime`, `GET /api/snapshot`, `POST /api/alerts/{id}/ack`, and `GET /health` while preserving `NewRouter`, `GET /`, and `POST /`.
- `server/internal/http/router_test.go` — Added `TestConfigStaleAfterAndMissionEpoch` and `TestRealtimeAndHealthEndpoints` covering `/health`, `/api/snapshot`, `/api/alerts/{id}/ack` (`200`, `400`, `404`), and `/ws/realtime` handshake + `"alert_ack"` broadcast.
- `server/cmd/server/main.go` — Wired `realtime.Hub`, `HydrateFromDB`, `ingester.SetPublisher(hub)`, `go hub.RunStaleMonitor(ctx, time.Second)`, `RouterDeps`, and graceful `hub.Close()` shutdown.

## Verification Outputs

### 1. `GOWORK=off go -C server vet ./...`

```text
(exit code 0, no warnings or errors)
```

### 2. `wails3 task server:check`

```text
task: [server:check] go vet ./...
task: [server:check] go test ./...
?   	hrig/server/cmd/server	[no test files]
?   	hrig/server/internal/config	[no test files]
?   	hrig/server/internal/db	[no test files]
ok  	hrig/server/internal/http	(cached)
ok  	hrig/server/internal/ingestion	(cached)
ok  	hrig/server/internal/realtime	3.531s
task: [server:check] go build -o ../bin/hrig-server.exe ./cmd/server
```

Exit code: `0`.

### 3. `go -C server test -race -count=1 -v ./internal/realtime/... ./internal/http/...`

CGO is enabled on Windows (`CGO_ENABLED=1`, `CC=gcc`), so the full Go race detector ran and passed with zero data races:

```text
=== RUN   TestFormatMET
=== RUN   TestFormatMET/zero
=== RUN   TestFormatMET/positive_multi_day
=== RUN   TestFormatMET/negative
--- PASS: TestFormatMET (0.00s)
    --- PASS: TestFormatMET/zero (0.00s)
    --- PASS: TestFormatMET/positive_multi_day (0.00s)
    --- PASS: TestFormatMET/negative (0.00s)
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
ok  	hrig/server/internal/realtime	3.933s
=== RUN   TestConfigDefaultsAndDatabaseURLFallback
--- PASS: TestConfigDefaultsAndDatabaseURLFallback (0.00s)
=== RUN   TestRootHelloWorld
--- PASS: TestRootHelloWorld (0.00s)
=== RUN   TestPostIngest_HTTPValidation
--- PASS: TestPostIngest_HTTPValidation (0.04s)
=== RUN   TestConfigStaleAfterAndMissionEpoch
--- PASS: TestConfigStaleAfterAndMissionEpoch (0.00s)
=== RUN   TestRealtimeAndHealthEndpoints
--- PASS: TestRealtimeAndHealthEndpoints (0.01s)
PASS
ok  	hrig/server/internal/http	3.974s
```

### 4. PostgreSQL 17 Integration Suite (`hrig_test`)

Started local PostgreSQL 17, created disposable `hrig_test`, ran `HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable go -C server test -v -count=1 ./...`, then dropped `hrig_test` and stopped PostgreSQL:

```text
ok  	hrig/server/internal/http	3.869s
=== RUN   TestClassifyReading_HeartRateBoundaries
--- PASS: TestClassifyReading_HeartRateBoundaries (0.00s)
=== RUN   TestClassifyReading_ManualMode
--- PASS: TestClassifyReading_ManualMode (0.00s)
=== RUN   TestClassifyReading_OneSidedAndScaleBounds
--- PASS: TestClassifyReading_OneSidedAndScaleBounds (0.00s)
=== RUN   TestIntegration_IngestAllBotSensorShapesAndClassification
--- PASS: TestIntegration_IngestAllBotSensorShapesAndClassification (0.70s)
=== RUN   TestIntegration_AtomicRollbackAndErrorCodes
--- PASS: TestIntegration_AtomicRollbackAndErrorCodes (0.66s)
=== RUN   TestIntegration_PostCommitPublisher_PublishesOnlyOnCommit
--- PASS: TestIntegration_PostCommitPublisher_PublishesOnlyOnCommit (0.57s)
PASS
ok  	hrig/server/internal/ingestion	2.120s
ok  	hrig/server/internal/realtime	3.612s
```

Exit code: `0`.

### 5. Live `:80` End-to-End Smoke Test (`bin/hrig-server.exe` + `bin/hrig-bot.exe` + Gorilla WS Consumer)

Started PostgreSQL 17 with a disposable `hrig_smoke` database seeded with `db/001_schema.sql`, `db/002_seed.sql`, and `db/003_demo_seed.sql`, launched `bin/hrig-server.exe` on `:80`, connected a Gorilla WebSocket consumer to `ws://127.0.0.1:80/ws/realtime`, started `bin/hrig-bot.exe` (`SENSOR_INTERVAL=1s`), queried `GET http://127.0.0.1:80/health` and `GET http://127.0.0.1:80/api/snapshot`, then stopped all processes, dropped `hrig_smoke`, and stopped PostgreSQL:

```text
=== WS SMOKE OUTPUT ===
WS_SNAPSHOT type=snapshot readings=39 alerts=0
WS_LIVE readings=15 alerts=0 unique_metrics=9 metrics=[body_mass core_temp heart_rate hrv_rmssd hydration_intake mood resp_rate sleep_hours spo2]
HTTP_HEALTH {"status":"ok","store":"postgres","timeUtc":"2026-11-15T21:17:34.5197789Z","activeAlerts":0,"realtimeClients":1,"ingestClients":1}

HTTP_SNAPSHOT readings=45 alerts=0 first_metric=body_mass met=T+000d 00:00:00
=== SERVER LOG ===
2026/11/15 15:17:33 hrig server listening on :80
=== BOT LOG ===
2026/11/15 15:17:34 simulando 7 sensores lógicos para 3 astronautas; POST http://127.0.0.1:80/ cada 1s
2026/11/15 15:17:34 lote de lecturas enviado correctamente
2026/11/15 15:17:35 lote de lecturas enviado correctamente
```

Verified `Get-Process -Name hrig-server,hrig-bot,postgres` returns no active processes afterward.

## Notes & Fixes Applied

- Updated `TestWebSocketSlowClient_DoesNotBlockPublishFrames` in `server/internal/realtime/ws_test.go` to drain `fastConn` concurrently in a reader goroutine across two bursts (`40` + `44` messages = `84 > sendBufferSize`), preventing a tight synchronous loop from overflowing the fast client's `64`-item send buffer before `writePump` could drain it while still guaranteeing the stalled `slowClient` is evicted without blocking `PublishFrames`.
- Used disposable `hrig_test` and `hrig_smoke` databases during verification so the persistent `hrig` database was not left with smoke-test artifacts, and dropped both databases and stopped PostgreSQL upon completion.
