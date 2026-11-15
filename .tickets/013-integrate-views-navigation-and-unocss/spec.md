# 013 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Assets, Composable & Live Telemetry Integration
- [x] `app/frontend/public/crew/sam-sonia.jpg` and `app/frontend/public/crew/patrick-cuckburn.jpg` exist and are copied to `app/frontend/dist/crew/` on build.
- [x] `app/frontend/src/composables/useTelemetryNavigation.ts` exists and exports `dispatchTelemetryStatus`, `registerTelemetryRouter`, `TELEMETRY_STATUS_EVENT`, and `useTelemetryNavigation` (`goToHome` -> `"home"`, `"abnormal"` -> `"support-alert"`, `"critical"` -> `"warning-alert"`).
- [x] `app/frontend/src/composables/useTelemetry.ts` calls `dispatchTelemetryStatus` when live `"reading"` or unacknowledged `"alert"` envelopes arrive with `abnormal` or `critical` status.
- [x] `app/frontend/src/components/CrewCard.vue` is keyboard/click interactive (`role="button"`, `tabindex="0"`, `@click`, `@keydown.enter.prevent`, `@keydown.space.prevent`) and calls `openPrimaryVitals(astronaut)`.

## UnoCSS Theme Alignment & Routing
- [x] `app/frontend/src/components/ViewNavBar.vue` and `app/frontend/src/components/DemoAlertStrip.vue` exist, use `uno.config.ts` tokens (`bg-ground`, `bg-surface`, `text-main`, `border-lineDim`, `text-warn`, `text-crit`, `text-ok`, `hud-pill`, `hud-value`, `button-side`), and render in `app/frontend/src/App.vue`.
- [x] `app/frontend/src/router/index.ts` registers `/`, `/dashboard-principal`, `/support-alert`, `/warning-alert`, and redirects `/pagina-principal` to `home`.
- [x] `app/frontend/src/views/dashboard-principal.vue`, `support-alert.vue`, and `warning-alert.vue` use `uno.config.ts` theme classes (`bg-ground`, `hud-card`, `hud-value`, `text-warn`/`border-warn`, `text-crit`/`border-crit`, `button-side`) and include a `Home` button calling `goToHome`.
- [x] `app/frontend/src/views/support-alert.vue` keeps `[UNVERIFIED]` comment and `"immediately"` spelled correctly (`rg -n "inmediately"` returns empty).

## Scope & Build Checks
- [x] `app/frontend/src/views/HomeView.vue`, `app/frontend/uno.config.ts`, `app/frontend/src/components/WindowBar.vue`, `app/frontend/src/components/EnvironmentCard.vue`, `app/frontend/src/components/MedicalManualButton.vue`, and `server/` have zero modifications.
- [x] Zero `new WebSocket` or `fetch(` calls exist in `app/frontend/src/`.
- [x] `bun run --cwd app/frontend type-check` exits `0`.
- [x] `bun run --cwd app/frontend build` exits `0` and outputs `app/frontend/dist/index.html`.
- [x] Go backend files (`app/app.go`, `app/client.go`, `app/client_test.go`, `app/main.go`) remain unmodified (`git diff -- app/*.go` is empty).
- [x] `.tickets/013-integrate-views-navigation-and-unocss/changelog.md` exists and records files touched and exact command results.
