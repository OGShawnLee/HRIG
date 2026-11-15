# 010 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Files & Routing
- [x] `app/frontend/src/views/support-alert.vue` exists.
- [x] `app/frontend/src/router/index.ts` imports `SupportAlert` from `@/views/support-alert.vue` and registers `{ path: "/support-alert", name: "support-alert", component: SupportAlert }`.
- [x] `app/frontend/src/router/index.ts` preserves all 4 pre-existing routes (`/`, `/warning-alert`, `/pagina-principal`, `/dashboard-principal`).

## Content & Grounding Checks
- [x] `app/frontend/src/views/support-alert.vue` contains the headings `SUPPORT`, `CAUTION`, `LOW BLOOD PRESSURE DETECTED`, and `RECOMMENDED ACTIONS`.
- [x] `app/frontend/src/views/support-alert.vue` contains an SVG caution triangle (`<svg`) with `aria-hidden="true"` paired with the `SUPPORT` / `CAUTION` labels.
- [x] `app/frontend/src/views/support-alert.vue` displays `86/55 mmHg` and `90–120 / 60–80 mmHg`.
- [x] `app/frontend/src/views/support-alert.vue` displays a source and `MET` + `UTC` timestamp line referencing `Blood pressure cuff (baumanometro_01)`.
- [x] `app/frontend/src/views/support-alert.vue` contains all 3 recommended action items with `"immediately"` spelled correctly (grep for `inmediately` returns empty).
- [x] `app/frontend/src/views/support-alert.vue` marks the mockup guidance copy with `[UNVERIFIED]`.
- [x] `app/frontend/src/views/support-alert.vue` applies `tabular-nums` (`font-variant-numeric: tabular-nums`) to numeric reading and timestamp elements.
- [x] `app/frontend/src/views/support-alert.vue` contains zero `fetch(`, `WebSocket`, or `@bindings/` calls (static prototype only).

## Build & Type-Check
- [x] `bun run --cwd app/frontend type-check` exits `0`.
- [x] `bun run --cwd app/frontend build` exits `0` and outputs `app/frontend/dist/index.html`.

## Scope & Ground Rules
- [x] `git status --short` shows modifications only in `app/frontend/src/router/index.ts`, `?? .tickets/010-add-support-alert-view/`, and `?? app/frontend/src/views/support-alert.vue`.
- [x] All code and `.tickets/010-add-support-alert-view/changelog.md` are written in English.
- [x] No unapproved git commits or pushes were executed.
- [x] `.tickets/010-add-support-alert-view/changelog.md` exists and records files touched and exact command results.
