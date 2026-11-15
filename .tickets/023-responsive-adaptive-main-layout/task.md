# 023 — Adaptive Full-Viewport Responsive Layout for Main Views

| Field | Value |
| :--- | :--- |
| **ID** | `023-responsive-adaptive-main-layout` |
| **Status** | `cancelled` |
| **Scope** | `app/frontend/src/views/HomeView.vue`, `app/frontend/src/components/CrewCard.vue`, `app/frontend/src/components/EnvironmentCard.vue`, `app/frontend/src/components/MedicalManualButton.vue`, `app/frontend/src/views/dashboard-principal.vue` |

## Context

Both `HomeView.vue` (Home / Crew Overview) and `dashboard-principal.vue` (Primary Vitals) currently constrain container widths (`max-w-xl`, `max-w-4xl`, `max-w-5xl`, `max-w-sm` on `CrewCard`, fixed `w-44 h-44` on `MedicalManualButton`, fixed `sm:w-92` on `EnvironmentCard`) and push bottom panels with `mt-auto`, leaving large empty margins on wide desktop viewports and awkward vertical gaps. Every card and section should adapt smoothly to screen width and height while staying legible and scrollable on narrow/mobile screens.

## Grounding Rules

- Modify only the 5 scoped Vue files above and `.tickets/023-responsive-adaptive-main-layout/`.
- Preserve ALL existing alert styles (`018`), help button links (`021`, `022`), card status resolution, keyboard accessibility, and telemetry bindings.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Update `app/frontend/src/views/HomeView.vue`:
   - Wrap content in a fluid viewport-filling layout (`w-full max-w-[1680px] mx-auto flex-1 flex flex-col gap-4 sm:gap-6`).
   - Expand the top Search trigger width (`w-full`) cleanly.
   - Make the crew grid fluid (`w-full flex-1 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4 sm:gap-5 items-stretch`) with `<CrewCard class="w-full h-full" />` (removing `max-w-sm` and `justify-items-center`).
   - Make the bottom row an adaptive 2-part grid (`w-full grid grid-cols-1 sm:grid-cols-[minmax(200px,280px)_1fr] lg:grid-cols-[minmax(240px,320px)_1fr] gap-4 sm:gap-5 items-stretch`) without `mt-auto` vertical void.
2. Update `app/frontend/src/components/MedicalManualButton.vue`:
   - Replace fixed `w-44 h-44` with `w-full h-full min-h-36 p-5 flex flex-col items-center justify-center gap-3 ...`.
3. Update `app/frontend/src/components/EnvironmentCard.vue`:
   - Replace `w-full sm:w-92` with `w-full h-full p-5 sm:p-6 flex flex-col justify-between gap-4`.
   - Render `<dl>` as a responsive multi-column metric grid (`grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-x-6 gap-y-3`) where each metric item pairs `<dt>` and `<dd>` via `flex items-baseline justify-between gap-3 border-b border-solid border-lineDim/40 pb-1.5 min-w-0`.
4. Update `app/frontend/src/components/CrewCard.vue`:
   - Use `justify-between` on `<article>`, `flex-1 content-center` on `<dl>`, and `justify-end` on `<dd>` so wide cards align labels left and telemetry values right cleanly.
5. Update `app/frontend/src/views/dashboard-principal.vue`:
   - Replace `max-w-4xl` with `w-full max-w-[1680px] mx-auto` on `<header>` and main grid.
   - Make the main grid `flex-1 grid grid-cols-1 md:grid-cols-[minmax(0,1fr)_240px] lg:grid-cols-[minmax(0,1fr)_280px] xl:grid-cols-[minmax(0,1fr)_320px] gap-4 sm:gap-6 items-stretch`.
   - Give the 5 vital cards `flex-1 min-h-24` and proportional waveform SVGs (`w-full flex-1 min-h-14 max-h-28 px-1`), and make the right astronaut `<aside>` full-height (`h-full justify-between`) with responsive silhouette sizing (`w-32 sm:w-40 lg:w-48 max-h-[380px] h-auto`).
6. Run `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks, write `changelog.md`, and report back.
