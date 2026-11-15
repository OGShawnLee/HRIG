# 011 — Change Log

## Files Created, Modified, and Regenerated

### Go Module & Backend (`app/`)
- **Modified:** `app/go.mod` — added `hrig/shared v0.0.0` with `replace hrig/shared => ../shared` and promoted `github.com/coder/websocket v1.8.14` to a direct dependency (no `github.com/gorilla/websocket` or `hrig/server` dependency). `app/go.sum` required no changes after `GOWORK=off go -C app mod tidy`.
- **Created:** `app/client.go` — implemented `resolveAPIBaseURL` (`HRIG_API_URL`, default `http://127.0.0.1:80`, deriving `ws`/`wss` + `/ws/realtime`), `ConnectionState`, `apiClient` typed REST methods (`GetHealth`, `GetAstronauts`, `GetMetrics`, `GetSnapshot`, `GetTrends`, `GetHistory`, `AckAlert`) with HTTP `>= 400` JSON `{"error":"..."}` decoding, `init()` `application.RegisterEvent` declarations for `"telemetry:envelope"` (`model.RealtimeEnvelope`) and `"telemetry:connection"` (`ConnectionState`), and `RunRealtimeLoop` using `github.com/coder/websocket` with bounded exponential backoff (`1s` -> `2s` -> `5s`).
- **Modified:** `app/app.go` — wired `*apiClient`, `ServiceStartup`, `ServiceShutdown`, `GetConnectionStatus`, `GetHealth`, `GetAstronauts`, `GetMetrics`, `GetSnapshot`, `GetTrends`, `GetHistory`, and `AckAlert`, while preserving `NewApp`, `SetContext`, `Ping`, and `app/app_window.go` (`OnMinimize`, `OnMaximize`, `OnClose`).
- **Created:** `app/client_test.go` — added unit and `httptest.Server` integration tests covering `resolveAPIBaseURL`, `backoffDuration`, all bound REST methods, 4xx JSON error decoding, validation on empty `alertID`, non-blocking `ServiceStartup` / clean `ServiceShutdown`, and `RunRealtimeLoop` WebSocket frame/event emission (`"telemetry:connection"`, `"telemetry:envelope"`).

### Regenerated TypeScript Bindings (`app/frontend/bindings/`)
- **Modified:** `app/frontend/bindings/hrig/app/app.ts` — exports `AckAlert`, `GetAstronauts`, `GetConnectionStatus`, `GetHealth`, `GetHistory`, `GetMetrics`, `GetSnapshot`, `GetTrends`, `OnClose`, `OnMaximize`, `OnMinimize`, `Ping`, `SetContext`.
- **Modified:** `app/frontend/bindings/hrig/app/index.ts`
- **Created:** `app/frontend/bindings/hrig/app/models.ts` — exports `ConnectionState`.
- **Created:** `app/frontend/bindings/hrig/shared/model/models.ts` and `app/frontend/bindings/hrig/shared/model/index.ts` — exports `AlertEvent`, `Astronaut`, `EnvironmentMode`, `HealthStatus`, `HistoryEntry`, `HistoryQuery`, `LatestReading`, `MediaKind`, `Metric`, `MetricThreshold`, `MetricTier`, `ReadingStatus`, `RealtimeEnvelope`, `Snapshot`, `TrendPoint`, `TrendQuery`, `ValueType`.
- **Modified:** `app/frontend/bindings/events/index.ts`, `app/frontend/bindings/github.com/wailsapp/wails/v3/internal/eventcreate.ts`, `app/frontend/bindings/github.com/wailsapp/wails/v3/internal/eventdata.d.ts` — generated custom event declarations (`TelemetryConnection`, `TelemetryEnvelope`).

### Temporary Artifacts
- Removed temporary `app/app.exe` produced by `go build .` inside `wails3 task client:check`.

## Verification Outputs

### 1. `$env:GOWORK='off'; go -C app mod tidy; go -C app vet .; Remove-Item Env:GOWORK`
Exited with code `0` (empty stdout/stderr).

### 2. `go -C app test -v -count=1 ./...`
```text
=== RUN   TestResolveAPIBaseURL
=== RUN   TestResolveAPIBaseURL/empty_uses_default
=== RUN   TestResolveAPIBaseURL/whitespace_uses_default
=== RUN   TestResolveAPIBaseURL/http_with_trailing_slash
=== RUN   TestResolveAPIBaseURL/https_with_multiple_trailing_slashes
=== RUN   TestResolveAPIBaseURL/uppercase_scheme_normalized
=== RUN   TestResolveAPIBaseURL/invalid_scheme
=== RUN   TestResolveAPIBaseURL/missing_host
=== RUN   TestResolveAPIBaseURL/malformed_url
--- PASS: TestResolveAPIBaseURL (0.00s)
    --- PASS: TestResolveAPIBaseURL/empty_uses_default (0.00s)
    --- PASS: TestResolveAPIBaseURL/whitespace_uses_default (0.00s)
    --- PASS: TestResolveAPIBaseURL/http_with_trailing_slash (0.00s)
    --- PASS: TestResolveAPIBaseURL/https_with_multiple_trailing_slashes (0.00s)
    --- PASS: TestResolveAPIBaseURL/uppercase_scheme_normalized (0.00s)
    --- PASS: TestResolveAPIBaseURL/invalid_scheme (0.00s)
    --- PASS: TestResolveAPIBaseURL/missing_host (0.00s)
    --- PASS: TestResolveAPIBaseURL/malformed_url (0.00s)
=== RUN   TestBackoffDuration
--- PASS: TestBackoffDuration (0.00s)
=== RUN   TestAppRESTBindingsAndErrorDecoding
--- PASS: TestAppRESTBindingsAndErrorDecoding (0.01s)
=== RUN   TestRunRealtimeLoopEmitsEventsAndShutsDownCleanly
--- PASS: TestRunRealtimeLoopEmitsEventsAndShutsDownCleanly (0.00s)
=== RUN   TestServiceStartupNonBlockingAndShutdown
--- PASS: TestServiceStartupNonBlockingAndShutdown (0.03s)
PASS
ok  	hrig/app	3.671s
```

### 3. `wails3 task client:check`
```text
task: [client:check] wails3 task check
task: [bindings] wails3 generate bindings -clean=true -ts
Wails v3.0.0 › Generate Bindings
 INFO  Processed: 267 Packages, 1 Service, 13 Methods, 5 Enums, 13 Models, 2 Events in 5.2464808s.
 INFO  Output directory: C:\Users\OGSha\OneDrive\Escritorio\Developer\Web Development\HRIG\app\frontend\bindings
task: [check] bun run --cwd frontend type-check
$ vue-tsc --noEmit
task: [check] bun run --cwd frontend build
$ vite build --mode production
vite v8.3.8 building client environment for production...
transforming...
✓ 61 modules transformed.
rendering chunks...
computing gzip size...
dist/index.html                   0.50 kB │ gzip:  0.34 kB
dist/assets/index-GjNYI1jp.css    6.67 kB │ gzip:  2.31 kB
dist/assets/index-D-iYOBjt.js   119.42 kB │ gzip: 43.95 kB

✓ built in 2.06s
task: [check] go build .

Need documentation? Run: wails3 docs
 ♥   If Wails is useful to you or your company, please consider sponsoring the project: wails3 sponsor

Need documentation? Run: wails3 docs
 ♥   If Wails is useful to you or your company, please consider sponsoring the project: wails3 sponsor
```

## Notes
- Registered `"telemetry:envelope"` (`model.RealtimeEnvelope`) and `"telemetry:connection"` (`ConnectionState`) in `init()` in `app/client.go` via `application.RegisterEvent` so `wails3 generate bindings` emits both `RealtimeEnvelope` in `app/frontend/bindings/hrig/shared/model/{index.ts,models.ts}` and typed event definitions in `app/frontend/bindings/events/index.ts`.
- Restored `go.work.sum` after an initial workspace-aware `go mod tidy` and ran `go -C app mod tidy` with `GOWORK=off` to keep `go.work.sum` untouched.
