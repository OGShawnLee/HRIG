# Changelog — 012 Add crew portrait images

## Files Created
- `app/frontend/public/crew/sam-sonia.jpg` — copied from `/home/sacred/.gemini/antigravity-cli/brain/a2f0364f-bba5-4e56-96cb-373fae8b90b7/sam_sonia_1794779841180.jpg` (603,300 bytes, JPEG 1024x1024).
- `app/frontend/public/crew/patrick-cuckburn.jpg` — copied from `/home/sacred/.gemini/antigravity-cli/brain/a2f0364f-bba5-4e56-96cb-373fae8b90b7/patrick_cuckburn_png_1794779848707.jpg` (703,527 bytes, JPEG 1024x1024).
- `.tickets/012-add-crew-portrait-images/changelog.md` — this execution log.

## Files Modified
- `app/frontend/src/views/pagina-principal.vue`:
  - Replaced Sam Sonia's `.crew-portrait.portrait-sam` + `.flag-stripe` placeholder inside `@click="openPrimaryVitals('sam')"` button with `<div class="crew-portrait"><img class="crew-portrait-img" src="/crew/sam-sonia.jpg" alt="Sam Sonia" /></div>`.
  - Replaced Patrick Cuckburn's `.crew-portrait.portrait-patrick` + `.flag-stripe` placeholder inside `@click="openPrimaryVitals('patrick')"` button with `<div class="crew-portrait"><img class="crew-portrait-img" src="/crew/patrick-cuckburn.jpg" alt="Patrick Cuckburn" /></div>`.
  - Updated `.crew-portrait` CSS to `border-radius: 4px; border: 1px solid #5b5b5b; background: #0b1320;` and added `.crew-portrait-img { display: block; width: 100%; height: 100%; object-fit: cover; object-position: center top; }`.
  - Removed unused `.portrait-patrick`, `.flag-stripe`, `.flag-blue`, `.flag-red`, and `@media (max-width: 620px)` `.flag-stripe` / `.flag-red` / `.flag-blue:last-child` CSS rules while keeping responsive `.crew-portrait` sizing (`flex-basis: 48px; width: 48px; height: 52px;`).

## Verification Commands & Results

1. **Copy & verify portrait assets (`file` & `ls -la`)** — exit code `0`:
   - `app/frontend/public/crew/sam-sonia.jpg`: `603300` bytes, `JPEG image data, JFIF standard 1.01, resolution (DPI), density 300x300, segment length 16, baseline, precision 8, 1024x1024, components 3`
   - `app/frontend/public/crew/patrick-cuckburn.jpg`: `703527` bytes, `JPEG image data, JFIF standard 1.01, resolution (DPI), density 300x300, segment length 16, baseline, precision 8, 1024x1024, components 3`

2. **`bun run --cwd app/frontend type-check`** — exit code `0`:
   - Output: `$ vue-tsc --noEmit`

3. **`bun run --cwd app/frontend build`** — exit code `0`:
   - Output: `$ vite build --mode production`, `✓ 78 modules transformed.`, `dist/index.html 0.50 kB`, `dist/assets/index-DCJn3aK4.css 22.42 kB`, `dist/assets/index-Dnj1yNb2.js 128.83 kB`, `✓ built in 1.05s`

4. **Verify built assets (`ls -la app/frontend/dist/crew/...`)** — exit code `0`:
   - `app/frontend/dist/crew/sam-sonia.jpg` (`603300` bytes)
   - `app/frontend/dist/crew/patrick-cuckburn.jpg` (`703527` bytes)

5. **Check view for external URLs or leftover placeholder selectors (`rg -n`)** — exit code `0`:
   - Pattern `http://|https://|flag-stripe|flag-blue|flag-red|portrait-sam|portrait-patrick` against `app/frontend/src/views/pagina-principal.vue` returned `NO_MATCHES`.

6. **Check repository status (`git status --short`)** — exit code `0`:
   - ` M app/frontend/src/views/pagina-principal.vue`
   - `?? .tickets/012-add-crew-portrait-images/`
   - `?? app/frontend/public/crew/`

## Spec Checklist
- [x] `app/frontend/public/crew/sam-sonia.jpg` exists and is a non-empty JPEG image — **Yes**
- [x] `app/frontend/public/crew/patrick-cuckburn.jpg` exists and is a non-empty JPEG image — **Yes**
- [x] `app/frontend/src/views/pagina-principal.vue` renders `<img` pointing to `/crew/sam-sonia.jpg` inside `openPrimaryVitals('sam')` button and `/crew/patrick-cuckburn.jpg` inside `openPrimaryVitals('patrick')` button — **Yes**
- [x] `app/frontend/src/views/pagina-principal.vue` defines `.crew-portrait-img` with `object-fit: cover` and no longer contains `.flag-stripe` placeholder CSS/markup — **Yes**
- [x] `app/frontend/src/views/pagina-principal.vue` contains zero external `http://` or `https://` image URLs — **Yes**
- [x] `bun run --cwd app/frontend type-check` exits `0` — **Yes**
- [x] `bun run --cwd app/frontend build` exits `0` — **Yes**
- [x] `app/frontend/dist/crew/sam-sonia.jpg` and `app/frontend/dist/crew/patrick-cuckburn.jpg` exist after build — **Yes**
- [x] `git status --short` shows changes only in `app/frontend/public/crew/`, `app/frontend/src/views/pagina-principal.vue`, and `.tickets/012-add-crew-portrait-images/` — **Yes**
- [x] No unapproved git commits or pushes were executed — **Yes**
- [x] `.tickets/012-add-crew-portrait-images/changelog.md` exists and records files touched and exact command results — **Yes**

## Observations & Blockers
- None. All scope and grounding constraints were respected; astronaut names, vital values, navigation handlers, and `dashboard-principal.vue` remain untouched.
