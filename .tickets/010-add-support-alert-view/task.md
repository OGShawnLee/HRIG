# 010 — Add Caution / Support low blood pressure view

| Field | Value |
| :--- | :--- |
| **ID** | `010` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Frontend Vue Prototype Builder |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `app/frontend/src/views/support-alert.vue`
  - `app/frontend/src/router/index.ts`
  - `.tickets/010-add-support-alert-view/changelog.md`
- **Must not touch:**
  - `.routing/**` (including `INPUT.md`, `CONTEXT.md`, `DOMAIN.md`, `SPECIFICATION.md`, `DESIGN.md`, `PROGRESS.md`)
  - `.claude/**`
  - `docs/**`
  - `db/**`, `server/**`, `bot/**`, `shared/**`
  - `app/frontend/src/views/HomeView.vue`
  - `app/frontend/src/views/warning-alert.vue`
  - `app/frontend/src/views/pagina-principal.vue`
  - `app/frontend/src/views/dashboard-principal.vue`
  - `app/frontend/src/App.vue`, `app/frontend/src/main.ts`, `app/frontend/uno.config.ts`, `app/frontend/package.json`
  - `.tickets/010-add-support-alert-view/task.md`
  - `.tickets/010-add-support-alert-view/spec.md`

## Inputs

Read these files before acting:
1. [`app/frontend/src/views/warning-alert.vue`](../../app/frontend/src/views/warning-alert.vue) (sibling alert view style, dark radial gradient, bordered card layout, responsive sizing)
2. [`app/frontend/src/router/index.ts`](../../app/frontend/src/router/index.ts) (existing routes; preserve uncommitted changes for `warning-alert`, `pagina-principal`, and `dashboard-principal`)
3. [`.routing/DESIGN.md`](../../.routing/DESIGN.md) (Sections 4, 5, 6, and 8: `tabular-nums`, contrast >= 6:1 / >= 10:1, icon + text label for `Caution`, source + timestamp display)
4. [`db/002_seed.sql`](../../db/002_seed.sql) (Lines 77–98: `baumanometro_01` / `Blood pressure cuff`, `bp_systolic` `90`–`120`, `bp_diastolic` `60`–`80`)
5. [`.routing/CONTEXT.md`](../../.routing/CONTEXT.md) (Ground Rules: English in code/tickets, no unapproved commits)

## Grounding & Anti-Hallucination Constraints

- Keep this view as a **static prototype** matching the architecture of `warning-alert.vue`, `pagina-principal.vue`, and `dashboard-principal.vue`. Do **not** add HTTP/WebSocket calls, stores, or backend API imports.
- Use the exact verified blood pressure demo nominal range from `db/002_seed.sql`:
  - Reading: `86/55 mmHg` (falls in the Caution/Abnormal band below `warning_min` `90/60` and above `critical_min` `70/40`)
  - Nominal Range label: `90–120 / 60–80 mmHg`
- Use the exact verified device catalog entry from `db/002_seed.sql` line 79 for the static source/timestamp placeholder:
  - Device: `Blood pressure cuff` (`baumanometro_01`)
  - Timestamp format: dual `MET` + `UTC` static demo placeholder (e.g., `MET 000:00:00 · 2026-11-15T21:42:00Z (UTC)`), clearly labeled as `Demo sample`.
- Mark the action guidance copy inside an HTML/TS comment in `support-alert.vue` as `[UNVERIFIED] Demo guidance copy from UI mockup — clinical review required` because `250ml` and `10 minutes` are mockup placeholders not yet verified against a NASA clinical source.
- Fix the mockup typo `"inmediately"` to `"immediately"` in Action 1.
- Do **not** invent extra action items, medication dosages, or additional vital readings beyond what is specified in the Task steps below.
- Do **not** remove or alter the existing routes already present in `app/frontend/src/router/index.ts`.
- Do **not** commit or push any git changes.
- If `bun run --cwd app/frontend type-check` or `bun run --cwd app/frontend build` fails, **stop**, do not guess workarounds outside scope, and record the exact command and output in `.tickets/010-add-support-alert-view/changelog.md`.

## Task

1. Create `app/frontend/src/views/support-alert.vue` using Vue 3 `<template>` and `<style scoped>` matching the dark clinical aesthetic of `app/frontend/src/views/warning-alert.vue` (dark `#000` / `#090909` background with subtle `rgba(5, 22, 42, 0.62)` radial glow, `1px solid` muted gray card borders, zero horizontal overflow, responsive padding down to `390px` width).
2. In the top `<section class="support-card">` of `support-alert.vue`, render:
   - Header row containing:
     - An inline SVG caution triangle icon (`▲`, `aria-hidden="true"`, yellow/amber `#f5c518`) paired with `<h1>` text `SUPPORT` in `#f5c518`.
     - An explicit status badge/label reading `CAUTION` (paired with the triangle icon to satisfy `DESIGN.md` §5 color + icon + text redundancy, contrast `>= 10:1` on dark background).
   - Bold condition heading: `LOW BLOOD PRESSURE DETECTED`
   - Reading paragraph with `font-variant-numeric: tabular-nums`:
     `Current Reading: 86/55 mmHg` followed by `<em>(Nominal Range: 90–120 / 60–80 mmHg)</em>`
   - Source and timestamp metadata line with `font-variant-numeric: tabular-nums` and contrast `>= 6:1` (e.g. `#d4d4d4` on `#000`):
     `Source: Blood pressure cuff (baumanometro_01) · Demo sample · MET 000:00:00 · 2026-11-15T21:42:00Z (UTC)`
   - Bold instruction lead paragraph:
     `Your blood pressure is below normal. Follow the option below that matches your current condition`
3. In the bottom `<section class="actions-card">` of `support-alert.vue`, render:
   - Heading: `RECOMMENDED ACTIONS`
   - An HTML comment directly above `<ol>` stating: `<!-- [UNVERIFIED] Demo guidance copy from UI mockup — clinical review required -->`
   - Ordered list (`<ol>`) with these exact 3 items:
     1. `<strong>If you feel dizzy, lightheaded, or have tunnel vision</strong> → Sit down or strap into a restraint immediately, drink 250ml of water with electrolytes, and remain seated for 10 minutes before re-measuring`
     2. `<strong>If you just transitioned to Lunar Gravity or stood up quickly</strong> → Review the “Orthostatic Intolerance Countermeasures”`
     3. `<strong>If you just finished intense exercise or an EVA</strong> → Check your Body Temperature and Heart Rate for dehydration signs.`
4. Ensure all text colors on `#000` / `rgba(0, 0, 0, 0.45)` meet `DESIGN.md` §4 (`>= 6:1` general text, `>= 10:1` for alert heading `#f5c518` and primary copy `#f5f5f5`), all tap/readability spacing works without horizontal scrolling (`overflow-x: hidden` / fluid widths), and `font-variant-numeric: tabular-nums` is applied to numeric readings and timestamps.
5. Update `app/frontend/src/router/index.ts` to import `SupportAlert` from `@/views/support-alert.vue` and add `{ path: "/support-alert", name: "support-alert", component: SupportAlert }` alongside the existing `home`, `warning-alert`, `pagina-principal`, and `dashboard-principal` routes. Keep imports sorted consistently with the existing file.
6. Run `bun run --cwd app/frontend type-check` and `bun run --cwd app/frontend build` from the repository root and confirm both exit `0`.
7. Write `.tickets/010-add-support-alert-view/changelog.md` documenting files created/modified, exact commands run with exit codes, and any observations.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/010-add-support-alert-view/changelog.md` listing:
- Files created and modified
- Verification commands executed and their exit codes/outputs
- Any blockers or assumptions avoided
