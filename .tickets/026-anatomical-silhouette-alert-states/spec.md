# Spec — 026 Interactive Anatomical Silhouette Alert States

## Functional Requirements

1. Temperature alert (`abnormal` / `critical`) highlights and blinks the astronaut **head** in yellow (`body-alert-warn`) or red (`body-alert-crit`).
2. Radiation alert (`abnormal` / `critical`) highlights and blinks the **entire body** silhouette in yellow or red.
3. Oxygenation alert (`abnormal` / `critical`) highlights and blinks the **lungs** in yellow or red.
4. Blood pressure alert (`abnormal` / `critical`) highlights and blinks the **representative brachial/vascular pressure region** in yellow or red.
5. Cardiac rhythm alert (`abnormal` / `critical`) highlights and blinks the **heart** in yellow or red.
6. Nominal cards (`normal` / `stale`) leave silhouette regions non-blinking, and vital cards on the left retain Ticket `018` rules (solid yellow on `abnormal`, red blinking on `critical`).

## Acceptance Criteria

- `rg -n "silhouetteAlertClass|body-alert-warn|body-alert-crit" app/frontend/src/views/dashboard-principal.vue` matches all 5 anatomical regions and CSS keyframes.
- `rg -n "vital-frame-blink-warn" app/frontend/src/` returns zero matches (`018` preserved).
- `rg -n "new WebSocket|fetch\(" app/frontend/src/` returns zero matches.
- `bun run --cwd app/frontend type-check` exits `0`.
- `bun run --cwd app/frontend build` exits `0`.
