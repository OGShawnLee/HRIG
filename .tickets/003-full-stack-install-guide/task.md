# 003 — Create full-stack installation guide (`docs/STACK_SETUP.md`)

| Field | Value |
| :--- | :--- |
| **ID** | `003` |
| **Status** | `completed` |
| **Worker** | `antigravity` / `gemini-4-hydrogen` (`thinking_level: medium`) |
| **Role** | Technical Documentation Writer |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `docs/STACK_SETUP.md`
  - `README.md` (only add a link to `docs/STACK_SETUP.md` under `### Core Project Docs`)
  - `.routing/CONTEXT.md` (only add `docs/STACK_SETUP.md` under `External project docs`)
  - `.tickets/003-full-stack-install-guide/changelog.md`
- **Must not touch:**
  - `app/**`, `server/**`, `shared/**`, `db/**`, `local/**`
  - `Taskfile.yml`, `.gitignore`, `go.work`
  - `.routing/INPUT.md`, `.routing/DOMAIN.md`, `.routing/SPECIFICATION.md`, `.routing/DESIGN.md`, `.routing/PROGRESS.md`, `.routing/BD.md`, `.routing/BDML.md`
  - `docs/METRICS_PRIORIZATION.md`
  - `.claude/**`, `PROJECT-REVIEW.md`
  - `.tickets/001-scaffold-wails3-app/**`, `.tickets/002-setup-portable-postgres-17/**`, `.tickets/004-scaffold-go-api/**`, `.tickets/005-postgres-db-integration/**`
  - `.tickets/003-full-stack-install-guide/task.md`, `.tickets/003-full-stack-install-guide/spec.md`

## Inputs

Read these files before acting (do not invent package versions, scripts, table names, or ports outside them):
1. [`app/frontend/package.json`](../../app/frontend/package.json)
2. [`app/frontend/.npmrc`](../../app/frontend/.npmrc)
3. [`app/frontend/vite.config.ts`](../../app/frontend/vite.config.ts)
4. [`app/frontend/uno.config.ts`](../../app/frontend/uno.config.ts)
5. [`app/frontend/src/main.ts`](../../app/frontend/src/main.ts)
6. [`app/Taskfile.yml`](../../app/Taskfile.yml)
7. [`Taskfile.yml`](../../Taskfile.yml)
8. [`app/go.mod`](../../app/go.mod)
9. [`db/README.md`](../../db/README.md)
10. [`db/001_schema.sql`](../../db/001_schema.sql)
11. [`.routing/BD.md`](../../.routing/BD.md)
12. [`README.md`](../../README.md)
13. [`.routing/CONTEXT.md`](../../.routing/CONTEXT.md)
14. [`.tickets/002-setup-portable-postgres-17/task.md`](../002-setup-portable-postgres-17/task.md)
15. [`.tickets/004-scaffold-go-api/task.md`](../004-scaffold-go-api/task.md)
16. [`.tickets/005-postgres-db-integration/task.md`](../005-postgres-db-integration/task.md)

## Grounding & Anti-Hallucination Constraints

- **Stack Layout:**
  - `app/` — Wails v3 desktop application (Go backend + Vue 3 / Vite / UnoCSS / TypeScript frontend).
  - `server/` — Go HTTP + WebSocket API (`module hrig/server`, Chi v5 + Gorilla WebSocket, scaffolded bare-bones in Ticket `004`).
  - `db/` — PostgreSQL 17+ SQL schema (`001_schema.sql`), catalog & threshold seeds (`002_seed.sql`), and 3-day demo dataset (`003_demo_seed.sql`).
- **Frontend & Wails v3 Stack (exact verified values from `app/`):**
  - Package manager: `bun` (verified host baseline `1.4.4`; note `app/frontend/.npmrc` sets `minimum-release-age=10080`).
  - Dependencies from `app/frontend/package.json`: `vue` (`^3.5.44`), `vue-router` (`^5.4.0`), `@wailsio/runtime` (`3.0.0`), `@unocss/reset` (`^66.12.0`), `unocss` (`^66.0.0`), `vite` (`^8.3.5`), `@vitejs/plugin-vue` (`^6.0.10`), `typescript` (`^5.7.3`), `vue-tsc` (`^2.2.8`), `lucide-vue-next` (`^1.0.0`), `reka-ui` (`^2.1.0`), `@vueuse/core` (`^15.0.0`), `@types/node` (`^22.13.4`).
  - Frontend scripts: `dev` (`vite`), `build:dev`, `build`, `type-check` (`vue-tsc --noEmit`), `preview`.
  - Vite dev server: `127.0.0.1:9245` (`WAILS_VITE_PORT`, `strictPort: true`), aliases `@` -> `./src` and `@bindings` -> `./bindings`.
  - UnoCSS setup: `presetWind3`, `presetWebFonts` (Fontshare `Satoshi`, non-fatal timeout warning during offline builds), `transformerDirectives` (`--uno`), and `"virtual:uno.css"` imported before `"@unocss/reset/tailwind-compat.css"`.
  - Desktop shell prerequisites: `Go` (`>= 1.26.0`), `Wails v3 CLI` (`v3.0.0`), native WebView (`WebView2` on Windows, `WKWebView` on macOS, `WebKitGTK` on Linux), and `wails3 task client:check` / `client:dev`. Always use `wails3 task` instead of standalone `task`.
- **Database Layer (PostgreSQL 17+ — exact SQL files from `db/`):**
  - Windows Option A (repo-local portable): PostgreSQL `17.11-5` inside `.gitignore`d `local/pgsql/` managed via `wails3 task db:start`, `db:stop`, `db:status`, `db:psql` on `127.0.0.1:5432` (database `hrig`, user `postgres`), per Ticket `002`.
  - Windows Option B (isolated Linux): Headless `debian/bookworm` VM or WSL2 running PostgreSQL 17 (if using a traditional VM, set `listen_addresses = '*'`, configure `pg_hba.conf`, and forward port `5432`).
  - macOS: OrbStack or Colima (`docker compose`), Homebrew (`postgresql@17`), or Postgres.app.
  - Linux: Docker Compose or native `postgresql-17`.
  - Schema & Seed execution (exact commands from `db/README.md`, run from repo root against database `hrig`):
    1. `psql -h 127.0.0.1 -p 5432 -U postgres -d hrig -v ON_ERROR_STOP=1 -f db/001_schema.sql` (non-idempotent DDL; run once on a fresh database)
    2. `psql -h 127.0.0.1 -p 5432 -U postgres -d hrig -v ON_ERROR_STOP=1 -f db/002_seed.sql` (idempotent catalogs: units, metrics, devices, `metric_threshold`)
    3. `psql -h 127.0.0.1 -p 5432 -U postgres -d hrig -v ON_ERROR_STOP=1 -f db/003_demo_seed.sql` (idempotent 3-day demo data: 3 astronauts, cabin & EVA suit telemetry)
  - Local connection string example: `DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable`.
  - Note if Ticket `005` has already added `db:migrate` / `db:seed` in `Taskfile.yml`, document those aliases alongside the direct `psql` commands; if not yet present in `Taskfile.yml`, label `db:migrate`/`db:seed` as `[Ticket 005]` or stick to the verified `psql` commands above.
- **Backend API Layer (`server/`):**
  - Document `server/.env.example` (`HRIG_API_ADDR=:8080`, `HRIG_ALLOWED_ORIGINS`, `DATABASE_URL`), `wails3 task server:check` / `server:dev`, and the baseline HTTP/WebSocket endpoints (`GET /health`, `GET /ws`).
- Write `docs/STACK_SETUP.md` in English.
- Do **not** commit or push any changes to git.

## Task

1. Create `docs/STACK_SETUP.md` with these sections in order:
   - **1. Architecture & Stack Overview:** Summary table covering Desktop/UI (`app/`: Wails v3.0.0, Vue 3, Vite, UnoCSS, Bun), Backend API (`server/`: Go >= 1.26, Chi v5, Gorilla WebSocket), and Database (`db/`: PostgreSQL 17+).
   - **2. Prerequisites by Operating System:** Windows (Go, Wails v3, Bun, WebView2), macOS (Go, Wails v3, Bun, Xcode CLT / WKWebView), and Linux (Go, Wails v3, Bun, WebKitGTK).
   - **3. Database Setup (PostgreSQL 17+):**
     - Windows Portable (`local/pgsql/` via `wails3 task db:start` / `db:stop` / `db:status` / `db:psql`) and Windows `debian/bookworm` headless VM / WSL2 notes (`listen_addresses`, `pg_hba.conf`, port `5432`).
     - macOS (`OrbStack` / `Colima`, `brew install postgresql@17`, or `Postgres.app`).
     - Linux (`docker` or native `postgresql-17`).
     - **Applying Schema & Seeds:** Exact `psql` commands for `db/001_schema.sql`, `db/002_seed.sql`, and `db/003_demo_seed.sql` (noting `001_schema.sql` runs once on a clean DB while `002_seed.sql` and `003_demo_seed.sql` are idempotent), plus `DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable`.
   - **4. Running the Backend API (`server/`):** `server/.env.example`, `wails3 task server:check` and `wails3 task server:dev`, and baseline HTTP/WebSocket endpoints (`GET /health`, `GET /ws`).
   - **5. Running the Desktop & Frontend App (`app/`):**
     - **Option A — Frontend-Only UI Dev:** `cd app/frontend && bun install`, `bun run dev` (`http://127.0.0.1:9245`), `bun run type-check`, `bun run build`, and note on pre-generated `@bindings/` vs. live desktop bridge calls.
     - **Option B — Full Desktop App (Wails v3):** `wails3 task client:check` and `wails3 task client:dev` from repo root.
   - **6. Conventions & Troubleshooting:** `@/` and `@bindings/` aliases, UnoCSS import order (`virtual:uno.css` before `@unocss/reset/tailwind-compat.css`), `.npmrc` `minimum-release-age=10080`, offline Fontshare `Satoshi` timeout warning, using `wails3 task` instead of standalone `task`, and running `go build .` (not `go build ./...`) inside `app/`.
2. Update `README.md` under `### Core Project Docs` to add a link and short description for [`docs/STACK_SETUP.md`](docs/STACK_SETUP.md).
3. Update `.routing/CONTEXT.md` under `External project docs` to add a link and short description for [`STACK_SETUP.md`](../docs/STACK_SETUP.md).
4. Write `.tickets/003-full-stack-install-guide/changelog.md` listing every file created or modified and every verification check executed.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/003-full-stack-install-guide/changelog.md` listing:
- Files created and modified
- Verification checks executed and their outputs
- Any blockers or assumptions avoided
