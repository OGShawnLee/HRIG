# 007 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Config, Bot & Taskfile
- [x] `server/internal/config/config.go` defaults `HRIG_API_ADDR` to `:80` and reads `DATABASE_URL` (with `HRIG_DATABASE_URL` fallback); `server/.env.example` documents both.
- [x] `bot/go.mod` exists (`module hrig/bot`) and `go vet ./...` + `go build -o ../bin/hrig-bot.exe .` inside `bot/` exit `0` without modifying `bot/main.go`.
- [x] `Taskfile.yml` defines `bot:dev`, `db:migrate`, and `db:seed` while preserving `client:*`, `server:*`, and existing `db:*` tasks.

## Transactional Ingestion & Threshold Classification (`GUIDE.md`)
- [x] `server/internal/http/router.go` preserves `GET /` (`200 OK` `{"message":"hello world"}`) and mounts `POST /`.
- [x] `server/cmd/server/main.go` requires `DatabaseURL`, connects and pings `pgxpool.Pool` on startup (failing fast if unavailable), and closes the pool cleanly on shutdown.
- [x] `server/internal/ingestion` validates JSON body size/syntax, `environment` (`iva`/`eva`), `recorded_at_utc` with timezone, non-empty `readings`, active `device_code` (`404`), unique active `astronaut_name` (`404`/`409`), active `metric_code` (`404`), `metric_type` vs `astronaut_id` rule (`400`), `value_type` payload matching (`400`), `unit_code` dimension matching (`404`/`400`), and duplicate `metric_code` (`409`).
- [x] `ClassifyReading` evaluates `critical` (`<= critical_min`, `>= critical_max`) before `abnormal` (`< warning_min`, `> warning_max`), returns `normal` for `manual` mode (`ultrasound`), and persists `status` inside the transaction.
- [x] `POST /` uses a single PostgreSQL transaction via `server/internal/db` and rolls back parent `measurement` if any reading fails.

## Automated Tests & Live End-to-End Smoke Verification
- [x] `wails3 task db:migrate` and `wails3 task db:seed` exit `0` against local `hrig` DB.
- [x] `HRIG_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:5432/hrig_test?sslmode=disable go test -v ./...` inside `server/` exits `0` with all unit and integration tests passing (covering all 7 `bot` sensor shapes and rollback checks), and `hrig_test` is dropped afterward.
- [x] Live smoke run with `hrig-server.exe` + `bot` completes at least one full 17-request batch (`lote de lecturas enviado correctamente`) with persisted rows confirmed in PostgreSQL `hrig`, and no background server/bot processes remain running.
- [x] `wails3 task server:check` and `wails3 task client:check` both exit `0`.

## Scope & Ground Rules
- [x] `git status --short` shows no modifications to `db/001_schema.sql`, `db/002_seed.sql`, `db/003_demo_seed.sql`, `db/queries/**`, `sqlc.yaml`, `server/internal/db/**`, `bot/main.go`, `app/**`, `.routing/**`, `docs/**`, `README.md`, `GUIDE.md`, or `.claude/**`.
- [x] No git commits or pushes were executed.
- [x] `.tickets/007-guide-sensor-ingestion/changelog.md` exists and records files touched and all command outputs.
