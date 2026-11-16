# 032 — Change Log

## Summary

Implemented astronaut manual daily check-ins for self-reported secondary metrics (`sleep_hours`, `mood`, `hydration_intake`, `body_mass`), stopped the telemetry bot from streaming those 4 daily metrics every 2 seconds, added a 24-hour staleness window for daily check-in readings in the realtime hub, exposed `SubmitCheckin` across the Wails Go backend and TypeScript bindings, added the yellow pending-check-in circle (`bg-warn`) next to astronaut names on the main menu (`CrewCard.vue`) and crew selector (`dashboard-secundario.vue`), and added inline check-in logging controls inside `dashboard-secundario.vue`.

## Modified Files

1. `bot/main.go`:
   - Removed `sleep_hours`, `mood`, `hydration_intake`, and `body_mass` from `biomonitor_01`'s `Metrics` slice so `biomonitor_01` only streams continuous wearable vitals (`spo2`, `resp_rate`, `heart_rate`, `hrv_rmssd`, `core_temp`).
2. `db/002_seed.sql`:
   - Added `('manual_checkin', 'Crew manual daily check-in', 'manual')` to `INSERT INTO device (code, name, kind)`.
3. `server/internal/realtime/hub.go`:
   - Added `dailyCheckinStaleAfter = 24 * time.Hour` and `isDailyCheckinReading(metricCode, deviceCode, source string) bool`.
   - Updated `HydrateFromDB` and `SweepStale()` to apply the 24-hour staleness threshold to daily check-in readings.
4. `app/client.go`:
   - Added `func (c *apiClient) SubmitCheckin(ctx context.Context, req model.IngestRequest) (model.IngestResult, error)` posting to `"/"` with defaults (`device_code = "manual_checkin"`, `environment = model.EnvIVA`, `recorded_at_utc = time.Now().UTC().Format(time.RFC3339)`).
5. `app/app.go`:
   - Added bound `func (a *App) SubmitCheckin(req model.IngestRequest) (model.IngestResult, error)`.
6. `app/frontend/bindings/hrig/app/app.ts`:
   - Exported `IngestReading`, `IngestRequest`, `IngestResult`, and `SubmitCheckin(req)` using Wails method ID `$Call.ByID(3333120644, req)`.
7. `app/frontend/src/composables/useTelemetry.ts`:
   - Added and exported `DAILY_CHECKIN_METRICS`, `DailyCheckinMetricCode`, `isDailyMetricLoggedToday`, `getPendingDailyMetrics`, `hasPendingDailyCheckin`, and `submitCrewDailyMetric`, and included them in `useTelemetry()`.
8. `app/frontend/src/components/CrewCard.vue`:
   - Added a yellow circle indicator (`bg-warn`) next to the astronaut's name when `hasPendingDailyCheckin(astronaut.id)` is `true`, with an accessible title/aria-label showing the number of remaining daily metrics.
9. `app/frontend/src/views/dashboard-secundario.vue`:
   - Marked `sleep_hours`, `mood`, `hydration_intake`, and `body_mass` as self-reported (`isSelfReported: true`, `isDisconnected: false`) so they never render as `"Sensor Disconnected · No Signal"`.
   - Added the `"PENDING TODAY"` yellow-dot badge/subline when `!row.isLoggedToday` and a yellow-dot indicator next to the crew selector when `hasPendingDailyCheckin(resolvedAstronautId)` is `true`.
   - Added inline check-in controls inside the 4 daily cards (`sleep_hours`, `hydration_intake`, and `body_mass` numeric inputs + `Save` button; `mood` 1..5 interactive buttons).

## Verification

- `go -C /home/sacred/bd/HRIG/bot test ./...` -> `ok hrig/bot` (exit 0)
- `go -C /home/sacred/bd/HRIG/server test ./internal/realtime/... ./internal/http/...` -> `ok hrig/server/internal/realtime`, `ok hrig/server/internal/http` (exit 0)
- `CGO_ENABLED=0 go -C /home/sacred/bd/HRIG/app test ./...` (and `go -C /home/sacred/bd/HRIG/app test -tags gtk3 ./...`) -> `ok hrig/app` (exit 0)
- `bun run --cwd /home/sacred/bd/HRIG/app/frontend type-check` -> exit 0
- `bun run --cwd /home/sacred/bd/HRIG/app/frontend build` -> exit 0
- `rg -n "new WebSocket|fetch\(" /home/sacred/bd/HRIG/app/frontend/src/` -> 0 matches
