# 015 — Changelog

## Files Touched

- `app/frontend/src/views/dashboard-principal.vue`
  - Restored per-vital nominal stroke/fill palette via `vitalChartStyles` (`cardiac`: `stroke-ok`, `oxygenation`: `stroke-ok`, `pressure`: `stroke-main`, `radiation`: `stroke-warn` + `fill-warn/20`, `temperature`: `stroke-main`) with escalation to `stroke-warn`/`fill-warn/20` on `StatusAbnormal`, `stroke-crit`/`fill-crit/20` on `StatusCritical`, and `stroke-stale`/`fill-stale/15` on `StatusStale`.
  - Replaced uniform polylines with real-time, `getCrewSeries`-driven SVG `<path>` generators (`buildCardiacChart`, `buildOxygenationChart`, `buildPressureChart`, `buildRadiationChart`, `buildTemperatureChart`) and phase advancement via `computeSeriesPhase` (without `setInterval`/`setTimeout` mock timers), while preserving `loadMetricTrends` (`GetTrends`), live WebSocket series updates, and `rangeLabel` (`tabular-nums`).
  - Added `cardFrameClass` and border/box-shadow `@keyframes` (`vital-frame-blink-warn`, `vital-frame-blink-crit`) so `<article class="hud-card ...">` frames switch to `border-warn` and blink smoothly on `StatusAbnormal` (1.3s) and to `border-crit` and blink faster on `StatusCritical` (0.65s) without fading inner text or numeric readouts.
- `.tickets/015-restore-chart-styles-with-live-data/changelog.md`
  - Recorded changes and verification results.

## Verification Results

- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`): exited `0`.
- `bun run --cwd app/frontend build` (`vite build --mode production`): exited `0`, transformed `1797` modules, generated `app/frontend/dist/index.html` (508 bytes).
- `grep -rnE "new WebSocket|fetch\(" app/frontend/src/`: `0` matches (exit `1`).
- `grep -nE "setInterval|setTimeout" app/frontend/src/views/dashboard-principal.vue`: `0` matches (exit `1`).
- `git status -s`: only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/015-restore-chart-styles-with-live-data/` modified for this ticket (existing ticket `014` working-tree changes left untouched; no commits or pushes made).
