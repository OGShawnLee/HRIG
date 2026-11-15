# 016E — Verification Spec

Every check is binary (`PASS` / `FAIL`). All checks must pass before ticket `016E` is accepted.

## 1. Satoshi Typography (`font-sans`)

- [x] `SearchView.vue` and `SearchFeaturedEntry.vue` contain zero occurrences of `font-mono`.
- [x] The Omnibox root container (`<main>`), `ComboboxInput` (`#hrig-omnibox-input`), `<aside>`, and every `<kbd>` element in `SearchView.vue` explicitly include `font-sans`.
- [x] `app/frontend/uno.config.ts` retains `sans: "Satoshi"` unmodified.

## 2. Rounded UI Geometry

- [x] The outer Omnibox container in `SearchView.vue` uses `rounded-2xl` and `overflow-hidden`.
- [x] `SearchView.vue` applies `rounded-xl` to category tab buttons, the `Close` button, the `Reset filters` button, and all four `ComboboxItem` groups (`Crew`, `Medical Procedures`, `NASA Hazards`, `Metrics`).
- [x] `SearchView.vue` applies `rounded-lg` or `rounded-md` to the `Clear` button, `Retry` button, metadata badges, and all `<kbd>` elements (`Ctrl + K`, `↑`, `↓`, `Enter`, `Esc`).
- [x] `SearchFeaturedEntry.vue` contains zero occurrences of `rounded-none` on any element (including all three `.button-main` CTA buttons).
- [x] `SearchFeaturedEntry.vue` applies `rounded-xl` to all inner detail/section panels and checklist step `<li>` cards, and `rounded-lg` / `rounded-md` to step index badges, equipment pills, escalation/status badges, and clickable metric/hazard filter chips.

## 3. All-Crew Metric Readouts in Search Results

- [x] `SearchView.vue` renders live per-member values + status icons for every astronaut in `activeAstronauts` inside crew `Metric` cards (`metricType !== 2`), and renders the live `Cabin / Suit` sensor reading inside environmental `Metric` cards (`metricType === 2`).
- [x] When `query.trim().length > 0`, `SearchView.vue` renders `Metrics` above `Medical Procedures` and keeps `entriesByKey` insertion order synchronized with visual group order so the first matching `Metric` is automatically highlighted in `SearchFeaturedEntry.vue` with all active crew readings visible.

## 4. Build, Type-Check & Functionality

- [x] `wails3 task client:check` exits with `0` (`bindings` -> `vue-tsc --noEmit` -> `vite build --mode production` -> `go build .`), with any untracked `app/app.exe` removed afterward.
- [x] Reka UI `ComboboxRoot`, `ComboboxInput`, `ComboboxContent`, `ComboboxViewport`, `ComboboxGroup`, `ComboboxLabel`, `ComboboxItem`, and `ComboboxEmpty` props/events remain intact in `SearchView.vue`.

## 5. Scope & Ground Rules

- [x] Only `app/frontend/src/views/SearchView.vue`, `app/frontend/src/components/SearchFeaturedEntry.vue`, and `.tickets/016E-omnibox-satoshi-and-rounded-ui/changelog.md` are created or modified (pre-existing untracked `bot-n/` and `DECISIONES_METRICAS.md` remain untouched).
- [x] `.tickets/016E-omnibox-satoshi-and-rounded-ui/changelog.md` is written and accurate.
