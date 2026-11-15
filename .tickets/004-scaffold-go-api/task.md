# 004 — Scaffold minimal Go server (`Chi` Hello World)

| Field | Value |
| :--- | :--- |
| **ID** | `004` |
| **Status** | `completed` |
| **Worker** | `antigravity` / `gemini-4-hydrogen` (`thinking_level: medium`) |
| **Role** | Go Backend Scaffolder |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `server/go.mod`, `server/go.sum`
  - `server/.env.example`
  - `server/cmd/server/main.go`
  - `server/internal/config/config.go`
  - `server/internal/http/router.go`
  - `server/internal/http/router_test.go`
  - `Taskfile.yml` (add/verify `server:check`, `server:test`, `server:build`, and `server:dev` tasks inside `dir: server`; keep `client:*` and `db:*` tasks intact)
  - `go.work` *(only if `go.work` exists at repo root: ensure `./server` is included in `use (...)` so `go` commands inside `server/` resolve cleanly)*
  - `.tickets/004-scaffold-go-api/changelog.md`
- **Must not touch:**
  - `app/**`
  - `shared/**`
  - `db/**`, `local/**`
  - `.gitignore`
  - `.routing/**`
  - `.claude/**`
  - `docs/**`, `README.md`, `PROJECT-REVIEW.md`
  - `.tickets/001-scaffold-wails3-app/**`, `.tickets/002-setup-portable-postgres-17/**`, `.tickets/003-full-stack-install-guide/**`, `.tickets/005-postgres-db-integration/**`
  - `.tickets/004-scaffold-go-api/task.md`, `.tickets/004-scaffold-go-api/spec.md`

## Inputs

Read these files before acting:
1. [`app/go.mod`](../../app/go.mod)
2. [`Taskfile.yml`](../../Taskfile.yml)
3. [`.routing/CONTEXT.md`](../../.routing/CONTEXT.md) (Ground Rules section)

## Grounding & Anti-Hallucination Constraints

- **Strictly Minimal (`Hello World` Baseline):**
  - Keep `server/` minimal so endpoints, WebSockets, services, interfaces, models, and DB layers can be built out one step at a time.
  - `server/go.mod` declares `module hrig/server` (`go 1.26.0`) and requires **only** `github.com/go-chi/chi/v5` and `github.com/go-chi/cors`.
  - Expose **only** the single `GET /` hello-world endpoint — do **not** add `/health`, `/ws`, `/api/*`, `gorilla/websocket`, `hrig/shared`, `pgx`, `store/`, `threshold/`, or any domain models/thresholds yet.
- **Workspace Safety (`go.work`):**
  - If `go.work` exists at the repository root, ensure `./server` is listed in `use (...)` so `go vet`, `go test`, and `go build` work inside `server/` without breaking `app/`.
- **Do NOT commit or push** any changes to Git.

## Task

1. Initialize `server/go.mod` (`module hrig/server`, `go 1.26.0`) requiring `github.com/go-chi/chi/v5` and `github.com/go-chi/cors`. If root `go.work` exists, ensure `./server` is listed in `use (...)`.
2. Create `server/.env.example` and `server/internal/config/config.go` loading environment variables with safe defaults:
   - `HRIG_API_ADDR` = `:8080`
   - `HRIG_ALLOWED_ORIGINS` = `http://127.0.0.1:9245,http://localhost:9245,wails://wails`
3. Create `server/internal/http/router.go` using `chi/v5` and `go-chi/cors`:
   - Apply `middleware.Recoverer` and `cors.Handler(...)` using `cfg.AllowedOrigins`.
   - Register **only** `GET /`, returning `200 OK` JSON (`Content-Type: application/json`): `{"message":"hello world"}`.
4. Create `server/cmd/server/main.go` wiring `config.Load`, `http.NewRouter`, `net/http.Server` with `ReadHeaderTimeout: 5 * time.Second`, and graceful shutdown on `SIGINT`/`SIGTERM`.
5. Create `server/internal/http/router_test.go` verifying `GET /` returns `200 OK` with `Content-Type: application/json` and `{"message":"hello world"}`.
6. Ensure root `Taskfile.yml` defines `server:check` (`go vet ./...`, `go test ./...`, `go build -o ../bin/hrig-server{{exeExt}} ./cmd/server`), `server:test` (`go test -v ./...`), `server:build`, and `server:dev` (`go run ./cmd/server`) inside `dir: server`, keeping `client:*` and `db:*` tasks intact.
7. Run `wails3 task server:check` and `wails3 task client:check` from the repository root, confirming both exit `0`.
8. Write `.tickets/004-scaffold-go-api/changelog.md`.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/004-scaffold-go-api/changelog.md` listing:
- Files created and modified
- Verification commands executed and their outputs
- Any blockers or assumptions avoided
