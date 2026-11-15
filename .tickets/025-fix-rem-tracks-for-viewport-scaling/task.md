# 025 — Convert Hardcoded `px` Tracks & Text in Dashboard to `rem` for Viewport Scaling

| Field | Value |
| :--- | :--- |
| **ID** | `025-fix-rem-tracks-for-viewport-scaling` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/views/dashboard-principal.vue`, `app/frontend/src/components/CrewCard.vue` |

## Context

Because `App.vue` scales the UI proportionally via root `font-size` (`rem`), hardcoded pixel grid tracks (`200px`, `140px`, `168px`, `36px`, `44px`) and pixel font sizes (`text-[11px]`, `text-[10px]`) in `dashboard-principal.vue` stayed fixed in pixels while `text-3xl` numbers and `w-36` silhouette SVGs scaled with `rem`. In fullscreen mode this caused numeric values (`9...`, `1...`, `0....`, `3...`) to truncate with ellipses and the astronaut silhouette to overflow its right-side card.

## Grounding Rules

- Modify only `app/frontend/src/views/dashboard-principal.vue`, `app/frontend/src/components/CrewCard.vue`, and `.tickets/025-fix-rem-tracks-for-viewport-scaling/`.
- Keep the exact `024` scaling geometry, `018` solid yellow / red blinking alert rules, and `021`/`022` centered help buttons intact.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Update `app/frontend/src/views/dashboard-principal.vue`:
   - Replace the aside track `md:grid-cols-[1fr_200px]` with `md:grid-cols-[minmax(0,1fr)_13rem]` so the right astronaut card scales proportionally in `rem` and always contains its `w-32 sm:w-36` silhouette with comfortable padding.
   - Replace `grid-cols-[minmax(0,1fr)_140px_36px] sm:grid-cols-[minmax(0,1fr)_168px_44px]` with `grid-cols-[minmax(0,1fr)_9.5rem_2.25rem] sm:grid-cols-[minmax(0,1fr)_11.5rem_2.75rem]` so the readout and `?` columns scale in `rem` alongside `App.vue` root scaling.
   - Convert micro label font sizes inside vital cards from fixed `text-[11px]` / `text-[10px]` to `rem`-based classes (`text-[0.6875rem]` and `text-[0.625rem]`).
   - On `<strong>{{ row.value }}</strong>`, use `text-2xl sm:text-[1.75rem] font-bold leading-none whitespace-nowrap` (removing `truncate`) so numeric readouts (`98`, `112/74`, `0.245`, `36.9`) are never clipped to ellipses at any screen resolution.
2. Update `app/frontend/src/components/CrewCard.vue`:
   - Convert `top-[29px]` and `text-[10px]` inside the crew portrait badge to `top-[1.8125rem]` and `text-[0.625rem]` so fallback initials and nationality badges scale cleanly with `rem` too.
3. Run `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks verifying zero `px]` structural tracks in `dashboard-principal.vue`, write `changelog.md`, and report back.
