# 002 — Set up portable PostgreSQL 17 inside ignored `local/pgsql/`

| Field | Value |
| :--- | :--- |
| **ID** | `002` |
| **Status** | `completed` |
| **Worker** | `antigravity` / `gemini-4-hydrogen` (`thinking_level: medium`) |
| **Role** | Local Database Setup Engineer |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `.gitignore` (only add `/local/pgsql/` to ignore the local portable PostgreSQL binaries, data directory, and logfile)
  - `Taskfile.yml` (only add `db:start`, `db:stop`, `db:status`, and `db:psql` tasks scoped to `platforms: [windows]`)
  - `local/pgsql/**` (untracked local runtime directory; must remain ignored by Git)
  - `.tickets/002-setup-portable-postgres-17/changelog.md`
- **Must not touch:**
  - `app/**`
  - `.routing/**`
  - `.claude/**`
  - `docs/**`, `README.md`, `PROJECT-REVIEW.md`
  - `.tickets/001-scaffold-wails3-app/**`
  - `.tickets/002-setup-portable-postgres-17/task.md`, `.tickets/002-setup-portable-postgres-17/spec.md`
  - Any other `.tickets/**` folders

## Inputs

Read these files before acting:
1. [`.gitignore`](../../.gitignore)
2. [`Taskfile.yml`](../../Taskfile.yml)
3. [`.routing/CONTEXT.md`](../../.routing/CONTEXT.md) (Ground Rules section)

## Grounding & Anti-Hallucination Constraints

- **Platform & Binary Source (verified EDB Windows x64 archive):**
  - URL: `https://get.enterprisedb.com/postgresql/postgresql-17.11-5-windows-x64-binaries.zip`
  - Expected SHA-256 checksum: `80379b2c04d51c30225532e0ae04509899141e9957ed096fe749d7fd9df8f82f`
  - Download the `.zip` into `$env:TEMP\postgresql-17.11-5-windows-x64-binaries.zip`, verify its SHA-256 matches `80379b2c04d51c30225532e0ae04509899141e9957ed096fe749d7fd9df8f82f` (using `Get-FileHash -Algorithm SHA256`), extract into `local\` (producing `local\pgsql\bin\postgres.exe`, `pg_ctl.exe`, `initdb.exe`, `psql.exe`, `pg_isready.exe`), and delete the temporary `.zip` from `$env:TEMP` after extraction.
  - If the download fails or the SHA-256 does not match, **stop**, do not guess another mirror, and record the exact error in `changelog.md`.
- **Cluster Initialization Parameters (local loopback dev cluster only):**
  - Data directory: `local/pgsql/data`
  - Superuser: `postgres`
  - Encoding & locale: `-E UTF8 --locale=C`
  - Local auth: `-A trust` (bound strictly to `127.0.0.1` port `5432`; never bind to `0.0.0.0`)
  - Default database to create: `hrig` (`createdb.exe -h 127.0.0.1 -p 5432 -U postgres hrig`)
- **Do NOT create SQL tables, schemas, or seed rows** — the DB specification has not been received yet.
- **Do NOT install a Windows Service** (`pg_ctl register`) or modify system-wide `PATH` / Windows Registry.
- **Taskfile Cross-Platform Safety:** All `db:*` tasks added to `Taskfile.yml` must declare `platforms: [windows]` so Linux/macOS teammates are unaffected, and existing `client:*` tasks must remain intact. Always invoke tasks via `wails3 task <name>`.
- **Do NOT commit or push** any changes to Git.

## Task

1. Update `.gitignore` to append `/local/pgsql/` on a new line while keeping `!app/build/` intact.
2. Create the `local` directory at the repository root and confirm `git check-ignore -v local/pgsql` matches `/local/pgsql/` in `.gitignore`.
3. Download `https://get.enterprisedb.com/postgresql/postgresql-17.11-5-windows-x64-binaries.zip` to `$env:TEMP\postgresql-17.11-5-windows-x64-binaries.zip`, verify `Get-FileHash` SHA-256 equals `80379b2c04d51c30225532e0ae04509899141e9957ed096fe749d7fd9df8f82f`, extract the archive into `local\` so `local\pgsql\bin\pg_ctl.exe` exists, and remove the `.zip` from `$env:TEMP`.
4. Initialize the PostgreSQL 17 cluster at `local\pgsql\data` by running:
   `.\local\pgsql\bin\initdb.exe -D .\local\pgsql\data -U postgres -E UTF8 --locale=C -A trust`
5. Append `listen_addresses = '127.0.0.1'` and `port = 5432` to `local\pgsql\data\postgresql.conf` so the server only binds to loopback.
6. Update `Taskfile.yml` to add these four Windows-only tasks without altering `client:*`:
   - `db:start`: runs `.\local\pgsql\bin\pg_ctl.exe -D .\local\pgsql\data -l .\local\pgsql\logfile -o "-h 127.0.0.1 -p 5432" start`
   - `db:stop`: runs `.\local\pgsql\bin\pg_ctl.exe -D .\local\pgsql\data stop -m fast`
   - `db:status`: runs `.\local\pgsql\bin\pg_ctl.exe -D .\local\pgsql\data status`
   - `db:psql`: runs `.\local\pgsql\bin\psql.exe -h 127.0.0.1 -p 5432 -U postgres -d hrig`
7. Run `wails3 task db:start`, verify `.\local\pgsql\bin\pg_isready.exe -h 127.0.0.1 -p 5432` exits `0`, create database `hrig` via `.\local\pgsql\bin\createdb.exe -h 127.0.0.1 -p 5432 -U postgres hrig`, verify `.\local\pgsql\bin\psql.exe -h 127.0.0.1 -p 5432 -U postgres -d hrig -c "SELECT version();"` outputs `PostgreSQL 17.11`, and stop the server cleanly with `wails3 task db:stop`.
8. Run `wails3 task client:check` to confirm existing tasks remain healthy, and run `git status --short` to confirm `local/pgsql/` does not appear as untracked or modified.
9. Write `.tickets/002-setup-portable-postgres-17/changelog.md` listing every file touched, exact command outputs, and the SHA-256 verification result.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/002-setup-portable-postgres-17/changelog.md` listing:
- Files created and modified
- Exact SHA-256 hash verified and download URL used
- Verification commands executed and their outputs
- Any blockers or assumptions avoided
