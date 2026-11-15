# Spec — 019 Remove Support & Warning Links from Top Navigation Bar

## Functional Requirements

1. `app/frontend/src/components/ViewNavBar.vue` renders only `Home`, `Search`, and `Primary Vitals`.
2. `Support` and `Warning` `<RouterLink>` items are removed from `ViewNavBar.vue`.
3. `DemoAlertStrip.vue` keeps `Simulate Abnormal`, `Simulate Critical`, `Reset`, `Home`, and the status pill untouched.

## Acceptance Criteria

- `rg -n "support-alert|warning-alert|AlertCircle|AlertTriangle" app/frontend/src/components/ViewNavBar.vue` returns zero matches.
- `rg -n "Simulate Abnormal|Simulate Critical" app/frontend/src/components/DemoAlertStrip.vue` matches both simulate buttons.
- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`) exits `0`.
- `bun run --cwd app/frontend build` (`vite build --mode production`) succeeds cleanly.
