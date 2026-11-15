# 022 — Changelog: Center & Enhance Visibility of Metric Help (`?`) Button

## Files Modified

- `app/frontend/src/views/dashboard-principal.vue`
  - Widened the right button grid column on `<article v-for="row in vitalRows">` from `28px`/`sm:32px` to `36px`/`sm:44px`.
  - Centered the help `<button type="button">` contents (`flex items-center justify-center` instead of `items-end pb-2`) and added a subtle left border and surface tint (`border-l border-solid border-lineDim bg-surface/50`).
  - Changed the nominal status class in `cardSecondaryTextClass(row.status, 'text-main')` from `'text-dull'` to `'text-main'` while preserving `text-warn` on `abnormal` and `text-crit vital-text-crit` on `critical`.
  - Enlarged `<HelpCircle>` from `w-3.5 h-3.5` to `class="w-5 h-5"` with `:stroke-width="2.2"`.
- `.tickets/022-center-and-highlight-help-button/changelog.md`
  - Created this changelog.

## Verification Results

- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`)
- `bun run --cwd app/frontend build` (`vite build --mode production`): **PASS** (exit code `0`; 2429 modules transformed, built in 2.15s: `dist/index.html` 0.50 kB, `dist/assets/index-DzKseksR.css` 21.05 kB, `dist/assets/index-WzxrT_4h.js` 312.22 kB)
- `rg -n "items-center justify-center.*border-l" app/frontend/src/views/dashboard-principal.vue`: **PASS** (1 match: line 984)
- `rg -n 'HelpCircle class="w-5 h-5"' app/frontend/src/views/dashboard-principal.vue`: **PASS** (1 match: line 990)
- `rg -n "new WebSocket|fetch\(" app/frontend/src/`: **PASS** (0 matches)
- Scope check (`git status --short`): Only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/022-center-and-highlight-help-button/` modified/untracked; no git commit or push performed.
