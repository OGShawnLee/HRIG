# 021 — Changelog: Link Metric Help Icons (`?`) to Relevant Medical Manual Section

## Files Modified

- `app/frontend/src/views/dashboard-principal.vue`
  - Imported `useRouter` from `"vue-router"` alongside `useRoute` and initialized `const router = useRouter()`.
  - Added `openMedicalManual(manualQuery: string)` to navigate via `void router.push({ name: "search", query: { tab: "procedures", q: manualQuery } })`.
  - Added `manualQuery: string` to `VitalRow` and populated each vital card row:
    - `cardiac` (`CARDIAC RHYTHM`): `"heart_rate"`
    - `oxygenation` (`OXYGENATION`): `"spo2"`
    - `pressure` (`PRESSURE`): `"pressure"`
    - `radiation` (`RADIATION`): `"radiation"`
    - `temperature` (`TEMPERATURE`): `"core_temp"`
  - Replaced the static `<div>` wrapper around `<HelpCircle>` on each vital card with an accessible `<button type="button">` bound to `@click="openMedicalManual(row.manualQuery)"`, preserving `:class="cardSecondaryTextClass(row.status, 'text-dull')"` and adding hover/focus feedback (`cursor-pointer transition-colors hover:bg-surface hover:text-main focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-line`), `:aria-label`, and `:title`.
- `.tickets/021-metric-help-links-to-medical-manual/changelog.md`
  - Created this changelog.

## Verification Results

- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`)
- `bun run --cwd app/frontend build` (`vite build --mode production`): **PASS** (exit code `0`; 2429 modules transformed, built in 2.07s: `dist/index.html` 0.50 kB, `dist/assets/index-CzngCAae.css` 21.04 kB, `dist/assets/index-CwP-DOxV.js` 312.16 kB)
- `rg -n "openMedicalManual" app/frontend/src/views/dashboard-principal.vue`: **PASS** (2 matches: line 51 function declaration, line 988 `@click` binding)
- `rg -n 'tab: "procedures"' app/frontend/src/views/dashboard-principal.vue`: **PASS** (1 match: line 54)
- `rg -n "new WebSocket|fetch\(" app/frontend/src/`: **PASS** (0 matches)
- Scope check (`git status -s`): Only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/021-metric-help-links-to-medical-manual/` modified/untracked; no git commit or push performed.
