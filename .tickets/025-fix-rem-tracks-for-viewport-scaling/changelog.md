# 025 — Changelog: Convert Hardcoded `px` Tracks & Text in Dashboard to `rem` for Viewport Scaling

## Files Modified

- `app/frontend/src/views/dashboard-principal.vue`
  - Converted aside layout grid track from `md:grid-cols-[1fr_200px]` to `md:grid-cols-[minmax(0,1fr)_13rem]` so the right astronaut silhouette card scales proportionally with `rem`.
  - Converted vital row card grid tracks from `grid-cols-[minmax(0,1fr)_140px_36px] sm:grid-cols-[minmax(0,1fr)_168px_44px]` to `grid-cols-[minmax(0,1fr)_9.5rem_2.25rem] sm:grid-cols-[minmax(0,1fr)_11.5rem_2.75rem]` so the readout and `?` manual button columns scale in `rem`.
  - Converted vital card micro typography classes from fixed `text-[11px]` and `text-[10px]` to `text-[0.6875rem]` and `text-[0.625rem]`.
  - Updated `<strong>{{ row.value }}</strong>` classes from `text-2xl sm:text-3xl font-bold leading-none truncate` to `text-2xl sm:text-[1.75rem] font-bold leading-none whitespace-nowrap` so full metric values render without ellipsis clipping.
- `app/frontend/src/components/CrewCard.vue`
  - Converted portrait badge fallback initials positioning from `top-[29px]` to `top-[1.8125rem]` and nationality badge text size from `text-[10px]` to `text-[0.625rem]` for uniform `rem` scaling.
- `.tickets/025-fix-rem-tracks-for-viewport-scaling/task.md`
  - Updated status from `in_progress` to `completed`.
- `.tickets/025-fix-rem-tracks-for-viewport-scaling/changelog.md`
  - Created this changelog.

## Verification Results

- `rg -n "grid-cols-.*px|text-\[[0-9]+px\]" app/frontend/src/views/dashboard-principal.vue`: **PASS** (0 matches)
- `rg -n "11\.5rem|13rem" app/frontend/src/views/dashboard-principal.vue`: **PASS** (matches on lines 914 and 919)
- `rg -n "top-\[[0-9]+px\]|text-\[[0-9]+px\]" app/frontend/src/components/CrewCard.vue`: **PASS** (0 matches)
- `rg -n "new WebSocket|fetch\(" app/frontend/src/`: **PASS** (0 matches)
- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`)
- `bun run --cwd app/frontend build` (`vite build --mode production`): **PASS** (exit code `0`; 2429 modules transformed, built in 2.45s: `dist/index.html` 0.50 kB, `dist/assets/index-Bs01nGIV.css` 21.20 kB, `dist/assets/index-_2D9xqbv.js` 312.69 kB)
- Scope check (`git status --short`): Only `app/frontend/src/views/dashboard-principal.vue`, `app/frontend/src/components/CrewCard.vue`, and `.tickets/025-fix-rem-tracks-for-viewport-scaling/` modified/untracked; no git commit or push performed.
