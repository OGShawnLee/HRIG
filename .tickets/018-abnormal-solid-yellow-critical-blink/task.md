# 018 — Make Abnormal Solid Yellow and Reserve Red Blinking Strictly for Critical

| Field | Value |
| :--- | :--- |
| **ID** | `018-abnormal-solid-yellow-critical-blink` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/views/dashboard-principal.vue` |

## Context

In `dashboard-principal.vue`, ticket 017 added blinking to both `.vital-frame-warn` (`abnormal`) and `.vital-frame-crit` (`critical`), and left `row.subtitle` (`BP`, `SENSOR`, `BODY`) and `row.rangeLabel` in default white/grey even during alerts. The user explicitly requested:
- When an alert is `abnormal`, both the card contour and all letters inside the card must be **solid yellow** (no blinking).
- **Red and blinking** contour and letters must occur **only** when the alert is `critical`.

## Grounding Rules

- Modify only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/018-abnormal-solid-yellow-critical-blink/`.
- Do NOT break nominal waveform morphology/colors when readings are `normal`.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Update `app/frontend/src/views/dashboard-principal.vue`:
   - In `.vital-frame-warn`, remove the blink animation (`vital-frame-blink-warn`) so an `abnormal` card contour is **solid, non-blinking yellow** (`border-width: 2px; border-color: #facc15; box-shadow: 0 0 0 1px rgba(250, 204, 21, 0.35), 0 0 10px rgba(250, 204, 21, 0.16);`).
   - Update all card text bindings in the template (`row.title`, `row.subtitle`, `row.rangeLabel`, `"BP"` label, `row.unit`, `row.value`, `row.subline`, and the right-hand help icon) so that:
     - When `row.status === ReadingStatus.StatusAbnormal`, every text/letter element on that card is yellow (`text-warn`) and **does not blink**.
     - When `row.status === ReadingStatus.StatusCritical`, every text/letter element on that card is red (`text-crit`) AND **blinks** (e.g., via `.vital-text-crit` animation matching `.vital-frame-crit`).
   - Keep `.vital-frame-crit` blinking red (`vital-frame-blink-crit`) strictly for `StatusCritical`.
2. Run `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks, write `changelog.md`, and report back.
