# 014 — Bot Anomaly Cancel/Control, Non-Looping Alert Navigation & Live Data Charts

| Field | Value |
| :--- | :--- |
| **ID** | `014-bot-alert-control-and-live-charts` |
| **Status** | `completed` |
| **Scope** | `bot/`, `app/frontend/src/` |

## Context

Two issues were observed during live testing of the telemetry stack (`PostgreSQL` -> `server/` -> `bot/` -> `app/frontend/`):
1. `bot/main.go` generated `core_temp` in `[36.4, 37.2]`, dipping below `db/002_seed.sql`'s `warning_min = 36.5` and continuously firing `abnormal` alerts without a manual way to cancel/restore nominal telemetry from the bot. Meanwhile, `useTelemetryNavigation.ts` repeatedly forced route transitions on every 2s reading batch instead of navigating once per alert episode while keeping the alert active until telemetry normalizes.
2. `app/frontend/src/views/dashboard-principal.vue` rendered hardcoded static SVG `wavePath` strings instead of plotting the actual `GetTrends` database history and live incoming `"reading"` stream.

## Grounding Rules

- Keep all `app/frontend/src/` data access strictly through `@bindings/hrig/app/app.js` and `@wailsio/runtime` (zero `fetch(` or `new WebSocket` calls in `app/frontend/src/`).
- Preserve `uno.config.ts`, `HomeView.vue`, `WindowBar.vue`, `EnvironmentCard.vue`, `MedicalManualButton.vue`, `server/`, and `app/*.go` untouched.
- Never invent clinical limits: keep `bot/` nominal bounds inside existing `db/002_seed.sql` thresholds (`core_temp` `[36.55, 37.20]`, `suit_pressure` `[28.1, 29.9]`, `wbc_count` `[4.6, 10.0]`), and keep `[UNVERIFIED]` copy in `support-alert.vue` untouched.

## Steps

1. Update `bot/main.go`:
   - Adjust nominal ranges (`core_temp` `36.55..37.20`, `suit_pressure` `28.1..29.9`, `wbc_count` `4.6..10.0`) so nominal simulation never triggers spurious `abnormal` readings against `db/002_seed.sql`.
   - Add thread-safe manual anomaly override state (`normal`, `abnormal`, `critical`, target `metric_code`, target `astronaut`) with:
     - Interactive stdin command loop (`abnormal [metric] [astronaut]`, `critical [metric] [astronaut]`, `cancel`/`normal`, `status`, `help`).
     - Local HTTP control listener (`BOT_CONTROL_ADDR`, default `127.0.0.1:8090`, disabled when `off`) exposing `GET /status`, `POST /override`, and `POST /cancel` (cancelling restores nominal values and dispatches a fresh nominal batch immediately).
   - Add `bot/main_test.go` covering nominal threshold compliance and override/cancel transitions.
2. Update `app/frontend/src/composables/useTelemetryNavigation.ts`:
   - Deduplicate alert episode navigation using a deterministic episode key (`${status}:${astronautId}:${metricCode}`), auto-navigating **at most once** per new episode or severity escalation (`abnormal` -> `critical`).
   - Suppress repeat redirects for the same episode when the user navigates back via `goToHome()` or `openPrimaryVitals()`, while keeping `lastTriggeredAlert` and `lastStatus` active as long as non-normal telemetry/unacknowledged alerts remain.
   - Expose `clearAlertIfResolved()` / normal resolution handling so `lastTriggeredAlert` and `lastStatus` return to `"normal"` when all live readings/alerts normalize.
3. Update `app/frontend/src/composables/useTelemetry.ts`:
   - Track a reactive time-series buffer (`seriesByKey`, capped at 40 points per key) fed by `"snapshot"` and `"reading"` envelopes, plus a `loadMetricTrends(astronautId, metricCode, limit)` helper calling `GetTrends` from `@bindings/hrig/app/app.js` and `getCrewSeries(astronautId, metricCode)`.
   - Re-evaluate active alert resolution when `"reading"` / `"alert_ack"` envelopes normalize so alerts disappear once telemetry returns to nominal, without fighting manual navigation while still abnormal/critical.
4. Update `app/frontend/src/components/DemoAlertStrip.vue`:
   - Make the active status pill clickable when `lastStatus` is `"abnormal"` or `"critical"` to jump back to `/support-alert` or `/warning-alert`.
5. Update `app/frontend/src/views/dashboard-principal.vue`:
   - Call `loadMetricTrends` when the selected astronaut changes for `heart_rate`, `spo2`, `bp_systolic`, `bp_diastolic`, `radiation_rate`, and `core_temp`.
   - Replace hardcoded static SVG `wavePath` curves with real data-driven SVG polyline + area fill + live latest-point dot computed dynamically from `getCrewSeries` (falling back gracefully only if zero telemetry points exist yet), showing min–max range labels (`tabular-nums`) and status-matched UnoCSS stroke/fill classes (`ok`, `warn`, `crit`, `stale`).
6. Run verification checks (`cd bot && go test -v ./...`, `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`), record results in `changelog.md`, and report back.
