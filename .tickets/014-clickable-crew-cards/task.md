# 014 — Make entire astronaut crew cards clickable in pagina-principal

| Field | Value |
| :--- | :--- |
| **ID** | `014` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Frontend Crew Card Accessibility Builder |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `app/frontend/src/views/pagina-principal.vue`
  - `.tickets/014-clickable-crew-cards/changelog.md`
- **Must not touch:**
  - `.routing/**` (including `INPUT.md`, `CONTEXT.md`, `DOMAIN.md`, `SPECIFICATION.md`, `DESIGN.md`, `PROGRESS.md`)
  - `.claude/**`
  - `docs/**`, `db/**`, `server/**`, `bot/**`, `shared/**`
  - `app/frontend/public/crew/**`
  - `app/frontend/src/components/**`
  - `app/frontend/src/composables/useTelemetryNavigation.ts`
  - `app/frontend/src/router/index.ts`
  - `app/frontend/src/views/HomeView.vue`, `app/frontend/src/views/dashboard-principal.vue`, `app/frontend/src/views/support-alert.vue`, `app/frontend/src/views/warning-alert.vue`
  - `.tickets/014-clickable-crew-cards/task.md`
  - `.tickets/014-clickable-crew-cards/spec.md`

## Inputs

Read this file before acting:
1. [`app/frontend/src/views/pagina-principal.vue`](../../app/frontend/src/views/pagina-principal.vue) (existing `.crew-card` markup, `.crew-portrait-btn`, and styles)

## Grounding & Anti-Hallucination Constraints

- Keep the exact astronaut names, portrait images (`/crew/sam-sonia.jpg`, `/crew/patrick-cuckburn.jpg`), and vital values unchanged.
- Do **not** nest interactive elements (`<button>` inside `role="button"`): replace the inner `<button class="crew-portrait-btn">` with direct card-level activation on `<article class="crew-card">`.
- Preserve keyboard accessibility (`tabindex="0"`, `role="button"`, `aria-label`, `@keydown.enter.prevent`, `@keydown.space.prevent`, and `:focus-visible` outline).
- Do **not** commit or push any git changes.
- If `bun run --cwd app/frontend type-check` or `bun run --cwd app/frontend build` fails, **stop** and record the exact command and output in `.tickets/014-clickable-crew-cards/changelog.md`.

## Task

1. Update both `<article class="crew-card">` elements in `app/frontend/src/views/pagina-principal.vue`:
   - On Sam Sonia's card:
     - Add `role="button"`, `tabindex="0"`, `aria-label="Open Sam Sonia primary vitals"`, `@click="openPrimaryVitals('sam')"`, `@keydown.enter.prevent="openPrimaryVitals('sam')"`, `@keydown.space.prevent="openPrimaryVitals('sam')"`.
     - Remove the inner `<button type="button" class="crew-portrait-btn" ...>` wrapper around `<div class="crew-portrait">` so `<div class="crew-portrait">` sits directly inside `<div class="crew-heading">`.
   - On Patrick Cuckburn's card:
     - Add `role="button"`, `tabindex="0"`, `aria-label="Open Patrick Cuckburn primary vitals"`, `@click="openPrimaryVitals('patrick')"`, `@keydown.enter.prevent="openPrimaryVitals('patrick')"`, `@keydown.space.prevent="openPrimaryVitals('patrick')"`.
     - Remove the inner `<button type="button" class="crew-portrait-btn" ...>` wrapper around `<div class="crew-portrait">`.
2. Update CSS in `app/frontend/src/views/pagina-principal.vue`:
   - Remove the unused `.crew-portrait-btn` and `.crew-portrait-btn:focus-visible` rules.
   - Add `cursor: pointer; transition: border-color 150ms ease, background-color 150ms ease;` to `.crew-card`.
   - Add `.crew-card:hover` (`border-color: #00a7ff; background: #0a121e;`) and `.crew-card:focus-visible` (`outline: 2px solid #00a7ff; outline-offset: 2px; border-color: #00a7ff;`).
3. Run `bun run --cwd app/frontend type-check` and `bun run --cwd app/frontend build` from repository root and confirm both exit `0`.
4. Write `.tickets/014-clickable-crew-cards/changelog.md` listing files created/modified, exact commands run with exit codes, and any observations.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/014-clickable-crew-cards/changelog.md` listing:
- Files created and modified
- Verification commands executed and their exit codes/outputs
- Any blockers or assumptions avoided
