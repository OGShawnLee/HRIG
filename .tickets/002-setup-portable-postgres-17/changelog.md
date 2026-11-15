# 002 — Changelog: Portable PostgreSQL 17 Setup

## Summary

Set up portable PostgreSQL 17.11 inside the git-ignored `local/pgsql/` directory, initialized a local loopback-only development cluster (`127.0.0.1:5432`, user `postgres`, database `hrig`), and added Windows-only `db:*` Taskfile commands.

## Files Created & Modified

### Modified (Tracked)
- `.gitignore`: Appended `/local/pgsql/` on line 7 while preserving `!app/build/`.
- `Taskfile.yml`: Added `db:start`, `db:stop`, `db:status`, and `db:psql` tasks scoped with `platforms: [windows]` while leaving all `client:*` tasks intact.

### Created (Untracked in Repo)
- `.tickets/002-setup-portable-postgres-17/changelog.md`: This verification and execution record.

### Created (Git-Ignored Runtime Directory)
- `local/pgsql/**`: Portable PostgreSQL 17.11 binaries, initialized cluster data directory (`local/pgsql/data`), and runtime logfile (`local/pgsql/logfile`).

## Binary Download & SHA-256 Verification

- **Download URL**: `https://get.enterprisedb.com/postgresql/postgresql-17.11-5-windows-x64-binaries.zip`
- **Temporary Archive**: `$env:TEMP\postgresql-17.11-5-windows-x64-binaries.zip` (deleted after extraction into `local\`)
- **Expected SHA-256**: `80379b2c04d51c30225532e0ae04509899141e9957ed096fe749d7fd9df8f82f`
- **Actual SHA-256 (`Get-FileHash -Algorithm SHA256`)**: `80379b2c04d51c30225532e0ae04509899141e9957ed096fe749d7fd9df8f82f` (**MATCH**)

## Cluster Initialization & Configuration

- Initialized cluster via:
  ```powershell
  .\local\pgsql\bin\initdb.exe -D .\local\pgsql\data -U postgres -E UTF8 --locale=C -A trust
  ```
- Appended loopback-only settings to `local\pgsql\data\postgresql.conf`:
  ```ini
  listen_addresses = '127.0.0.1'
  port = 5432
  ```
- Verified `local\pgsql\data\PG_VERSION` contains `17`.

## Verification Commands & Outputs

### 1. Git Ignore Checks
```text
> git check-ignore -v local/pgsql/bin/postgres.exe
.gitignore:7:/local/pgsql/	local/pgsql/bin/postgres.exe

> git check-ignore -v local/pgsql/data
.gitignore:7:/local/pgsql/	local/pgsql/data
```
Both exited `0`.

### 2. PostgreSQL Cluster Lifecycle & Query Verification
Executed inside a single session (`wails3 task db:start` -> readiness -> status -> `createdb` -> `psql` -> `wails3 task db:stop`):
```text
task: [db:start] ./local/pgsql/bin/pg_ctl.exe -D ./local/pgsql/data -l ./local/pgsql/logfile -o "-h 127.0.0.1 -p 5432" start
esperando que el servidor se inicie..... listo
servidor iniciado
db:start exit: 0

127.0.0.1:5432 - aceptando conexiones
pg_isready exit: 0

task: [db:status] ./local/pgsql/bin/pg_ctl.exe -D ./local/pgsql/data status
pg_ctl: el servidor está en ejecución (PID: 13884)
C:/Users/OGSha/OneDrive/Escritorio/Developer/Web Development/HRIG/local/pgsql/bin/postgres.exe "-D" "local/pgsql/data" "-h" "127.0.0.1" "-p" "5432"
db:status exit: 0

createdb exit: 0

                                 version
--------------------------------------------------------------------------
 PostgreSQL 17.11 on x86_64-windows, compiled by msvc-19.44.35229, 64-bit
(1 fila)
psql exit: 0

task: [db:stop] ./local/pgsql/bin/pg_ctl.exe -D ./local/pgsql/data stop -m fast
esperando que el servidor se detenga...... listo
servidor detenido
db:stop exit: 0

Port 5432 connections after stop: 0
postgres processes after stop: 0
```

### 3. Existing Taskfile Health (`wails3 task client:check`)
```text
task: [client:check] wails3 task check
task: [bindings] wails3 generate bindings -clean=true -ts
Wails v3.0.0 › Generate Bindings
 INFO  Processed: 261 Packages, 1 Service, 5 Methods, 0 Enums, 0 Models, 0 Events in 45.1689671s.
 INFO  Output directory: C:\Users\OGSha\OneDrive\Escritorio\Developer\Web Development\HRIG\app\frontend\bindings
task: [check] bun run --cwd frontend type-check
$ vue-tsc --noEmit
task: [check] bun run --cwd frontend build
$ vite build --mode production
✓ 59 modules transformed.
✓ built in 1.97s
task: [check] go build .
client:check exit: 0
```
*Note*: `wails3 task client:check` regenerated `app/frontend/bindings/hrig/app/app.ts` as a side effect of `wails3 generate bindings -clean=true -ts`; it was restored immediately via `git checkout -- app/frontend/bindings/hrig/app/app.ts` so `app/**` remains untouched by this ticket.

### 4. Git Status & Scope Check
```text
> git status --short
 M .gitignore
RM docs/bd.md -> .routing/BD.md
RM docs/bdml.md -> .routing/BDML.md
 M .routing/CONTEXT.md
 M .routing/PROGRESS.md
 M README.md
 M Taskfile.yml
 M db/001_schema.sql
 M db/002_seed.sql
?? .routing/TEMP-DB-TESTING.md
?? .tickets/002-setup-portable-postgres-17/
?? .tickets/003-full-stack-install-guide/
?? .tickets/004-scaffold-go-api/
?? .tickets/005-postgres-db-integration/
```
No entries under `local/` appear in `git status --short`. Changes outside `.gitignore`, `Taskfile.yml`, and `.tickets/002-setup-portable-postgres-17/` were pre-existing in the working tree and were not modified.

## Ground Rules & Constraints Compliance

- No SQL tables, schemas, migrations, or seed rows were created.
- No Windows Services (`pg_ctl register`), system `PATH` edits, or Registry changes were made (`Get-Service *postgres*` returned empty).
- Forward-slash paths (`./local/pgsql/...`) were used in `Taskfile.yml` so `wails3 task` (`mvdan.cc/sh`) executes paths cleanly without backslash escaping issues on Windows.
- No `git commit` or `git push` commands were executed.
