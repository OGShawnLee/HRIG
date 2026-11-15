# 027 — Automatically Close Active Alert Screen When Vitals Return to Normal

| Field | Value |
| :--- | :--- |
| **ID** | `027-auto-close-alert-on-normal-recovery` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/composables/useTelemetryNavigation.ts` |

## Context

Currently, when telemetry readings return to nominal (`StatusNormal`), `reconcileActiveAlerts()` in `useTelemetry.ts` calls `clearAlertIfResolved(null, activeEpisodeKeys)`, which resets `lastStatus` to `"normal"` and clears `lastTriggeredAlert` / `handledEpisodeKeys`, **without** dismissing the active alert route (`/support-alert` or `/warning-alert`). As a result, a user on the alert screen remains on the alert screen even after vitals stabilize. Similarly, `resetDemoStatus()` resets status without closing an open alert screen.

## Grounding Rules

- Modify only `app/frontend/src/composables/useTelemetryNavigation.ts` and `.tickets/027-auto-close-alert-on-normal-recovery/`.
- Preserve `previousNonAlertPath` restoration (`closeActiveAlertWithRouter`), episode deduplication (`handledEpisodeKeys`), and alert-to-alert transitions cleanly.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Update `app/frontend/src/composables/useTelemetryNavigation.ts`:
   - Extract a small helper `isAlertRouteName(name: unknown): boolean` returning `true` for `"support-alert"` or `"warning-alert"`.
   - In `clearAlertIfResolved(remainingAlert, activeEpisodeKeys)`:
     - When `!remainingAlert`: after setting `lastStatus.value = "normal"`, `lastTriggeredAlert.value = null`, and `handledEpisodeKeys.clear()`, check the active router route; if currently on `"support-alert"` or `"warning-alert"`, call `closeActiveAlertWithRouter(targetRouter)` so the alert view automatically closes and returns the user to `previousNonAlertPath` (or `"home"`).
     - When `remainingAlert` exists (e.g., critical downgraded to abnormal while on `"warning-alert"`), if currently on `"warning-alert"` and `remainingAlert.status === "abnormal"`, navigate (`void targetRouter.push({ name: "support-alert" })`) without overwriting `previousNonAlertPath`.
   - In `resetDemoStatus()`: after resetting `lastStatus`, `lastTriggeredAlert`, and `handledEpisodeKeys`, if currently on `"support-alert"` or `"warning-alert"`, call `closeActiveAlertWithRouter(targetRouter)` as well.
2. Run `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks, write `changelog.md`, and report back.
