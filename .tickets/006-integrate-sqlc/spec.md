# 006 — Verification Spec

All checks must pass with binary Yes/No evidence.

## SQLC Configuration & Queries
- [x] `sqlc.yaml` exists at repo root (`version: "2"`, `engine: "postgresql"`, `schema: ["db/001_schema.sql"]`, `queries: ["db/queries"]`, `out: "server/internal/db"`, `sql_package: "pgx/v5"`).
- [x] `db/queries/astronauts.sql`, `db/queries/catalogs.sql`, and `db/queries/telemetry.sql` exist, include active lookup queries (`ListActiveAstronautsByFullName`, `GetActiveDeviceByCode`, `GetActiveMetricByCode`, `GetActiveUnitByCode`) and write queries (`CreateMeasurement`, `CreateMeasurementValue`), and reference only tables/views/columns defined in `db/001_schema.sql`.
- [x] `db/001_schema.sql`, `db/002_seed.sql`, and `db/003_demo_seed.sql` are completely unmodified.

## Code Generation & Build Pipeline
- [x] `Taskfile.yml` defines `db:sqlc` (`sqlc generate`) while preserving `client:*`, `server:*`, and `db:*` tasks.
- [x] `wails3 task db:sqlc` exits `0` and generates `server/internal/db/db.go`, `models.go`, `querier.go`, `astronauts.sql.go`, `catalogs.sql.go`, and `telemetry.sql.go`.
- [x] `server/go.mod` includes `github.com/jackc/pgx/v5` (and no ORM dependencies).
- [x] `wails3 task server:check` exits `0` (`go vet ./...`, `go test ./...`, and `go build` all pass).
- [x] `wails3 task client:check` exits `0`.

## Scope & Ground Rules
- [x] `git status --short` shows no changes in `app/**`, `db/001_schema.sql`, `db/002_seed.sql`, `db/003_demo_seed.sql`, `server/cmd/server/main.go`, `server/internal/http/**`, `.routing/**`, `docs/**`, `README.md`, `GUIDE.md`, or `.claude/**`.
- [x] No git commits or pushes were executed.
- [x] `.tickets/006-integrate-sqlc/changelog.md` exists and records files touched, `sqlc version`, and command outputs.
