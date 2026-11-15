# 017 — Highlight Graph Card Contour & Numeric Value While Alert Readings Persist

| Field | Value |
| :--- | :--- |
| **ID** | `017-alert-card-contour-and-value-highlight` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/views/dashboard-principal.vue` |

## Context

Even after dismissing/closing the full alert view (`Close Alert`), while abnormal or critical values continue arriving (whether from live crew metrics, paired environmental metrics displayed on the card such as `cabin_pressure`/`cabin_o2`/`cabin_temp`, active `alerts`, or `lastTriggeredAlert`), the contour enclosing that specific metric graph card and the numeric readout representing it in `dashboard-principal.vue` must clearly display yellow (`abnormal`) or red (`critical`).

## Grounding Rules

- Modify only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/017-alert-card-contour-and-value-highlight/`.
- Do NOT break existing waveform morphology/colors when readings are `normal`.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Update `app/frontend/src/views/dashboard-principal.vue`:
   - Compute an effective per-card severity (`resolveCardStatus`) combining:
     - The card's primary metric reading status (`heart_rate`, `spo2`, `bp_systolic`/`bp_diastolic`, `radiation_rate`, `core_temp`),
     - Any paired subline metric reading shown in that card (`hrv_rmssd` for `cardiac`, `cabin_o2` for `oxygenation`, `cabin_pressure` for `pressure`, `radiation_dose_cum` for `radiation`, `cabin_temp` for `temperature`),
     - Any active unacknowledged alert in `alerts` or active `lastTriggeredAlert` matching any of the card's metrics (for the selected astronaut, environmental metrics, or demo alerts without an `astronautId`), taking the highest severity (`critical` > `abnormal` > `stale` > `normal` > `fallback`).
   - Also bind the active `row.status` reliably to:
     - The enclosing `<article>` frame (`cardFrameClass(row.status)` -> bright blinking yellow `border-2 border-warn vital-frame-warn` for `abnormal`, bright blinking red `border-2 border-crit vital-frame-crit` for `critical`), and
     - The displayed metric numeric readout (`<strong>{{ row.value }}</strong>`) + unit/title (`statusTextClass(row.status)` -> `text-warn` for `abnormal`, `text-crit` for `critical`).
   - Ensure `.vital-frame-warn` and `.vital-frame-crit` use `border-width: 2px` with crisp yellow (`#facc15`) / red (`#f87171`) border & glow `@keyframes` so the card contour clearly stands out over `.hud-card` while keeping inner numbers solid and legible.
2. Run `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks, write `changelog.md`, and report back.
