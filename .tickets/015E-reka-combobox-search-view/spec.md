# 015E — Verification Spec

Every check is binary (`PASS` / `FAIL`). All checks must pass before ticket `015E` is accepted.

## 1. Wails Go Client, Bindings & Tests

- [x] `app/client.go` implements `GetHazards`, `GetProcedures`, and `Search`, and `app/app.go` exposes bound `GetHazards`, `GetProcedures`, and `Search` methods returning `[]model.SpaceHazard`, `[]model.MedicalProcedure`, and `model.SearchResults` (with non-nil slices).
- [x] `app/frontend/bindings/hrig/app/app.ts` and `app/frontend/bindings/hrig/shared/model/models.ts` include regenerated TypeScript bindings for `GetHazards`, `GetProcedures`, `Search`, `SpaceHazard`, `MedicalProcedure`, `SearchResults`, and `EscalationLevel`.
- [x] `go -C app vet .` and `go -C app test -v -count=1 ./...` exit with `0` (covering `GetHazards`, `GetProcedures`, and `Search` against `httptest.Server`).

## 2. Reka UI Omnibox Search Page & Triggers

- [x] `app/frontend/src/router/index.ts` defines route `/search` (`name: "search"`) pointing to `SearchView.vue` while preserving `/`, `/dashboard-principal`, `/support-alert`, `/warning-alert`, and `/pagina-principal`.
- [x] `App.vue` preserves `WindowBar`, `ViewNavBar`, `DemoAlertStrip`, and `TELEMETRY_STATUS_EVENT` while handling global `Ctrl+K` / `Cmd+K` to navigate to `/search` (or dispatch `hrig:focus-omnibox` when already on `/search`).
- [x] `ViewNavBar.vue` includes a `Search` link to `/search`, and `HomeView.vue` navigates to `/search` on clicking/focusing the Search bar and `/search?tab=procedures` on clicking `Medical Manual`.
- [x] `app/frontend/src/views/SearchView.vue` imports and uses `ComboboxRoot`, `ComboboxInput`, `ComboboxContent`, `ComboboxViewport`, `ComboboxGroup`, `ComboboxLabel`, `ComboboxItem`, and `ComboboxEmpty` from `"reka-ui"` with `:open="true"`, `:ignore-filter="true"`, `position="inline"`, and discriminated `ComboboxItem` values.
- [x] `SearchView.vue` + `SearchFeaturedEntry.vue` render HRIG's dark HUD palette (`bg-ground` `#000000`, `border-line`/`border-lineDim`, semantic status tokens), tabs (`All`, `Crew`, `Procedures`, `Hazards & Metrics`), left `FEATURED ENTRY` preview defaulting to the first visible result and updating on `@highlight`/`Enter` (showing numbered procedure steps, equipment, warnings, escalation badge + icon, or live crew vitals + `/dashboard-principal?astronaut=<id>` link), right grouped results with zero-padded counts, and bottom footer (`↑ ↓ Navigate`, `Enter Select`, `Esc Close`).
- [x] No hardcoded clinical threshold numbers or medical step arrays live in `app/frontend/src/**`.

## 3. Build & Live Integration Verification

- [x] `wails3 task client:check` exits with `0` (`bindings` -> `vue-tsc --noEmit` -> `vite build` -> `go build .`).
- [x] Live check with `postgres`, `bin/hrig-server.exe`, and `bot` confirms `GET /api/hazards` returns `5` NASA hazards, `GET /api/procedures` returns `20` procedures, and `GET /api/search?q=hypoxia&tab=all` returns matched entries, with all processes cleanly stopped afterward.

## 4. Scope & Ground Rules

- [x] No files outside `May create or edit` in `task.md` (`db/`, `server/`, `shared/`, `bot/`, `bot-n/`, `.routing/INPUT.md`, `CrewCard.vue`, `EnvironmentCard.vue`, `DemoAlertStrip.vue`, `dashboard-principal.vue`, `support-alert.vue`, `warning-alert.vue`) are modified.
- [x] `.tickets/015E-reka-combobox-search-view/changelog.md` is written and accurate.
