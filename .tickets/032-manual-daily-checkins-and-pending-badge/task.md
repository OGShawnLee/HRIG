# 032 — Astronaut Manual Daily Check-Ins & Main Menu Pending Yellow Circle

| Field | Value |
| :--- | :--- |
| **ID** | `032-manual-daily-checkins-and-pending-badge` |
| **Status** | `completed` |
| **Scope** | `bot/main.go`, `db/002_seed.sql`, `server/internal/realtime/hub.go`, `app/client.go`, `app/app.go`, `app/frontend/bindings/hrig/app/app.ts`, `app/frontend/src/composables/useTelemetry.ts`, `app/frontend/src/components/CrewCard.vue`, `app/frontend/src/views/dashboard-secundario.vue` |

## Context

Daily self-reported statistics such as sleep hours (`sleep_hours`), mood (`mood`), fluid intake (`hydration_intake`), and body mass (`body_mass`) must be logged by the astronaut themselves rather than streamed every 2 seconds by the bot. When an astronaut still has unfilled daily check-in statistics today, a small yellow circle (`bg-warn`) must appear next to their name on the main menu (`CrewCard.vue` in `HomeView.vue`).

## Grounding Rules

- Modify only the 9 scoped files above and `.tickets/032-manual-daily-checkins-and-pending-badge/`.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.
- Do NOT stop running background tasks or run `git commit` / `git push`.

## Steps

1. **`bot/main.go`**:
   - Remove `sleep_hours`, `mood`, `hydration_intake`, and `body_mass` from `biomonitor_01`'s `Metrics` slice so the bot only streams continuous wearable vitals (`spo2`, `resp_rate`, `heart_rate`, `hrv_rmssd`, `core_temp`).
   - Also update `controlPanelHTML` device description if `sleep_hours` is listed in static text, while keeping `anomalyTargets` intact so existing override tests pass.
2. **`db/002_seed.sql` & `server/internal/realtime/hub.go`**:
   - In `db/002_seed.sql`, add `('manual_checkin', 'Crew manual daily check-in', 'manual')` to `INSERT INTO device (code, name, kind)`.
   - In `server/internal/realtime/hub.go`, add `const dailyCheckinStaleAfter = 24 * time.Hour` and `func isDailyCheckinReading(metricCode, deviceCode, source string) bool` (returning `true` when `deviceCode == "manual_checkin"` or `source == "manual"` or `metricCode` is `"sleep_hours"`, `"mood"`, `"hydration_intake"`, or `"body_mass"`). Use `dailyCheckinStaleAfter` (`24h`) for those readings in `HydrateFromDB` and `SweepStale()` so astronaut-logged daily check-ins do not turn `StatusStale` after 15 seconds.
3. **`app/client.go`, `app/app.go`, & `app/frontend/bindings/hrig/app/app.ts`**:
   - In `app/client.go`, add `func (c *apiClient) SubmitCheckin(ctx context.Context, req model.IngestRequest) (model.IngestResult, error)` posting `req` to `"/"` via `c.doJSON(ctx, http.MethodPost, "/", req, &out)` (defaulting `req.DeviceCode` to `"manual_checkin"`, `req.Environment` to `model.EnvIVA`, and `req.RecordedAtUTC` to `time.Now().UTC().Format(time.RFC3339)` when empty).
   - In `app/app.go`, add `func (a *App) SubmitCheckin(req model.IngestRequest) (model.IngestResult, error)` calling `client.SubmitCheckin(ctx, req)`.
   - Run `wails3 generate bindings -d app/frontend/bindings` from `/home/sacred/bd/HRIG/app` (or update `app/frontend/bindings/hrig/app/app.ts` with `SubmitCheckin` using its generated `$Call.ByID(...)` ID and `model$0.IngestResult.createFrom`).
4. **`app/frontend/src/composables/useTelemetry.ts`**:
   - Import `SubmitCheckin` from `@bindings/hrig/app/app.js` and `EnvironmentMode`, `IngestReading`, `IngestRequest` from `@bindings/hrig/shared/model/models.js`.
   - Export `DAILY_CHECKIN_METRICS = ["sleep_hours", "mood", "hydration_intake", "body_mass"] as const` and `type DailyCheckinMetricCode = (typeof DAILY_CHECKIN_METRICS)[number]`.
   - Export `isDailyMetricLoggedToday(astronautId: string, metricCode: string): boolean`:
     - Returns `true` only if `getCrewReading(astronautId, metricCode)` exists, has `reading.deviceCode === "manual_checkin"` or `reading.source === "manual"`, and its `recordedAtUtc` falls on today's local date (or within the last 20 hours).
   - Export `getPendingDailyMetrics(astronautId: string): DailyCheckinMetricCode[]` and `hasPendingDailyCheckin(astronautId: string): boolean`.
   - Export `submitCrewDailyMetric(astronaut: { id: string; fullName: string }, metricCode: DailyCheckinMetricCode, numericOrScaleValue: number): Promise<void>`:
     - Constructs an `IngestRequest` (`deviceCode: "manual_checkin"`, `environment: EnvironmentMode.EnvIVA`, `recordedAtUtc: new Date().toISOString()`, `astronautName: astronaut.fullName`, `readings: [...]`) with appropriate `unitCode` (`"h"`, `"{score}"`, `"mL"`, `"kg"`) and `valueNumeric` (or `valueScale` for `"mood"`), calls `await SubmitCheckin(req)`, and refreshes telemetry + trends (`await Promise.all([refresh(), loadMetricTrends(astronaut.id, metricCode, 40)])`).
   - Include `DAILY_CHECKIN_METRICS`, `isDailyMetricLoggedToday`, `getPendingDailyMetrics`, `hasPendingDailyCheckin`, and `submitCrewDailyMetric` in `useTelemetry()`.
5. **`app/frontend/src/components/CrewCard.vue`**:
   - Destructure `hasPendingDailyCheckin` and `getPendingDailyMetrics` from `useTelemetry()`.
   - Compute `hasPendingCheckin` and `pendingCheckinLabel` (e.g. `"Pending daily check-in (N remaining)"`).
   - Next to the astronaut's name (`<h2>`) in `CrewCard.vue`, render a small yellow circle (`<span v-if="hasPendingCheckin" class="inline-block w-2.5 h-2.5 rounded-full bg-warn ring-2 ring-warn/30 shrink-0" :title="pendingCheckinLabel" :aria-label="pendingCheckinLabel" />`) right beside the astronaut's name.
6. **`app/frontend/src/views/dashboard-secundario.vue`**:
   - For the 4 astronaut-logged daily metrics (`sleep_hours`, `mood`, `hydration_intake`, `body_mass`):
     - Do NOT mark them as `isDisconnected` ("Sensor Disconnected · No Signal") when `isStale` or when only older seed data exists; instead track `isSelfReported: true` and `isLoggedToday = isDailyMetricLoggedToday(id, metricCode)`.
     - When `!row.isLoggedToday`, display a small yellow circle (`w-2 h-2 rounded-full bg-warn`) + `"PENDING TODAY"` in the card header/subline.
     - Render clean inline logging controls inside each of the 4 self-reported cards so the astronaut can enter and save their value right there:
       - `sleep_hours`: number input (`min="0" max="16" step="0.5"`) + `Save` button.
       - `mood`: 5 clickable scale buttons (`1`..`5`) that log the selected mood score on click.
       - `hydration_intake`: number input (`min="0" max="6000" step="100"`) + `Save` button.
       - `body_mass`: number input (`min="35" max="140" step="0.1"`) + `Save` button.
     - Also show the small yellow circle indicator next to the crew selector in `dashboard-secundario.vue` when `hasPendingDailyCheckin(resolvedAstronautId)` is true.
7. Run verification checks (`go -C bot test ./...`, `go -C server test ./...`, `go -C app test ./...`, `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks), mark `task.md` `completed`, and write `changelog.md`.
