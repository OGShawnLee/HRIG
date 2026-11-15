# Changelog — 011 Prepare mockup navigation and global abnormal/critical alert routing

## Files Created

- `app/frontend/src/composables/useTelemetryNavigation.ts`
  - Exports `ReadingStatus` (`"normal" | "abnormal" | "critical" | "stale"`), `AlertDispatchPayload`, `CREW_MEMBERS`, `CrewMemberId`, `TELEMETRY_STATUS_EVENT` (`"hrig:telemetry-status"`), `registerTelemetryRouter`, `dispatchTelemetryStatus`, `handleIncomingStatus`, and `useTelemetryNavigation`.
  - Maintains shared module-level reactive state (`selectedAstronautId`, `selectedAstronautName`, `lastStatus`, `lastTriggeredAlert`) without Pinia.
  - Routes `"abnormal"` to `"support-alert"`, `"critical"` to `"warning-alert"`, clears `lastTriggeredAlert` on `"normal"` without forced navigation, and leaves the current view intact on `"stale"`.
- `app/frontend/src/components/DemoAlertStrip.vue`
  - Compact dark-mode bottom simulator and navigation strip (`#080b10` background, `1px solid #475569` border-top, `#f5f5f5` text, `flex-wrap: wrap`, `font-variant-numeric: tabular-nums`).
  - Displays icon + text status pill (`NOMINAL`, `CAUTION / ABNORMAL`, `CRITICAL`, `STALE`) and `Crew Overview`, `▲ Simulate Abnormal`, `⬣ Simulate Critical`, and `Reset` controls.
- `.tickets/011-prepare-mockup-navigation-and-alert-routing/changelog.md`
  - This change log.

## Files Modified

- `app/frontend/src/App.vue`
  - Registers the Vue Router instance via `registerTelemetryRouter(router)`.
  - Attaches and cleans up the global `window` listener for `TELEMETRY_STATUS_EVENT` (`hrig:telemetry-status`) validating `detail.status` before calling `dispatchTelemetryStatus`.
  - Renders `<DemoAlertStrip class="shrink-0 sticky bottom-0 z-40" />` at the bottom of the root flex layout.
- `app/frontend/src/views/HomeView.vue`
  - Imports `RouterLink` and adds `<RouterLink :to="{ name: 'pagina-principal' }" class="button-side">Open Crew Monitor</RouterLink>` alongside the existing `Test Go Bridge` button.
- `app/frontend/src/views/pagina-principal.vue`
  - Wraps both astronaut portraits (Sam Sonia and Patrick Cuckburn) in accessible `<button type="button" class="crew-portrait-btn">` elements invoking `openPrimaryVitals('sam')` and `openPrimaryVitals('patrick')` with `#00a7ff` `:focus-visible` outlines.
- `app/frontend/src/views/dashboard-principal.vue`
  - Watches `route.query.astronaut` immediately to sync `setSelectedAstronaut(...)`.
  - Binds the astronaut panel heading to `{{ selectedAstronautName }}`.
  - Adds a high-contrast `← Crew Overview` button (`min-height: 36px`) calling `goToCrewOverview`.
- `app/frontend/src/views/support-alert.vue`
  - Adds a top action bar with a `← Crew Overview` button (`min-height: 36px`) calling `goToCrewOverview`.
- `app/frontend/src/views/warning-alert.vue`
  - Adds a top action bar with a `← Crew Overview` button (`min-height: 36px`) calling `goToCrewOverview`.

## Verification Commands & Results

1. `bun run --cwd app/frontend type-check`
   - Exit code: `0`
   - Output: `$ vue-tsc --noEmit` (clean, 0 errors)
2. `bun run --cwd app/frontend build`
   - Exit code: `0`
   - Output:
     - `76 modules transformed.`
     - `dist/index.html                   0.50 kB │ gzip:  0.34 kB`
     - `dist/assets/index-CluLYOzH.css   22.86 kB │ gzip:  5.82 kB`
     - `dist/assets/index-n9rjQjSR.js   128.98 kB │ gzip: 47.55 kB`
     - `✓ built in 1.03s`
3. `rg -n "new WebSocket|fetch\(" app/frontend/src/ || echo "NO_MATCHES"`
   - Exit code: `0`
   - Output: `NO_MATCHES`
4. `ls -la app/frontend/dist/index.html`
   - Exit code: `0`
   - Output: `-rw-r--r-- 1 sacred sacred 508 nov 15 15:48 app/frontend/dist/index.html`
5. `git diff --stat app/frontend/src/components/WindowBar.vue`
   - Exit code: `0`
   - Output: empty (unmodified)
6. `git status --short`
   - Exit code: `0`
   - Only ticket-011 files modified/created (plus pre-existing uncommitted `.routing/PROGRESS.md`, `app/frontend/src/router/index.ts`, `.tickets/010-add-support-alert-view/`, and `app/frontend/src/views/support-alert.vue` from prior ticket state).

## Observations & Scope Hygiene

- No WebSocket connections, `fetch()` calls, Go/Wails backend files, `WindowBar.vue`, `router/index.ts`, `main.ts`, `uno.config.ts`, or `package.json` were modified or added by this ticket.
- Existing astronaut roster (`Sam Sonia`, `Patrick Cuckburn`) and visual designs of all mockup views were preserved.
- Note: `app/frontend/src/router/index.ts` and `.routing/PROGRESS.md` had pre-existing uncommitted modifications in the working tree before ticket 011 started and were left untouched.
