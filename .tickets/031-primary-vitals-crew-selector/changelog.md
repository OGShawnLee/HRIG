# 031 Changelog — Add Crew Member Selector to Primary Vitals Dashboard

## Summary

- Added `crewPool` computed property in `app/frontend/src/views/dashboard-principal.vue` (`activeAstronauts.value.length > 0 ? activeAstronauts.value : astronauts.value`) and reused it in `activeAstronaut`.
- Added `onAstronautSelect(event: Event)` handler in `app/frontend/src/views/dashboard-principal.vue` to update `selectedAstronautId` (`setSelectedAstronaut`) and sync the route query (`?astronaut=<id>`) via `router.replace`.
- Rendered `<select id="primary-crew-select">` inside `<header>` immediately after `<h1>PRIMARY VITALS</h1>` when `crewPool.length > 0`, matching the markup, styling, and accessibility of `dashboard-secundario.vue`.

## Verification

- `bun run --cwd /home/sacred/bd/HRIG/app/frontend type-check` (`vue-tsc --noEmit`): passed (`exit 0`).
- `bun run --cwd /home/sacred/bd/HRIG/app/frontend build` (`vite build --mode production`): passed (`exit 0`).
- `rg -n "primary-crew-select|onAstronautSelect" /home/sacred/bd/HRIG/app/frontend/src/views/dashboard-principal.vue`: matched lines 95, 1001, 1003, 1007.
- `rg -n "new WebSocket|fetch\(" /home/sacred/bd/HRIG/app/frontend/src/`: 0 matches.
