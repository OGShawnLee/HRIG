# 010 — Expose server read/catalog/trend/history REST endpoints

| Field | Value |
| :--- | :--- |
| **ID** | `010` |
| **Status** | `completed` |
| **Worker** | `self` (`inherit`) |
| **Role** | Go Server & SQLC Read API Engineer |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `shared/model/telemetry.go`, `shared/model/telemetry_test.go` (add `BaseUnitDimension`/`BaseUnitDecimals` to `Metric`, `Offset` to `HistoryQuery`, plus `TrendPoint` and `HistoryEntry` DTOs)
  - `server/internal/query/*.go` (`service.go`, `service_test.go`, `service_integration_test.go` — read service mapping `db.Querier` rows into `hrig/shared/model` DTOs)
  - `server/internal/http/router.go`, `server/internal/http/router_test.go` (extend `RouterDeps` with `Reader` and mount `GET /api/astronauts`, `GET /api/metrics`, `GET /api/trends`, `GET /api/history`)
  - `server/cmd/server/main.go` (construct `query.NewService(pool, cfg.MissionEpochUTC)` and wire into `RouterDeps`)
  - `.tickets/010-server-read-endpoints/changelog.md`
- **Must not touch:**
  - `bot/**`, `app/**`, `local/**`, `docs/**`
  - `db/001_schema.sql`, `db/002_seed.sql`, `db/003_demo_seed.sql`, `db/README.md`, `db/queries/**`
  - `sqlc.yaml`, `server/internal/db/*.go`
  - `server/internal/ingestion/**`, `server/internal/realtime/**`, `server/internal/config/**`
  - `Taskfile.yml`, `go.work`
  - `.gitignore`, `.routing/**`, `.claude/**`, `README.md`, `PROJECT-REVIEW.md`, `GUIDE.md`
  - `.tickets/001-*` .. `.tickets/009-*`, `.tickets/011-*`, `.tickets/012-*`
  - `.tickets/010-server-read-endpoints/task.md`, `.tickets/010-server-read-endpoints/spec.md`

## Inputs

Read these files before acting:
1. [`shared/model/telemetry.go`](../../shared/model/telemetry.go)
2. [`shared/model/telemetry_test.go`](../../shared/model/telemetry_test.go)
3. [`server/internal/db/querier.go`](../../server/internal/db/querier.go)
4. [`server/internal/db/astronauts.sql.go`](../../server/internal/db/astronauts.sql.go)
5. [`server/internal/db/catalogs.sql.go`](../../server/internal/db/catalogs.sql.go)
6. [`server/internal/db/telemetry.sql.go`](../../server/internal/db/telemetry.sql.go)
7. [`server/internal/realtime/met.go`](../../server/internal/realtime/met.go)
8. [`server/internal/http/router.go`](../../server/internal/http/router.go)
9. [`server/internal/http/router_test.go`](../../server/internal/http/router_test.go)
10. [`server/cmd/server/main.go`](../../server/cmd/server/main.go)

## Grounding & Anti-Hallucination Constraints

- **Reuse Existing SQLC Queries Only:**
  - Do **not** modify `db/queries/*.sql` or `server/internal/db/*.go`. Every read endpoint maps directly from existing `db.Querier` methods:
    - `GET /api/astronauts` → `ListActiveAstronauts(ctx)`
    - `GET /api/metrics` → `ListActiveMetrics(ctx)`
    - `GET /api/trends` → `ListTrendsByMetric(ctx, arg)`
    - `GET /api/history` → `ListHistory(ctx, arg)`
- **Shared Contract Discipline:**
  - All JSON response payloads must use `hrig/shared/model` structs (`[]model.Astronaut`, `[]model.Metric`, `[]model.TrendPoint`, `[]model.HistoryEntry`) so the Wails Go client in ticket `011` can decode and re-export them as TypeScript bindings.
  - Reuse `realtime.FormatMET(recordedAtUTC, missionEpochUTC)` for `MET` strings on `TrendPoint` and `HistoryEntry`.
  - Never expose or persist `"stale"` from PostgreSQL queries (`db.ReadingStatus` only stores `normal`, `abnormal`, `critical`).
- **Preserve Existing Behavior:**
  - Keep `NewRouter(cfg, ingester)` signature and existing routes (`GET /`, `POST /`, `GET /health`, `GET /ws/realtime`, `GET /api/snapshot`, `POST /api/alerts/{id}/ack`) working identically.
  - Keep `GET /api/*` routes returning JSON error objects `{"error":"..."}` (`400`, `500`, `503` when `Reader == nil`) matching `writeError`.
- **No Unsolicited Git / Daemon Side Effects:**
  - Do **not** commit or push changes.
  - Do **not** leave `hrig-server.exe` or `postgres.exe` running after verification, and drop `hrig_test` if created.

## Task

1. **Extend `shared/model/telemetry.go` & `shared/model/telemetry_test.go`:**
   - Add two optional catalog fields to `model.Metric`:
     - `BaseUnitDimension *string` (`json:"baseUnitDimension,omitempty"`)
     - `BaseUnitDecimals *int16` (`json:"baseUnitDecimals,omitempty"`)
   - Add `Offset int` (`json:"offset"`) to `model.HistoryQuery`.
   - Add `TrendPoint` struct to `shared/model/telemetry.go`:
     - `ID string` (`json:"id"`)
     - `MeasurementID string` (`json:"measurementId"`)
     - `MetricID string` (`json:"metricId"`)
     - `MetricCode string` (`json:"metricCode"`)
     - `AstronautID *string` (`json:"astronautId,omitempty"`)
     - `DeviceID string` (`json:"deviceId"`)
     - `Environment EnvironmentMode` (`json:"environment"`)
     - `UnitCode *string` (`json:"unitCode,omitempty"`)
     - `UnitSymbol *string` (`json:"unitSymbol,omitempty"`)
     - `ValueNumeric *float64` (`json:"valueNumeric,omitempty"`)
     - `ValueScale *int16` (`json:"valueScale,omitempty"`)
     - `ValueText *string` (`json:"valueText,omitempty"`)
     - `Status ReadingStatus` (`json:"status"`)
     - `RecordedAtUTC time.Time` (`json:"recordedAtUtc"`)
     - `MET string` (`json:"met"`)
   - Add `HistoryEntry` struct to `shared/model/telemetry.go`:
     - `MeasurementID string` (`json:"measurementId"`)
     - `ValueID string` (`json:"valueId"`)
     - `AstronautID *string` (`json:"astronautId,omitempty"`)
     - `AstronautName *string` (`json:"astronautName,omitempty"`)
     - `DeviceCode string` (`json:"deviceCode"`)
     - `DeviceName string` (`json:"deviceName"`)
     - `Environment EnvironmentMode` (`json:"environment"`)
     - `MetricCode string` (`json:"metricCode"`)
     - `MetricName string` (`json:"metricName"`)
     - `ValueType ValueType` (`json:"valueType"`)
     - `UnitCode *string` (`json:"unitCode,omitempty"`)
     - `UnitSymbol *string` (`json:"unitSymbol,omitempty"`)
     - `DisplayUnitSymbol *string` (`json:"displayUnitSymbol,omitempty"`)
     - `ValueNumeric *float64` (`json:"valueNumeric,omitempty"`)
     - `ValueScale *int16` (`json:"valueScale,omitempty"`)
     - `ValueText *string` (`json:"valueText,omitempty"`)
     - `Status ReadingStatus` (`json:"status"`)
     - `RecordedAtUTC time.Time` (`json:"recordedAtUtc"`)
     - `MET string` (`json:"met"`)
     - `MeasurementNote *string` (`json:"measurementNote,omitempty"`)
     - `ValueNote *string` (`json:"valueNote,omitempty"`)
     - `MediaAssetID *string` (`json:"mediaAssetId,omitempty"`)
     - `MediaKind *MediaKind` (`json:"mediaKind,omitempty"`)
     - `MediaFilePath *string` (`json:"mediaFilePath,omitempty"`)
     - `MediaMimeType *string` (`json:"mediaMimeType,omitempty"`)
     - `MediaFindings *string` (`json:"mediaFindings,omitempty"`)
     - `MediaCapturedAtUTC *time.Time` (`json:"mediaCapturedAtUtc,omitempty"`)
   - Add JSON round-trip test coverage in `shared/model/telemetry_test.go` for `TrendPoint`, `HistoryEntry`, and the updated `Metric` fields, then run `GOWORK=off go -C shared test ./...`.
2. **Create Read Service in `server/internal/query/service.go`:**
   - Define `TrendsFilter` (`MetricCode string`, `AstronautID *string`, `SinceUTC *time.Time`, `Limit int`) and `HistoryFilter` (`AstronautID *string`, `Limit int`, `Offset int`), plus sentinel `ErrBadRequest`.
   - Implement `Service` backed by `db.Querier` and `missionEpochUTC time.Time`:
     - `NewService(pool *pgxpool.Pool, missionEpochUTC time.Time) *Service` and `NewServiceWithQuerier(q db.Querier, missionEpochUTC time.Time) *Service`.
     - `ListAstronauts(ctx context.Context) ([]model.Astronaut, error)` calling `q.ListActiveAstronauts(ctx)` and mapping UUIDs (`pgtype.UUID`), dates (`pgtype.Date`), timestamps (`pgtype.Timestamptz`), `Age`, `FullName`, `NationalityCode`, `NationalityLabel`, `IsActive`.
     - `ListMetrics(ctx context.Context) ([]model.Metric, error)` calling `q.ListActiveMetrics(ctx)` and mapping all metric fields, unit metadata (`BaseUnitCode`, `BaseUnitSymbol`, `BaseUnitDimension`, `BaseUnitDecimals`, `DisplayUnitSymbol`), `DisplayOrder` (default `0` if `nil`), and populating `Threshold *model.MetricThreshold` whenever threshold fields (`ClassificationMode != nil` or any min/max numeric is valid).
     - `ListTrends(ctx context.Context, filter TrendsFilter) ([]model.TrendPoint, error)`:
       - Validate `metric_code` is non-empty after trim (`ErrBadRequest`).
       - If `filter.AstronautID` is non-empty, parse as UUID (`pgtype.UUID.Scan`); return `ErrBadRequest` on invalid UUID.
       - Normalize `limit`: if `<= 0` default to `200`; cap at `1000` (if `> 1000`, clamp or return `ErrBadRequest` — clamp at `1000` when `> 1000` or enforce `1..1000`; return `ErrBadRequest` if explicitly negative). Wait: enforce explicit bounds cleanly: if `limit == 0` use `200`; if `limit < 0 || limit > 1000` return `ErrBadRequest`. If `len(rows) > limit`, return the trailing `limit` entries (`rows[len(rows)-limit:]`) so chronological `ASC` order contains the latest `limit` points.
       - Convert `pgtype.Numeric` via `Float64Value()` and compute `MET` via `realtime.FormatMET`.
     - `ListHistory(ctx context.Context, filter HistoryFilter) ([]model.HistoryEntry, error)`:
       - Validate optional `AstronautID` UUID (`ErrBadRequest` if malformed).
       - Validate bounds: if `limit == 0` default `50`; if `limit < 0 || limit > 500` return `ErrBadRequest`; if `offset < 0` return `ErrBadRequest`.
       - Call `q.ListHistory(ctx, db.ListHistoryParams{...})` and map all columns cleanly, converting nullable media fields (`MediaAssetID`, `MediaKind`, `MediaFilePath`, `MediaMimeType`, `MediaFindings`, `MediaCapturedAtUtc`) only when `MediaAssetID.Valid`.
3. **Mount Read Endpoints in `server/internal/http/router.go` & Wire `server/cmd/server/main.go`:**
   - Define `TelemetryReader` interface in `server/internal/http/router.go`:
     - `ListAstronauts(ctx context.Context) ([]model.Astronaut, error)`
     - `ListMetrics(ctx context.Context) ([]model.Metric, error)`
     - `ListTrends(ctx context.Context, filter query.TrendsFilter) ([]model.TrendPoint, error)`
     - `ListHistory(ctx context.Context, filter query.HistoryFilter) ([]model.HistoryEntry, error)`
   - Add `Reader TelemetryReader` to `RouterDeps`.
   - Mount:
     - `GET /api/astronauts` → `200 OK` JSON array `[]model.Astronaut` (`503` if `Reader == nil`).
     - `GET /api/metrics` → `200 OK` JSON array `[]model.Metric` (`503` if `Reader == nil`).
     - `GET /api/trends` → parse query params `metric_code`, `astronaut_id`, `since` (RFC3339/RFC3339Nano), `limit` (integer); return `400` on malformed query parameters, `200 OK` JSON array `[]model.TrendPoint`.
     - `GET /api/history` → parse query params `astronaut_id`, `limit` (integer), `offset` (integer); return `400` on malformed query parameters, `200 OK` JSON array `[]model.HistoryEntry`.
   - Update `server/cmd/server/main.go` to construct `reader := query.NewService(pool, cfg.MissionEpochUTC)` and pass `Reader: reader` in `RouterDeps`.
4. **Write Unit & Integration Tests and Run Verification:**
   - Unit tests in `server/internal/query/service_test.go` using a mock `db.Querier`: verify astronaut mapping, metric + threshold + `BaseUnitDecimals` mapping, trend slicing/UUID validation, history pagination/media mapping.
   - HTTP tests in `server/internal/http/router_test.go`: verify `GET /api/astronauts`, `GET /api/metrics`, `GET /api/trends` (`200` and `400` cases), `GET /api/history` (`200` and `400` cases), and `503` when `Reader == nil`.
   - PostgreSQL integration test in `server/internal/query/service_integration_test.go` against `HRIG_TEST_DATABASE_URL` (`001_schema.sql` + `002_seed.sql` + `003_demo_seed.sql`): verify `ListAstronauts` returns the 3 active seeded astronauts (`Ana Torres`, `Liam Okafor`, `Sofia Petrov`) with valid UUIDs and derived `Age`, `ListMetrics` returns all 23 active metrics with thresholds and unit decimals, `ListTrends` returns chronological readings for `spo2` & `cabin_o2`, and `ListHistory` honors `limit`/`offset` and optional `astronaut_id`.
   - Run `GOWORK=off go -C shared test ./...`, `GOWORK=off go -C server vet ./...`, `wails3 task server:check`, and PostgreSQL `hrig_test` integration tests.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/010-server-read-endpoints/changelog.md` listing:
- Files created and modified
- Shared, unit, HTTP, PostgreSQL integration (`hrig_test`) test outputs
- Any blockers or assumptions avoided
