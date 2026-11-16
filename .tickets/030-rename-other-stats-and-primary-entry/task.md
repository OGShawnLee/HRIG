# 030 — Rename Secondary Stats to "Other Stats" and Enter from Primary Vitals

| Field | Value |
| :--- | :--- |
| **ID** | `030-rename-other-stats-and-primary-entry` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/components/ViewNavBar.vue`, `app/frontend/src/views/dashboard-principal.vue`, `app/frontend/src/views/dashboard-secundario.vue`, `app/frontend/src/router/index.ts`, `app/frontend/src/composables/useTelemetryNavigation.ts` |

## Context

The user clarified that:
1. "Secondary Stats" must be renamed to **"Other Stats"**.
2. **"Other Stats" must NOT appear in the main top navigation bar** (`ViewNavBar.vue`), which keeps only `Home`, `Search`, and `Primary Vitals`.
3. Each astronaut's **Other Stats** page is entered by clicking an **`Other Stats`** button inside **Primary Vitals** (`dashboard-principal.vue`), passing the active astronaut (`activeAstronaut` / `resolvedAstronautId`).

## Grounding Rules

- Modify only the 5 scoped files above and `.tickets/030-rename-other-stats-and-primary-entry/`.
- Preserve all existing behavior in `dashboard-principal.vue` (Ticket `018` alerts, `021`/`022` help `?` button, `024`/`025` `rem` scaling, `026` anatomical silhouette alerts, `028` disconnected sensor / no-signal chart state) and in `dashboard-secundario.vue` (continuous vs periodic cards, crew selector, cadence filter, alerts, disconnected state, `?` Medical Manual links).
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Update `app/frontend/src/components/ViewNavBar.vue`:
   - Remove the `<RouterLink>` for `secondary-stats` / `Secondary Stats` so the main navigation bar contains only `Home`, `Search`, and `Primary Vitals`.
2. Update `app/frontend/src/router/index.ts`:
   - Change the secondary route to `{ path: "/other-stats", name: "other-stats", component: DashboardSecundario }` and add `{ path: "/secondary-stats", redirect: { name: "other-stats" } }`.
3. Update `app/frontend/src/composables/useTelemetryNavigation.ts`:
   - Replace `"secondary-stats"` with `"other-stats"` in `NON_ALERT_ROUTE_NAMES`.
   - Rename `openSecondaryStatsWithRouter` / `openSecondaryStats` to push `{ name: "other-stats", query: id ? { astronaut: id } : undefined }` (also export `openOtherStats` alongside alias `openSecondaryStats` if helpful, using `openOtherStats` cleanly).
4. Update `app/frontend/src/views/dashboard-principal.vue`:
   - Destructure `openOtherStats` from `useTelemetryNavigation()` and import `Activity` (or `BarChart2` / `SlidersHorizontal`) from `"lucide-vue-next"`.
   - Add an **`Other Stats`** button in the `<header>` right next to the search box (using `h-10 px-3.5 text-xs gap-1.5 shrink-0` `button-main` or `button-side`, e.g., `button-side` with clear high-contrast border/label `Other Stats`, `aria-label="Open Other Stats for selected astronaut"`) that invokes `openOtherStats(activeAstronaut ?? resolvedAstronautId)`.
5. Update `app/frontend/src/views/dashboard-secundario.vue`:
   - Update heading from `SECONDARY STATS` to `OTHER STATS` and `aria-label` from `Secondary Statistics Dashboard` to `Other Stats Dashboard`.
   - Add a `Primary Vitals` back button (`ArrowLeft` + `"Primary Vitals"`, calling `openPrimaryVitals(activeAstronaut ?? resolvedAstronautId)`) right next to `Home` in the header so users can jump straight back to that astronaut's Primary Vitals.
   - Update search placeholder/aria-label from `"secondary metrics"` to `"other stats metrics"` / `"Filter other metrics"`.
6. Run `bun run --cwd /home/sacred/bd/HRIG/app/frontend type-check`, `bun run --cwd /home/sacred/bd/HRIG/app/frontend build`, and `rg` checks, mark `task.md` as `completed`, and write `changelog.md`.
