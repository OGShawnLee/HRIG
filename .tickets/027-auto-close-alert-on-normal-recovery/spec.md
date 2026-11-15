# Spec — 027 Auto-Close Alert on Normal Recovery

## Functional Requirements

1. When `clearAlertIfResolved(null, ...)` runs (all active vital/environmental alerts resolved back to `normal`) and the router is currently displaying `"support-alert"` or `"warning-alert"`, the route closes automatically and navigates back to `previousNonAlertPath` (or `"home"`).
2. When `resetDemoStatus()` runs while displaying `"support-alert"` or `"warning-alert"`, it also closes back to `previousNonAlertPath` (or `"home"`).
3. When a `critical` alert downgrades to `abnormal` while viewing `"warning-alert"`, the route transitions cleanly to `"support-alert"` while keeping `previousNonAlertPath`.

## Acceptance Criteria

- `rg -n "isAlertRouteName|closeActiveAlertWithRouter" app/frontend/src/composables/useTelemetryNavigation.ts` confirms auto-close wiring inside `clearAlertIfResolved` and `resetDemoStatus`.
- `rg -n "new WebSocket|fetch\(" app/frontend/src/` returns zero matches.
- `bun run --cwd app/frontend type-check` exits `0`.
- `bun run --cwd app/frontend build` exits `0`.
