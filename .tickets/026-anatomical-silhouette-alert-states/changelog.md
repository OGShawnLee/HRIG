# 026 — Changelog: Interactive Anatomical Alert States on Astronaut Silhouette

## Files Modified

- `app/frontend/src/views/dashboard-principal.vue`
  - Extracted the 5 vital card severity evaluations (`cardiac`, `oxygenation`, `pressure`, `radiation`, `temperature`) into standalone `computed` `cardStatuses` using `resolveCardStatus`, shared with `vitalRows` so the astronaut silhouette updates independently of `metricFilter`.
  - Added `silhouetteAlertClass(status: CardStatus): string` returning `"body-alert-warn"` for `StatusAbnormal`, `"body-alert-crit"` for `StatusCritical`, and `""` otherwise.
  - Updated the `<aside>` astronaut silhouette SVG (`viewBox="0 0 160 360"`) with anatomical low-poly regions bound to live severity:
    - **Whole Body (`radiation`)**: outer silhouette and wireframe paths apply `silhouetteAlertClass(cardStatuses.radiation)` (wireframe preserves `fill: none !important` via `.silhouette-wire`).
    - **Head (`temperature`)**: faceted head path applies `silhouetteAlertClass(cardStatuses.temperature) || silhouetteAlertClass(cardStatuses.radiation)`.
    - **Lungs (`oxygenation`)**: left/right low-poly lung polygons (`x=57..77` & `83..103`, `y=96..139`) and airway lines (`M80 84v17l-7 8M80 101l7 8`) apply `silhouetteAlertClass(cardStatuses.oxygenation)`.
    - **Pressure Region (`pressure`)**: bilateral brachial cuff bands (`x=34..51` & `115..132`, `y=114..136`) and central vascular torso band (`y=142..151`) apply `silhouetteAlertClass(cardStatuses.pressure)`.
    - **Heart (`cardiac`)**: faceted low-poly heart polygon at `(77..96, 107..129)` applies `silhouetteAlertClass(cardStatuses.cardiac)`.
  - Added scoped CSS classes `.silhouette-organ`, `.silhouette-wire`, `.body-alert-warn` (`@keyframes body-blink-warn 1.2s ease-in-out infinite`, `#facc15`), and `.body-alert-crit` (`@keyframes body-blink-crit 0.65s ease-in-out infinite`, `#f87171`), leaving Ticket `018` `.vital-frame-warn` / `.vital-frame-crit` rules intact.
- `.tickets/026-anatomical-silhouette-alert-states/task.md`
  - Updated status from `in_progress` to `completed`.
- `.tickets/026-anatomical-silhouette-alert-states/changelog.md`
  - Created this changelog.

## Verification Results

- `rg -n "silhouetteAlertClass|body-alert-warn|body-alert-crit" app/frontend/src/views/dashboard-principal.vue`: **PASS** (matched helper, all 5 anatomical regions at lines 1034, 1038, 1045, 1051, 1061, 1071, and CSS classes `.body-alert-warn` / `.body-alert-crit` at lines 1175, 1181)
- `rg -n "vital-frame-blink-warn" app/frontend/src/`: **PASS** (0 matches, Ticket `018` preserved)
- `rg -n "new WebSocket|fetch\(" app/frontend/src/`: **PASS** (0 matches)
- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`)
- `bun run --cwd app/frontend build` (`vite build --mode production`): **PASS** (exit code `0`; 2429 modules transformed, built in 3.06s: `dist/index.html` 0.50 kB, `dist/assets/index-BAc3xwH2.css` 22.12 kB, `dist/assets/index-DbqhDEJh.js` 314.18 kB)
- Scope check (`git status --short`): Only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/026-anatomical-silhouette-alert-states/` modified/untracked; no git commit or push performed.
