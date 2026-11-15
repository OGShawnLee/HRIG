# 024 — Changelog: Cancel Adaptive Grid Reflow and Apply Uniform Proportional Viewport UI Scaling

## Files Modified

- `app/frontend/src/views/HomeView.vue`
- `app/frontend/src/components/CrewCard.vue`
- `app/frontend/src/components/EnvironmentCard.vue`
- `app/frontend/src/components/MedicalManualButton.vue`
- `app/frontend/src/views/dashboard-principal.vue`
  - Restored all 5 layout files back to commit `b7a217a` (`git checkout b7a217a -- ...`), reverting ticket `023` adaptive grid reflow while preserving ticket `018` abnormal/critical styles and tickets `021`/`022` centered medical manual help buttons.
- `app/frontend/src/App.vue`
  - Added proportional root `rem` viewport scaling via `updateViewportScale()` using `DESIGN_WIDTH = 1120`, `DESIGN_HEIGHT = 860`, `BASE_FONT_SIZE = 16`, clamped between `MIN_SCALE = 0.7` and `MAX_SCALE = 2.25`.
  - When `window.innerWidth < 768` (`MOBILE_BREAKPOINT`), clears `document.documentElement.style.fontSize = ""` so mobile stacks cleanly at default typography; otherwise sets `document.documentElement.style.fontSize = "${(BASE_FONT_SIZE * scale).toFixed(2)}px"`.
  - Invokes `updateViewportScale()` in `onMounted`, binds `resize` listener on `window`, and cleans up the listener and resets `document.documentElement.style.fontSize = ""` in `onBeforeUnmount`.
- `.tickets/024-viewport-proportional-ui-scaling/task.md`
  - Updated status from `in_progress` to `completed`.
- `.tickets/024-viewport-proportional-ui-scaling/changelog.md`
  - Created this changelog.

## Verification Results

- `git diff b7a217a -- app/frontend/src/views/HomeView.vue app/frontend/src/components/CrewCard.vue app/frontend/src/components/EnvironmentCard.vue app/frontend/src/components/MedicalManualButton.vue app/frontend/src/views/dashboard-principal.vue`: **PASS** (empty output)
- `rg -n "updateViewportScale|DESIGN_WIDTH|DESIGN_HEIGHT" app/frontend/src/App.vue`: **PASS** (matches on lines 15, 16, 22, 29, 64, 65, 71)
- `rg -n "new WebSocket|fetch\(" app/frontend/src/`: **PASS** (0 matches)
- `rg -n "critical|border-warn|text-warn|MedicalManualButton" app/frontend/src/views/HomeView.vue app/frontend/src/views/dashboard-principal.vue app/frontend/src/components/CrewCard.vue app/frontend/src/components/EnvironmentCard.vue app/frontend/src/components/MedicalManualButton.vue`: **PASS** (`018` and `021`/`022` markers confirmed intact)
- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`)
- `bun run --cwd app/frontend build` (`vite build --mode production`): **PASS** (exit code `0`; 2429 modules transformed, built in 1.96s: `dist/index.html` 0.50 kB, `dist/assets/index-BIEtyGdy.css` 21.08 kB, `dist/assets/index-CR40SZfr.js` 312.63 kB)
- Scope check (`git status --short`): Only the 5 restored Vue files, `app/frontend/src/App.vue`, `.tickets/024-viewport-proportional-ui-scaling/`, and pre-existing `.tickets/023-responsive-adaptive-main-layout/task.md` modified/untracked; no git commit or push performed.
