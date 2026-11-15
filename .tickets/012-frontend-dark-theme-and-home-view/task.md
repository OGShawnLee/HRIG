# 012 — Configure minimal dark UnoCSS theme & implement Figma HomeView

| Field | Value |
| :--- | :--- |
| **ID** | `012` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Wails v3 Vue 3 & UnoCSS UI Engineer |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `app/main.go` (change `BackgroundColour` from `application.NewRGB(255, 255, 255)` to `application.NewRGB(0, 0, 0)`)
  - `app/frontend/index.html` (dark `bg-ground text-main` body styling)
  - `app/frontend/uno.config.ts` (minimal dark HUD color tokens, clinical status tokens, `hud-card` / `hud-pill` / dark button & field shortcuts)
  - `app/frontend/src/App.vue` (root `bg-ground text-main` shell)
  - `app/frontend/src/components/WindowBar.vue` (dark header bar + live WebSocket connection status indicator)
  - `app/frontend/src/composables/useTelemetry.ts` (loads initial `GetAstronauts`, `GetMetrics`, `GetSnapshot`, `GetConnectionStatus` from `@bindings/hrig/app/app.js`, subscribes to `"telemetry:envelope"` and `"telemetry:connection"` via `Events.On` from `@wailsio/runtime`, formats values/units/BP pair/worst status)
  - `app/frontend/src/components/CrewCard.vue` (Figma-aligned crew card with portrait/avatar + 6 primary physiological rows)
  - `app/frontend/src/components/EnvironmentCard.vue` (Figma-aligned `Environment` card with 4 cabin rows)
  - `app/frontend/src/components/MedicalManualButton.vue` (Figma-aligned `Medical Manual` button card)
  - `app/frontend/src/views/HomeView.vue` (initial Home dashboard matching the Figma layout: top Search bar + dynamic Crew cards + bottom Medical Manual & Environment row)
  - `.tickets/012-frontend-dark-theme-and-home-view/changelog.md`
- **Must not touch:**
  - `server/**`, `bot/**`, `shared/**`, `db/**`, `local/**`, `docs/**`
  - `app/app.go`, `app/app_window.go`, `app/client.go`, `app/client_test.go`, `app/go.mod`, `app/go.sum`, `app/Taskfile.yml`, `app/build/**`
  - `app/frontend/package.json`, `app/frontend/vite.config.ts`, `app/frontend/tsconfig.json`, `app/frontend/src/main.ts`, `app/frontend/src/router/index.ts`, `app/frontend/bindings/**`
  - `Taskfile.yml`, `go.work`
  - `.gitignore`, `.routing/**`, `.claude/**`, `README.md`
  - `.tickets/001-*` .. `.tickets/011-*`
  - `.tickets/012-frontend-dark-theme-and-home-view/task.md`, `.tickets/012-frontend-dark-theme-and-home-view/spec.md`

## Inputs

Read these files before acting:
1. [`.routing/DESIGN.md`](../../.routing/DESIGN.md)
2. [`db/002_seed.sql`](../../db/002_seed.sql) (metrics catalog lines 37–54 and unit decimals lines 9–25)
3. [`app/main.go`](../../app/main.go)
4. [`app/frontend/uno.config.ts`](../../app/frontend/uno.config.ts)
5. [`app/frontend/index.html`](../../app/frontend/index.html)
6. [`app/frontend/src/App.vue`](../../app/frontend/src/App.vue)
7. [`app/frontend/src/components/WindowBar.vue`](../../app/frontend/src/components/WindowBar.vue)
8. [`app/frontend/src/views/HomeView.vue`](../../app/frontend/src/views/HomeView.vue)
9. [`app/frontend/bindings/hrig/app/app.ts`](../../app/frontend/bindings/hrig/app/app.ts)
10. [`app/frontend/bindings/hrig/shared/model/models.ts`](../../app/frontend/bindings/hrig/shared/model/models.ts)

## Grounding & Anti-Hallucination Constraints

- **Wait for Ticket `011` Bindings:**
  - Ticket `012` assumes `010` and `011` are completed so `@bindings/hrig/app/app.js` and `@bindings/hrig/shared/model/models.js` exist with `GetAstronauts`, `GetMetrics`, `GetSnapshot`, `GetConnectionStatus`, and runtime events `"telemetry:envelope"` / `"telemetry:connection"`.
- **Database Metric Truth (Zero Invented Metrics or Hardcoded Thresholds):**
  - Crew Card 6 physiological rows map strictly to active `metric_type = 1` codes from `db/002_seed.sql`:
    1. `Oxygenation:` ← `spo2`
    2. `Blood Pressure:` ← `bp_systolic` + `/` + `bp_diastolic` (unit `mmHg`, status = worst severity across both: `critical` > `stale` > `abnormal` > `normal`)
    3. `Rad. Exposure:` ← `radiation_rate` (use catalog unit symbol `µSv/h`, never rename to `mSv`)
    4. `Acc. Radiation:` ← `radiation_dose_cum` (`mSv`)
    5. `Heart Beat:` ← `heart_rate` (`bpm`)
    6. `Temperature:` ← `core_temp` (`°C`)
  - `Environment` Card 4 rows map strictly to active `metric_type = 2` cabin codes from `db/002_seed.sql`:
    1. `O₂:` ← `cabin_o2` (`%`)
    2. `Cabin Pressure:` ← `cabin_pressure` (`kPa`)
    3. `Cabin ppCO₂:` ← `cabin_ppco2` (`mmHg`) *(Note: the Figma mockup labeled row 3 `Radiation Exposure`, but `db/002_seed.sql` only tracks `radiation_rate` per astronaut (`metric_type=1`) and tracks `cabin_ppco2` for cabin environment (`metric_type=2`). Do **not** invent a fake environmental radiation metric.)*
    4. `Temperature:` ← `cabin_temp` (`°C`)
  - Never hardcode clinical thresholds in Vue/TypeScript; always render status colors directly from server-classified `reading.status` (`normal`, `abnormal`, `critical`, `stale`).
- **Dark Theme & Visual Redundancy (`.routing/DESIGN.md`):**
  - Canvas background must be `#000000` (`ground: "#000000"`), primary border/text `#FFFFFF`, secondary text `>= 6:1` contrast (`#D4D4D8` / `#9CA3AF`).
  - Numeric telemetry values must use `tabular-nums`.
  - Pair every non-normal status (`abnormal`, `critical`, `stale`) with both color AND a `lucide-vue-next` icon (`AlertTriangle`, `AlertCircle`, `WifiOff`) plus `aria-label` / `title` so color is never the sole indicator.
- **No Unsolicited Packages / No Unsolicited Git Ops:**
  - Do **not** run `bun add` or modify `package.json`; use only already-installed dependencies (`vue`, `lucide-vue-next`, `reka-ui`, `@vueuse/core`, `@wailsio/runtime`, `unocss`).
  - Do **not** commit or push changes.

## Task

1. **Set Window Dark Background in `app/main.go` & `app/frontend/index.html`:**
   - Update `BackgroundColour: application.NewRGB(0, 0, 0)` in `app/main.go`.
   - Update `app/frontend/index.html` `<body>` classes to `bg-ground text-main font-sans antialiased min-h-screen`.
2. **Create Minimal Dark Theme in `app/frontend/uno.config.ts`:**
   - Keep `presetWind3()`, `presetWebFonts({ provider: "fontshare", fonts: { sans: "Satoshi" } })`, and `transformerDirectives({ applyVariable: "--uno" })`.
   - Configure minimal dark `theme.colors`:
     - `ground: "#000000"`
     - `surface: "#0A0A0A"`
     - `main: "#FFFFFF"`
     - `side: "#D4D4D8"`
     - `dull: "#9CA3AF"`
     - `line: "#FFFFFF"`
     - `lineDim: "#262626"`
     - `ok: "#4ADE80"`
     - `warn: "#FACC15"`
     - `crit: "#F87171"`
     - `stale: "#9CA3AF"`
   - Define minimal `shortcuts`:
     - `hud-card`: `"bg-ground border border-solid border-line rounded-2xl"`
     - `hud-pill`: `"bg-ground border border-solid border-line rounded-xl"`
     - `hud-value`: `"font-medium tabular-nums"`
     - `button`: `"h-11 min-w-fit px-4 border border-solid border-line rounded-xl inline-flex items-center justify-center transition-colors disabled:opacity-50"`
     - `button-main`: `"button bg-main font-medium text-ground hover:bg-side"`
     - `button-side`: `"button bg-ground text-main hover:bg-surface"`
     - `field`: `"h-11 px-4 bg-ground text-main border border-solid border-line rounded-xl font-medium placeholder-dull focus:outline-none disabled:opacity-50"`
3. **Update `app/frontend/src/App.vue` & `app/frontend/src/components/WindowBar.vue`:**
   - In `App.vue`, apply `bg-ground text-main` on the root shell div.
   - In `WindowBar.vue`, replace light classes (`bg-ground border-light`) with dark bar classes (`bg-ground border-b border-solid border-lineDim`), preserve mac-style window control buttons (`OnClose`, `OnMinimize`, `OnMaximize`) and centered `HRIG`, and render a compact right-aligned live stream indicator (`Live` green / `Reconnecting` yellow or gray) driven by `useTelemetry()`.
4. **Create `app/frontend/src/composables/useTelemetry.ts`:**
   - Manage reactive state (`astronauts`, `metricsByCode`, `readingsByKey`, `alerts`, `connectionState`, `loading`, `error`, `lastUpdatedUtc`) via shared module-level reactive refs initialized once (plus manual `refresh()`).
   - Call `GetAstronauts()`, `GetMetrics()`, `GetSnapshot()`, and `GetConnectionStatus()` from `@bindings/hrig/app/app.js`.
   - Subscribe with `Events.On("telemetry:envelope", ...)` handling `"snapshot"`, `"reading"`, `"stale"`, `"alert"`, and `"alert_ack"`, and `Events.On("telemetry:connection", ...)` updating `connectionState`.
   - Provide helper functions:
     - `getCrewReading(astronautId: string, metricCode: string): LatestReading | undefined`
     - `getEnvReading(metricCode: string): LatestReading | undefined`
     - `formatMetricValue(reading: LatestReading | undefined, metric: Metric | undefined): { display: string; status: ReadingStatus | "unknown" }` (honoring `metric.baseUnitDecimals` and `metric.displayUnitSymbol ?? metric.baseUnitSymbol ?? reading.unitCode`)
     - `formatBloodPressure(astronautId: string): { display: string; status: ReadingStatus | "unknown" }` combining `bp_systolic` and `bp_diastolic` into `<sys>/<dia> mmHg` with worst status severity.
5. **Create Modular Home Components & Implement `HomeView.vue`:**
   - `CrewCard.vue`: accepts `astronaut: Astronaut` and reads formatted values via props or composable; renders crisp rounded white-border dark card (`hud-card`), offline-first NASA-style crew avatar/initials badge on top-left + split/centered full name on top-right, and the exact 6 rows (`Oxygenation:`, `Blood Pressure:`, `Rad. Exposure:`, `Acc. Radiation:`, `Heart Beat:`, `Temperature:`) with status color + accessible status icon for `abnormal`/`critical`/`stale`.
   - `MedicalManualButton.vue`: renders rounded white-border square card with `BookOpen` / medical cross-style Lucide composition + `"Medical Manual"` label, emitting `click` (toggles an inline accessible stub notice `"Medical Manual view will be added in an upcoming UI step"` cleanly so clicking gives clear feedback with zero dead ends).
   - `EnvironmentCard.vue`: renders `Environment` rounded white-border card (`hud-card`) with header `"Environment"` and the 4 rows (`O₂:`, `Cabin Pressure:`, `Cabin ppCO₂:`, `Temperature:`) + status color and icon redundancy.
   - `HomeView.vue`:
     - Centered Search bar at top (`max-w-xl`, white border, `Search` icon + input with `Ctrl+K` keyboard shortcut focus) that filters visible crew cards by name (or shows a subtle clear button when non-empty).
     - Centered responsive Crew Cards grid (`grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6` max-w-5xl) rendering all active crew from PostgreSQL (`Ana Torres`, `Liam Okafor`, `Sofia Petrov` when seeded), plus a clear non-crashing empty/offline state with Retry button if server is unreachable.
     - Bottom row (`max-w-4xl w-full flex flex-col sm:flex-row items-center sm:items-stretch justify-between gap-6`) placing `MedicalManualButton` on the left and `EnvironmentCard` on the right, matching the Figma composition cleanly with zero horizontal overflow at `1280x820` and `390x680`.
6. **Verify Type-Check, Vite Build, Go Build & Live Full-Stack Check:**
   - Run `wails3 task client:check` (`bindings` -> `bun run --cwd frontend type-check` -> `bun run --cwd frontend build` -> `go build .`) and verify exit code `0`. Remove any stray `app/app.exe` created by `go build .` if needed, or verify `wails3 task client:build` builds cleanly to `app/bin/hrig.exe`.
   - Start PostgreSQL (`wails3 task db:start`), `server`, and `bot` briefly alongside a headless/dev Vite preview or `wails3 task client:start` check to verify REST/WS payloads bind cleanly without runtime console/Go errors, then stop all processes cleanly (`db:stop`, kill server/bot).

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/012-frontend-dark-theme-and-home-view/changelog.md` listing:
- Files created and modified
- `wails3 task client:check` and live smoke verification outputs
- Any blockers or assumptions avoided
