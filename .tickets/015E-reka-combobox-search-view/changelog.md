# 015E — Change Log

## Summary

Implemented Wails Go REST client methods and App RPC bindings for medical hazards, procedures, and unified search (`GetHazards`, `GetProcedures`, `Search`), regenerated TypeScript bindings, wired global `Ctrl+K` / `Cmd+K` and navigation/home triggers to `/search`, and built the dark-HUD Reka UI `Combobox` Omnibox search page (`SearchView.vue`) with inline `Featured Entry` checklists and clinical previews (`SearchFeaturedEntry.vue`).

## Files Created or Modified

### Go Client, App Bindings & Tests
- `app/client.go` — Added `GetHazards(ctx, q)`, `GetProcedures(ctx, q, category, metricCode)`, and `Search(ctx, q, tab)` calling `GET /api/hazards`, `GET /api/procedures`, and `GET /api/search` with URL-escaped query parameters and normalized non-nil slices.
- `app/app.go` — Exposed bound Wails RPCs `GetHazards(q string)`, `GetProcedures(q string, category string, metricCode string)`, and `Search(q string, tab string)` on `App`, guaranteeing non-nil slices on both empty and error returns.
- `app/client_test.go` — Added `TestAppSearchManualBindings` (`httptest.Server`) verifying query parameter escaping, JSON decoding, nil-slice normalization, and HTTP error message propagation for `GetHazards`, `GetProcedures`, and `Search`.

### Regenerated Wails Bindings
- `app/frontend/bindings/hrig/app/app.ts` — Regenerated via `wails3 task client:bindings` (`GetHazards`, `GetProcedures`, `Search`).
- `app/frontend/bindings/hrig/shared/model/index.ts` — Regenerated model barrel exports.
- `app/frontend/bindings/hrig/shared/model/models.ts` — Regenerated TypeScript models (`EscalationLevel`, `MedicalProcedure`, `SearchResults`, `SpaceHazard`).

### Frontend Router, Shell & Views
- `app/frontend/src/router/index.ts` — Registered `{ path: "/search", name: "search", component: SearchView }` while preserving `/`, `/dashboard-principal`, `/support-alert`, `/warning-alert`, and `/pagina-principal`.
- `app/frontend/src/App.vue` — Added global `Ctrl+K` / `Cmd+K` keydown listener (routing to `/search` or dispatching `hrig:focus-omnibox` when already on `/search`) while keeping `WindowBar`, `ViewNavBar`, `DemoAlertStrip`, and `TELEMETRY_STATUS_EVENT` intact.
- `app/frontend/src/components/ViewNavBar.vue` — Added compact `Search` + `Ctrl+K` `<RouterLink>` alongside `Home`, `Primary Vitals`, `Support`, and `Warning`.
- `app/frontend/src/views/HomeView.vue` — Replaced the client-side input filter and Medical Manual notice stub with a `.field`-styled `Search` trigger button navigating to `/search` (or `/search?q=<key>` on printable key press) and wired `MedicalManualButton` directly to `/search?tab=procedures`.
- `app/frontend/src/components/SearchFeaturedEntry.vue` *(created)* — Implemented the `FEATURED ENTRY` detail pane supporting:
  - **Crew:** Live vitals (`SpO₂`, `Blood Pressure`, `Heart Rate`, `Core Temp`, `Radiation Rate`, `Acc. Radiation`) from `useTelemetry`, unacknowledged alert callout, status icons + labels, and `"Open Primary Vitals"` button (`/dashboard-principal?astronaut=<id>`).
  - **Medical Procedure (Option A):** Full checklist inline inside `/search` with `ACTIVE CHECKLIST` badge when pinned, `category` / `bodySystem` / `environment` badges, `escalation` badge with Lucide icon + label (`Self-Care`, `Crew Medical Officer (CMO)`, `Flight Surgeon`), `~N min`, `When to Use`, `Equipment Needed`, numbered `Steps`, `Warnings & Escalation Triggers`, clickable `relatedMetricCodes` & `hazardCodes` chips, and `source`.
  - **Space Hazard:** NASA HRP `summary` (*What it is*), `clinicalImpact` (*Why it matters*), clickable linked metric chips, `"View Matching Procedures"` action, and `source`.
  - **Metric:** Tier, unit, catalog `warningMin`/`warningMax` & `criticalMin`/`criticalMax` thresholds from `MetricThreshold`, live crew/cabin readings from `useTelemetry`, and `"View Procedures for <metric>"` button.
- `app/frontend/src/views/SearchView.vue` *(created)* — Implemented the Reka UI `ComboboxRoot` (`:open="true"`, `:ignore-filter="true"`, `:reset-search-term-on-blur="false"`, `:reset-search-term-on-select="false"`), `ComboboxInput`, `ComboboxContent` (`position="inline"`), `ComboboxViewport`, `ComboboxGroup`, `ComboboxLabel` with zero-padded counts, `ComboboxItem` with discriminated keys (`crew:<id>`, `proc:<code>`, `hazard:<code>`, `metric:<code>`), and `ComboboxEmpty` with `"Reset filters"`, plus category tabs (`All`, `Crew`, `Procedures`, `Hazards & Metrics`), 120ms debounced `Search` RPC with monotonic sequence guard, `@highlight` + `pinnedKey` preview logic, `Esc` close handling, and bottom keyboard hint footer.

## Verification

| Check | Command | Exit Code | Notes |
| :--- | :--- | :--- | :--- |
| Go Vet | `go -C app vet .` | `0` | Clean |
| Go Unit Tests | `go -C app test -v -count=1 ./...` | `0` | All 6 test suites (`TestResolveAPIBaseURL`, `TestBackoffDuration`, `TestAppRESTBindingsAndErrorDecoding`, `TestRunRealtimeLoopEmitsEventsAndShutsDownCleanly`, `TestServiceStartupNonBlockingAndShutdown`, `TestAppSearchManualBindings`) passed |
| Wails Client Check | `wails3 task client:check` | `0` | Regenerated bindings (`16 Methods`, `6 Enums`, `16 Models`, `2 Events`), `vue-tsc --noEmit`, `vite build --mode production`, and `go build .` all succeeded; `app/app.exe` removed afterward |
| Live Integration Smoke Test | `postgres` + `bin/hrig-server.exe` + `go -C bot run .` | `0` | `GET /health` returned `store: "postgres"`; `bot` populated `57` live snapshot readings; `GET /api/hazards` returned `5` hazards; `GET /api/procedures` returned `20` procedures; `GET /api/search?q=hypoxia&tab=all` returned `2` matched procedures (`hypoxia_cabin_o2`); `GET /api/search?q=&tab=all` returned `3` astronauts, `20` procedures, `5` hazards, and `23` metrics. All processes stopped cleanly afterward (`db:status` confirmed PostgreSQL stopped). |

## UX & Edge-Case Notes

- **HomeView Search Trigger Focus Loop Prevention:** Rendering the `HomeView.vue` search trigger as a `<button>` styled with `.field` (instead of an `<input>` navigating on `focus`) prevents unintended redirect loops when pressing `Esc` in `/search` to return via `router.back()` and browsers restore focus to the last active element.
- **Reka UI Combobox Single-Select Toggle:** `ListboxRoot` defaults to `selectionBehavior: "toggle"`, meaning activating the same already-selected item a second time emits `undefined`. `SearchView.vue` handles `undefined` in `@update:model-value` by falling back to `highlightedKey` so re-selecting or pressing `Enter` on a procedure keeps it pinned cleanly.
- **Zero Hardcoded Clinical Data:** All procedure steps, warnings, equipment lists, escalation levels, hazard summaries, and metric warning/critical bounds are sourced dynamically from backend RPCs (`Search`, `GetMetrics`, `GetAstronauts`, `GetSnapshot`).
