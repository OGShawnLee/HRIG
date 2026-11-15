# 013 — Add Home buttons and cross-window navigation bar

| Field | Value |
| :--- | :--- |
| **ID** | `013` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Frontend Navigation Bar Builder |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `app/frontend/src/components/ViewNavBar.vue`
  - `app/frontend/src/App.vue`
  - `app/frontend/src/components/DemoAlertStrip.vue`
  - `app/frontend/src/views/HomeView.vue`
  - `app/frontend/src/views/dashboard-principal.vue`
  - `app/frontend/src/views/support-alert.vue`
  - `app/frontend/src/views/warning-alert.vue`
  - `.tickets/013-add-home-and-window-navigation/changelog.md`
- **Must not touch:**
  - `.routing/**` (including `INPUT.md`, `CONTEXT.md`, `DOMAIN.md`, `SPECIFICATION.md`, `DESIGN.md`, `PROGRESS.md`)
  - `.claude/**`
  - `docs/**`, `db/**`, `server/**`, `bot/**`, `shared/**`
  - `app/frontend/public/crew/**`
  - `app/frontend/src/components/WindowBar.vue`
  - `app/frontend/src/router/index.ts`
  - `app/frontend/src/composables/useTelemetryNavigation.ts`
  - `app/frontend/src/views/pagina-principal.vue`
  - `.tickets/013-add-home-and-window-navigation/task.md`
  - `.tickets/013-add-home-and-window-navigation/spec.md`

## Inputs

Read these files before acting:
1. [`app/frontend/src/App.vue`](../../app/frontend/src/App.vue) (shell layout with `WindowBar`, `RouterView`, and `DemoAlertStrip`)
2. [`app/frontend/src/router/index.ts`](../../app/frontend/src/router/index.ts) (route names: `home`, `pagina-principal`, `dashboard-principal`, `support-alert`, `warning-alert`)
3. [`app/frontend/src/composables/useTelemetryNavigation.ts`](../../app/frontend/src/composables/useTelemetryNavigation.ts) (`goToCrewOverview` navigates to `pagina-principal`)
4. [`app/frontend/src/components/DemoAlertStrip.vue`](../../app/frontend/src/components/DemoAlertStrip.vue) (existing dark-mode bar styles)
5. [`app/frontend/src/views/HomeView.vue`](../../app/frontend/src/views/HomeView.vue)
6. [`app/frontend/src/views/dashboard-principal.vue`](../../app/frontend/src/views/dashboard-principal.vue)
7. [`app/frontend/src/views/support-alert.vue`](../../app/frontend/src/views/support-alert.vue)
8. [`app/frontend/src/views/warning-alert.vue`](../../app/frontend/src/views/warning-alert.vue)
9. [`.routing/DESIGN.md`](../../.routing/DESIGN.md) (Sections 4 and 5: high contrast `>= 6:1`, zero horizontal scroll)

## Grounding & Anti-Hallucination Constraints

- `Home` navigation throughout this ticket targets `{ name: "pagina-principal" }` (`/pagina-principal`, the Crew Monitor main page). Keep `{ name: "home" }` (`/`) accessible via a `Bridge` link in `ViewNavBar.vue` so `HomeView.vue` remains reachable without breaking Wails bridge testing.
- Do **not** modify `WindowBar.vue`, `router/index.ts`, `useTelemetryNavigation.ts`, or `pagina-principal.vue`.
- Do **not** add WebSocket or `fetch()` calls.
- Keep `flex-wrap: wrap` and zero horizontal overflow down to `390px` viewport width.
- Do **not** commit or push any git changes.
- If `bun run --cwd app/frontend type-check` or `bun run --cwd app/frontend build` fails, **stop** and record the exact command and output in `.tickets/013-add-home-and-window-navigation/changelog.md`.

## Task

1. Create `app/frontend/src/components/ViewNavBar.vue`:
   - Render a `<nav class="view-nav-bar" aria-label="Window navigation">` styled with dark `#080b10` background, `1px solid #475569` bottom border, `#f5f5f5` text, `padding: 6px 12px`, `display: flex; flex-wrap: wrap; align-items: center; gap: 6px;`.
   - Include `<RouterLink>` buttons (`min-height: 32px`, `padding: 4px 10px`, `border: 1px solid #64748b`, active state `.is-active` when `route.name === targetName` highlighted with `border-color: #00a7ff; color: #38bdf8; background: #0f1b2d;`):
     1. `{ name: "pagina-principal" }` with inline `<svg viewBox="0 0 24 24" aria-hidden="true">` Home house icon + label `Home` (`aria-label="Home (Crew Monitor)"`).
     2. `{ name: "dashboard-principal" }` with label `Primary Vitals`.
     3. `{ name: "support-alert" }` with label `Support`.
     4. `{ name: "warning-alert" }` with label `Warning`.
     5. `{ name: "home" }` with label `Bridge` (placed at the end with `margin-left: auto` on wide viewports).
2. Update `app/frontend/src/App.vue` to import and render `<ViewNavBar class="shrink-0" />` directly below `<WindowBar />` and above `<div class="flex-1 flex flex-col"><RouterView /></div>`.
3. Update `app/frontend/src/views/HomeView.vue`:
   - Update the existing `<RouterLink :to="{ name: 'pagina-principal' }" class="button-side">` text to `Home (Crew Monitor)`.
4. Update `app/frontend/src/views/dashboard-principal.vue`, `app/frontend/src/views/support-alert.vue`, and `app/frontend/src/views/warning-alert.vue`:
   - Update each view's `<button type="button" class="back-btn" @click="goToCrewOverview">` so it renders an inline Home SVG icon (`viewBox="0 0 24 24"`, `width="14"`, `height="14"`, `aria-hidden="true"`) + `<span>Home</span>` (with `aria-label="Return to Home"`, `display: inline-flex; align-items: center; gap: 6px;`), replacing the `← Crew Overview` text.
5. Update `app/frontend/src/components/DemoAlertStrip.vue`:
   - Rename the `Crew Overview` button text to `Home` (keeping `@click="goToCrewOverview()"` and adding the same inline Home SVG icon `aria-hidden="true"`).
6. Run `bun run --cwd app/frontend type-check` and `bun run --cwd app/frontend build` from repository root and confirm both exit `0`.
7. Write `.tickets/013-add-home-and-window-navigation/changelog.md` listing files created/modified, exact commands run with exit codes, and any observations.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/013-add-home-and-window-navigation/changelog.md` listing:
- Files created and modified
- Verification commands executed and their exit codes/outputs
- Any blockers or assumptions avoided
