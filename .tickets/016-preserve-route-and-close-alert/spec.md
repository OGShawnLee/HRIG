# 016 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Return Route Preservation & Close Action
- [x] `app/frontend/src/composables/useTelemetryNavigation.ts` records `previousNonAlertPath` (with full path + query params) before entering `support-alert` / `warning-alert` via both `registerTelemetryRouter` (`beforeEach`) and `dispatchTelemetryStatus` / `openActiveAlert`.
- [x] `app/frontend/src/composables/useTelemetryNavigation.ts` exports `closeActiveAlert`, which suppresses the active episode from re-triggering auto-navigation and navigates back to `previousNonAlertPath` (or `"home"` fallback).
- [x] `app/frontend/src/views/support-alert.vue` and `app/frontend/src/views/warning-alert.vue` include a `Close Alert` button (`X` icon) wired to `closeActiveAlert()` alongside the `Home` button.

## Scope & Build Checks
- [x] `app/frontend/src/views/support-alert.vue` preserves `[UNVERIFIED]` comment and `"immediately"` (`rg -n "inmediately" app/frontend/src/` returns empty).
- [x] Only `useTelemetryNavigation.ts`, `support-alert.vue`, `warning-alert.vue`, and `.tickets/016-preserve-route-and-close-alert/` are modified for this ticket.
- [x] Zero `new WebSocket` or `fetch(` calls exist in `app/frontend/src/`.
- [x] `bun run --cwd app/frontend type-check` exits `0`.
- [x] `bun run --cwd app/frontend build` exits `0` and outputs `app/frontend/dist/index.html`.
- [x] `.tickets/016-preserve-route-and-close-alert/changelog.md` exists and records files touched and exact command results.
