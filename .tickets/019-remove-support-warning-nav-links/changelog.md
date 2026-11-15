# 019 — Changelog: Remove Support & Warning Links from Top Navigation Bar

## Files Modified

- `app/frontend/src/components/ViewNavBar.vue`
  - Removed the `<RouterLink :to="{ name: 'support-alert' }">` (`Support`) and `<RouterLink :to="{ name: 'warning-alert' }">` (`Warning`) buttons from the `<nav>` bar, leaving only `Home`, `Search`, and `Primary Vitals`.
  - Removed the unused `AlertCircle` and `AlertTriangle` icon imports from `lucide-vue-next`.
- `.tickets/019-remove-support-warning-nav-links/changelog.md`
  - Created this changelog.

## Verification Results

- `rg -n "support-alert|warning-alert|AlertCircle|AlertTriangle" app/frontend/src/components/ViewNavBar.vue`: **PASS** (0 matches, exit code `1`).
- `rg -n "Simulate Abnormal|Simulate Critical" app/frontend/src/components/DemoAlertStrip.vue`: **PASS** (exit code `0`; matched line 83 `<span>Simulate Abnormal</span>` and line 97 `<span>Simulate Critical</span>`).
- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`).
- `bun run --cwd app/frontend build` (`vite build --mode production`): **Deferred** — skipped per orchestrator instructions to avoid concurrent `dist/` writes with ticket 018; to be executed in the combined build once both 018 and 019 finish.
- Scope check: Only `app/frontend/src/components/ViewNavBar.vue` and `.tickets/019-remove-support-warning-nav-links/changelog.md` were created/modified for ticket 019 (`DemoAlertStrip.vue` and router `/support-alert` / `/warning-alert` routes left intact; no git commit or push performed).
