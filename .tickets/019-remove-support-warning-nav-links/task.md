# 019 — Remove Support & Warning Links from Top Navigation Bar

| Field | Value |
| :--- | :--- |
| **ID** | `019-remove-support-warning-nav-links` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/components/ViewNavBar.vue` |

## Context

`ViewNavBar.vue` renders manual `Support` and `Warning` navigation buttons alongside `Home`, `Search`, and `Primary Vitals`. Alert views should only appear when triggered by live or simulated alerts (via the bottom `DemoAlertStrip.vue` simulate controls or live telemetry), not as permanent tabs in the top navigation bar.

## Grounding Rules

- Modify only `app/frontend/src/components/ViewNavBar.vue` and `.tickets/019-remove-support-warning-nav-links/`.
- Keep `DemoAlertStrip.vue` (`Simulate Abnormal`, `Simulate Critical`, `Reset`, status pill) and the Vue Router `/support-alert` and `/warning-alert` routes intact.
- Remove unused icon imports in `ViewNavBar.vue` so TypeScript strict mode (`vue-tsc --noEmit`) stays clean.

## Steps

1. Edit `app/frontend/src/components/ViewNavBar.vue`:
   - Remove the `<RouterLink :to="{ name: 'support-alert' }">` (`Support`) and `<RouterLink :to="{ name: 'warning-alert' }">` (`Warning`) buttons.
   - Keep `Home`, `Search`, and `Primary Vitals`.
   - Remove unused `AlertCircle` and `AlertTriangle` imports from `lucide-vue-next`.
2. Run `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks on `ViewNavBar.vue` and `DemoAlertStrip.vue`, write `changelog.md`, and report back.
