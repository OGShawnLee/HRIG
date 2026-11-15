# 006 — Changelog

## Summary

Integrated `sqlc` (`v1.28.0`) with the PostgreSQL 17+ schema (`db/001_schema.sql`), added typed SQL query definitions under `db/queries/`, added `github.com/jackc/pgx/v5` (`v5.13.0`) to `server/go.mod`, and generated Go database bindings under `server/internal/db/`.

## Environment

- `sqlc version`: `v1.28.0`

## Files Created & Modified

### Created
- `sqlc.yaml` — SQLC v2 configuration (`engine: "postgresql"`, `schema: ["db/001_schema.sql"]`, `queries: ["db/queries"]`, `gen.go.package: "db"`, `gen.go.out: "server/internal/db"`, `gen.go.sql_package: "pgx/v5"`, `emit_json_tags: true`, `emit_interface: true`, `emit_empty_slices: true`, `emit_pointers_for_null_types: true`).
- `db/queries/astronauts.sql` — `ListActiveAstronauts`, `GetAstronautByID`, `ListActiveAstronautsByFullName`.
- `db/queries/catalogs.sql` — `ListUnits`, `GetActiveUnitByCode`, `ListActiveMetrics`, `GetActiveMetricByCode`, `GetThresholdByMetricCode`, `ListActiveDevices`, `GetActiveDeviceByCode`.
- `db/queries/telemetry.sql` — `CreateMeasurement`, `CreateMeasurementValue`, `CreateMediaAsset`, `ListLatestReadingsByMetric`, `ListTrendsByMetric`, `ListHistory`.
- Generated files under `server/internal/db/`:
  - `server/internal/db/db.go`
  - `server/internal/db/models.go`
  - `server/internal/db/querier.go`
  - `server/internal/db/astronauts.sql.go`
  - `server/internal/db/catalogs.sql.go`
  - `server/internal/db/telemetry.sql.go`
- `.tickets/006-integrate-sqlc/changelog.md` — this changelog.

### Modified
- `server/go.mod`, `server/go.sum` — added `github.com/jackc/pgx/v5 v5.13.0` and its transitive dependencies (`github.com/jackc/pgpassfile`, `github.com/jackc/pgservicefile`, `golang.org/x/text`).
- `go.work.sum` — updated by Go workspace resolution after adding `pgx/v5` to `server/go.mod`.

### Preserved (Unmodified)
- `Taskfile.yml` — verified `db:sqlc` (`sqlc generate`) alongside `client:*`, `server:*`, and `db:*` tasks; no edits required.
- `db/001_schema.sql`, `db/002_seed.sql`, `db/003_demo_seed.sql`, `app/**`, `server/cmd/server/main.go`, `server/internal/http/**` — untouched. (Note: `app/frontend/bindings/hrig/app/app.ts` dirtied by `wails3 task client:check` was restored via `git checkout -- app/frontend/bindings/`.)

## Verification Commands & Outputs

1. **`sqlc version`**
   - Output: `v1.28.0`

2. **`wails3 task db:sqlc`** (exit code `0`)
   ```text
   task: [db:sqlc] sqlc generate
   ```

3. **`go mod tidy`** (inside `server/`, exit code `0`)
   - Resolved `github.com/jackc/pgx/v5 v5.13.0` cleanly with no ORM dependencies.

4. **`wails3 task server:check`** (exit code `0`)
   ```text
   task: [server:check] go vet ./...
   task: [server:check] go test ./...
   ?   	hrig/server/cmd/server	[no test files]
   ?   	hrig/server/internal/config	[no test files]
   ?   	hrig/server/internal/db	[no test files]
   ok  	hrig/server/internal/http	(cached)
   task: [server:check] go build -o ../bin/hrig-server.exe ./cmd/server
   ```

5. **`wails3 task client:check`** (exit code `0`)
   ```text
   task: [client:check] wails3 task check
   task: [bindings] wails3 generate bindings -clean=true -ts
   Wails v3.0.0 › Generate Bindings
    INFO  Processed: 261 Packages, 1 Service, 5 Methods, 0 Enums, 0 Models, 0 Events in 15.7788063s.
    INFO  Output directory: C:\Users\OGSha\OneDrive\Escritorio\Developer\Web Development\HRIG\app\frontend\bindings
   task: [check] bun run --cwd frontend type-check
   $ vue-tsc --noEmit
   task: [check] bun run --cwd frontend build
   $ vite build --mode production
   ...
   ✓ built in 7.12s
   task: [check] go build .
   ```

6. **`git status --short`** (after `git checkout -- app/frontend/bindings/`)
   ```text
    M .routing/PROGRESS.md
    M .tickets/006-integrate-sqlc/spec.md
    M .tickets/006-integrate-sqlc/task.md
    M go.work.sum
    M server/go.mod
    M server/go.sum
   ?? GUIDE.md
   ?? bot/
   ?? db/queries/
   ?? server/internal/db/
   ?? sqlc.yaml
   ```
   (Pre-existing uncommitted entries `.routing/PROGRESS.md`, `spec.md`, `task.md`, `GUIDE.md`, and `bot/` were not modified or created by this ticket.)

## Assumptions & Blockers

- None. All queries map 1:1 to `db/001_schema.sql` tables, views, enums, and columns without schema modifications.
- Transient `go get` memory allocation failure on Windows while running concurrently with `sqlc` was avoided by running `go mod tidy` sequentially with `GOMAXPROCS=2`.
