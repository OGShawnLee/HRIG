# 017 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Card Contour & Numeric Highlight While Alert Persists
- [x] `app/frontend/src/views/dashboard-principal.vue` evaluates worst severity across each card's primary metric, paired subline metric (`hrv_rmssd`, `cabin_o2`, `cabin_pressure`, `radiation_dose_cum`, `cabin_temp`), active `alerts`, and `lastTriggeredAlert`.
- [x] While `abnormal` values persist (even after closing the alert notification via `Close Alert`), the matching graph card contour renders in yellow (`border-warn` + `.vital-frame-warn` with `border-width: 2px`) AND its numeric value renders in yellow (`text-warn`).
- [x] While `critical` values persist, the matching graph card contour renders in red (`border-crit` + `.vital-frame-crit` with `border-width: 2px`) AND its numeric value renders in red (`text-crit`).

## Scope & Build Checks
- [x] Only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/017-alert-card-contour-and-value-highlight/` are modified for this ticket.
- [x] Zero `new WebSocket` or `fetch(` calls exist in `app/frontend/src/`.
- [x] `bun run --cwd app/frontend type-check` exits `0`.
- [x] `bun run --cwd app/frontend build` exits `0` and outputs `app/frontend/dist/index.html`.
- [x] `.tickets/017-alert-card-contour-and-value-highlight/changelog.md` exists and records files touched and exact command results.
