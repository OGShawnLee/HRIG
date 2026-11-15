# Spec — 024 Cancel Adaptive Reflow & Apply Viewport Scaling

## Functional Requirements

1. `HomeView.vue`, `CrewCard.vue`, `EnvironmentCard.vue`, `MedicalManualButton.vue`, and `dashboard-principal.vue` match `b7a217a` layout geometry verbatim (keeping `018` abnormal/critical alert styles and `021`/`022` centered help buttons).
2. `App.vue` dynamically updates `document.documentElement.style.fontSize` based on `Math.min(window.innerWidth / 1120, window.innerHeight / 860)` (clamped `[0.7, 2.25]`) for viewports `>= 768px`, and resets to default when `< 768px` or unmounted.
3. Because UnoCSS dimensions, spacing, icons, and typography use `rem`, the entire interface (`WindowBar`, `ViewNavBar`, `HomeView`, `dashboard-principal`, `SearchView`, alerts, `DemoAlertStrip`) scales uniformly while preserving exact aspect proportions and crisp vector rendering.

## Acceptance Criteria

- `git diff b7a217a -- app/frontend/src/views/HomeView.vue app/frontend/src/components/CrewCard.vue app/frontend/src/components/EnvironmentCard.vue app/frontend/src/components/MedicalManualButton.vue app/frontend/src/views/dashboard-principal.vue` produces empty output.
- `rg -n "updateViewportScale|DESIGN_WIDTH" app/frontend/src/App.vue` matches.
- `rg -n "new WebSocket|fetch\(" app/frontend/src/` returns zero matches.
- `bun run --cwd app/frontend type-check` exits `0`.
- `bun run --cwd app/frontend build` exits `0`.
