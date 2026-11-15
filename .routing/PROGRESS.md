# Project Progress

## Current Task
- [ ] **Execute [`.tickets/006-integrate-sqlc/`](../.tickets/006-integrate-sqlc/task.md)** *(Draft — Ready to Dispatch)*
  - `006`: Configure `sqlc.yaml` (`pgx/v5`), starter `db/queries/*.sql`, generated `server/internal/db/` bindings, and `wails3 task db:sqlc`.
- [ ] **Design & Layer Backend Services, Models & Routes Incrementally** *(Next — starting from minimal `server/` Hello World baseline)*
  - Define domain interfaces/models/routes one slice at a time (aligned with `db/001_schema.sql`, `002_seed.sql`, and [`SENSOR-READINGS.md`](../SENSOR-READINGS.md)), then update and unblock [`.tickets/005-postgres-db-integration/`](../.tickets/005-postgres-db-integration/task.md).
- [ ] **Reconcile [`SPECIFICATION.md`](SPECIFICATION.md) & [`DOMAIN.md`](DOMAIN.md) with Reduced DB Scope & [`PROJECT-REVIEW.md`](../PROJECT-REVIEW.md)** *(Pending)*
  - **Status enum (`RF01`, [`DOMAIN.md`](DOMAIN.md) vs. [`BD.md`](BD.md) `reading_status`):** DB persists 3 states (`normal`, `abnormal`, `critical`); decide whether to merge **Aviso** + **Valor Anómalo** into `Valor Anómalo` (`abnormal`) and mark **Dato Obsoleto** (`stale`, `RNF10`) explicitly as runtime-only.
  - **Operating modes (`RF15` vs. `environment_mode`):** Trim 4 habitats (`Nave / Estación Espacial / Rover / Base Lunar-Marte`) down to **`iva` / `eva`** modes.
  - **Out-of-DB scope cuts (`RF06`–`RF09`, `RF14`, `RF16`, `RNF12`):** DB excludes auth/external systems (`RF09`/`RNF12`), missions/locations, persisted alerts, symptoms, procedures, and scheduled reminders (`RF16`), and simplifies crew profiles (`RF14`) to `astronaut` (`full_name`, `birth_date`, derived `age`, `nationality_code`, `nationality_label`). Update specs accordingly and keep alerts/procedures/symptoms in app/server config.
  - **Metric catalog (`DOMAIN.md` vs. [`db/002_seed.sql`](../db/002_seed.sql)):** DB drops `cortisol` and `caloric_intake`, keeps `rbc_count` & `wbc_count`, and splits environmental vs. astronaut readings (`metric_type = 2` enforces `astronaut_id IS NULL`). Align `DOMAIN.md` *Datos Secundarios* / *Métricas Excluidas*.
  - **Server thresholds (`metric_threshold` vs. `RF06`/`RF07`, [`TEMP-DB-TESTING.md`](TEMP-DB-TESTING.md)):** Document server-evaluated limits (`<= critical_min` / `>= critical_max` → `critical`, `< warning_min` / `> warning_max` → `abnormal`, `manual` mode for `ultrasound`), keep mockup ranges tagged as unverified (`'demo - clinical review required'`), and confirm `EVA` vs. `IVA` threshold handling.
  - **Timestamps (`RNF11` vs. `001_schema.sql`):** DB stores `recorded_at_utc` (`timestamptz`) only; update `RNF11` to derive `MET` in the API/UI from a mission epoch or remove `MET` persistence.
  - **Engine version note (`BD.md`/`BDML.md` vs. `001_schema.sql`):** Align `BD.md` line 296 & `BDML.md` line 5 (`PostgreSQL 18`) with `db/001_schema.sql` (`PostgreSQL 17+`).
- [ ] **Review [`PROJECT-REVIEW.md`](../PROJECT-REVIEW.md)** *(Pending — Team Review)*
  - Confirm remaining non-DB scope cuts in [`DOMAIN.md`](DOMAIN.md) and [`SPECIFICATION.md`](SPECIFICATION.md).
- [ ] **Working on the UI & UI Flow** *(In Progress)*
  - General UI flow defined; designing the core interactive visualization prototype (primary/secondary dashboards, responsive anatomical figure, alerts, and guidance views).
  - Adding a persistent **Command Palette / Universal Search** ("Dynamic Island" / Omnibox) rendered at all times—accessible via touch/click or `Ctrl+K`—for fast system-wide search and navigation.

## Completed
- [x] Documented challenge brief in [`INPUT.md`](INPUT.md)
- [x] Created and updated [`README.md`](../README.md) with repository navigation (and fixed broken `DATABASE_DESIGN.md` link to point to `db/README.md`, `.routing/BD.md`, and `.routing/BDML.md`)
- [x] Added [`CONTEXT.md`](CONTEXT.md) with Conventional Commits (English) ground rule and indexed `BD.md`, `BDML.md`, and `db/`
- [x] Added NASA Five Hazards reminder notes in [`reference/`](reference/)
- [x] Defined minimal health metric prioritization in [`docs/METRICS_PRIORIZATION.md`](../docs/METRICS_PRIORIZATION.md)
- [x] Added simplified UI/UX guidelines in [`DESIGN.md`](DESIGN.md)
- [x] Refined domain glossary ([`DOMAIN.md`](DOMAIN.md)) and initial full specification (`RF01`–`RF16`, `RNF01`–`RNF12` in [`SPECIFICATION.md`](SPECIFICATION.md))
- [x] Outlined general UI flow and persistent **Command Palette / Universal Search** (`Ctrl+K` / Dynamic Island) concept
- [x] Adapted workflow skills (`router`, `ticket`, `venture-killer`) in [`.claude/skills/`](../.claude/skills/)
- [x] Generated cold scope & feasibility audit in [`PROJECT-REVIEW.md`](../PROJECT-REVIEW.md)
- [x] Completed [`.tickets/001-scaffold-wails3-app/`](../.tickets/001-scaffold-wails3-app/) — bare-bones Wails v3 + Go + Vue 3 + Vite + Bun + UnoCSS scaffold verified via `wails3 task client:check`
- [x] Pulled & moved teammate's PostgreSQL DB design docs into context ([`BD.md`](BD.md), [`BDML.md`](BDML.md)) alongside [`db/001_schema.sql`](../db/001_schema.sql), [`db/002_seed.sql`](../db/002_seed.sql), [`db/003_demo_seed.sql`](../db/003_demo_seed.sql), [`db/README.md`](../db/README.md), and [`TEMP-DB-TESTING.md`](TEMP-DB-TESTING.md)
- [x] Locked initial backend & database ticket definitions: [`.tickets/003-full-stack-install-guide/`](../.tickets/003-full-stack-install-guide/), [`.tickets/004-scaffold-go-api/`](../.tickets/004-scaffold-go-api/) *(later reset to minimal `Chi` Hello World)*, and [`.tickets/005-postgres-db-integration/`](../.tickets/005-postgres-db-integration/) *(pending re-alignment with incremental API design)*
- [x] Completed [`.tickets/002-setup-portable-postgres-17/`](../.tickets/002-setup-portable-postgres-17/) — portable PostgreSQL `17.11` in ignored `local/pgsql/` (`hrig` DB) + Windows `db:*` tasks in [`Taskfile.yml`](../Taskfile.yml)
- [x] Completed [`.tickets/003-full-stack-install-guide/`](../.tickets/003-full-stack-install-guide/) — created [`docs/STACK_SETUP.md`](../docs/STACK_SETUP.md) and indexed it in [`README.md`](../README.md) & [`CONTEXT.md`](CONTEXT.md)
- [x] Completed [`.tickets/004-scaffold-go-api/`](../.tickets/004-scaffold-go-api/) — minimal `Chi v5` Hello World Go server (`GET /` -> `{"message":"hello world"}`, config, CORS, graceful shutdown, `server:*` tasks)

## Upcoming
- [ ] Retire [`TEMP-DB-TESTING.md`](TEMP-DB-TESTING.md) once `005` passes automated integration tests on disposable `hrig_test`
- [ ] Draft & execute 3D anatomical figure ticket (interactive 3D low-poly anatomical figure + left floating procedures/diagnosis modal)
- [ ] Connect `app/` Wails v3 service + TS bindings to `hrig/shared/model` and `server/` HTTP/WS API
- [ ] Translate [`SPECIFICATION.md`](SPECIFICATION.md) and [`DOMAIN.md`](DOMAIN.md) to English and finalize delivery docs
