# 016 — Changelog: Preserve Previous Route & Close Alert Return Action

## Files Modified

- `app/frontend/src/composables/useTelemetryNavigation.ts`
  - Added `previousNonAlertPath` (`ref<string | null>(null)`), `ALERT_ROUTE_NAMES`, `NON_ALERT_ROUTE_NAMES` (including `"home"`, `"search"`, and `"dashboard-principal"`), `isAlertRouteName`, `isNonAlertRouteName`, and `capturePreviousNonAlertRoute`.
  - Updated `registerTelemetryRouter(router)` to attach a `router.beforeEach` guard once per router (tracked via `guardedRouters` `WeakSet`) so navigating from `home`, `search`, or `dashboard-principal` to `support-alert` or `warning-alert` records `from.fullPath`.
  - Updated `openActiveAlertWithRouter` and `dispatchTelemetryStatusWithRouter` to call `capturePreviousNonAlertRoute(targetRouter)` before pushing an alert route.
  - Added `closeActiveAlertWithRouter` and exported `closeActiveAlert` (both at module level and from `useTelemetryNavigation()` alongside `previousNonAlertPath`), which calls `suppressCurrentEpisode()`, clears `previousNonAlertPath`, and navigates to the captured `returnPath` or `{ name: "home" }` fallback while leaving `selectedAstronautId` intact.
- `app/frontend/src/views/support-alert.vue`
  - Imported `X` from `lucide-vue-next` and destructured `closeActiveAlert` from `useTelemetryNavigation()`.
  - Added a `Close Alert` button (`button-side border-warn text-warn`) with `aria-label="Close alert and return to previous page"` next to the `Home` button.
  - Preserved the `[UNVERIFIED]` comment and `"immediately"` spelling.
- `app/frontend/src/views/warning-alert.vue`
  - Imported `X` from `lucide-vue-next` and destructured `closeActiveAlert` from `useTelemetryNavigation()`.
  - Added a `Close Alert` button (`button-side border-crit text-crit`) with `aria-label="Close alert and return to previous page"` next to the `Home` button.
- `.tickets/016-preserve-route-and-close-alert/changelog.md`
  - Created this changelog.

## Verification Results

- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`)
- `bun run --cwd app/frontend build` (`vite build --mode production`): **PASS** (exit code `0`; `dist/index.html` 0.50 kB, `dist/assets/index-Bcv0zglI.css` 20.81 kB, `dist/assets/index-DXNGsgDQ.js` 310.60 kB; `test -f app/frontend/dist/index.html` passed)
- `rg -n "inmediately" app/frontend/src/`: **PASS** (0 matches, exit code `1`)
- `rg -n "new WebSocket|fetch\(" app/frontend/src/`: **PASS** (0 matches, exit code `1`)
- `rg -n "\[UNVERIFIED\]|immediately" app/frontend/src/views/support-alert.vue`: **PASS** (both line 132 `[UNVERIFIED]` comment and line 135 `"immediately"` preserved)
