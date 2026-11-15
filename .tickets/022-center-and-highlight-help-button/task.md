# 022 — Center & Enhance Visibility of Metric Help (`?`) Button

| Field | Value |
| :--- | :--- |
| **ID** | `022-center-and-highlight-help-button` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/views/dashboard-principal.vue` |

## Context

In `dashboard-principal.vue`, the `<HelpCircle>` (`?`) button on each vital card currently aligns at the bottom (`items-end pb-2`), uses muted `text-dull` at `w-3.5 h-3.5`, and sits in a narrow `28px`/`32px` column without a separator, making it look off-balance and hard to spot. The user wants it centered inside a clearer, more visible button.

## Grounding Rules

- Modify only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/022-center-and-highlight-help-button/`.
- Keep `cardSecondaryTextClass(row.status, 'text-main')` so nominal cards show high-contrast centered `?`, `abnormal` shows solid yellow `?`, and `critical` shows blinking red `?`.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Update `app/frontend/src/views/dashboard-principal.vue`:
   - Widen the right button track in `<article v-for="row in vitalRows">` from `28px`/`sm:32px` to `36px`/`sm:44px`.
   - Center `<HelpCircle>` inside the `<button type="button">` (`flex items-center justify-center` instead of `items-end pb-2`), add a subtle left border and surface tint (`border-l border-solid border-lineDim bg-surface/50`), switch nominal class in `cardSecondaryTextClass(row.status, 'text-main')` for high visibility, and enlarge `<HelpCircle>` to `class="w-5 h-5"` (`:stroke-width="2.2"`).
2. Run `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`), `bun run --cwd app/frontend build`, and `rg` checks, write `changelog.md`, and report back.
