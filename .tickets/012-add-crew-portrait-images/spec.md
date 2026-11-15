# 012 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Portrait Assets
- [x] `app/frontend/public/crew/sam-sonia.jpg` exists and is a non-empty JPEG image.
- [x] `app/frontend/public/crew/patrick-cuckburn.jpg` exists and is a non-empty JPEG image.

## View Integration
- [x] `app/frontend/src/views/pagina-principal.vue` renders `<img` pointing to `/crew/sam-sonia.jpg` inside `openPrimaryVitals('sam')` button and `/crew/patrick-cuckburn.jpg` inside `openPrimaryVitals('patrick')` button.
- [x] `app/frontend/src/views/pagina-principal.vue` defines `.crew-portrait-img` with `object-fit: cover` and no longer contains `.flag-stripe` placeholder CSS/markup.
- [x] `app/frontend/src/views/pagina-principal.vue` contains zero external `http://` or `https://` image URLs.

## Build & Type-Check
- [x] `bun run --cwd app/frontend type-check` exits `0`.
- [x] `bun run --cwd app/frontend build` exits `0`.
- [x] `app/frontend/dist/crew/sam-sonia.jpg` and `app/frontend/dist/crew/patrick-cuckburn.jpg` exist after build.

## Scope & Ground Rules
- [x] `git status --short` shows changes only in `app/frontend/public/crew/`, `app/frontend/src/views/pagina-principal.vue`, and `.tickets/012-add-crew-portrait-images/`.
- [x] No unapproved git commits or pushes were executed.
- [x] `.tickets/012-add-crew-portrait-images/changelog.md` exists and records files touched and exact command results.
