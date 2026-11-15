# 016 — Preserve Previous Route & Close Alert Return Action

| Field | Value |
| :--- | :--- |
| **ID** | `016-preserve-route-and-close-alert` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/composables/useTelemetryNavigation.ts`, `app/frontend/src/views/support-alert.vue`, `app/frontend/src/views/warning-alert.vue` |

## Context

When an `abnormal` (or `critical`) alert opens `/support-alert` (or `/warning-alert`), the user loses track of the page they were viewing (for example `/dashboard-principal?astronaut=<id>`) because the alert views only offer a `Home` button (`/`). The user needs the previous page preserved so closing the alert returns them directly to where they were before.

## Grounding Rules

- Modify only `app/frontend/src/composables/useTelemetryNavigation.ts`, `app/frontend/src/views/support-alert.vue`, `app/frontend/src/views/warning-alert.vue`, and `.tickets/016-preserve-route-and-close-alert/`.
- Preserve `[UNVERIFIED]` comment and exact `"immediately"` spelling in `app/frontend/src/views/support-alert.vue`.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Update `app/frontend/src/composables/useTelemetryNavigation.ts`:
   - Track `previousNonAlertPath` (`ref<string | null>(null)`) capturing the exact `fullPath` of the last non-alert route (`home` or `dashboard-principal`, including query params such as `?astronaut=...`).
   - In `registerTelemetryRouter(router)`, attach a `router.beforeEach((to, from) => ...)` guard (registered once per router) so whenever navigating from a non-alert route (`home` or `dashboard-principal`) into an alert route (`support-alert` or `warning-alert`), `previousNonAlertPath` records `from.fullPath`. Also capture `targetRouter.currentRoute.value.fullPath` before pushing an alert route in `dispatchTelemetryStatusWithRouter` / `openActiveAlertWithRouter` when currently on a non-alert route.
   - Add `closeActiveAlert()`: suppresses repeat auto-navigation for the active episode (`suppressCurrentEpisode()`) so the alert does not immediately re-open, and pushes `previousNonAlertPath` if available (otherwise `{ name: "home" }`). Export `previousNonAlertPath` and `closeActiveAlert` from `useTelemetryNavigation()`.
   - Keep `selectedAstronautId` preserved across alert transitions.
2. Update `app/frontend/src/views/support-alert.vue` and `app/frontend/src/views/warning-alert.vue`:
   - Add a `Close Alert` button (`X` icon from `lucide-vue-next`, styled with `uno.config.ts` `button-side` / `border-warn text-warn` or `border-crit text-crit`) alongside the `Home` button in the top action bar calling `closeActiveAlert()`.
   - Ensure `aria-label` clearly indicates closing the alert and returning to the previous page, and keep `[UNVERIFIED]` + `"immediately"` untouched in `support-alert.vue`.
3. Run `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks, write `changelog.md`, and report back.
