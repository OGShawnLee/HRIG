# 014 — Change Log

## Files Modified
- `app/frontend/src/views/pagina-principal.vue`
  - Added `role="button"`, `tabindex="0"`, `aria-label="Open Sam Sonia primary vitals"`, `@click="openPrimaryVitals('sam')"`, `@keydown.enter.prevent="openPrimaryVitals('sam')"`, and `@keydown.space.prevent="openPrimaryVitals('sam')"` directly on Sam Sonia's `<article class="crew-card">`.
  - Added `role="button"`, `tabindex="0"`, `aria-label="Open Patrick Cuckburn primary vitals"`, `@click="openPrimaryVitals('patrick')"`, `@keydown.enter.prevent="openPrimaryVitals('patrick')"`, and `@keydown.space.prevent="openPrimaryVitals('patrick')"` directly on Patrick Cuckburn's `<article class="crew-card">`.
  - Removed both inner `<button type="button" class="crew-portrait-btn">` wrappers so `<div class="crew-portrait">` sits directly inside `<div class="crew-heading">` with no nested interactive controls, keeping `/crew/sam-sonia.jpg`, `/crew/patrick-cuckburn.jpg`, astronaut names, and vital values intact.
  - Removed `.crew-portrait-btn` and `.crew-portrait-btn:focus-visible` CSS rules.
  - Added `cursor: pointer; transition: border-color 150ms ease, background-color 150ms ease;` to `.crew-card` and added `.crew-card:hover` and `.crew-card:focus-visible` rules.

## Files Created
- `.tickets/014-clickable-crew-cards/changelog.md`

## Verification Commands & Results

1. `bun run --cwd app/frontend type-check` (from `/home/sacred/bd/HRIG`)
   - Exit code: `0`
   - Output:
     ```text
     $ vue-tsc --noEmit
     ```

2. `bun run --cwd app/frontend build` (from `/home/sacred/bd/HRIG`)
   - Exit code: `0`
   - Output:
     ```text
     $ vite build --mode production
     vite v8.3.8 building client environment for production...
     transforming (8)  vite/modulepreload-polyfill.js
     transforming (9) src/App.vue
     transforming (28) bindings/hrig/app/app.ts
     transforming (29) src/composables/useTelemetryNavigation.ts
     ✓ 81 modules transformed.
     rendering chunks (1)...
     computing gzip size...
     dist/index.html                   0.50 kB │ gzip:  0.34 kB
     dist/assets/index-BUqPF-PS.css   24.07 kB │ gzip:  5.94 kB
     dist/assets/index-DI4Io2Fe.js   132.47 kB │ gzip: 48.37 kB
     ✓ built in 942ms
     ```
   - `test -f app/frontend/dist/index.html` confirmed `app/frontend/dist/index.html` exists.

3. `grep -c "crew-portrait-btn" app/frontend/src/views/pagina-principal.vue`
   - Output: `0` (zero remaining elements or CSS rules for `crew-portrait-btn`).

4. `git status --short`
   - Confirmed no files outside `app/frontend/src/views/pagina-principal.vue` and `.tickets/014-clickable-crew-cards/` were modified by this ticket (existing workspace changes from earlier tickets remain untouched).
   - No `git commit` or `git push` commands were executed.

## Spec Checklist
- [x] `app/frontend/src/views/pagina-principal.vue` binds `@click="openPrimaryVitals('sam')"`, `role="button"`, `tabindex="0"`, `@keydown.enter.prevent`, and `@keydown.space.prevent` on Sam Sonia's `<article class="crew-card">` (`Yes`).
- [x] `app/frontend/src/views/pagina-principal.vue` binds `@click="openPrimaryVitals('patrick')"`, `role="button"`, `tabindex="0"`, `@keydown.enter.prevent`, and `@keydown.space.prevent` on Patrick Cuckburn's `<article class="crew-card">` (`Yes`).
- [x] `app/frontend/src/views/pagina-principal.vue` contains zero `crew-portrait-btn` elements or CSS rules (`Yes`).
- [x] `app/frontend/src/views/pagina-principal.vue` defines `cursor: pointer`, `:hover`, and `:focus-visible` styles on `.crew-card` (`Yes`).
- [x] Both `/crew/sam-sonia.jpg` and `/crew/patrick-cuckburn.jpg` `<img>` tags remain intact inside their respective `.crew-card` (`Yes`).
- [x] `bun run --cwd app/frontend type-check` exits `0` (`Yes`).
- [x] `bun run --cwd app/frontend build` exits `0` and outputs `app/frontend/dist/index.html` (`Yes`).
- [x] No files outside `app/frontend/src/views/pagina-principal.vue` and `.tickets/014-clickable-crew-cards/` were modified by this ticket (`Yes`).
- [x] No unapproved git commits or pushes were executed (`Yes`).
- [x] `.tickets/014-clickable-crew-cards/changelog.md` exists and records files touched and exact command results (`Yes`).
