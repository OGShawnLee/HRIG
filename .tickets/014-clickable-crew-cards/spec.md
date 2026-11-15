# 014 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Card Interactivity & Accessibility
- [x] `app/frontend/src/views/pagina-principal.vue` binds `@click="openPrimaryVitals('sam')"`, `role="button"`, `tabindex="0"`, `@keydown.enter.prevent`, and `@keydown.space.prevent` on Sam Sonia's `<article class="crew-card">`.
- [x] `app/frontend/src/views/pagina-principal.vue` binds `@click="openPrimaryVitals('patrick')"`, `role="button"`, `tabindex="0"`, `@keydown.enter.prevent`, and `@keydown.space.prevent` on Patrick Cuckburn's `<article class="crew-card">`.
- [x] `app/frontend/src/views/pagina-principal.vue` contains zero `crew-portrait-btn` elements or CSS rules (no nested interactive controls).
- [x] `app/frontend/src/views/pagina-principal.vue` defines `cursor: pointer`, `:hover`, and `:focus-visible` styles on `.crew-card`.
- [x] Both `/crew/sam-sonia.jpg` and `/crew/patrick-cuckburn.jpg` `<img>` tags remain intact inside their respective `.crew-card`.

## Build, Type-Check & Scope
- [x] `bun run --cwd app/frontend type-check` exits `0`.
- [x] `bun run --cwd app/frontend build` exits `0` and outputs `app/frontend/dist/index.html`.
- [x] No files outside `app/frontend/src/views/pagina-principal.vue` and `.tickets/014-clickable-crew-cards/` were modified by this ticket.
- [x] No unapproved git commits or pushes were executed.
- [x] `.tickets/014-clickable-crew-cards/changelog.md` exists and records files touched and exact command results.
