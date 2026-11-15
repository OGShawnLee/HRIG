# 012 — Add crew portrait images for Sam Sonia and Patrick Cuckburn

| Field | Value |
| :--- | :--- |
| **ID** | `012` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Frontend Crew Asset Integrator |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `app/frontend/public/crew/sam-sonia.jpg`
  - `app/frontend/public/crew/patrick-cuckburn.jpg`
  - `app/frontend/src/views/pagina-principal.vue`
  - `.tickets/012-add-crew-portrait-images/changelog.md`
- **Must not touch:**
  - `.routing/**` (including `INPUT.md`, `CONTEXT.md`, `DOMAIN.md`, `SPECIFICATION.md`, `DESIGN.md`, `PROGRESS.md`)
  - `.claude/**`
  - `docs/**`, `db/**`, `server/**`, `bot/**`, `shared/**`
  - `app/frontend/src/views/dashboard-principal.vue`
  - `app/frontend/src/views/support-alert.vue`
  - `app/frontend/src/views/warning-alert.vue`
  - `app/frontend/src/views/HomeView.vue`
  - `app/frontend/src/App.vue`, `app/frontend/src/router/index.ts`, `app/frontend/src/composables/useTelemetryNavigation.ts`
  - `.tickets/012-add-crew-portrait-images/task.md`
  - `.tickets/012-add-crew-portrait-images/spec.md`

## Inputs

Read these files before acting:
1. [`app/frontend/src/views/pagina-principal.vue`](../../app/frontend/src/views/pagina-principal.vue) (crew portrait buttons, `.crew-portrait`, responsive rules at lines 17–62 and 152–275)
2. [`app/frontend/vite.config.ts`](../../app/frontend/vite.config.ts) (verify static `public/` asset serving)

Source generated portrait files to copy into `app/frontend/public/crew/`:
- Sam Sonia: `/home/sacred/.gemini/antigravity-cli/brain/a2f0364f-bba5-4e56-96cb-373fae8b90b7/sam_sonia_1794779841180.jpg`
- Patrick Cuckburn: `/home/sacred/.gemini/antigravity-cli/brain/a2f0364f-bba5-4e56-96cb-373fae8b90b7/patrick_cuckburn_png_1794779848707.jpg`

## Grounding & Anti-Hallucination Constraints

- Use only local static assets inside `app/frontend/public/crew/` (offline-first per `RNF09`; zero external image URLs).
- Keep the existing `openPrimaryVitals('sam')` and `openPrimaryVitals('patrick')` `<button type="button" class="crew-portrait-btn">` navigation behavior intact.
- Keep `dashboard-principal.vue` untouched (preserve the anatomical astronaut figure).
- Do **not** modify astronaut names or vital values in `pagina-principal.vue`.
- Do **not** commit or push any git changes.
- If `bun run --cwd app/frontend type-check` or `bun run --cwd app/frontend build` fails, **stop** and record the exact command and output in `.tickets/012-add-crew-portrait-images/changelog.md`.

## Task

1. Create `app/frontend/public/crew/` and copy the two generated portrait files into it:
   - Copy `/home/sacred/.gemini/antigravity-cli/brain/a2f0364f-bba5-4e56-96cb-373fae8b90b7/sam_sonia_1794779841180.jpg` to `app/frontend/public/crew/sam-sonia.jpg`.
   - Copy `/home/sacred/.gemini/antigravity-cli/brain/a2f0364f-bba5-4e56-96cb-373fae8b90b7/patrick_cuckburn_png_1794779848707.jpg` to `app/frontend/public/crew/patrick-cuckburn.jpg`.
2. Update `app/frontend/src/views/pagina-principal.vue`:
   - Replace Sam Sonia's placeholder `.crew-portrait.portrait-sam` and `.flag-stripe` spans inside `<button type="button" class="crew-portrait-btn" ... @click="openPrimaryVitals('sam')">` with:
     ```html
     <div class="crew-portrait">
       <img class="crew-portrait-img" src="/crew/sam-sonia.jpg" alt="Sam Sonia" />
     </div>
     ```
   - Replace Patrick Cuckburn's placeholder `.crew-portrait.portrait-patrick` and `.flag-stripe` spans inside `<button type="button" class="crew-portrait-btn" ... @click="openPrimaryVitals('patrick')">` with:
     ```html
     <div class="crew-portrait">
       <img class="crew-portrait-img" src="/crew/patrick-cuckburn.jpg" alt="Patrick Cuckburn" />
     </div>
     ```
   - Update `.crew-portrait` CSS to remove the unused `.portrait-patrick` gradient and `.flag-stripe` / `.flag-blue` / `.flag-red` rules, add `border-radius: 4px; border: 1px solid #5b5b5b; background: #0b1320;` on `.crew-portrait`, and add `.crew-portrait-img { display: block; width: 100%; height: 100%; object-fit: cover; object-position: center top; }`.
   - Remove any unused `@media (max-width: 620px)` `.flag-stripe` / `.flag-red` / `.flag-blue` overrides while keeping `.crew-portrait` responsive sizing (`width: 48px; height: 52px; flex-basis: 48px;`).
3. Run `bun run --cwd app/frontend type-check` and `bun run --cwd app/frontend build` from repository root and confirm both exit `0`, and verify `app/frontend/dist/crew/sam-sonia.jpg` and `app/frontend/dist/crew/patrick-cuckburn.jpg` exist in the built output.
4. Write `.tickets/012-add-crew-portrait-images/changelog.md` listing files created/modified, exact commands run with exit codes, and any observations.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/012-add-crew-portrait-images/changelog.md` listing:
- Files created and modified
- Verification commands executed and their exit codes/outputs
- Any blockers or assumptions avoided
