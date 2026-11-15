# 011 — Prepare navigational mockup and global abnormal/critical alert routing

| Field | Value |
| :--- | :--- |
| **ID** | `011` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Frontend Navigation & Alert Routing Builder |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `app/frontend/src/composables/useTelemetryNavigation.ts`
  - `app/frontend/src/components/DemoAlertStrip.vue`
  - `app/frontend/src/App.vue`
  - `app/frontend/src/views/HomeView.vue`
  - `app/frontend/src/views/pagina-principal.vue`
  - `app/frontend/src/views/dashboard-principal.vue`
  - `app/frontend/src/views/support-alert.vue`
  - `app/frontend/src/views/warning-alert.vue`
  - `.tickets/011-prepare-mockup-navigation-and-alert-routing/changelog.md`
- **Must not touch:**
  - `.routing/**` (including `INPUT.md`, `CONTEXT.md`, `DOMAIN.md`, `SPECIFICATION.md`, `DESIGN.md`, `PROGRESS.md`)
  - `.claude/**`
  - `docs/**`, `db/**`, `server/**`, `bot/**`, `shared/**`
  - `app/main.go`, `app/app.go`, `app/app_window.go`, `app/frontend/src/components/WindowBar.vue`
  - `app/frontend/src/router/index.ts`, `app/frontend/src/main.ts`, `app/frontend/uno.config.ts`, `app/frontend/package.json`
  - `.tickets/011-prepare-mockup-navigation-and-alert-routing/task.md`
  - `.tickets/011-prepare-mockup-navigation-and-alert-routing/spec.md`

## Inputs

Read these files before acting:
1. [`shared/model/telemetry.go`](../../shared/model/telemetry.go) (Lines 5–15: `ReadingStatus` values `"normal"`, `"abnormal"`, `"critical"`, `"stale"`; lines 200–305: `TelemetryFrame`, `AlertEvent`, `RealtimeEnvelope`)
2. [`app/frontend/src/router/index.ts`](../../app/frontend/src/router/index.ts) (Route names: `home`, `warning-alert`, `support-alert`, `pagina-principal`, `dashboard-principal`)
3. [`app/frontend/src/App.vue`](../../app/frontend/src/App.vue) (Root layout shell)
4. [`app/frontend/src/views/HomeView.vue`](../../app/frontend/src/views/HomeView.vue) (Root `/` entry view)
5. [`app/frontend/src/views/pagina-principal.vue`](../../app/frontend/src/views/pagina-principal.vue) (Crew overview with `Sam Sonia` and `Patrick Cuckburn` portraits)
6. [`app/frontend/src/views/dashboard-principal.vue`](../../app/frontend/src/views/dashboard-principal.vue) (`PRIMARY VITALS` view with `Sam Sonia` panel)
7. [`app/frontend/src/views/support-alert.vue`](../../app/frontend/src/views/support-alert.vue) (`SUPPORT` / `CAUTION` view)
8. [`app/frontend/src/views/warning-alert.vue`](../../app/frontend/src/views/warning-alert.vue) (`WARNING` critical view)
9. [`.routing/DESIGN.md`](../../.routing/DESIGN.md) (Sections 4 and 5: high contrast, color + icon + text redundancy, zero horizontal scroll)

## Grounding & Anti-Hallucination Constraints

- Do **not** open WebSocket connections, call `fetch()`, or modify `server/`. The backend realtime feed is being developed separately; this ticket provides the frontend navigation layer and typed entrypoints (`dispatchTelemetryStatus`, `handleIncomingStatus`, and the `hrig:telemetry-status` window `CustomEvent`) ready for future wiring.
- Align `ReadingStatus` strictly with `shared/model/telemetry.go`: `"normal" | "abnormal" | "critical" | "stale"`.
- Keep the existing astronaut roster in `pagina-principal.vue` (`Sam Sonia` and `Patrick Cuckburn`). Do **not** invent extra crew members or vital readings.
- Preserve the existing visual design of `pagina-principal.vue`, `dashboard-principal.vue`, `support-alert.vue`, and `warning-alert.vue`.
- Do **not** modify `WindowBar.vue` or Wails Go bindings.
- Do **not** commit or push any git changes.
- If `bun run --cwd app/frontend type-check` or `bun run --cwd app/frontend build` fails, **stop**, do not guess workarounds outside scope, and record the exact command and output in `.tickets/011-prepare-mockup-navigation-and-alert-routing/changelog.md`.

## Task

1. Create `app/frontend/src/composables/useTelemetryNavigation.ts`:
   - Export `type ReadingStatus = "normal" | "abnormal" | "critical" | "stale"`.
   - Export `interface AlertDispatchPayload { status: ReadingStatus; astronautId?: string; metricCode?: string; source?: string }`.
   - Export `CREW_MEMBERS` constant mapping `"sam"` -> `"Sam Sonia"` and `"patrick"` -> `"Patrick Cuckburn"`.
   - Maintain module-level reactive refs so state is shared across all views without adding Pinia:
     - `selectedAstronautId = ref<"sam" | "patrick">("sam")`
     - `selectedAstronautName = computed(() => CREW_MEMBERS[selectedAstronautId.value])`
     - `lastStatus = ref<ReadingStatus>("normal")`
     - `lastTriggeredAlert = ref<AlertDispatchPayload | null>(null)`
   - Inside `useTelemetryNavigation()`:
     - Obtain `router = useRouter()` (only when called inside Vue setup/component context; also store a module-level `registeredRouter: Router | null = null` or pass `router` via `registerTelemetryRouter(router: Router)` called from `App.vue` so `dispatchTelemetryStatus` works both from components and from global window listeners).
     - `setSelectedAstronaut(id: string)`: updates `selectedAstronautId.value` when `id === "sam" || id === "patrick"`.
     - `openPrimaryVitals(astronautId: "sam" | "patrick")`: sets `selectedAstronautId.value = astronautId` and navigates to `{ name: "dashboard-principal", query: { astronaut: astronautId } }`.
     - `goToCrewOverview()`: navigates to `{ name: "pagina-principal" }`.
     - `dispatchTelemetryStatus(payload: AlertDispatchPayload)`:
       - Sets `lastStatus.value = payload.status`.
       - If `payload.status === "critical"`: sets `lastTriggeredAlert.value = payload` and navigates to `{ name: "warning-alert" }` (unless `router.currentRoute.value.name === "warning-alert"`).
       - Else if `payload.status === "abnormal"`: sets `lastTriggeredAlert.value = payload` and navigates to `{ name: "support-alert" }` (unless `router.currentRoute.value.name === "support-alert"`).
       - Else (`"normal"` or `"stale"`): if `payload.status === "normal"`, clears `lastTriggeredAlert.value = null`; does not force navigation away from the current view.
     - `handleIncomingStatus(status: ReadingStatus, context?: Omit<AlertDispatchPayload, "status">)`: calls `dispatchTelemetryStatus({ status, ...context })`.
     - `resetDemoStatus()`: sets `lastStatus.value = "normal"` and `lastTriggeredAlert.value = null`.
   - Export `TELEMETRY_STATUS_EVENT = "hrig:telemetry-status"` for global `CustomEvent<AlertDispatchPayload>` dispatching.
2. Create `app/frontend/src/components/DemoAlertStrip.vue`:
   - Render a compact, discreet dark-mode navigation and demo trigger bar (`role="region"`, `aria-label="Mockup navigation and alert simulator"`), styled with `#080b10` background, `1px solid #475569` top border, `#f5f5f5` text (`>= 6:1` contrast), `flex-wrap: wrap` (zero horizontal scroll down to `390px` width), and `font-variant-numeric: tabular-nums`.
   - Include:
     - Current status indicator pill showing icon + text (`NOMINAL`, `CAUTION / ABNORMAL`, `CRITICAL`, or `STALE`).
     - Button `Crew Overview` (`type="button"`) calling `goToCrewOverview()`.
     - Button `▲ Simulate Abnormal` (`type="button"`, yellow/amber `#f5c518` border & text) calling `dispatchTelemetryStatus({ status: "abnormal", metricCode: "bp_systolic", source: "mockup-demo" })`.
     - Button `⬣ Simulate Critical` (`type="button"`, red `#ff4d4d` border & text) calling `dispatchTelemetryStatus({ status: "critical", metricCode: "cabin_pressure", source: "mockup-demo" })`.
     - Button `Reset` (`type="button"`) calling `resetDemoStatus()`.
3. Update `app/frontend/src/App.vue`:
   - Call `useRouter()` and register it via `registerTelemetryRouter(router)`.
   - Register a global `window.addEventListener(TELEMETRY_STATUS_EVENT, onTelemetryStatusEvent)` on `onMounted` and remove it on `onBeforeUnmount` so any `CustomEvent<AlertDispatchPayload>` with `detail.status` routes immediately from any view via `dispatchTelemetryStatus`.
   - Render `<DemoAlertStrip />` at the bottom of the root flex layout (`shrink-0 sticky bottom-0 z-40`).
4. Update `app/frontend/src/views/HomeView.vue`:
   - Add a `<RouterLink :to="{ name: 'pagina-principal' }" class="button-side">Open Crew Monitor</RouterLink>` alongside the existing `Test Go Bridge` button so users can enter the mockup flow directly from `/`.
5. Update `app/frontend/src/views/pagina-principal.vue`:
   - Add `<script setup lang="ts">` importing `useTelemetryNavigation`.
   - Wrap each astronaut portrait (`.crew-portrait`) in an accessible `<button type="button" class="crew-portrait-btn" ...>` with `aria-label="Open Sam Sonia primary vitals"` on Sam's portrait (`@click="openPrimaryVitals('sam')"`) and `aria-label="Open Patrick Cuckburn primary vitals"` on Patrick's portrait (`@click="openPrimaryVitals('patrick')"`). Style `.crew-portrait-btn` with `padding: 0; border: 0; background: transparent; cursor: pointer;` plus a visible `:focus-visible` outline (`2px solid #00a7ff`) so the portrait looks identical while being clickable and keyboard-accessible.
6. Update `app/frontend/src/views/dashboard-principal.vue`:
   - Add `<script setup lang="ts">` importing `useRoute`, `watch`, and `useTelemetryNavigation`.
   - Sync `route.query.astronaut` via `setSelectedAstronaut(...)` on immediate watch so direct links/refreshes preserve the selected astronaut.
   - Replace hardcoded `<h2>Sam Sonia</h2>` in `<aside class="astronaut-panel">` with `<h2>{{ selectedAstronautName }}</h2>`.
   - Add a compact `<button type="button" class="back-btn" @click="goToCrewOverview">← Crew Overview</button>` in `.dashboard-header` before `<h1>PRIMARY VITALS</h1>`, styled consistently with the dark high-contrast HUD (`border: 1px solid #8a8a8a; background: #0a0a0a; color: #f5f5f5; min-height: 36px; padding: 0 10px; font-size: 12px; font-weight: 700; cursor: pointer;`).
7. Update `app/frontend/src/views/support-alert.vue` and `app/frontend/src/views/warning-alert.vue`:
   - Add `<script setup lang="ts">` importing `useTelemetryNavigation` and expose `goToCrewOverview`.
   - Add a compact top action bar containing `<button type="button" class="back-btn" @click="goToCrewOverview">← Crew Overview</button>` above the alert card in both views, styled identically with high-contrast dark HUD styling and `min-height: 36px` so the user can return cleanly from any alert view.
8. Run `bun run --cwd app/frontend type-check` and `bun run --cwd app/frontend build` from repository root and verify both exit `0`.
9. Write `.tickets/011-prepare-mockup-navigation-and-alert-routing/changelog.md` listing files created/modified, exact commands run with exit codes, and any observations.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/011-prepare-mockup-navigation-and-alert-routing/changelog.md` listing:
- Files created and modified
- Verification commands executed and their exit codes/outputs
- Any blockers or assumptions avoided
