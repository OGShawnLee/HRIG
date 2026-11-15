# Spec — 023 Adaptive Full-Viewport Responsive Layout

## Functional Requirements

1. `HomeView.vue` stretches smoothly up to `1680px` wide and fills available viewport height without `mt-auto` vertical voids or `max-w-sm` crew card caps.
2. `CrewCard.vue`, `EnvironmentCard.vue`, and `MedicalManualButton.vue` stretch proportionally to fill their grid tracks cleanly across mobile (`<640px`), tablet (`md`), laptop (`lg`), and desktop (`xl`+) viewports.
3. `dashboard-principal.vue` adapts both horizontally (`max-w-[1680px]`, wider right aside at `lg`/`xl`) and vertically (`flex-1` vital cards and proportional waveform SVGs) while preserving all `018` alert frame/text rules and `021`/`022` centered help buttons.

## Acceptance Criteria

- `rg -n "max-w-sm|w-44 h-44|sm:w-92" app/frontend/src/views/HomeView.vue app/frontend/src/components/MedicalManualButton.vue app/frontend/src/components/EnvironmentCard.vue` returns zero matches.
- `rg -n "max-w-4xl" app/frontend/src/views/HomeView.vue app/frontend/src/views/dashboard-principal.vue` returns zero matches.
- `rg -n "new WebSocket|fetch\(" app/frontend/src/` returns zero matches.
- `bun run --cwd app/frontend type-check` exits `0`.
- `bun run --cwd app/frontend build` exits `0`.
