# 024 — Cancel Adaptive Grid Reflow and Apply Uniform Proportional Viewport UI Scaling

| Field | Value |
| :--- | :--- |
| **ID** | `024-viewport-proportional-ui-scaling` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/views/HomeView.vue`, `app/frontend/src/components/CrewCard.vue`, `app/frontend/src/components/EnvironmentCard.vue`, `app/frontend/src/components/MedicalManualButton.vue`, `app/frontend/src/views/dashboard-principal.vue`, `app/frontend/src/App.vue` |

## Context

The user prefers the exact original fixed-proportion HUD layout (`b7a217a`, preserving tickets `018`, `021`, and `022`) rather than the adaptive grid reflow from ticket `023`, and wants the entire UI to scale proportionally to fit the screen size so the interface always keeps the exact same proportions regardless of display resolution.

## Grounding Rules

- Revert ONLY the `023` layout changes in `HomeView.vue`, `CrewCard.vue`, `EnvironmentCard.vue`, `MedicalManualButton.vue`, and `dashboard-principal.vue` back to `b7a217a` (keeping `018` solid yellow/critical blink styles and `021`/`022` centered help button intact).
- Implement uniform root-`rem` proportional scaling in `app/frontend/src/App.vue` without touching `bot/` or `server/`.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Restore `app/frontend/src/views/HomeView.vue`, `app/frontend/src/components/CrewCard.vue`, `app/frontend/src/components/EnvironmentCard.vue`, `app/frontend/src/components/MedicalManualButton.vue`, and `app/frontend/src/views/dashboard-principal.vue` to commit `b7a217a` (`git checkout b7a217a -- ...`).
2. Update `app/frontend/src/App.vue` to compute and apply proportional root `font-size` scaling on mount, window `resize`, and cleanup on unmount:
   - Reference canvas: `DESIGN_WIDTH = 1120`, `DESIGN_HEIGHT = 860`, base `16px`.
   - When `window.innerWidth < 768` (mobile/narrow layout), clear custom `fontSize` (`""`) so mobile stacks cleanly at readable default typography.
   - When `window.innerWidth >= 768`, compute `scale = Math.max(0.7, Math.min(2.25, Math.min(window.innerWidth / DESIGN_WIDTH, window.innerHeight / DESIGN_HEIGHT)))` and set `document.documentElement.style.fontSize = "${(16 * scale).toFixed(2)}px"`.
   - Restore `document.documentElement.style.fontSize = ""` in `onBeforeUnmount`.
3. Run `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks verifying `018`, `021`, `022` remain intact and `024` root scaling is active, write `changelog.md`, and report back.
