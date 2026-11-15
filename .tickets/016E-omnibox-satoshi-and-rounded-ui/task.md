# 016E — Omnibox Satoshi Typography, Rounded HUD & All-Crew Metric Readouts

| Field | Value |
| :--- | :--- |
| **ID** | `016E` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Vue 3 & UnoCSS Frontend Engineer |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `app/frontend/src/views/SearchView.vue`
  - `app/frontend/src/components/SearchFeaturedEntry.vue`
  - `.tickets/016E-omnibox-satoshi-and-rounded-ui/changelog.md`
- **Must not touch:**
  - `.routing/INPUT.md`
  - `db/`, `server/`, `shared/`, `bot/`, `bot-n/`
  - `app/app.go`, `app/client.go`, `app/client_test.go`, `app/frontend/bindings/`
  - `app/frontend/uno.config.ts`, `app/frontend/index.html`, `app/frontend/src/App.vue`
  - `app/frontend/src/views/HomeView.vue`, `app/frontend/src/views/dashboard-principal.vue`, `app/frontend/src/views/support-alert.vue`, `app/frontend/src/views/warning-alert.vue`
  - `app/frontend/src/components/ViewNavBar.vue`, `app/frontend/src/components/CrewCard.vue`, `app/frontend/src/components/EnvironmentCard.vue`, `app/frontend/src/components/DemoAlertStrip.vue`

## Inputs

- [`app/frontend/uno.config.ts`](../../app/frontend/uno.config.ts) *(defines `Satoshi` via `font-sans`, `hud-card` `rounded-2xl`, `hud-pill` / `button` / `field` `rounded-xl`, and color tokens)*
- [`app/frontend/index.html`](../../app/frontend/index.html) *(`font-sans` body root)*
- [`app/frontend/src/views/HomeView.vue`](../../app/frontend/src/views/HomeView.vue) *(reference rounded HUD radii: `rounded-2xl`, `rounded-xl`, `rounded-md`)*
- [`app/frontend/src/components/ViewNavBar.vue`](../../app/frontend/src/components/ViewNavBar.vue) *(reference rounded navigation buttons and `<kbd>` styling)*
- [`app/frontend/src/composables/useTelemetry.ts`](../../app/frontend/src/composables/useTelemetry.ts) *(`activeAstronauts`, `getCrewReading`, `getEnvReading`, `formatMetricValue`, `resolveReadingStatus`)*
- [`app/frontend/src/views/SearchView.vue`](../../app/frontend/src/views/SearchView.vue)
- [`app/frontend/src/components/SearchFeaturedEntry.vue`](../../app/frontend/src/components/SearchFeaturedEntry.vue)

## Grounding & Anti-Hallucination Constraints

1. **Preserve Reka UI & Search Behavior:** Keep all Reka UI `Combobox*` bindings, RPC calls (`Search`, `useTelemetry`, `useTelemetryNavigation`), keyboard shortcuts (`Ctrl+K`, `↑`, `↓`, `Enter`, `Esc`), and routing queries intact. Do not add nested `<button>` elements inside `<ComboboxItem>` cards so Reka UI keyboard and click selection never conflict.
2. **Enforce Satoshi (`font-sans`) Everywhere in the Omnibox:** `uno.config.ts` maps `font-sans` to `Satoshi` via Fontshare. Browsers and WebView2 apply a default monospace font to `<kbd>` unless overridden, and several Omnibox tags currently use `font-mono`. Remove every `font-mono` class from `SearchView.vue` and `SearchFeaturedEntry.vue`, apply `font-sans` to the Omnibox root shell, `<ComboboxInput>`, `<button>` controls, and every `<kbd>` element, and keep `tabular-nums` where numeric alignment is needed.
3. **Match HRIG Rounded Geometry:** Match the exact radius hierarchy used in `uno.config.ts`, `HomeView.vue`, and `ViewNavBar.vue`:
   - Outer Omnibox card container: `rounded-2xl overflow-hidden` (matches `hud-card`)
   - Interactive cards (`ComboboxItem`), tabs, secondary buttons, primary `.button-main` CTA buttons, and inner content panels: `rounded-xl` (remove every `rounded-none`)
   - Equipment pills, metric/hazard filter chips, escalation/duration pills, step index badges (`1..N`), and icon-only Clear button: `rounded-lg`
   - Compact inline status/tier/category badges and `<kbd>` keycaps: `rounded-md`
4. **Live All-Crew Metric Readouts (`useTelemetry`, Zero Hardcoded Values):**
   - For crew metrics (`metricType !== 2`), render every active crew member from `activeAstronauts` with their own live `formatMetricValue(getCrewReading(astro.id, metric.code), metric)` value and redundant status icon + color (using `resolveReadingStatus` / `formatted.status`). Never summarize a crew metric by only showing the first astronaut's value.
   - For environmental metrics (`metricType === 2`, e.g., `cabin_o2`, `cabin_pressure`, `cabin_temp`), render `Cabin / Suit Sensor` with its live `formatMetricValue(getEnvReading(metric.code), metric)` value and status icon.
5. **Keep Dark HUD Palette Intact:** Do not change colors (`bg-ground`, `bg-surface`, `text-main`, `text-side`, `text-dull`, `border-line`, `border-lineDim`, `ok`, `warn`, `crit`, `stale`).

## Task

1. **Refactor Typography & Radii in `app/frontend/src/views/SearchView.vue`:**
   - Add `font-sans` to `<main>` and explicitly add `font-sans` to `ComboboxInput` (`#hrig-omnibox-input`), all `<button>` controls, and all `<kbd>` elements (`Ctrl + K`, `↑`, `↓`, `Enter`, `Esc`).
   - Remove all occurrences of `font-mono` and replace them with `font-sans` (preserving `tabular-nums` on numbers).
   - Round the outer Omnibox container (`ref="omniboxContainerRef"`) with `rounded-2xl overflow-hidden`.
   - Apply rounded corners consistently:
     - Clear (`×`) and `Retry` buttons: `rounded-lg`
     - `Close` button, category tabs, `Reset filters` button, and all four `ComboboxItem` groups: `rounded-xl` (preserving `data-[highlighted]:border-line data-[highlighted]:bg-surface data-[highlighted]:ring-1 data-[highlighted]:ring-line`)
     - Inline metadata tags and all `<kbd>` elements: `rounded-md` or `rounded-lg`

2. **Show All-Crew Live Telemetry Directly Inside Metric Search Results (`app/frontend/src/views/SearchView.vue`):**
   - Replace the single-value `metricLiveSummary(metric)` helper with a per-metric breakdown helper `metricLiveRows(metric)` returning:
     - If `metric.metricType === 2`: single row `{ key: "env", label: "Cabin / Suit", display, status }`.
     - Otherwise (`metricType !== 2`): one row per astronaut in `activeAstronauts.value` (`{ key: astro.id, label: astro.fullName, display, status }`).
   - Update the `Metrics` `<ComboboxItem>` cards to display the metric name, tier, code/unit, and the compact live breakdown rows for **all** crew members (or `Cabin / Suit` for environmental metrics), each showing member name, tabular value, and status icon (`CheckCircle2`, `AlertTriangle`, `AlertCircle`, `WifiOff`, `Minus`) + color (`text-ok`/`text-warn`/`text-crit`/`text-stale`/`text-dull`).
   - When `query.trim().length > 0` (`hasActiveQuery`), render the **Metrics** `<ComboboxGroup>` directly after **Crew** and before **Medical Procedures** and **NASA Hazards**, and match `entriesByKey` insertion order to the exact visual order (`astronauts` -> `metrics` -> `procedures` -> `hazards` when `hasActiveQuery`, and `astronauts` -> `procedures` -> `hazards` -> `metrics` when `query.trim() === ""`). That way typing `"oxygen"` (or `"heart"`, `"temp"`, `"pressure"`, `"radiation"`) automatically highlights the first matching `Metric` on the left and displays all crew members' live readings immediately on both panes.

3. **Refactor Typography & Radii in `app/frontend/src/components/SearchFeaturedEntry.vue`:**
   - Add `font-sans` on the `<aside>` root so every heading, label, value, and button inherits `Satoshi`.
   - Remove all occurrences of `font-mono` (related metric chips, source/procedure code footer line, hazard code, metric code subtitle, `MET` timestamp) and replace with `font-sans` (`tabular-nums` on `MET` / numeric values).
   - Remove `rounded-none` from all three `.button-main` CTA buttons (`Open Primary Vitals`, `View Matching Procedures`, `View Procedures for <metric>`) so they inherit `.button-main`'s `rounded-xl`.
   - Apply rounded corners across all left-pane panels, lists, pills, and tags:
     - Inner section containers (`Active Alert Callout`, `Live Vitals Summary`, `When to Use`, `Warnings & Escalation Triggers`, `What It Is`, `Why It Matters`, `Catalog Thresholds`, `Live Telemetry`): `rounded-xl`
     - Procedure checklist step `<li>` cards: `rounded-xl`; numeric step index badges (`1..N`): `rounded-lg`
     - Escalation badge, `~N min` pill, equipment `<li>` pills, clickable `relatedMetricCodes` & `hazardCodes` chips, hazard `relatedMetricCodes` chips, and live telemetry status badge (`statusBadgeClass`): `rounded-lg`
     - Compact metadata tags (`Active Checklist`, `Pinned`, `Crew`, `Active`/`Inactive`, `category`/`bodySystem`/`environment`, `NASA HRP Hazard`, `Crew Metric`/`Environment Metric`, `Tier`, unit tag): `rounded-md`

4. **Verify:**
   - Run all checks in [spec.md](spec.md) (`wails3 task client:check`, pattern/grep checks, scope check). Remove any temporary `app/app.exe` binary created by `client:check` if untracked.

5. **Change log:**
   - Write `.tickets/016E-omnibox-satoshi-and-rounded-ui/changelog.md` listing exact files modified, `wails3 task client:check` output/exit code, and verification results.

## Verification

Run every check in [spec.md](spec.md).

## Change log

Write `.tickets/016E-omnibox-satoshi-and-rounded-ui/changelog.md` after completing verification.
