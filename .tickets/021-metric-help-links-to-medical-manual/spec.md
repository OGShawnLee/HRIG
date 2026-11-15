# Spec — 021 Metric Help Icons to Medical Manual

## Functional Requirements

1. Every vital card in `app/frontend/src/views/dashboard-principal.vue` exposes a keyboard-accessible `<button type="button">` around its `<HelpCircle>` (`?`) icon.
2. Clicking `?` navigates via Vue Router to `{ name: "search", query: { tab: "procedures", q: row.manualQuery } }` so the Medical Manual opens filtered to procedures related to that vital sign (`pressure` for Pressure, `heart_rate` for Cardiac Rhythm, `spo2` for Oxygenation, `radiation` for Radiation, `core_temp` for Temperature).
3. Existing card colors/blinking rules (`018`) remain completely intact.

## Acceptance Criteria

- `rg -n "openMedicalManual" app/frontend/src/views/dashboard-principal.vue` confirms the handler and button binding exist.
- `rg -n 'tab: "procedures"' app/frontend/src/views/dashboard-principal.vue` confirms navigation targets the Medical Manual procedures tab.
- `rg -n "new WebSocket|fetch\(" app/frontend/src/` returns zero matches.
- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`) exits `0`.
- `bun run --cwd app/frontend build` (`vite build --mode production`) succeeds cleanly.
