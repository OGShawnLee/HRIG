# 030 — Changelog: Rename Secondary Stats to "Other Stats" and Enter from Primary Vitals

## Files Modified & Created

- `app/frontend/src/components/ViewNavBar.vue`
  - Removed the `<RouterLink>` for `secondary-stats` / `Secondary Stats` so the top navigation bar keeps only `Home`, `Search`, and `Primary Vitals`.
- `app/frontend/src/router/index.ts`
  - Changed the secondary statistics route to `{ path: "/other-stats", name: "other-stats", component: DashboardSecundario }` and added `{ path: "/secondary-stats", redirect: { name: "other-stats" } }`.
- `app/frontend/src/composables/useTelemetryNavigation.ts`
  - Replaced `"secondary-stats"` with `"other-stats"` in `NON_ALERT_ROUTE_NAMES`.
  - Renamed `openSecondaryStatsWithRouter` to `openOtherStatsWithRouter` navigating to `{ name: "other-stats", query: id ? { astronaut: id } : undefined }`, exported `openOtherStats` (with `openSecondaryStats` kept as an alias), and exposed `openOtherStats` (plus `openSecondaryStats` alias) in `useTelemetryNavigation()`.
- `app/frontend/src/views/dashboard-principal.vue`
  - Imported `Activity` from `lucide-vue-next` and destructured `openOtherStats` from `useTelemetryNavigation()`.
  - Added the `Other Stats` button in the `<header>` next to the metric search input (`aria-label="Open Other Stats for selected astronaut"`, calling `openOtherStats(activeAstronaut ?? resolvedAstronautId)`).
- `app/frontend/src/views/dashboard-secundario.vue`
  - Updated `aria-label` to `"Other Stats Dashboard"` and heading from `SECONDARY STATS` to `OTHER STATS`.
  - Imported `ArrowLeft` from `lucide-vue-next`, destructured `openPrimaryVitals` from `useTelemetryNavigation()`, and added a `Primary Vitals` return button next to `Home` calling `openPrimaryVitals(activeAstronaut ?? resolvedAstronautId)`.
  - Updated `onAstronautSelect` `router.replace` route name to `"other-stats"`, updated the search input `aria-label` / `placeholder` to `"Filter other metrics"` / `"Search other stats metrics"`, and updated the empty-filter message to `"No other stats metrics match the current filter."`.
- `.tickets/030-rename-other-stats-and-primary-entry/task.md`
  - Updated status from `in_progress` to `completed`.
- `.tickets/030-rename-other-stats-and-primary-entry/changelog.md`
  - Created this changelog.

## Verification Results

- `bun run --cwd /home/sacred/bd/HRIG/app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`).
- `bun run --cwd /home/sacred/bd/HRIG/app/frontend build` (`vite build --mode production`): **PASS** (exit code `0`; 2432 modules transformed, built in 2.72s: `dist/index.html` 0.50 kB, `dist/assets/index-C7xLSfK4.css` 24.29 kB, `dist/assets/index-D6B120bT.js` 348.60 kB).
- `rg -n "Secondary Stats|SECONDARY STATS|secondary-stats" /home/sacred/bd/HRIG/app/frontend/src/components/ViewNavBar.vue`: **PASS** (0 matches, exit code `1`).
- `rg -n "Other Stats|OTHER STATS" /home/sacred/bd/HRIG/app/frontend/src/views/dashboard-principal.vue /home/sacred/bd/HRIG/app/frontend/src/views/dashboard-secundario.vue`: **PASS** (matches in both files, exit code `0`).
- `rg -n "new WebSocket|fetch\(" /home/sacred/bd/HRIG/app/frontend/src/`: **PASS** (0 matches, exit code `1`).
- Scope check: Only the 5 scoped frontend files and `.tickets/030-rename-other-stats-and-primary-entry/` were modified; no git commit or push performed.
