# 013 — Integrate seth views, navigation, alert routing, and UnoCSS dark theme onto main

| Field | Value |
| :--- | :--- |
| **ID** | `013` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Frontend Integration & UnoCSS Styling Builder |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `app/frontend/public/crew/sam-sonia.jpg`
  - `app/frontend/public/crew/patrick-cuckburn.jpg`
  - `app/frontend/src/composables/useTelemetryNavigation.ts`
  - `app/frontend/src/composables/useTelemetry.ts`
  - `app/frontend/src/components/CrewCard.vue`
  - `app/frontend/src/components/ViewNavBar.vue`
  - `app/frontend/src/components/DemoAlertStrip.vue`
  - `app/frontend/src/App.vue`
  - `app/frontend/src/router/index.ts`
  - `app/frontend/src/views/dashboard-principal.vue`
  - `app/frontend/src/views/support-alert.vue`
  - `app/frontend/src/views/warning-alert.vue`
  - `.tickets/013-integrate-views-navigation-and-unocss/changelog.md`
- **Must not touch:**
  - `.routing/**` (including `PROGRESS.md` — router updates `PROGRESS.md` upon acceptance)
  - `.claude/**`, `docs/**`, `db/**`, `server/**`, `bot/**`, `shared/**`
  - `app/app.go`, `app/client.go`, `app/client_test.go`, `app/main.go`, `app/go.mod`, `app/frontend/bindings/**`
  - `app/frontend/uno.config.ts`
  - `app/frontend/src/components/WindowBar.vue`
  - `app/frontend/src/components/EnvironmentCard.vue`
  - `app/frontend/src/components/MedicalManualButton.vue`
  - `app/frontend/src/views/HomeView.vue`
  - `.tickets/010-server-read-endpoints/**`, `.tickets/011-wails-backend-api-bindings/**`, `.tickets/012-frontend-dark-theme-and-home-view/**`
  - `.tickets/013-integrate-views-navigation-and-unocss/task.md`, `spec.md`

## Inputs

Read these files before acting:
1. [`app/frontend/uno.config.ts`](../../app/frontend/uno.config.ts) (theme colors: `ground`, `surface`, `main`, `side`, `dull`, `line`, `lineDim`, `ok`, `warn`, `crit`, `stale`; shortcuts: `hud-card`, `hud-pill`, `hud-value`, `button-main`, `button-side`)
2. [`app/frontend/src/composables/useTelemetry.ts`](../../app/frontend/src/composables/useTelemetry.ts) (live Wails `"telemetry:envelope"` listener, `ReadingStatus`, `Astronaut`, `LatestReading`, `AlertEvent`, `resolveReadingStatus`, `formatMetricValue`, `formatBloodPressure`)
3. [`app/frontend/src/components/CrewCard.vue`](../../app/frontend/src/components/CrewCard.vue) (live crew card component rendered inside `HomeView.vue`)
4. [`app/frontend/src/views/HomeView.vue`](../../app/frontend/src/views/HomeView.vue) (main `/` Home view to preserve untouched)
5. [`app/frontend/src/App.vue`](../../app/frontend/src/App.vue) and [`app/frontend/src/router/index.ts`](../../app/frontend/src/router/index.ts)
6. [`app/frontend/src/views/dashboard-principal.vue`](../../app/frontend/src/views/dashboard-principal.vue) and [`app/frontend/src/views/warning-alert.vue`](../../app/frontend/src/views/warning-alert.vue)
7. [`db/002_seed.sql`](../../db/002_seed.sql) (Lines 79, 97–98: `baumanometro_01`, `bp_systolic` `90–120`, `bp_diastolic` `60–80`)
8. Git reference from branch `seth` (`git show seth:app/frontend/src/views/support-alert.vue`) for the Caution / Support low blood pressure view copy

## Grounding & Anti-Hallucination Constraints

- Preserve `HomeView.vue`, `WindowBar.vue`, `EnvironmentCard.vue`, `MedicalManualButton.vue`, and `uno.config.ts` untouched. `/` (`HomeView.vue`) is the single Home / Crew Overview view.
- Use `uno.config.ts` semantic classes (`bg-ground`, `bg-surface`, `text-main`, `text-side`, `text-dull`, `border-line`, `border-lineDim`, `text-ok`, `border-ok`, `text-warn`, `border-warn`, `text-crit`, `border-crit`, `text-stale`, `hud-card`, `hud-pill`, `hud-value`, `button-side`, `button-main`) and `lucide-vue-next` icons instead of ad-hoc hex CSS in the new/updated views and components.
- Do **not** add direct browser `new WebSocket` or `fetch()` calls in `app/frontend/src/`; live telemetry already flows through `useTelemetry.ts` via `@wailsio/runtime` `Events.On("telemetry:envelope", ...)`.
- In `support-alert.vue`, keep `"immediately"` spelled correctly and keep the `<!-- [UNVERIFIED] Demo guidance copy from UI mockup — clinical review required -->` comment above the 3 fallback action items.
- Do **not** commit or push any git changes.
- If `bun run --cwd app/frontend type-check` or `bun run --cwd app/frontend build` fails, **stop** and record exact output in `.tickets/013-integrate-views-navigation-and-unocss/changelog.md`.

## Task

1. Restore the two offline crew portrait JPEGs from `seth` into `app/frontend/public/crew/`:
   - `mkdir -p app/frontend/public/crew`
   - `git show seth:app/frontend/public/crew/sam-sonia.jpg > app/frontend/public/crew/sam-sonia.jpg`
   - `git show seth:app/frontend/public/crew/patrick-cuckburn.jpg > app/frontend/public/crew/patrick-cuckburn.jpg`
2. Create `app/frontend/src/composables/useTelemetryNavigation.ts`:
   - Export `type NavReadingStatus = "normal" | "abnormal" | "critical" | "stale"`.
   - Export `interface AlertDispatchPayload { status: NavReadingStatus; astronautId?: string; astronautName?: string; metricCode?: string; source?: string; deviceCode?: string; valueDisplay?: string; message?: string; met?: string; timestampUtc?: string }`.
   - Export `TELEMETRY_STATUS_EVENT = "hrig:telemetry-status"`.
   - Maintain shared module-level reactive refs:
     - `selectedAstronautId = ref<string>("")`
     - `selectedAstronautFallbackName = ref<string>("Sam Sonia")`
     - `lastStatus = ref<NavReadingStatus>("normal")`
     - `lastTriggeredAlert = ref<AlertDispatchPayload | null>(null)`
   - Export `registerTelemetryRouter(router: Router): void`.
   - In `useTelemetryNavigation()`:
     - `setSelectedAstronaut(id: string, name?: string)`: updates `selectedAstronautId.value = id` and when `name` is non-empty updates `selectedAstronautFallbackName.value = name`.
     - `openPrimaryVitals(astronaut: { id: string; fullName?: string } | string)`: extracts `id` and optional `fullName`, calls `setSelectedAstronaut`, and navigates to `{ name: "dashboard-principal", query: { astronaut: id } }`.
     - `goToHome()`: navigates to `{ name: "home" }` (`/`).
     - `dispatchTelemetryStatus(payload: AlertDispatchPayload)`: updates `lastStatus.value = payload.status`; if `"critical"` sets `lastTriggeredAlert.value = payload` and routes to `{ name: "warning-alert" }` (unless already on `"warning-alert"`); if `"abnormal"` sets `lastTriggeredAlert.value = payload` and routes to `{ name: "support-alert" }` (unless already on `"support-alert"`); if `"normal"` clears `lastTriggeredAlert.value = null`.
     - `resetDemoStatus()`: sets `lastStatus.value = "normal"` and `lastTriggeredAlert.value = null`.
   - Also export standalone `dispatchTelemetryStatus(payload: AlertDispatchPayload): void` so `useTelemetry.ts` can call it outside Vue component `setup()`.
3. Update `app/frontend/src/composables/useTelemetry.ts`:
   - Import `dispatchTelemetryStatus` and `type NavReadingStatus` from `@/composables/useTelemetryNavigation`.
   - Add a small helper `toNavStatus(status: TelemetryDisplayStatus): NavReadingStatus | null` mapping `ReadingStatus.StatusNormal` -> `"normal"`, `ReadingStatus.StatusAbnormal` -> `"abnormal"`, `ReadingStatus.StatusCritical` -> `"critical"`, `ReadingStatus.StatusStale` -> `"stale"`, otherwise `null`.
   - Inside `handleEnvelope`:
     - In `case "reading":` (and NOT `"stale"` or `"snapshot"`), if `envelope.reading` has resolved status `"abnormal"` or `"critical"`, call `dispatchTelemetryStatus(...)` with `status`, `astronautId: envelope.reading.astronautId`, `metricCode: envelope.reading.metricCode`, `source: envelope.reading.source`, `deviceCode: envelope.reading.deviceCode`, `met: envelope.reading.met`, `timestampUtc: envelope.reading.recordedAtUtc`, and formatted `valueDisplay` (from `formatMetricValue(envelope.reading, metricsByCode.value[envelope.reading.metricCode]).display`).
     - In `case "alert":` (not `"alert_ack"`), if `envelope.alert && !envelope.alert.acknowledged`, and `envelope.alert.status` maps to `"abnormal"` or `"critical"`, call `dispatchTelemetryStatus(...)` with `status`, `astronautId`, `metricCode`, `source`, `deviceCode`, `message`, `met`, `timestampUtc: envelope.alert.triggeredAtUtc`.
4. Update `app/frontend/src/components/CrewCard.vue`:
   - Import `useTelemetryNavigation` and expose `openPrimaryVitals`.
   - Make the root `<article class="hud-card ...">` interactive and keyboard-accessible: add `role="button"`, `tabindex="0"`, `cursor-pointer transition-colors hover:bg-surface focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-line`, `:aria-label="`Open ${astronaut.fullName} primary vitals`"`, `@click="openPrimaryVitals(astronaut)"`, `@keydown.enter.prevent="openPrimaryVitals(astronaut)"`, `@keydown.space.prevent="openPrimaryVitals(astronaut)"`.
   - Add a `portraitUrl` computed checking if `props.astronaut.fullName.toLowerCase()` includes `"sam sonia"` -> `"/crew/sam-sonia.jpg"` or `"patrick"` -> `"/crew/patrick-cuckburn.jpg"`, rendering `<img v-if="portraitUrl" :src="portraitUrl" :alt="astronaut.fullName" class="w-full h-full object-cover object-top" />` inside the existing portrait container, keeping the existing SVG initials badge as fallback for all other crew members.
5. Create `app/frontend/src/components/ViewNavBar.vue` using `uno.config.ts` tokens (`bg-ground border-b border-solid border-lineDim px-4 py-2 flex flex-wrap items-center gap-2 text-xs text-main`):
   - Render `<RouterLink>` items (`button-side h-9 px-3 text-xs gap-1.5` with active route class `bg-surface border-line text-main ring-1 ring-line`):
     - `{ name: 'home' }`: `<Home class="w-3.5 h-3.5" aria-hidden="true" /><span>Home</span>`
     - `{ name: 'dashboard-principal' }`: `<span>Primary Vitals</span>`
     - `{ name: 'support-alert' }`: `<AlertTriangle class="w-3.5 h-3.5 text-warn" aria-hidden="true" /><span>Support</span>`
     - `{ name: 'warning-alert' }`: `<AlertCircle class="w-3.5 h-3.5 text-crit" aria-hidden="true" /><span>Warning</span>`
6. Create `app/frontend/src/components/DemoAlertStrip.vue` using `uno.config.ts` tokens (`bg-ground border-t border-solid border-lineDim px-4 py-2 flex flex-wrap items-center justify-between gap-2 text-xs text-main hud-value`):
   - Status pill (`hud-pill px-3 py-1.5 inline-flex items-center gap-1.5 border-lineDim`) using `text-ok` (`NOMINAL`), `text-warn border-warn` (`CAUTION / ABNORMAL`), `text-crit border-crit` (`CRITICAL`), `text-stale` (`STALE`).
   - Buttons: `Home` (`goToHome()`), `Simulate Abnormal` (`border-warn text-warn` with `AlertTriangle`), `Simulate Critical` (`border-crit text-crit` with `AlertCircle`), and `Reset` (`resetDemoStatus()`).
7. Update `app/frontend/src/App.vue`:
   - Register `router` via `registerTelemetryRouter(router)`, register/cleanup `window` `TELEMETRY_STATUS_EVENT` listener (`onMounted` / `onBeforeUnmount`), and render `<WindowBar />`, `<ViewNavBar class="shrink-0" />`, `<div class="flex-1 flex flex-col"><RouterView /></div>`, and `<DemoAlertStrip class="shrink-0 sticky bottom-0 z-40" />`.
8. Update `app/frontend/src/router/index.ts` to register:
   - `{ path: "/", name: "home", component: HomeView }`
   - `{ path: "/dashboard-principal", name: "dashboard-principal", component: DashboardPrincipal }`
   - `{ path: "/support-alert", name: "support-alert", component: SupportAlert }`
   - `{ path: "/warning-alert", name: "warning-alert", component: WarningAlert }`
   - `{ path: "/pagina-principal", redirect: { name: "home" } }`
9. Refactor `app/frontend/src/views/dashboard-principal.vue`, `app/frontend/src/views/support-alert.vue`, and `app/frontend/src/views/warning-alert.vue` to use `uno.config.ts` tokens (`bg-ground`, `bg-surface`, `text-main`, `text-side`, `text-dull`, `border-line`, `border-lineDim`, `text-ok`, `text-warn`, `border-warn`, `text-crit`, `border-crit`, `hud-card`, `hud-pill`, `hud-value`, `button-side`) and `lucide-vue-next` icons:
   - In `dashboard-principal.vue`: include `Home` button (`goToHome`), sync `route.query.astronaut` + `astronauts` from `useTelemetry()` to resolve the displayed astronaut name (`selectedAstronautName`) and live primary readings (`heart_rate`, `hrv_rmssd`, `spo2`, `cabin_o2`, `bp_systolic`/`bp_diastolic`, `cabin_pressure`, `radiation_rate`, `radiation_dose_cum`, `core_temp`, `cabin_temp`) with clean demo/offline fallbacks, and preserve the anatomical astronaut SVG + `Caution` (`text-warn`) / `Warning` (`text-crit`) legend.
   - In `support-alert.vue`: include `Home` button (`goToHome`), yellow `AlertTriangle` (`text-warn`) + `SUPPORT` + `CAUTION` badge (`border-warn text-warn`), render `lastTriggeredAlert` metric/value/source/timestamp if triggered dynamically with fallback to `LOW BLOOD PRESSURE DETECTED` (`86/55 mmHg`, `Nominal Range: 90–120 / 60–80 mmHg`, `Source: Blood pressure cuff (baumanometro_01) · Demo sample · MET 000:00:00 · 2026-11-15T21:42:00Z (UTC)`), and the 3 `RECOMMENDED ACTIONS` with `<!-- [UNVERIFIED] Demo guidance copy from UI mockup — clinical review required -->` and `"immediately"` fixed.
   - In `warning-alert.vue`: include `Home` button (`goToHome`), red `AlertCircle` (`text-crit`) + `WARNING` + `CRITICAL` badge (`border-crit text-crit`), render `lastTriggeredAlert` metric/value/message/timestamp if triggered dynamically with fallback to `Pressure: 78 kPa (Drop below 90 kPa)` / `Decompression Sickness (DCS)`, and `REQUIRED ACTIONS` / `IMMEDIATE EMERGENCY PROCEDURES`.
10. Run `bun run --cwd app/frontend type-check` and `bun run --cwd app/frontend build`, confirming both exit `0` (note: do not modify any `.go` files; `go test ./...` in `app/` requires system `webkitgtk-6.0` on Linux which is unavailable in this environment).
11. Write `.tickets/013-integrate-views-navigation-and-unocss/changelog.md` with files created/modified and exact verification outputs.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/013-integrate-views-navigation-and-unocss/changelog.md` listing:
- Files created and modified
- Verification commands executed and their exit codes/outputs
- Any blockers or assumptions avoided
