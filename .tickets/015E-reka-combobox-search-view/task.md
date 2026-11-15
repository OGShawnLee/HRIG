# 015E — Wails Search Bindings and Reka UI Omnibox Search Page

| Field | Value |
| :--- | :--- |
| **ID** | `015E` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Wails v3 & Vue 3 / Reka UI Frontend Engineer |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `app/client.go`
  - `app/app.go`
  - `app/client_test.go`
  - `app/frontend/bindings/` *(only generated files from `wails3 task bindings`)*
  - `app/frontend/uno.config.ts` *(optional utility shortcuts only; keep existing color/font palette intact)*
  - `app/frontend/src/App.vue` *(keep `WindowBar`, `ViewNavBar`, `DemoAlertStrip`, and `TELEMETRY_STATUS_EVENT` intact while adding global `Ctrl+K` / `Cmd+K` routing to `/search`)*
  - `app/frontend/src/components/ViewNavBar.vue` *(add a compact `Search` + `Ctrl+K` nav link alongside `Home`, `Primary Vitals`, `Support`, `Warning`)*
  - `app/frontend/src/router/index.ts`
  - `app/frontend/src/views/HomeView.vue`
  - `app/frontend/src/views/SearchView.vue`
  - `app/frontend/src/components/SearchFeaturedEntry.vue`
  - `app/frontend/src/composables/useTelemetry.ts`
  - `.tickets/015E-reka-combobox-search-view/changelog.md`
- **Must not touch:**
  - `.routing/INPUT.md`
  - `db/`, `server/`, `shared/`, `bot/`, `bot-n/`
  - `app/frontend/src/components/CrewCard.vue`, `EnvironmentCard.vue`, `DemoAlertStrip.vue`
  - `app/frontend/src/views/dashboard-principal.vue`, `support-alert.vue`, `warning-alert.vue`

## Inputs

- [`shared/model/telemetry.go`](../../shared/model/telemetry.go) *(for `SpaceHazard`, `MedicalProcedure`, `SearchResults`, `EscalationLevel` added in Ticket 014E)*
- [`.routing/DESIGN.md`](../../.routing/DESIGN.md) *(dark HUD palette, contrast >= 6:1, redundant status icon + text, non-physician clarity)*
- [`app/client.go`](../../app/client.go)
- [`app/app.go`](../../app/app.go)
- [`app/frontend/uno.config.ts`](../../app/frontend/uno.config.ts)
- [`app/frontend/src/App.vue`](../../app/frontend/src/App.vue)
- [`app/frontend/src/components/ViewNavBar.vue`](../../app/frontend/src/components/ViewNavBar.vue)
- [`app/frontend/src/composables/useTelemetryNavigation.ts`](../../app/frontend/src/composables/useTelemetryNavigation.ts)
- [`app/frontend/src/views/HomeView.vue`](../../app/frontend/src/views/HomeView.vue)
- [`app/frontend/src/composables/useTelemetry.ts`](../../app/frontend/src/composables/useTelemetry.ts)
- [`app/frontend/node_modules/reka-ui/src/Combobox/ComboboxRoot.vue`](../../app/frontend/node_modules/reka-ui/src/Combobox/ComboboxRoot.vue)

## Grounding & Anti-Hallucination Constraints

1. **Prerequisite:** This ticket builds directly on Ticket `014E`. Verify `GET /api/hazards`, `GET /api/procedures`, and `GET /api/search` exist in `server/` before starting.
2. **Keep HRIG Dark HUD Theme:** Match the Omnibox/Command Palette structure (top search header, rectangular category tabs, left `Featured Entry` preview panel, right grouped results pane with group counts, bottom keyboard hint footer) strictly using HRIG's dark HUD palette (`bg-ground` `#000000`, `bg-surface` `#0A0A0A`, `border-line`/`border-lineDim`, `text-main` `#FFFFFF`, `text-side`, `text-dull`, `ok`/`warn`/`crit`/`stale`). Never switch to a white background.
3. **Use Real Reka UI `Combobox` Primitives:** Import and use `ComboboxRoot`, `ComboboxInput`, `ComboboxContent`, `ComboboxViewport`, `ComboboxGroup`, `ComboboxLabel`, `ComboboxItem`, and `ComboboxEmpty` from `reka-ui`. Do not replace `Combobox*` with a plain `<input>` + `<ul>`.
   - Pass `:open="true"`, `:ignore-filter="true"`, `:reset-search-term-on-blur="false"`, `:reset-search-term-on-select="false"` on `ComboboxRoot`, and bind `position="inline"` on `ComboboxContent`.
   - Use a stable discriminated union key string (`crew:<id>`, `proc:<code>`, `hazard:<code>`, `metric:<code>`) for each `ComboboxItem` `:value` so Reka UI highlighting and selection work reliably across mixed result types.
4. **No Hardcoded Medical Content or Thresholds in Vue/TS:** All astronauts, procedures, hazards, metrics, thresholds, and procedure steps must come from backend RPCs (`Search`, `GetProcedures`, `GetHazards`, `GetAstronauts`, `GetMetrics`, `GetSnapshot`).
5. **Integrate with Teammate Crew Views (`dashboard-principal.vue`):** Selecting a **Crew** entry (`Enter` or explicit "Open Primary Vitals" action) routes via `useTelemetryNavigation().openPrimaryVitals(astronaut)` (`/dashboard-principal?astronaut=<id>`).

## Task

1. **Extend Wails Go Client & App Bindings (`app/client.go`, `app/app.go`, `app/client_test.go`):**
   - Add `GetHazards(ctx, q string) ([]model.SpaceHazard, error)` -> `GET /api/hazards`.
   - Add `GetProcedures(ctx, q, category, metricCode string) ([]model.MedicalProcedure, error)` -> `GET /api/procedures`.
   - Add `Search(ctx, q, tab string) (model.SearchResults, error)` -> `GET /api/search`.
   - Expose bound RPCs `GetHazards(q string)`, `GetProcedures(q string, category string, metricCode string)`, and `Search(q string, tab string)` on `App` in `app/app.go` with normalized empty slices (`[]...{}`) so bindings never return `null` slices.
   - Add `httptest.Server` unit tests in `app/client_test.go` verifying query escaping and JSON decoding for `GetHazards`, `GetProcedures`, and `Search`.
   - Run `wails3 task bindings` from the repo root to regenerate `app/frontend/bindings/` (`App.Search`, `App.GetProcedures`, `App.GetHazards`, `SpaceHazard`, `MedicalProcedure`, `SearchResults`, `EscalationLevel`).
2. **Wire Global `Ctrl+K` & Triggers (`app/frontend/src/router/index.ts`, `App.vue`, `ViewNavBar.vue`, `HomeView.vue`):**
   - Register `{ path: "/search", name: "search", component: SearchView }` in `app/frontend/src/router/index.ts` (preserving `/`, `/dashboard-principal`, `/support-alert`, `/warning-alert`, `/pagina-principal`).
   - Add a global `Ctrl+K` / `Cmd+K` keydown listener in `App.vue` (cleaning up on `onBeforeUnmount`):
     - Call `event.preventDefault()`.
     - If not on `/search`, store/remember previous route or push `{ name: "search" }`.
     - If already on `/search`, dispatch `window.dispatchEvent(new CustomEvent("hrig:focus-omnibox"))` so `SearchView.vue` focuses and selects its `ComboboxInput`.
   - Update `ViewNavBar.vue` to add a compact `Search` + `Ctrl+K` link alongside `Home`, `Primary Vitals`, `Support`, `Warning` (and ensure any `<kbd>` uses subtle `border-lineDim` styling).
   - Update `HomeView.vue`:
     - Clicking or focusing the top `Search` bar navigates to `{ name: "search" }`, and typing a printable key while focused immediately navigates to `{ name: "search", query: { q: <typed> } }` (or render the HomeView search bar as a clickable trigger button styled with `.field` that opens `{ name: "search" }` cleanly!).
     - Clicking `Medical Manual` navigates directly to `{ name: "search", query: { tab: "procedures" } }` (remove the old stub `showManualNotice` banner).
3. **Implement `SearchView.vue` & `SearchFeaturedEntry.vue` (Reka UI Omnibox Layout):**
   - **Omnibox Frame:** Full-page dark HUD panel (`max-w-6xl w-full mx-auto border border-solid border-line bg-ground flex flex-col`) inside `/search` so it matches the attached command-palette structure cleanly inside HRIG's window shell.
   - **Top Input Row:**
     - Left magnifying glass (`Search` icon) + `ComboboxInput` (`placeholder="Search crew, medical procedures, hazards, or metrics..."`, autofocused on mount via `nextTick` and on `hrig:focus-omnibox` event).
     - Right `×` clear button when `query` is non-empty (clears `query` and refocuses input), vertical hairline divider, `<kbd>Ctrl + K</kbd>` badge, and subtle `Close` / `Esc` action returning to previous route (or `/`).
   - **Category Tabs Row (`All` | `Crew` | `Procedures` | `Hazards & Metrics`):**
     - Map tabs directly to backend values: `"all"`, `"crew"`, `"procedures"`, `"hazards_metrics"`.
     - Initialize `activeTab` and `query` from `route.query.tab` and `route.query.q` (reacting if route query changes while mounted) and sync tab/query changes cleanly without duplicate fetches.
     - Active tab styled inverted (`bg-main text-ground border-line font-semibold`), inactive tabs styled `bg-ground text-side border-lineDim hover:border-line hover:text-main`.
   - **Search Execution & Fallback State:**
     - Call `Search(query, activeTab)` from `@bindings/hrig/app/app.js` with ~120ms debounce and a monotonic request sequence counter so older responses never overwrite newer ones.
     - Maintain `highlightedKey` (`crew:<id>`, `proc:<code>`, `hazard:<code>`, `metric:<code>`) and `pinnedKey`:
       - `@highlight` on `ComboboxRoot` updates `highlightedKey` immediately as the user presses `↑ ↓` or hovers items.
       - Default `featuredKey` = `pinnedKey ?? highlightedKey ?? firstVisibleKey`, automatically resetting `highlightedKey`/`pinnedKey` to the first visible result whenever the current key disappears after filtering/tab changes so the left `FEATURED ENTRY` pane is never blank when results exist.
     - Selecting an item via `Enter` or click (`@update:model-value`):
       - **Crew:** Calls `useTelemetryNavigation().openPrimaryVitals(astronaut)` (`/dashboard-principal?astronaut=<id>`).
       - **Medical Procedure:** Pins that procedure in the left `FEATURED ENTRY` pane (displaying an `ACTIVE CHECKLIST` badge and scrolling/focusing the detail pane smoothly on narrow screens).
       - **Space Hazard / Metric:** Pins that entry in the left `FEATURED ENTRY` pane; also provide one-click `"View Matching Procedures"` chip inside `SearchFeaturedEntry.vue` that switches `activeTab = "procedures"` and sets `query` to the first linked metric code!
     - `Esc` handling (`@escape-key-down.prevent` on `ComboboxContent` + keydown on `ComboboxInput`):
       - Always closes `/search` (navigates back via `window.history.length > 1 ? router.back() : router.push({ name: "home" })`), matching the `Esc Close` footer in the mockup (while the `×` button clears the input).
   - **Left Pane (`FEATURED ENTRY` — `SearchFeaturedEntry.vue`):**
     - **Crew:** Full name, nationality code/label, active status, live vitals summary from `useTelemetry` (SpO₂, BP, Heart Rate, Temp, Radiation) paired with status icons (`normal`/`abnormal`/`critical`/`stale`), plus prominent `"Open Primary Vitals"` button.
     - **Medical Procedure:** Title, `category` + `bodySystem` pills, `escalation` badge with Lucide icon + label (`Self-Care`, `Crew Medical Officer (CMO)`, `Flight Surgeon`), `~N min`, **When to Use**, **Equipment Needed**, **Numbered Steps (`1..N`)**, **Warnings / Escalation Triggers**, clickable linked `relatedMetricCodes` chips (clicking filters matching metric/procedures), `hazardCodes`, and `source`.
     - **Space Hazard:** Hazard name, NASA HRP summary (*What it is*), clinical impact (*Why it matters*), linked metric chips, and `source`.
     - **Metric:** Metric name, `code`, tier, display unit, DB `warning_min/max` & `critical_min/max` thresholds from catalog, plus live per-crew or cabin values from `useTelemetry` and `"View Procedures for <metric>"` button.
   - **Right Pane (`ComboboxContent position="inline"` + `ComboboxViewport`):**
     - Render each non-empty section with `ComboboxGroup`, `ComboboxLabel` + zero-padded count (`03`, `12`, `05`):
       - **Crew** (`Crew`): single-column rows showing full name, nationality label, and summary of current telemetry/alert status.
       - **Medical Procedures** (`Medical Procedures`): 2-column grid (`grid-cols-1 sm:grid-cols-2 gap-3`), each card showing title, concise `whenToUse`/`summary`, escalation badge, and metric tags; `data-[highlighted]` styled with `border-line bg-surface ring-1 ring-line`.
       - **NASA Hazards** (`NASA Hazards`): 2-column grid showing hazard title, summary excerpt, and linked metric count.
       - **Metrics** (`Metrics`): 2-column grid or rows showing metric name, code, tier, unit, and live value/status if available.
     - Render `ComboboxEmpty` when zero results match, offering a `"Reset filters"` button (`query = ""`, `activeTab = "all"`).
   - **Bottom Footer Bar:**
     - Left: `<totalResults> results` or `<totalResults> results for '<query>'`.
     - Right: `<kbd>↑</kbd> <kbd>↓</kbd> Navigate`, `<kbd>Enter</kbd> Select`, and clickable `<kbd>Esc</kbd> Close` button.
4. **Verify Everything:**
   - `go -C app vet .`
   - `go -C app test -v -count=1 ./...`
   - `wails3 task client:check` (`bindings` -> `vue-tsc --noEmit` -> `vite build` -> `go build .`)
   - Live smoke test with `postgres`, `bin/hrig-server.exe`, and `bot` confirming `/api/hazards` (`5`), `/api/procedures` (`20`), and `/api/search?q=hypoxia&tab=all`, and stopping all background processes cleanly afterward.

## Verification

Run every check in [spec.md](spec.md).

## Change log

Write `.tickets/015E-reka-combobox-search-view/changelog.md` listing every file touched, verification commands/exit codes, and any UX/edge-case notes.
