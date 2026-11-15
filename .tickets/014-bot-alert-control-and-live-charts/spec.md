# 014 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Bot Nominal Stability & Manual Control
- [x] `bot/main.go` sets `core_temp` nominal range inside `[36.55, 37.20]`, `suit_pressure` `[28.1, 29.9]`, and `wbc_count` `[4.6, 10.0]` (`db/002_seed.sql` nominal compliance).
- [x] `bot/main.go` supports manual anomaly injection and cancellation via both stdin commands (`abnormal`, `critical`, `cancel`/`normal`, `status`) and local HTTP endpoints (`GET /status`, `POST /override`, `POST /cancel` on `BOT_CONTROL_ADDR` default `127.0.0.1:8090`).
- [x] `cd bot && go test -v ./...` exits `0`.

## Non-Looping Alert Navigation & Persistence
- [x] `app/frontend/src/composables/useTelemetryNavigation.ts` auto-navigates once per alert episode/escalation, suppresses repeat redirects after `goToHome()` / `openPrimaryVitals()` for the same episode, keeps `lastTriggeredAlert`/`lastStatus` active while anomaly persists, and clears when resolved to normal.
- [x] `app/frontend/src/composables/useTelemetry.ts` tracks live `seriesByKey` points, exposes `loadMetricTrends` (calling `GetTrends`) and `getCrewSeries`, and resolves active alert state when telemetry returns to normal.
- [x] `app/frontend/src/components/DemoAlertStrip.vue` keeps active alert indicator visible and lets the user navigate to `/support-alert` or `/warning-alert` from the status badge when abnormal/critical.

## Live Data Charts in Primary Vitals
- [x] `app/frontend/src/views/dashboard-principal.vue` loads historical `TrendPoint`s via `loadMetricTrends` and renders dynamic SVG polyline/area paths + point markers derived from real incoming telemetry values (zero hardcoded mock waveform paths for live metrics) with status-driven UnoCSS styling (`ok`/`warn`/`crit`/`stale`).

## Scope & Build Checks
- [x] Protected files (`app/frontend/uno.config.ts`, `HomeView.vue`, `WindowBar.vue`, `EnvironmentCard.vue`, `MedicalManualButton.vue`, `server/`, `app/*.go`) have zero modifications.
- [x] Zero `new WebSocket` or `fetch(` calls exist in `app/frontend/src/`.
- [x] `bun run --cwd app/frontend type-check` exits `0`.
- [x] `bun run --cwd app/frontend build` exits `0` and outputs `app/frontend/dist/index.html`.
- [x] `.tickets/014-bot-alert-control-and-live-charts/changelog.md` exists and records files touched and exact command results.
