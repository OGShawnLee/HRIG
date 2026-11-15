# Changelog — 003 Full-Stack Installation Guide

## Files Created & Modified

- **Created:** [`docs/STACK_SETUP.md`](../../docs/STACK_SETUP.md) — Full-stack English installation and development guide covering:
  1. Architecture & Stack Overview (`app/`, `server/`, `db/`, incremental module approach, and exact versions from `app/frontend/package.json` and `app/go.mod`).
  2. Prerequisites by Operating System (Windows, macOS, Linux, and using `wails3 task` instead of standalone `task`).
  3. Database Setup (PostgreSQL 17+ across Windows portable `local/pgsql/`, Windows `debian/bookworm` VM / WSL2, macOS `OrbStack`/`Colima`/`Homebrew`/`Postgres.app`, Linux `Docker`/`postgresql-17`, exact `psql` commands for `db/001_schema.sql`, `db/002_seed.sql`, `db/003_demo_seed.sql`, and `DATABASE_URL`).
  4. Running the Backend API (`server/`, `server/.env.example`, `wails3 task server:check` / `server:dev`, and baseline `GET /health` + `GET /ws`).
  5. Running the Desktop & Frontend App (`app/` Option A frontend-only UI dev on `http://127.0.0.1:9245` with `@bindings/` note, and Option B full Wails v3 desktop via `wails3 task client:check` / `client:dev`).
  6. Conventions & Troubleshooting (`@/` and `@bindings/` aliases, UnoCSS import order, `.npmrc` `minimum-release-age=10080`, offline Fontshare `Satoshi` timeout warning, `wails3 task`, and `go build .`).
- **Modified:** [`README.md`](../../README.md) — Added [`docs/STACK_SETUP.md`](../../docs/STACK_SETUP.md) under `### Core Project Docs`.
- **Modified:** [`.routing/CONTEXT.md`](../../.routing/CONTEXT.md) — Added [`STACK_SETUP.md`](../../docs/STACK_SETUP.md) under `External project docs`.
- **Created:** [`.tickets/003-full-stack-install-guide/changelog.md`](changelog.md) — Ticket execution record.

## Verification Checks & Outputs

1. **Content & Grounding Verification (PowerShell check against `docs/STACK_SETUP.md`, `README.md`, and `.routing/CONTEXT.md`):**
   - `No [PENDING DB SPEC]: True`
   - `Versions: vue ^3.5.44: True`
   - `Versions: vue-router ^5.4.0: True`
   - `Versions: @wailsio/runtime 3.0.0: True`
   - `Versions: @unocss/reset ^66.12.0: True`
   - `Versions: unocss ^66.0.0: True`
   - `Versions: vite ^8.3.5: True`
   - `Versions: @vitejs/plugin-vue ^6.0.10: True`
   - `Versions: typescript ^5.7.3: True`
   - `Versions: vue-tsc ^2.2.8: True`
   - `Versions: lucide-vue-next ^1.0.0: True`
   - `Versions: reka-ui ^2.1.0: True`
   - `Versions: @vueuse/core ^15.0.0: True`
   - `Versions: @types/node ^22.13.4: True`
   - `Versions: go 1.26.0: True`
   - `Three-layer overview (`app/`, `server/`, `db/`): True`
   - `OSes & DB runtimes: True`
   - `SQL scripts & idempotency: True`
   - `Port, aliases, npmrc, wails3 task: True`
   - `README link: True`
   - `CONTEXT link: True`
2. **Scope & Git Status Check (`git status --short`):**
   - Verified edits made by this ticket are strictly inside `docs/STACK_SETUP.md`, `README.md`, `.routing/CONTEXT.md`, and `.tickets/003-full-stack-install-guide/changelog.md`.
   - `app/**`, `db/**`, `Taskfile.yml`, `.gitignore`, `.claude/**`, `PROJECT-REVIEW.md`, `docs/METRICS_PRIORIZATION.md`, and other `.tickets/**` were not modified by this ticket.
   - No git commits or pushes were executed.

## Blockers & Assumptions Avoided

- Updated Sections 1 & 4 after Ticket `004` reset so `docs/STACK_SETUP.md` does not reference deleted `server/` routes or unimplemented `shared/model` packages.
- Verified `Taskfile.yml` currently contains `client:*` and `db:start`/`db:stop`/`db:status`/`db:psql` only; labeled `db:migrate` and `db:seed` as `[Ticket 005]` alongside the verified `psql` commands rather than presenting unmerged task aliases as active.
- Did not invent table names, package versions, ports, or clinical thresholds outside the input files.
