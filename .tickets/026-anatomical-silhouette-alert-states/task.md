# 026 — Interactive Anatomical Alert States on Astronaut Silhouette

| Field | Value |
| :--- | :--- |
| **ID** | `026-anatomical-silhouette-alert-states` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/views/dashboard-principal.vue` |

## Context

In `dashboard-principal.vue`, the right-side `<aside>` shows a static low-poly astronaut silhouette. The user requested that the human figure visually reflect active alert states per vital system, glowing and blinking in **yellow** (`abnormal`) or **red** (`critical`) according to alert severity:
- **Temperature (`core_temp` / `cabin_temp`)**: the **head** turns yellow/red and blinks.
- **Radiation (`radiation_rate` / `radiation_dose_cum`)**: the **entire body** turns yellow/red and blinks.
- **Oxygenation (`spo2` / `cabin_o2`)**: the **lungs** turn yellow/red and blink.
- **Blood Pressure (`bp_systolic` / `bp_diastolic` / `cabin_pressure`)**: the **representative pressure region** (bilateral upper-arm cuff bands + central vascular pressure band) turns yellow/red and blinks.
- **Cardiac Rhythm (`heart_rate` / `hrv_rmssd`)**: the **heart** turns yellow/red and blinks.

## Grounding Rules

- Modify only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/026-anatomical-silhouette-alert-states/`.
- Preserve Ticket `018` styling on the 5 left vital cards (solid yellow on `abnormal`, red blinking on `critical`), Ticket `021`/`022` centered `?` links, and Ticket `025` `rem`-based tracks.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Update `app/frontend/src/views/dashboard-principal.vue` `<script setup>`:
   - Extract the 5 effective vital statuses into a standalone `computed` (`cardStatuses` returning `{ cardiac, oxygenation, pressure, radiation, temperature }` via `resolveCardStatus`) so the astronaut silhouette reads live severity independently of `metricFilter`.
   - Add a helper `silhouetteAlertClass(status: CardStatus): string` returning `"body-alert-warn"` for `StatusAbnormal`, `"body-alert-crit"` for `StatusCritical`, and `""` otherwise.
2. Update the astronaut silhouette SVG (`viewBox="0 0 160 360"`) in `<aside>`:
   - Add anatomical low-poly groups inside the chest/arms aligned to the existing wireframe (`x=80` center):
     - **Whole Body (`radiation`)**: applies `silhouetteAlertClass(cardStatuses.radiation)` to the full-body outer silhouette & wireframe paths.
     - **Head (`temperature`)**: applies `silhouetteAlertClass(cardStatuses.temperature)` to the faceted head path (`M68 12h30l10 10-2 24-7 25-21 10-21-10-7-25-2-24 10-10Z`).
     - **Lungs (`oxygenation`)**: crisp left/right low-poly lung polygons (`x=57..77` & `83..103`, `y=96..139`) + subtle airway lines using `silhouetteAlertClass(cardStatuses.oxygenation)` (dim wireframe when nominal).
     - **Pressure Region (`pressure`)**: bilateral brachial cuff bands (`x=34..51` & `115..132`, `y=114..136`) and central vascular torso band (`y=142..151`) using `silhouetteAlertClass(cardStatuses.pressure)` (dim wireframe when nominal).
     - **Heart (`cardiac`)**: faceted low-poly heart polygon at `(77..96, 107..129)` using `silhouetteAlertClass(cardStatuses.cardiac)` (dim wireframe when nominal).
3. Add scoped CSS keyframes & classes `.body-alert-warn` (`@keyframes body-blink-warn 1.2s ease-in-out infinite`) and `.body-alert-crit` (`@keyframes body-blink-crit 0.65s ease-in-out infinite`) so active body regions glow and blink yellow (`#facc15`) on `abnormal` and red (`#f87171`) on `critical`, without disturbing `.vital-frame-warn` on the vital cards.
4. Run `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks, write `changelog.md`, and report back.
