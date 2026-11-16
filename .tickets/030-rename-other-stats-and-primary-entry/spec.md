# 030 Spec — Rename to Other Stats & Primary Vitals Entry

## Acceptance Criteria

1. `ViewNavBar.vue` contains only `Home`, `Search`, and `Primary Vitals` (no `Secondary Stats` and no `Other Stats` link in the main navigation bar).
2. `router/index.ts` defines route `name: "other-stats"` (`path: "/other-stats"`) rendering `dashboard-secundario.vue`, plus redirect from `/secondary-stats`.
3. `useTelemetryNavigation.ts` includes `"other-stats"` in `NON_ALERT_ROUTE_NAMES` and navigates to `name: "other-stats"` with the selected `astronaut` query parameter.
4. `dashboard-principal.vue` renders an **`Other Stats`** button that opens `/other-stats?astronaut=<id>` for the currently active astronaut (`activeAstronaut ?? resolvedAstronautId`).
5. `dashboard-secundario.vue` displays **`OTHER STATS`** as its title, **`Primary Vitals`** return button for the active astronaut, and preserves all continuous + periodic cards and alert/stale behaviors.
6. Verification after final edit:
   - `bun run --cwd /home/sacred/bd/HRIG/app/frontend type-check` -> exit 0
   - `bun run --cwd /home/sacred/bd/HRIG/app/frontend build` -> exit 0
   - `rg -n "Secondary Stats|SECONDARY STATS|secondary-stats" /home/sacred/bd/HRIG/app/frontend/src/components/ViewNavBar.vue` -> 0 matches
   - `rg -n "Other Stats|OTHER STATS" /home/sacred/bd/HRIG/app/frontend/src/views/dashboard-principal.vue /home/sacred/bd/HRIG/app/frontend/src/views/dashboard-secundario.vue` -> matches in both files
   - `rg -n "new WebSocket|fetch\(" /home/sacred/bd/HRIG/app/frontend/src/` -> 0 matches
