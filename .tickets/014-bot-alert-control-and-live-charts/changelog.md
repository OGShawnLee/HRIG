# 014 — Changelog

## Files Touched

- `bot/main.go`
- `bot/main_test.go` (new)
- `app/frontend/src/composables/useTelemetryNavigation.ts`
- `app/frontend/src/composables/useTelemetry.ts`
- `app/frontend/src/components/DemoAlertStrip.vue`
- `app/frontend/src/views/dashboard-principal.vue`
- `.tickets/014-bot-alert-control-and-live-charts/changelog.md` (new)

## Summary of Changes

1. **Bot Nominal Ranges & Manual Anomaly Control (`bot/main.go`, `bot/main_test.go`)**
   - Adjusted nominal ranges to comply with `db/002_seed.sql`: `core_temp` `[36.55, 37.20]`, `suit_pressure` `[28.1, 29.9]`, `wbc_count` `[4.6, 10.0]`, and clamped nominal numeric generation inside `[m.Min, m.Max]`.
   - Added mutex-protected override state (`normal`, `abnormal`, `critical`, `metric_code`, `astronaut`) with `anomalyTargets` aligned to `db/002_seed.sql` thresholds.
   - Added interactive stdin command loop (`abnormal [metric] [astronaut]`, `critical [metric] [astronaut]`, `cancel`/`normal`, `status`, `help`) and local HTTP control server (`BOT_CONTROL_ADDR`, default `127.0.0.1:8090`, disabled when `off`) exposing `GET /status`, `POST /override`, and `POST /cancel` (which resets out-of-bounds previous values and immediately triggers a fresh nominal batch).
   - Added `bot/main_test.go` verifying nominal threshold compliance, override/cancel transitions, HTTP endpoints, and stdin commands.

2. **Non-Looping Episode Navigation & Alert Resolution (`app/frontend/src/composables/useTelemetryNavigation.ts`, `app/frontend/src/composables/useTelemetry.ts`)**
   - Deduplicated auto-navigation in `useTelemetryNavigation.ts` using deterministic `${status}:${astronautId}:${metricCode}` episode keys (`buildAlertEpisodeKey`), navigating at most once per episode or severity escalation (`abnormal` -> `critical`), and suppressing repeat redirects after `goToHome()` / `openPrimaryVitals()` while keeping `lastStatus` and `lastTriggeredAlert` active.
   - Added `openActiveAlert()` and `clearAlertIfResolved()` to `useTelemetryNavigation.ts`.
   - Updated `useTelemetry.ts` with `reconcileActiveAlerts()` so normalized `"reading"`, `"alert_ack"`, and `"snapshot"` envelopes resolve active alerts back to `"normal"` once all live readings and active alerts normalize (without interfering with `mockup-demo` triggers).

3. **Clickable Alert Indicator (`app/frontend/src/components/DemoAlertStrip.vue`)**
   - Rendered the active status badge as an accessible `<button>` when `lastStatus` is `"abnormal"` or `"critical"` to jump back to `/support-alert` or `/warning-alert` via `openActiveAlert()`.

4. **Time-Series Buffer & Live Primary Vitals Charts (`app/frontend/src/composables/useTelemetry.ts`, `app/frontend/src/views/dashboard-principal.vue`)**
   - Added reactive `seriesByKey` (capped at 40 points per key), fed by `"snapshot"` and `"reading"` envelopes, plus `loadMetricTrends(astronautId, metricCode, limit)` (calling `GetTrends`) and `getCrewSeries(astronautId, metricCode)`.
   - Updated `dashboard-principal.vue` to load historical trends via `loadMetricTrends` when the selected astronaut changes (`heart_rate`, `spo2`, `bp_systolic`, `bp_diastolic`, `radiation_rate`, `core_temp`) and replaced all hardcoded static `wavePath` curves with dynamic SVG `<polyline>`, area `<path>`, latest-point `<circle>` markers, `tabular-nums` min–max range labels, and status-driven UnoCSS classes (`ok`/`warn`/`crit`/`stale`).

## Verification Results

- `cd bot && go test -v ./...` -> exit `0` (`TestNominalRangesComplyWithSeedThresholds`, `TestOverrideAndCancelTransitions`, `TestControlHTTPEndpointsAndStdinCommands` all `PASS`).
- `bun run --cwd app/frontend type-check` -> exit `0` (`vue-tsc --noEmit`).
- `bun run --cwd app/frontend build` -> exit `0` (`vite build --mode production`, generated `app/frontend/dist/index.html`, `dist/assets/index-GXH9aaTr.css`, `dist/assets/index-De7DYYnZ.js`).
- `grep -rn "new WebSocket\|fetch(" app/frontend/src/` -> exit `1` (0 matches).
- `git status --short` -> confirmed protected files (`app/frontend/uno.config.ts`, `HomeView.vue`, `WindowBar.vue`, `EnvironmentCard.vue`, `MedicalManualButton.vue`, `support-alert.vue`, `server/`, `app/*.go`) are unmodified.
