# 011 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Files Created & Composable Contract
- [x] `app/frontend/src/composables/useTelemetryNavigation.ts` exists and exports `ReadingStatus`, `AlertDispatchPayload`, `TELEMETRY_STATUS_EVENT`, `registerTelemetryRouter`, and `useTelemetryNavigation`.
- [x] `useTelemetryNavigation.ts` maps `"abnormal"` to route `"support-alert"` and `"critical"` to route `"warning-alert"`.
- [x] `app/frontend/src/components/DemoAlertStrip.vue` exists and renders `Crew Overview`, `Simulate Abnormal`, `Simulate Critical`, and `Reset` controls with icon + text status labels.

## Global Listener & View Navigation
- [x] `app/frontend/src/App.vue` registers the router via `registerTelemetryRouter`, attaches and cleans up the `TELEMETRY_STATUS_EVENT` window listener, and renders `<DemoAlertStrip />`.
- [x] `app/frontend/src/views/HomeView.vue` includes a `<RouterLink>` to `{ name: 'pagina-principal' }` (`Open Crew Monitor`).
- [x] `app/frontend/src/views/pagina-principal.vue` wraps both astronaut portraits in `<button type="button">` elements calling `openPrimaryVitals('sam')` and `openPrimaryVitals('patrick')`.
- [x] `app/frontend/src/views/dashboard-principal.vue` renders `{{ selectedAstronautName }}`, syncs `route.query.astronaut`, and includes a `← Crew Overview` button calling `goToCrewOverview`.
- [x] `app/frontend/src/views/support-alert.vue` includes a `← Crew Overview` button calling `goToCrewOverview`.
- [x] `app/frontend/src/views/warning-alert.vue` includes a `← Crew Overview` button calling `goToCrewOverview`.

## Anti-Hallucination & Scope Hygiene
- [x] Zero WebSocket (`new WebSocket`) or `fetch(` calls exist in `app/frontend/src/`.
- [x] `app/frontend/src/components/WindowBar.vue` and `app/frontend/src/router/index.ts` are unmodified by this ticket.
- [x] `git status --short` shows modifications/additions only in the files listed under `May create or edit` (plus pre-existing uncommitted `app/frontend/src/router/index.ts` and `app/frontend/src/views/support-alert.vue` / `.tickets/010-add-support-alert-view/` if uncommitted).

## Build & Type-Check
- [x] `bun run --cwd app/frontend type-check` exits `0`.
- [x] `bun run --cwd app/frontend build` exits `0` and outputs `app/frontend/dist/index.html`.
- [x] `.tickets/011-prepare-mockup-navigation-and-alert-routing/changelog.md` exists and records files touched and exact command results.
