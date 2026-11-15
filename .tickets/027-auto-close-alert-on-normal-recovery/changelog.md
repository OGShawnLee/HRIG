# 027 — Changelog: Automatically Close Active Alert Screen When Vitals Return to Normal

## Files Modified

- `app/frontend/src/composables/useTelemetryNavigation.ts`
  - Preserved `isAlertRouteName(name: unknown): boolean` (`"support-alert"` or `"warning-alert"`).
  - Extracted `clearAlertIfResolvedWithRouter(router, remainingAlert, activeEpisodeKeys)` and wired both exported `clearAlertIfResolved`, `dispatchTelemetryStatusWithRouter` (`"normal"` branch), and `useTelemetryNavigation().clearAlertIfResolved` through it.
  - Inside `clearAlertIfResolvedWithRouter`:
    - When `!remainingAlert`: after resetting `lastStatus.value = "normal"`, `lastTriggeredAlert.value = null`, and `handledEpisodeKeys.clear()`, checks `resolveRouter(router)` and if `isAlertRouteName(targetRouter.currentRoute.value.name)`, calls `closeActiveAlertWithRouter(targetRouter)` to restore `previousNonAlertPath` (or `"home"`).
    - When `remainingAlert` exists and the active route is `"warning-alert"` with `remainingAlert.status === "abnormal"`, navigates via `void targetRouter.push({ name: "support-alert" })` without overwriting `previousNonAlertPath`.
  - Extracted `resetDemoStatusWithRouter(router)` and wired `useTelemetryNavigation().resetDemoStatus` through it: after resetting `lastStatus`, `lastTriggeredAlert`, and `handledEpisodeKeys`, checks `resolveRouter(router)` and calls `closeActiveAlertWithRouter(targetRouter)` when currently on `"support-alert"` or `"warning-alert"`.
- `.tickets/027-auto-close-alert-on-normal-recovery/task.md`
  - Updated status from `in_progress` to `completed`.
- `.tickets/027-auto-close-alert-on-normal-recovery/changelog.md`
  - Created this changelog.

## Verification Results

- `rg -n "isAlertRouteName|closeActiveAlertWithRouter" app/frontend/src/composables/useTelemetryNavigation.ts`: **PASS** (matched `isAlertRouteName` at line 61, router guard at line 81, `closeActiveAlertWithRouter` at line 141, `closeActiveAlert` at lines 159 & 281, `clearAlertIfResolvedWithRouter` at lines 227–228, and `resetDemoStatusWithRouter` at lines 259–260).
- `rg -n "new WebSocket|fetch\(" app/frontend/src/`: **PASS** (exit code `1`, 0 matches).
- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`).
- `bun run --cwd app/frontend build` (`vite build --mode production`): **PASS** (exit code `0`; 2429 modules transformed, built in 2.16s: `dist/index.html` 0.50 kB, `dist/assets/index-BAc3xwH2.css` 22.12 kB, `dist/assets/index-BYCtJPik.js` 314.48 kB).
- Scope check (`git status --short`): Only `app/frontend/src/composables/useTelemetryNavigation.ts` and `.tickets/027-auto-close-alert-on-normal-recovery/` modified/untracked; no git commit or push performed.
