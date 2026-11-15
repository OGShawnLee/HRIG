# 017 — Changelog: Highlight Graph Card Contour & Numeric Value While Alert Readings Persist

## Files Modified

- `app/frontend/src/views/dashboard-principal.vue`
  - Imported `AlertEvent` type from `@bindings/hrig/shared/model/models.js`, `alerts` and `resolveReadingStatus` from `useTelemetry()`, and `lastTriggeredAlert` from `useTelemetryNavigation()`.
  - Added `CardStatus`, `CARD_STATUS_SEVERITY`, `pickWorstCardStatus`, `normalizeAlertStatus`, `matchesAstronautScope`, `isAlertSupersededByNormalReading`, `isAlertSeverity`, and `resolveCardStatus` to compute the worst severity (`critical > abnormal > stale > normal > fallback`) per card across:
    - Primary metrics (`heart_rate`, `spo2`, `bp_systolic`/`bp_diastolic`, `radiation_rate`, `core_temp`),
    - Paired subline metrics (`hrv_rmssd` for `cardiac`, `cabin_o2` for `oxygenation`, `cabin_pressure` for `pressure`, `radiation_dose_cum` for `radiation`, `cabin_temp` for `temperature`),
    - Active unacknowledged alerts in `alerts` (not superseded by a newer normal reading) matching the selected astronaut, environmental metrics, or alerts without `astronautId`,
    - Active `lastTriggeredAlert` matching any of the card's metrics (including demo alerts without `astronautId`).
  - Bound the resolved per-card status (`hrCardStatus`, `spo2CardStatus`, `bpCardStatus`, `radCardStatus`, `tempCardStatus`) to `VitalRow.status`, `vitalChartStyles`, `cardFrameClass(row.status)`, and `statusTextClass(row.status)`.
  - Updated `cardFrameClass(status)` to return `"border-2 border-warn vital-frame-warn"` for `abnormal` and `"border-2 border-crit vital-frame-crit"` for `critical`.
  - Bound `statusTextClass(row.status)` directly on `<strong>{{ row.value }}</strong>` and applied `statusTextClass(row.status)` to the unit/BP header and subline when `isAlertSeverity(row.status)` is true while keeping normal/fallback styling untouched.
  - Added `border-width: 2px;` to `.vital-frame-warn` and `.vital-frame-crit` alongside their crisp yellow (`#facc15`) and red (`#f87171`) border/glow `@keyframes`.
- `.tickets/017-alert-card-contour-and-value-highlight/changelog.md`
  - Created this changelog.

## Verification Results

- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`)
- `bun run --cwd app/frontend build` (`vite build --mode production`): **PASS** (exit code `0`; `dist/index.html` 0.50 kB, `dist/assets/index-D3Mx12M8.css` 18.13 kB, `dist/assets/index-YUdlxo_m.js` 179.30 kB; verified `app/frontend/dist/index.html` exists)
- `rg -n "new WebSocket|fetch\(" app/frontend/src/`: **PASS** (0 matches, exit code `1`)
- Scope check (`git status --short` / `git diff`): Only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/017-alert-card-contour-and-value-highlight/changelog.md` were modified for ticket 017 (no git commit or push performed).
