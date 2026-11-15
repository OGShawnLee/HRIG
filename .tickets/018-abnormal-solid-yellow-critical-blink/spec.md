# Spec — 018 Abnormal Solid Yellow vs Critical Red Blinking

## Functional Requirements

1. **Abnormal (`ReadingStatus.StatusAbnormal`) styling in `app/frontend/src/views/dashboard-principal.vue`:**
   - Card contour uses `border-2 border-warn vital-frame-warn` with a solid yellow border (`#facc15`) and NO CSS animation/blinking.
   - All text labels and values inside the card (`row.title`, `row.subtitle`, `row.rangeLabel`, `"BP"` subtitle, `row.unit`, `row.value`, `row.subline`, and the help icon) use `text-warn` (yellow) and do NOT blink.
2. **Critical (`ReadingStatus.StatusCritical`) styling in `app/frontend/src/views/dashboard-principal.vue`:**
   - Card contour uses `border-2 border-crit vital-frame-crit` with a blinking red border (`@keyframes vital-frame-blink-crit`).
   - All text labels and values inside the card use `text-crit` AND `.vital-text-crit` with a red blinking animation (`@keyframes vital-text-blink-crit`), ensuring only `critical` cards blink red (both contour and letters).
3. **Normal & Stale isolation:**
   - Cards in `normal`, `stale`, or `fallback` states keep their existing non-blinking colors and unhighlighted `.hud-card` contour.

## Acceptance Criteria

- `rg -n "vital-frame-blink-warn" app/frontend/src/` returns no matches.
- `rg -n "vital-frame-warn" app/frontend/src/views/dashboard-principal.vue` has no `animation:` property.
- `rg -n "vital-text-crit" app/frontend/src/views/dashboard-principal.vue` confirms critical text blink styling exists and is applied exclusively when `row.status === ReadingStatus.StatusCritical`.
- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`) exits `0`.
- `bun run --cwd app/frontend build` (`vite build --mode production`) succeeds cleanly.
- Working frontend has zero `new WebSocket` or direct `fetch(` calls in `app/frontend/src/`.
