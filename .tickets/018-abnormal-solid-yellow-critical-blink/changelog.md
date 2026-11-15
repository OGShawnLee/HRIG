# 018 — Changelog: Make Abnormal Solid Yellow and Reserve Red Blinking Strictly for Critical

## Files Modified

- `app/frontend/src/views/dashboard-principal.vue`
  - Updated `statusTextClass(status)` to return `"text-crit vital-text-crit"` for `ReadingStatus.StatusCritical` and keep `"text-warn"` (no blink class) for `ReadingStatus.StatusAbnormal`.
  - Replaced `isAlertSeverity` with `cardSecondaryTextClass(status, nominalClass)` so secondary card text/icons return `"text-warn"` for `StatusAbnormal`, `"text-crit vital-text-crit"` for `StatusCritical`, and preserve `nominalClass` for `normal`, `stale`, `fallback`, and `unknown` states.
  - Bound `cardSecondaryTextClass` across all secondary card text and icon elements in `<article v-for="row in vitalRows">`: `row.subtitle` (`nominalClass: 'text-main'`), `row.rangeLabel` (`nominalClass: 'text-dull'`), the `"BP"`/`row.unit` header (`nominalClass: 'text-side'`), `row.subline` (`nominalClass: 'text-side'`), and the right-hand `HelpCircle` wrapper (`nominalClass: 'text-dull'`).
  - Removed `@keyframes vital-frame-blink-warn` and the `animation:` property from `.vital-frame-warn`, giving `abnormal` cards a solid non-blinking yellow contour (`border-width: 2px; border-color: #facc15; box-shadow: 0 0 0 1px rgba(250, 204, 21, 0.35), 0 0 10px rgba(250, 204, 21, 0.16);`).
  - Added `@keyframes vital-text-blink-crit` and `.vital-text-crit` (`0.65s ease-in-out infinite`) synchronized with `.vital-frame-crit` so only `critical` cards blink red on both contour and letters.
- `.tickets/018-abnormal-solid-yellow-critical-blink/changelog.md`
  - Created this changelog.

## Verification Results

- `rg -n "vital-frame-blink-warn" app/frontend/src/`: **PASS** (0 matches)
- `rg -n -A 5 "vital-frame-warn" app/frontend/src/views/dashboard-principal.vue`: **PASS** (`.vital-frame-warn` has no `animation:` property)
- `rg -n "vital-text-crit" app/frontend/src/views/dashboard-principal.vue`: **PASS** (matches only `ReadingStatus.StatusCritical` branches in `statusTextClass` and `cardSecondaryTextClass`, plus the `.vital-text-crit` CSS rule)
- `rg -n "new WebSocket|fetch\(" app/frontend/src/`: **PASS** (0 matches)
- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`)
- `bun run --cwd app/frontend build` (`vite build --mode production`): **PASS** (exit code `0`; `dist/index.html` 0.50 kB, `dist/assets/index-BQ5mXUqb.css` 21.01 kB, `dist/assets/index-DixfJ_0U.js` 310.10 kB)
- Scope check: Only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/018-abnormal-solid-yellow-critical-blink/changelog.md` were modified for ticket 018 (no git commit or push performed).
