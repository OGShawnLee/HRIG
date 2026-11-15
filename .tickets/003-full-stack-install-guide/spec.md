# 003 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Guide Coverage & Accuracy
- [ ] `docs/STACK_SETUP.md` exists and is written in English.
- [ ] `docs/STACK_SETUP.md` lists exact frontend/desktop package versions matching `app/frontend/package.json` and `app/go.mod`.
- [ ] `docs/STACK_SETUP.md` documents the monorepo Go workspace (`app/`, `server/`, `shared/`) and `shared/model` role for Wails v3 TypeScript bindings.
- [ ] `docs/STACK_SETUP.md` covers all three layers: Desktop/Frontend (`app/`), Backend API (`server/`), and PostgreSQL 17+ across Windows (portable `local/pgsql/` + `debian/bookworm` VM/WSL2), macOS (`OrbStack`/`Colima`, `Homebrew`, `Postgres.app`), and Linux (`Docker` / native `postgresql-17`).
- [ ] `docs/STACK_SETUP.md` documents exact commands for `db/001_schema.sql`, `db/002_seed.sql`, and `db/003_demo_seed.sql` (and notes `001_schema.sql` runs once on a fresh DB while seeds are idempotent).
- [ ] `docs/STACK_SETUP.md` documents `127.0.0.1:9245`, `@/` and `@bindings/` aliases, `minimum-release-age=10080`, `wails3 task db:*` / `server:*` / `client:*`, and using `wails3 task` instead of standalone `task`.
- [ ] `docs/STACK_SETUP.md` contains no `[PENDING DB SPEC]` placeholders and no invented table names or unverified clinical thresholds.

## Navigation & Context Updates
- [ ] `README.md` links to `docs/STACK_SETUP.md` under `### Core Project Docs`.
- [ ] `.routing/CONTEXT.md` links to `../docs/STACK_SETUP.md` under `External project docs`.

## Scope & Ground Rules
- [ ] `git status --short` shows modifications only inside `docs/STACK_SETUP.md`, `README.md`, `.routing/CONTEXT.md`, and `.tickets/003-full-stack-install-guide/`.
- [ ] `app/**`, `server/**`, `shared/**`, `db/**`, `Taskfile.yml`, `.gitignore`, `.claude/**`, `PROJECT-REVIEW.md`, `docs/METRICS_PRIORIZATION.md`, and other `.tickets/**` are untouched.
- [ ] No unapproved git commits or pushes were executed.
- [ ] `.tickets/003-full-stack-install-guide/changelog.md` exists and records files touched and verification results.
