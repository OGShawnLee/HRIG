# 013 — Changelog

## Files Created
- `app/frontend/src/components/ViewNavBar.vue`: Added top cross-window `<nav class="view-nav-bar" aria-label="Window navigation">` with `flex-wrap: wrap`, `#080b10` background, `1px solid #475569` bottom border, `#f5f5f5` text, and `<RouterLink>` items for `pagina-principal` (`Home` with inline `<svg viewBox="0 0 24 24">` house icon and `aria-label="Home (Crew Monitor)"`), `dashboard-principal` (`Primary Vitals`), `support-alert` (`Support`), `warning-alert` (`Warning`), and `home` (`Bridge`, pushed to the right with `margin-left: auto` on viewports `>= 560px`). Active route links receive `.is-active` (`border-color: #00a7ff; color: #38bdf8; background: #0f1b2d;`).

## Files Modified
- `app/frontend/src/App.vue`: Imported `ViewNavBar` and rendered `<ViewNavBar class="shrink-0" />` directly below `<WindowBar />` and above `<div class="flex-1 flex flex-col"><RouterView /></div>`.
- `app/frontend/src/views/HomeView.vue`: Updated `<RouterLink :to="{ name: 'pagina-principal' }" class="button-side">` label to `Home (Crew Monitor)`.
- `app/frontend/src/views/dashboard-principal.vue`: Replaced `← Crew Overview` text inside `<button type="button" class="back-btn" @click="goToCrewOverview">` with `aria-label="Return to Home"`, inline Home SVG icon (`viewBox="0 0 24 24"`, `width="14"`, `height="14"`, `aria-hidden="true"`), and `<span>Home</span>`, plus `display: inline-flex; align-items: center; justify-content: center; gap: 6px;` on `.back-btn`.
- `app/frontend/src/views/support-alert.vue`: Replaced `← Crew Overview` text inside `<button type="button" class="back-btn" @click="goToCrewOverview">` with `aria-label="Return to Home"`, inline Home SVG icon (`viewBox="0 0 24 24"`, `width="14"`, `height="14"`, `aria-hidden="true"`), and `<span>Home</span>`, plus `display: inline-flex; align-items: center; gap: 6px;` on `.back-btn`.
- `app/frontend/src/views/warning-alert.vue`: Replaced `← Crew Overview` text inside `<button type="button" class="back-btn" @click="goToCrewOverview">` with `aria-label="Return to Home"`, inline Home SVG icon (`viewBox="0 0 24 24"`, `width="14"`, `height="14"`, `aria-hidden="true"`), and `<span>Home</span>`, plus `display: inline-flex; align-items: center; gap: 6px;` on `.back-btn`.
- `app/frontend/src/components/DemoAlertStrip.vue`: Updated `Crew Overview` button (calling `goToCrewOverview()`) to render the inline Home SVG icon (`aria-hidden="true"`) + `<span>Home</span>`, and added `display: inline-flex; align-items: center; justify-content: center; gap: 6px;` to `.strip-btn`.

## Verification Commands & Outputs

1. `cd /home/sacred/bd/HRIG && bun run --cwd app/frontend type-check`
   - Exit code: `0`
   - Output:
     ```text
     $ vue-tsc --noEmit
     ```

2. `cd /home/sacred/bd/HRIG && bun run --cwd app/frontend build`
   - Exit code: `0`
   - Output:
     ```text
     $ vite build --mode production
     vite v8.3.8 building client environment for production...
     transforming (10) src/views/warning-alert.vue
     transforming (28) src/composables/useTelemetryNavigation.ts
     transforming (29) bindings/hrig/app/app.ts
     transforming (30) src/components/WindowBar.vue
     ✓ 81 modules transformed.
     rendering chunks (1)...
     computing gzip size...
     dist/index.html                   0.50 kB │ gzip:  0.34 kB
     dist/assets/index-Cw5LOiqq.css   23.72 kB │ gzip:  5.82 kB
     dist/assets/index-eTpeVdpL.js   131.43 kB │ gzip: 48.00 kB
     ✓ built in 966ms
     ```

3. `cd /home/sacred/bd/HRIG && git status --short && ls -la app/frontend/dist/index.html`
   - Exit code: `0`
   - Verified `app/frontend/dist/index.html` exists (`508` bytes).
   - Verified `WindowBar.vue`, `router/index.ts`, `useTelemetryNavigation.ts`, and `pagina-principal.vue` received zero modifications from this ticket (note: `pagina-principal.vue` and `.routing/PROGRESS.md` carry prior working-tree changes from an earlier ticket).
   - No git commits or pushes were executed.

## Spec Checklist Verification
- [x] `app/frontend/src/components/ViewNavBar.vue` exists and renders `<RouterLink>` items for `pagina-principal` (`Home` with inline `<svg` icon), `dashboard-principal` (`Primary Vitals`), `support-alert` (`Support`), `warning-alert` (`Warning`), and `home` (`Bridge`). — **Yes**
- [x] `app/frontend/src/App.vue` imports and renders `<ViewNavBar` below `<WindowBar />` and above `<RouterView />`. — **Yes**
- [x] `app/frontend/src/views/HomeView.vue` links to `{ name: 'pagina-principal' }` with label `Home (Crew Monitor)`. — **Yes**
- [x] `app/frontend/src/views/dashboard-principal.vue` renders a `<button type="button" class="back-btn"` containing an inline `<svg` Home icon and `Home` calling `goToCrewOverview`. — **Yes**
- [x] `app/frontend/src/views/support-alert.vue` renders a `<button type="button" class="back-btn"` containing an inline `<svg` Home icon and `Home` calling `goToCrewOverview`. — **Yes**
- [x] `app/frontend/src/views/warning-alert.vue` renders a `<button type="button" class="back-btn"` containing an inline `<svg` Home icon and `Home` calling `goToCrewOverview`. — **Yes**
- [x] `app/frontend/src/components/DemoAlertStrip.vue` renders a `Home` button calling `goToCrewOverview()`. — **Yes**
- [x] `bun run --cwd app/frontend type-check` exits `0`. — **Yes**
- [x] `bun run --cwd app/frontend build` exits `0` and outputs `app/frontend/dist/index.html`. — **Yes**
- [x] `app/frontend/src/components/WindowBar.vue`, `app/frontend/src/router/index.ts`, `app/frontend/src/composables/useTelemetryNavigation.ts`, and `app/frontend/src/views/pagina-principal.vue` have zero new modifications from this ticket. — **Yes**
- [x] No unapproved git commits or pushes were executed. — **Yes**
- [x] `.tickets/013-add-home-and-window-navigation/changelog.md` exists and records files touched and exact command results. — **Yes**

## Blockers / Assumptions Avoided
- None. Preserved `{ name: "home" }` (`/`) as `Bridge` in `ViewNavBar.vue` while routing all `Home` buttons/links to `{ name: "pagina-principal" }`. No WebSocket or `fetch()` calls were added.
