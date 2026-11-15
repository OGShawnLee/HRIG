# HRIG Full-Stack Installation & Development Guide

This guide covers setting up, running, and verifying the full **HRIG (Resource Integration Gear)** stack across Windows, macOS, and Linux: the Wails v3 desktop and Vue 3 frontend (`app/`), shared Go domain models (`shared/`), the Go HTTP + WebSocket API (`server/`), and PostgreSQL 17+ (`db/`).

---

## 1. Architecture & Stack Overview

HRIG is organized into three focused layers — a Wails v3 + Vue 3 desktop/UI app, a Go HTTP + WebSocket API server (`[Ticket 004]`), and PostgreSQL 17+ SQL scripts:

| Layer | Directory | Module / Package | Core Stack & Versions |
| :--- | :--- | :--- | :--- |
| **Desktop & UI** | `app/` | `hrig/app` / `hrig-app-frontend` | Go `1.26.0`, Wails v3 (`github.com/wailsapp/wails/v3 v3.0.0`), `@wailsio/runtime 3.0.0`, Vue `^3.5.44`, Vue Router `^5.4.0`, Vite `^8.3.5`, UnoCSS `^66.0.0`, Bun (baseline `1.4.4`) |
| **Backend API** | `server/` | `hrig/server` *(Ticket `004`)* | Go `>= 1.26.0`, Chi v5 (`github.com/go-chi/chi/v5`), `github.com/go-chi/cors`, Gorilla WebSocket (`github.com/gorilla/websocket`) |
| **Database** | `db/` | — | PostgreSQL 17+ (`001_schema.sql`, `002_seed.sql`, `003_demo_seed.sql`) |

> **Incremental Design Note:** `app/` and `server/` are scaffolded as standalone Go modules first (`001` and `004`) before layering shared domain models, services, and PostgreSQL integration (`[Ticket 005]+`) incrementally.

### Exact Frontend & Desktop Package Versions

From `app/go.mod`:
- `go`: `1.26.0`
- `github.com/wailsapp/wails/v3`: `v3.0.0`

From `app/frontend/package.json` (`hrig-app-frontend` v`0.1.0`):
- **Dependencies:**
  - `@unocss/reset`: `^66.12.0`
  - `@vueuse/core`: `^15.0.0`
  - `@wailsio/runtime`: `3.0.0`
  - `lucide-vue-next`: `^1.0.0`
  - `reka-ui`: `^2.1.0`
  - `vue`: `^3.5.44`
  - `vue-router`: `^5.4.0`
- **DevDependencies:**
  - `@types/node`: `^22.13.4`
  - `@vitejs/plugin-vue`: `^6.0.10`
  - `typescript`: `^5.7.3`
  - `unocss`: `^66.0.0`
  - `vite`: `^8.3.5`
  - `vue-tsc`: `^2.2.8`

---

## 2. Prerequisites by Operating System

All platforms require:
1. **Go** `>= 1.26.0`
2. **Wails v3 CLI** `v3.0.0` (`go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0`)
3. **Bun** (verified host baseline `1.4.4`)

> **Important:** Always invoke repository tasks with **`wails3 task <name>`** instead of standalone `task`. Wails v3 bundles Task v3 directly so all contributors use the same task runner without a separate installation.

### Windows
- **Go** `>= 1.26.0`, **Wails v3 CLI** `v3.0.0`, **Bun** (`1.4.4` baseline)
- **WebView:** Microsoft Edge **WebView2** Runtime (pre-installed on Windows 11 and modern Windows 10).
- **Database:** Portable PostgreSQL `17.11-5` in `.gitignore`d `local/pgsql/` (Option A) or WSL2 / headless Linux VM (Option B).

### macOS
- **Go** `>= 1.26.0`, **Wails v3 CLI** `v3.0.0`, **Bun** (`1.4.4` baseline)
- **Toolchain & WebView:** Xcode Command Line Tools (`xcode-select --install`) and native **WKWebView** (included with macOS).
- **Database:** OrbStack, Colima (`docker compose`), Homebrew (`postgresql@17`), or Postgres.app.

### Linux
- **Go** `>= 1.26.0`, **Wails v3 CLI** `v3.0.0`, **Bun** (`1.4.4` baseline)
- **Toolchain & WebView:** `gcc`, `pkg-config`, and **WebKitGTK** development libraries (plus GTK3 headers required by Wails v3 on Linux).
- **Database:** Docker / Docker Compose or native `postgresql-17`.

---

## 3. Database Setup (PostgreSQL 17+)

HRIG targets **PostgreSQL 17+** listening on `127.0.0.1:5432` with database `hrig`.

### Windows — Option A: Portable PostgreSQL (`local/pgsql/`)

Windows contributors can run the repo-local portable PostgreSQL `17.11-5` cluster inside `.gitignore`d `local/pgsql/` (configured in Ticket `002`, bound strictly to `127.0.0.1:5432` with user `postgres` and database `hrig`). No Windows Service or global `PATH` change is required.

Manage the local cluster from the repository root using `wails3 task` (`platforms: [windows]` in `Taskfile.yml`):

```powershell
wails3 task db:start   # Starts PostgreSQL on 127.0.0.1:5432 (logs to local/pgsql/logfile)
wails3 task db:status  # Checks cluster status
wails3 task db:psql    # Opens psql connected to database hrig as user postgres
wails3 task db:stop    # Cleanly stops the server (-m fast)
```

### Windows — Option B: Isolated Linux (`debian/bookworm` VM or WSL2)

If running PostgreSQL 17 inside WSL2 or a headless `debian/bookworm` VM:
- Install `postgresql-17` inside the guest and create database `hrig`.
- **WSL2:** `127.0.0.1:5432` typically forwards automatically from WSL2 to Windows, or connect via the WSL instance IP.
- **Traditional VM (`debian/bookworm`):**
  1. Set `listen_addresses = '*'` in `postgresql.conf`.
  2. Add a host entry in `pg_hba.conf` permitting the Windows host/NAT subnet (e.g., `host all all 10.0.2.0/24 scram-sha-256` or `trust` for local isolated dev only).
  3. Forward guest port `5432` to host `127.0.0.1:5432` in the hypervisor/VM network configuration and restart PostgreSQL.

### macOS

Choose any PostgreSQL 17+ runtime exposing `127.0.0.1:5432` with database `hrig`:
- **OrbStack or Colima (Docker Compose):** Run a `postgres:17` container with `POSTGRES_USER=postgres`, `POSTGRES_DB=hrig`, and port `127.0.0.1:5432:5432`.
- **Homebrew:**
  ```bash
  brew install postgresql@17
  brew services start postgresql@17
  createdb -h 127.0.0.1 -p 5432 -U postgres hrig
  ```
- **Postgres.app:** Select PostgreSQL 17, initialize a server on port `5432`, and create the `hrig` database.

### Linux

- **Docker / Docker Compose:** Run a `postgres:17` container bound to `127.0.0.1:5432:5432` with `POSTGRES_USER=postgres` and `POSTGRES_DB=hrig`.
- **Native `postgresql-17`:** Install `postgresql-17` from the PostgreSQL APT/package repository, ensure it listens on `127.0.0.1:5432`, and create the `hrig` database (`sudo -u postgres createdb hrig`).

### Applying Schema & Seeds

Run these commands from the **repository root** against a running PostgreSQL 17+ server with database `hrig`:

```bash
# 1. Schema DDL (non-idempotent — run ONCE on a fresh/empty database)
psql -h 127.0.0.1 -p 5432 -U postgres -d hrig -v ON_ERROR_STOP=1 -f db/001_schema.sql

# 2. Catalog & threshold seed (idempotent — safe to re-run)
psql -h 127.0.0.1 -p 5432 -U postgres -d hrig -v ON_ERROR_STOP=1 -f db/002_seed.sql

# 3. 3-day demo dataset: 3 astronauts, cabin & EVA suit telemetry (idempotent — safe to re-run)
psql -h 127.0.0.1 -p 5432 -U postgres -d hrig -v ON_ERROR_STOP=1 -f db/003_demo_seed.sql
```

> **Windows Portable PG Note:** If `psql` is not on your `PATH`, replace `psql` above with `.\local\pgsql\bin\psql.exe`.
> **Taskfile Aliases (`[Ticket 005]`):** `wails3 task db:migrate` and `wails3 task db:seed` are planned Windows Taskfile wrappers in Ticket `005` (`db/001_schema.sql`, `db/002_seed.sql`, `db/003_demo_seed.sql`). Until Ticket `005` is merged, run the direct `psql` commands above.

- **`db/001_schema.sql`** creates the PostgreSQL enums, tables (`astronaut`, `unit`, `metric`, `metric_threshold`, `device`, `measurement`, `measurement_value`, `media_asset`), triggers, and `v_astronaut_age` view. Because it is non-idempotent DDL, run it only once on a clean `hrig` database.
- **`db/002_seed.sql`** populates the idempotent `unit`, `metric`, `device`, and `metric_threshold` catalogs.
- **`db/003_demo_seed.sql`** populates the idempotent 3-day demo dataset (3 astronauts plus physiological, cabin, and EVA suit readings).

### Local Connection String

```env
DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable
```

---

## 4. Running the Backend API (`server/`)

The backend API lives in `server/` (`module hrig/server`, scaffolded as a bare-bones Chi v5 + Gorilla WebSocket server in Ticket `004`).

### Environment Configuration (`server/.env.example`)

Copy `server/.env.example` to `server/.env` (or export variables in your shell) when configuring the backend:

| Variable | Default / Example | Description |
| :--- | :--- | :--- |
| `HRIG_API_ADDR` | `:8080` | HTTP and WebSocket bind address |
| `HRIG_ALLOWED_ORIGINS` | `http://127.0.0.1:9245,http://localhost:9245,wails://wails` | Allowed CORS / WebSocket origins for Vite dev and Wails desktop |
| `DATABASE_URL` | `postgres://postgres@127.0.0.1:5432/hrig?sslmode=disable` | PostgreSQL connection string *(wired in Ticket `005`)* |

### Verification & Run Commands

From the repository root (`[Ticket 004]`):

```bash
wails3 task server:check  # Runs go vet, unit tests, and builds server
wails3 task server:dev    # Starts the Go HTTP + WebSocket API server
```

### Baseline HTTP & WebSocket Surface

| Protocol / Method | Path | Purpose |
| :--- | :--- | :--- |
| `GET` | `/health` | Basic server health check (`{"status":"ok"}`) |
| `GET` / `WS` | `/ws` | Bare-bones Gorilla WebSocket handshake & ping/echo test endpoint |

> Additional domain endpoints, services, models, and PostgreSQL integration are added incrementally after the bare-bones scaffold in `Ticket 004`.

---

## 5. Running the Desktop & Frontend App (`app/`)

### Option A — Frontend-Only UI Dev

Use this workflow when iterating on Vue 3 components, routing, or UnoCSS styling without compiling the Wails Go shell:

```bash
cd app/frontend
bun install
bun run dev
```

Available scripts in `app/frontend/package.json`:
- `bun run dev` — Starts Vite on `http://127.0.0.1:9245` (`WAILS_VITE_PORT`, `strictPort: true`).
- `bun run type-check` — Runs `vue-tsc --noEmit`.
- `bun run build` — Production bundle (`vite build --mode production`).
- `bun run build:dev` — Unminified development bundle (`vite build --minify false --mode development`).
- `bun run preview` — Serves the built bundle locally via `vite preview`.

> **Pre-generated `@bindings/` Note:** `app/frontend/bindings/` holds pre-generated TypeScript bindings (`@bindings` alias) so `bun run dev`, `type-check`, and `build` work standalone. However, live native desktop bridge calls require running inside the Wails v3 desktop shell (**Option B**).

### Option B — Full Desktop App (Wails v3)

Run from the **repository root**:

```bash
wails3 task client:check  # Regenerates TS bindings, runs type-check + vite build + go build .
wails3 task client:dev    # Launches Wails v3 desktop window with Vite hot-reload (127.0.0.1:9245)
```

Additional root `client:*` tasks in `Taskfile.yml`:
- `wails3 task client:bindings` — Regenerates TypeScript bindings in `app/frontend/bindings` from Go (`wails3 generate bindings -clean=true -ts`).
- `wails3 task client:build` — Builds the desktop binary for the current OS.
- `wails3 task client:start` — Builds and launches the compiled binary.

---

## 6. Conventions & Troubleshooting

- **Always use `wails3 task` instead of standalone `task`:** Root `Taskfile.yml` and `app/Taskfile.yml` are invoked via `wails3 task <name>` so every contributor uses the Task v3 runner embedded in Wails v3.
- **Vite Path Aliases (`app/frontend/vite.config.ts`):**
  - `@/` resolves to `app/frontend/src/`
  - `@bindings/` resolves to `app/frontend/bindings/`
- **Vite Dev Port (`127.0.0.1:9245`):** `app/frontend/vite.config.ts` binds strictly to `127.0.0.1` on `WAILS_VITE_PORT` (default `9245`, `strictPort: true`). If `9245` is already occupied, stop the conflicting process or override `WAILS_VITE_PORT` consistently for both Vite and Wails.
- **UnoCSS Import Order (`app/frontend/src/main.ts`):** `"virtual:uno.css"` **must** be imported **before** `"@unocss/reset/tailwind-compat.css"` so utility styles and reset rules cascade as configured. `app/frontend/uno.config.ts` enables `presetWind3`, `presetWebFonts` (Fontshare `Satoshi`), and `transformerDirectives` (`--uno`).
- **`minimum-release-age=10080` (`app/frontend/.npmrc`):** Bun enforces a 7-day (`10080` minutes) minimum package release age. If `bun install` or `bun add` rejects a freshly published package version, wait until the release is at least 7 days old or pin an older verified version.
- **Offline Fontshare `Satoshi` Timeout Warning:** `presetWebFonts` queries Fontshare for `Satoshi` during builds. In offline environments, UnoCSS prints a non-fatal timeout warning and continues; the build still succeeds.
- **Building Go in `app/` (`go build .` vs. `go build ./...`):** Inside `app/`, run `go build .` (not `go build ./...`), matching `app/Taskfile.yml` (`check` target), or run `wails3 task client:check` from the repository root.
