# Project Context

## Files

- [`CONTEXT.md`](CONTEXT.md) — Project context, ground rules, and index of `.routing` files.
- [`ARCHITECTURE.md`](ARCHITECTURE.md) — End-to-end system architecture and data flow from sensors and manual check-ins to the Vue 3 HUD.
- [`INPUT.md`](INPUT.md) — NASA hackathon challenge prompt, objectives, hazards overview, and radiation monitoring notes.
- [`DOMAIN.md`](DOMAIN.md) — Ubiquitous Language glossary defining core domain concepts, metric tiers, severity states, and clinical roles.
- [`SPECIFICATION.md`](SPECIFICATION.md) — Refined software functional (`RF01`–`RF16`) and non-functional (`RNF01`–`RNF12`) requirements in Spanish until final delivery.
- [`DESIGN.md`](DESIGN.md) — Simplified UI/UX design guidelines, status encoding, body map rules, and human-factors references.
- [`BD.md`](BD.md) — PostgreSQL database design specification (tables, constraints, triggers, seed units/metrics).
- [`BDML.md`](BDML.md) — DBML schema representation matching `BD.md`.
- [`PROGRESS.md`](PROGRESS.md) — Current task status, completed milestones, and upcoming work.
- [`image.png`](image.png) — Reference visual summarizing NASA's five hazards of human spaceflight.
- `reference/` — Minimal reminder notes on NASA's five hazards of human spaceflight:
  - [`space-radiation.md`](reference/space-radiation.md)
  - [`isolation-and-confinement.md`](reference/isolation-and-confinement.md)
  - [`distance-from-earth.md`](reference/distance-from-earth.md)
  - [`gravity-fields.md`](reference/gravity-fields.md)
  - [`hostile-closed-environments.md`](reference/hostile-closed-environments.md)
- External project docs:
  - [`README.md`](../README.md) — Minimal hackathon project overview and repository navigation.
  - [`PROJECT-REVIEW.md`](../PROJECT-REVIEW.md) — Cold scope & feasibility review (`venture-killer`) pending team review.
  - [`METRICS_PRIORIZATION.md`](../docs/METRICS_PRIORIZATION.md) — Minimal health metric tiers (Primary, Secondary, Excluded) and NASA references.
  - [`STACK_SETUP.md`](../docs/STACK_SETUP.md) — Full-stack installation and local setup guide (Wails v3, Vue 3, Go API, PostgreSQL 17+).
  - [`db/`](../db/) — PostgreSQL SQL schema (`001_schema.sql`), catalog seeds (`002_seed.sql`), and demo data (`003_demo_seed.sql`).
  - [`.claude/skills/`](../.claude/skills/) — Project workflow skills (`router`, `ticket`, `venture-killer`).
  - [`.tickets/`](../.tickets/) — Implementation tickets (`task.md`, `spec.md`, `changelog.md`).
  - [`app/`](../app/) — Wails v3 desktop application (Go backend + Vue 3 / UnoCSS / TypeScript frontend).

**HRIG** is a NASA hackathon project to build health monitoring software for astronauts on long-duration space missions (see [`INPUT.md`](INPUT.md)).

## Ground Rules

### Git & Commits
- **Conventional Commits in English**: All commit messages must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification and be written strictly in English.
  - Format: `<type>[optional scope]: <description>`
  - Common types: `feat`, `fix`, `docs`, `chore`, `refactor`, `style`, `test`, `build`, `ci`
  - Keep subject lines clear and imperative; include a minimal body only when helpful context is needed.

## Current Architecture & Handoff Notes (for Router / Subagents)

### Routes & Navigation (`app/frontend/src/router/index.ts`, `ViewNavBar.vue`, `useTelemetryNavigation.ts`)
- **Top Navigation Bar (`ViewNavBar.vue`)**: Contains **only** `Home` (`/`), `Search` (`/search`, `Ctrl+K`), and `Primary Vitals` (`/dashboard-principal`). Do **not** add `Support`, `Warning`, or `Other Stats` links to `ViewNavBar.vue`.
- **Primary Vitals (`dashboard-principal.vue`, `/dashboard-principal`)**:
  - Includes a crew selector (`<select id="primary-crew-select">`) to switch astronauts in-place without returning to `Home`, and an **`Other Stats`** button (`openOtherStats(activeAstronaut ?? resolvedAstronautId)`) that navigates to `/other-stats?astronaut=<id>`.
  - Each card has a `?` (`HelpCircle`) button opening `/search?tab=procedures&q=<manualQuery>`.
  - **Alert & Stale States**: `abnormal` renders a solid non-blinking yellow contour (`.vital-frame-warn`) and yellow text (`text-warn`); `critical` renders a blinking red contour (`.vital-frame-crit`) and blinking red text (`.vital-text-crit`); `StatusStale` renders a dashed contour (`border-dashed border-stale/75`), flat dashed SVG baseline + `"Sensor Disconnected · No Signal"` (`WifiOff`) overlay, `"—"` value, and `"NO SIGNAL"`.
  - **Anatomical Silhouette (`dashboard-principal.vue` `<aside>`)**: Head (`temperature`), full body (`radiation`), lungs (`oxygenation`), brachial cuffs + central arterial band (`pressure`), and heart (`cardiac`) highlight and blink yellow (`.body-alert-warn`) or red (`.body-alert-crit`) from live `cardStatuses`.
- **Other Stats (`dashboard-secundario.vue`, `/other-stats`, with `/secondary-stats` redirect)**:
  - Reached via the **`Other Stats`** button inside `Primary Vitals`; includes a `Primary Vitals` back button and `<select id="secondary-crew-select">` crew switcher.
  - Divided into two sections:
    1. **`CONTINUOUS MONITORING · LIVE STREAM`**: Live SVG waveforms for `resp_rate`, `hrv_rmssd`, `blood_glucose`, `cabin_ppco2` (paired with `cabin_o2`), `suit_pressure` (paired with `suit_temp`), and `radiation_dose_cum` (paired with `radiation_rate`), using `getCrewSeries` and `getEnvSeries`.
    2. **`PERIODIC & CHECK-IN STATS · NON-SECOND CADENCE`**: Self-reported daily astronaut check-ins (`sleep_hours`, `mood` 1–5, `hydration_intake`, `body_mass` with inline inputs/buttons calling `submitCrewDailyMetric` -> Wails `SubmitCheckin` `$Call.ByID(3333120644, req)` -> `POST /` with `device_code: "manual_checkin"`), plus periodic lab/imaging summaries (`rbc_count` + `wbc_count`, `ultrasound`).
- **Pending Daily Check-In Yellow Circle (`CrewCard.vue`, `useTelemetry.ts`)**:
  - `DAILY_CHECKIN_METRICS = ["sleep_hours", "mood", "hydration_intake", "body_mass"]` are no longer sent by `biomonitor_01` in `bot/main.go`.
  - `hasPendingDailyCheckin(astronautId)` checks whether all 4 metrics have been logged today via `device_code == "manual_checkin"` (`source == "manual"`). While any of the 4 remain unlogged today, a small yellow circle (`bg-warn`) appears beside the astronaut's name on `CrewCard.vue` (`HomeView.vue`) and in `dashboard-secundario.vue`.
  - `server/internal/realtime/hub.go` uses a `24h` staleness window (`dailyCheckinStaleAfter`) for daily check-in metrics so manual logs do not become `StatusStale` after 15 seconds.
- **Alert Auto-Recovery (`useTelemetryNavigation.ts`)**:
  - `NON_ALERT_ROUTE_NAMES` includes `"home"`, `"search"`, `"dashboard-principal"`, and `"other-stats"`.
  - When all readings return to `normal`, active alert routes (`/support-alert`, `/warning-alert`) automatically close and return to `previousNonAlertPath`.
- **Proportional Viewport Scaling (`App.vue`)**:
  - `App.vue` scales `document.documentElement.style.fontSize` proportionally against `1120x860` (`0.7`–`2.25` for viewports `>= 768px`). Always use `rem` / standard Tailwind classes (avoid fixed `px` grid tracks) so fullscreen scaling never clips numbers.
- **Bot Control (`bot/main.go`, `:8090`)**:
  - Supports anomaly overrides (`POST /override`, `POST /cancel`) and per-device disconnect/reconnect simulation (`POST /disconnect`, `POST /reconnect`, `GET /status`, and dark-HUD UI at `GET /`).
