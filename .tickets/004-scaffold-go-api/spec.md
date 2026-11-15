# 004 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Minimal Server Scaffold (`server/`)
- [x] `server/go.mod` declares `module hrig/server` (`go 1.26.0`) and requires only `github.com/go-chi/chi/v5` and `github.com/go-chi/cors` (no `hrig/shared`, no `gorilla/websocket`, no SQL/pgx driver).
- [x] `server/.env.example` exists and documents `HRIG_API_ADDR` and `HRIG_ALLOWED_ORIGINS`.
- [x] `server/internal/config/config.go` loads `HRIG_API_ADDR` (`:8080` default) and `HRIG_ALLOWED_ORIGINS`.
- [x] `server/internal/http/router.go` exposes only `GET /` returning `200 OK` JSON `{"message":"hello world"}` (with `middleware.Recoverer` and `cors.Handler`).
- [x] No extra routes, WebSockets, stores, thresholds, or domain models exist in `server/`.
- [x] `server/cmd/server/main.go` boots `net/http.Server` and handles graceful shutdown on `SIGINT`/`SIGTERM`.
- [x] `server/internal/http/router_test.go` verifies `GET /` returns `200 OK` and `{"message":"hello world"}`.

## Taskfile & Build Pipeline
- [x] `Taskfile.yml` includes `server:check`, `server:test`, `server:build`, and `server:dev` while preserving `client:*` and `db:*` tasks.
- [x] `wails3 task server:check` exits `0` (`go vet`, `go test`, and `go build` all pass).
- [x] `wails3 task client:check` exits `0`.

## Scope & Ground Rules
- [x] `git status --short` shows no changes in `app/**`, `db/**`, `.routing/**`, `docs/**`, `README.md`, `PROJECT-REVIEW.md`, or `.claude/**`.
- [x] No git commits or pushes were executed.
- [x] `.tickets/004-scaffold-go-api/changelog.md` exists and records files touched and command outputs.
