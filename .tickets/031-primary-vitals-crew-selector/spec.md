# 031 Spec — Crew Member Selector in Primary Vitals

## Acceptance Criteria

1. `dashboard-principal.vue` renders `<select id="primary-crew-select">` in `<header>` right after `PRIMARY VITALS` when `crewPool.length > 0`.
2. Selecting a crew member updates `selectedAstronautId` (`setSelectedAstronaut`), updates the route query (`?astronaut=<id>`), updates `selectedAstronautName` in `<aside>`, reloads trends via the existing `watch(resolvedAstronautId, ...)`, and passes the newly selected astronaut when clicking `Other Stats`.
3. Verification commands pass after final edit:
   - `bun run --cwd /home/sacred/bd/HRIG/app/frontend type-check` -> exit 0
   - `bun run --cwd /home/sacred/bd/HRIG/app/frontend build` -> exit 0
   - `rg -n "primary-crew-select|onAstronautSelect" /home/sacred/bd/HRIG/app/frontend/src/views/dashboard-principal.vue` -> matches
   - `rg -n "new WebSocket|fetch\(" /home/sacred/bd/HRIG/app/frontend/src/` -> 0 matches
