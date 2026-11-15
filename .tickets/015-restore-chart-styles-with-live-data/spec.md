# 015 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Chart Visuals & Real-Time Coupling
- [x] `app/frontend/src/views/dashboard-principal.vue` renders each vital card with its original nominal colors (`cardiac`: `stroke-ok`, `oxygenation`: `stroke-ok`, `pressure`: `stroke-main`, `radiation`: `stroke-warn` + `fill-warn/20`, `temperature`: `stroke-main`) and escalates on `abnormal`/`critical`/`stale`.
- [x] All 5 chart waveforms (`cardiac`, `oxygenation`, `pressure`, `radiation`, `temperature`) preserve their clinical shapes (ECG QRS spikes, plethysmograph wave, arterial wave, sawtooth + fill, thermal baseline trace) AND are computed dynamically from real `getCrewSeries` (`GetTrends` + live WebSocket readings) without `setInterval` mock animation.
- [x] Live `rangeLabel` (`tabular-nums`) remains displayed in each chart header.
- [x] Each `<article>` chart card frame switches from white to yellow (`border-warn`) and blinks on `abnormal`, and to red (`border-crit`) and blinks on `critical`, keeping numeric readouts solid and legible.

## Scope & Build Checks
- [x] Only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/015-restore-chart-styles-with-live-data/` are modified for this ticket.
- [x] Zero `new WebSocket` or `fetch(` calls exist in `app/frontend/src/`.
- [x] `bun run --cwd app/frontend type-check` exits `0`.
- [x] `bun run --cwd app/frontend build` exits `0` and outputs `app/frontend/dist/index.html`.
- [x] `.tickets/015-restore-chart-styles-with-live-data/changelog.md` exists and records files touched and exact command results.
