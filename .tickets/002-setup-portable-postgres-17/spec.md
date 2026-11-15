# 002 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Git Ignore & Placement
- [ ] `.gitignore` contains `/local/pgsql/` and still preserves `!app/build/`.
- [ ] `local/pgsql/bin/postgres.exe`, `local/pgsql/bin/initdb.exe`, `local/pgsql/bin/pg_ctl.exe`, `local/pgsql/bin/psql.exe`, and `local/pgsql/data/PG_VERSION` exist.
- [ ] `local/pgsql/data/PG_VERSION` contains `17`.
- [ ] `git check-ignore -v local/pgsql/bin/postgres.exe` and `git check-ignore -v local/pgsql/data` exit `0` and match `.gitignore`.
- [ ] `git status --short` outputs no entries under `local/`.

## Cluster & Taskfile Operation
- [ ] `Taskfile.yml` defines `db:start`, `db:stop`, `db:status`, and `db:psql` with `platforms: [windows]` while keeping all `client:*` tasks intact.
- [ ] `wails3 task db:start` exits `0`.
- [ ] `.\local\pgsql\bin\pg_isready.exe -h 127.0.0.1 -p 5432` exits `0` (`accepting connections`).
- [ ] `.\local\pgsql\bin\psql.exe -h 127.0.0.1 -p 5432 -U postgres -d hrig -c "SELECT version();"` exits `0` and prints `PostgreSQL 17.11`.
- [ ] `wails3 task db:status` exits `0` while running, and `wails3 task db:stop` exits `0` leaving no running `postgres.exe` bound to `5432`.
- [ ] `wails3 task client:check` exits `0`.

## Scope & Ground Rules
- [ ] `git status --short` shows modifications only in `.gitignore`, `Taskfile.yml`, and `.tickets/002-setup-portable-postgres-17/changelog.md`.
- [ ] No SQL tables, migration files, or Windows Services were created.
- [ ] No git commits or pushes were executed.
- [ ] `.tickets/002-setup-portable-postgres-17/changelog.md` exists and records SHA-256 verification and command outputs.
