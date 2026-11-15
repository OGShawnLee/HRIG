# 013 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Navigation Bar & App Shell
- [x] `app/frontend/src/components/ViewNavBar.vue` exists and renders `<RouterLink>` items for `pagina-principal` (`Home` with inline `<svg` icon), `dashboard-principal` (`Primary Vitals`), `support-alert` (`Support`), `warning-alert` (`Warning`), and `home` (`Bridge`).
- [x] `app/frontend/src/App.vue` imports and renders `<ViewNavBar` below `<WindowBar />` and above `<RouterView />`.

## In-View Home Buttons
- [x] `app/frontend/src/views/HomeView.vue` links to `{ name: 'pagina-principal' }` with label `Home (Crew Monitor)`.
- [x] `app/frontend/src/views/dashboard-principal.vue` renders a `<button type="button" class="back-btn"` containing an inline `<svg` Home icon and `Home` calling `goToCrewOverview`.
- [x] `app/frontend/src/views/support-alert.vue` renders a `<button type="button" class="back-btn"` containing an inline `<svg` Home icon and `Home` calling `goToCrewOverview`.
- [x] `app/frontend/src/views/warning-alert.vue` renders a `<button type="button" class="back-btn"` containing an inline `<svg` Home icon and `Home` calling `goToCrewOverview`.
- [x] `app/frontend/src/components/DemoAlertStrip.vue` renders a `Home` button calling `goToCrewOverview()`.

## Build, Type-Check & Scope
- [x] `bun run --cwd app/frontend type-check` exits `0`.
- [x] `bun run --cwd app/frontend build` exits `0` and outputs `app/frontend/dist/index.html`.
- [x] `app/frontend/src/components/WindowBar.vue`, `app/frontend/src/router/index.ts`, `app/frontend/src/composables/useTelemetryNavigation.ts`, and `app/frontend/src/views/pagina-principal.vue` have zero new modifications from this ticket.
- [x] No unapproved git commits or pushes were executed.
- [x] `.tickets/013-add-home-and-window-navigation/changelog.md` exists and records files touched and exact command results.
