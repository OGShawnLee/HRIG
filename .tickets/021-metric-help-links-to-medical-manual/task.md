# 021 — Link Metric Help Icons (`?`) to Relevant Medical Manual Section

| Field | Value |
| :--- | :--- |
| **ID** | `021-metric-help-links-to-medical-manual` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/views/dashboard-principal.vue` |

## Context

Each vital sign card in `dashboard-principal.vue` displays a `<HelpCircle>` (`?`) icon on its right edge, currently rendered as a non-interactive `<div>`. Clicking `?` should navigate directly to the Medical Manual (`search` route with `tab=procedures` and `q` set to the metric-relevant query, e.g., `pressure` for `PRESSURE`, `heart_rate` for `CARDIAC RHYTHM`, `spo2` for `OXYGENATION`, `radiation` for `RADIATION`, `core_temp` for `TEMPERATURE`).

## Grounding Rules

- Modify only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/021-metric-help-links-to-medical-manual/`.
- Preserve existing alert severity coloring (`cardSecondaryTextClass(row.status, 'text-dull')`), waveform rendering, and `018` styling behavior.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Update `app/frontend/src/views/dashboard-principal.vue`:
   - Import `useRouter` from `"vue-router"` alongside `useRoute` and initialize `const router = useRouter()`.
   - Add `manualQuery: string` to `VitalRow` and populate each card row:
     - `cardiac`: `"heart_rate"`
     - `oxygenation`: `"spo2"`
     - `pressure`: `"pressure"`
     - `radiation`: `"radiation"`
     - `temperature`: `"core_temp"`
   - Add `openMedicalManual(manualQuery: string)` which navigates via `void router.push({ name: "search", query: { tab: "procedures", q: manualQuery } })`.
   - Replace the static `<div ...><HelpCircle ... /></div>` wrapper on each vital card with an accessible `<button type="button" ... @click="openMedicalManual(row.manualQuery)">` preserving `cardSecondaryTextClass(row.status, 'text-dull')`, adding hover/focus feedback, descriptive `aria-label`, and `title`.
2. Run `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks, write `changelog.md`, and report back.
