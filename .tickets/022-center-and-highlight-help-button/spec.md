# Spec — 022 Center & Highlight Metric Help Button

## Functional Requirements

1. The `?` (`HelpCircle`) button on every vital sign card in `dashboard-principal.vue` is vertically and horizontally centered (`items-center justify-center`).
2. The button column is visually distinct and legible (`36px`/`sm:44px` track, `border-l border-solid border-lineDim`, `w-5 h-5`, nominal color `text-main`), while inheriting yellow (`text-warn`) on `abnormal` and blinking red (`text-crit vital-text-crit`) on `critical`.
3. Clicking `?` continues to navigate to `{ name: "search", query: { tab: "procedures", q: row.manualQuery } }`.

## Acceptance Criteria

- `rg -n "items-center justify-center.*border-l" app/frontend/src/views/dashboard-principal.vue` matches the help button.
- `rg -n "HelpCircle class=\"w-5 h-5\"" app/frontend/src/views/dashboard-principal.vue` matches.
- `bun run --cwd app/frontend type-check` exits `0`.
- `bun run --cwd app/frontend build` exits `0`.
