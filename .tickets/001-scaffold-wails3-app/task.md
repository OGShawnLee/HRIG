# 001 — Scaffold bare-bones Wails v3 desktop application

| Field | Value |
| :--- | :--- |
| **ID** | `001` |
| **Status** | `completed` |
| **Worker** | `antigravity` / `gemini-4-hydrogen` (`thinking_level: medium`) |
| **Role** | Desktop App Scaffolder |
| **Spec** | [spec.md](spec.md) |

## Scope

- **May create or edit:**
  - `.gitignore`
  - `Taskfile.yml`
  - `app/go.mod`
  - `app/go.sum`
  - `app/main.go`
  - `app/app.go`
  - `app/app_window.go`
  - `app/Taskfile.yml`
  - `app/build/**`
  - `app/frontend/.npmrc`
  - `app/frontend/package.json`
  - `app/frontend/bun.lock`
  - `app/frontend/tsconfig.json`
  - `app/frontend/vite.config.ts`
  - `app/frontend/uno.config.ts`
  - `app/frontend/index.html`
  - `app/frontend/public/wails.png`
  - `app/frontend/bindings/**`
  - `app/frontend/dist/**`
  - `app/frontend/src/vite-env.d.ts`
  - `app/frontend/src/main.ts`
  - `app/frontend/src/App.vue`
  - `app/frontend/src/router/index.ts`
  - `app/frontend/src/components/WindowBar.vue`
  - `app/frontend/src/views/HomeView.vue`
  - `.tickets/001-scaffold-wails3-app/changelog.md`
- **Must not touch:**
  - `.routing/**` (including `.routing/INPUT.md`, `.routing/CONTEXT.md`, `.routing/DOMAIN.md`, `.routing/SPECIFICATION.md`, `.routing/DESIGN.md`, `.routing/PROGRESS.md`)
  - `.claude/**`
  - `docs/**`
  - `README.md`, `PROJECT-REVIEW.md`, `wails3_barebones_scaffold_guide.md`
  - `.tickets/001-scaffold-wails3-app/task.md`
  - `.tickets/001-scaffold-wails3-app/spec.md`

## Inputs

Read these files before acting:
1. [`wails3_barebones_scaffold_guide.md`](../../wails3_barebones_scaffold_guide.md) (Sections 1–6)
2. [`.routing/CONTEXT.md`](../../.routing/CONTEXT.md) (Ground Rules section)

## Grounding & Anti-Hallucination Constraints

- Use these exact project replacements consistently across all generated files:
  - `<project>` = `hrig`
  - `<ProjectName>` = `HRIG`
  - `<Project description>` = `Resource Integration Gear — Astronaut Health Monitoring Software`
  - `<com.example.app>` = `com.hrig.app`
  - `<app_binary>` = `hrig`
  - `<Your Company>` = `HRIG Team`
  - Window `Title` in `app/main.go` = `HRIG — Resource Integration Gear`
- Use exact pinned versions and dependencies from `wails3_barebones_scaffold_guide.md` (`@wailsio/runtime` `3.0.0`, `typescript ^5.7.3`, `vue ^3.5.44`, `vue-router ^5.4.0`, `unocss ^66.0.0`, `@unocss/reset ^66.12.0`, `vite ^8.3.5`, `vue-tsc ^2.2.8`, `@vitejs/plugin-vue ^6.0.10`, `@vueuse/core ^15.0.0`, `lucide-vue-next ^1.0.0`, `reka-ui ^2.1.0`, `@types/node ^22.13.4`).
- Do **not** create `go.work`; `app/go.mod` must remain a standalone module (`module hrig/app`).
- Do **not** pass `-d app` to `wails3 init` (it would create a nested `app/app/` folder).
- Do **not** run `go build ./...` inside `app/` (`app/build/*` has non-main packaging stubs); always run `go build .` inside `app/`.
- Do **not** run standalone `task` directly; always invoke via `wails3 task <name>`.
- In `app/frontend/src/main.ts`, import `"virtual:uno.css"` **before** `"@unocss/reset/tailwind-compat.css"`.
- Do **not** add HRIG domain views, health metrics, anatomical maps, or dark-mode theme overrides in this ticket; keep this scaffold strictly bare-bones as specified in `wails3_barebones_scaffold_guide.md`.
- Do **not** commit or push any changes to git.
- If any CLI command (`wails3`, `bun`, `go`) fails or a dependency fails to resolve, **stop**, do not invent workarounds, and record the exact command and stderr in `.tickets/001-scaffold-wails3-app/changelog.md`.

## Task

1. Run `wails3 init -n app -t vue -mod hrig/app -skipgomodtidy -q` from the repository root (`HRIG/`). Confirm `app/go.mod` and `app/frontend/` exist directly under `HRIG/app/` (not `HRIG/app/app/`).
2. Prune dead template files inside `app/`:
   - Delete `app/greetservice.go`.
   - Delete `app/build/ios/`, `app/build/android/`, and `app/build/docker/`.
   - Delete all files and subdirectories inside `app/frontend/src/`.
   - Delete `app/frontend/public/style.css`, `app/frontend/public/bg-desktop.jpg`, `app/frontend/public/bg-mobile.jpg`, `app/frontend/public/vue.svg`, `app/frontend/public/Inter-Medium.ttf`, `app/frontend/Inter Font License.txt`, and `app/frontend/README.md` (if present). Keep `app/frontend/public/wails.png`.
3. Create root `.gitignore` containing:
   ```gitignore
   node_modules/
   bin/
   dist/
   build/
   !app/build/
   *.exe
   ```
4. Update `app/go.mod` to declare `module hrig/app`, `go 1.24.0`, and `require ( github.com/wailsapp/wails/v3 v3.0.0 )`.
5. Write `app/app.go` with package `main`, `App` struct (`ctx context.Context`), `NewApp() *App`, `SetContext(ctx context.Context)`, and `Ping() string` returning `"pong"`.
6. Write `app/app_window.go` with package `main` implementing `OnMinimize()`, `OnMaximize()`, and `OnClose()` on `*App` using `github.com/wailsapp/wails/v3/pkg/application`.
7. Write `app/main.go` matching Section 4.4 of `wails3_barebones_scaffold_guide.md` with `Name: "HRIG"`, `Description: "Resource Integration Gear — Astronaut Health Monitoring Software"`, `Title: "HRIG — Resource Integration Gear"`, `Frameless: true`, `Width: 1280`, `Height: 820`, `MinWidth: 390`, `MinHeight: 680`, and `//go:embed all:frontend/dist`.
8. Create `app/frontend/.npmrc` with `minimum-release-age=10080`.
9. Replace `app/frontend/package.json` with the exact manifest from Section 5.2 of `wails3_barebones_scaffold_guide.md` using `"name": "hrig-app-frontend"`.
10. Replace `app/frontend/tsconfig.json`, `app/frontend/vite.config.ts`, `app/frontend/uno.config.ts`, and `app/frontend/index.html` using Sections 5.3–5.6 of `wails3_barebones_scaffold_guide.md` (with `<title>HRIG</title>` in `index.html`).
11. Create the minimal Vue 3 source files inside `app/frontend/src/` matching Sections 6.1–6.6 of `wails3_barebones_scaffold_guide.md`:
    - `app/frontend/src/vite-env.d.ts`
    - `app/frontend/src/main.ts` (`"virtual:uno.css"` before `"@unocss/reset/tailwind-compat.css"`)
    - `app/frontend/src/router/index.ts` (`createWebHashHistory()`)
    - `app/frontend/src/components/WindowBar.vue` (importing `OnClose`, `OnMaximize`, `OnMinimize` from `@bindings/hrig/app/app.js` and displaying `HRIG`)
    - `app/frontend/src/App.vue`
    - `app/frontend/src/views/HomeView.vue` (importing `Ping` from `@bindings/hrig/app/app.js`, displaying `<h1>HRIG</h1>` and the **Test Go Bridge** button)
12. Replace `app/Taskfile.yml` using Section 7.1 of `wails3_barebones_scaffold_guide.md` with `APP_NAME: "hrig"`.
13. Create root `Taskfile.yml` delegating `client:check`, `client:build`, `client:start`, `client:dev`, and `client:bindings` to `dir: app` via `wails3 task <name>`.
14. Update the `info:` block in `app/build/config.yml` with `companyName: "HRIG Team"`, `productName: "HRIG"`, `productIdentifier: "com.hrig.app"`, `description: "Resource Integration Gear — Astronaut Health Monitoring Software"`, `copyright: "(c) 2026, HRIG Team"`, and `version: "0.1.0"`.
15. Run `bun install` inside `app/frontend/`.
16. Run `go mod tidy` inside `app/`.
17. Run `wails3 task check` inside `app/` (executes `bindings` -> `bun run --cwd frontend type-check` -> `bun run --cwd frontend build` -> `go build .`) and verify exit code `0`. Remove any leftover `app/app.exe` generated by `go build .` if `bin/` wasn't used, and verify `wails3 task client:check` from repo root also exits `0`.
18. Write `.tickets/001-scaffold-wails3-app/changelog.md` documenting all created/deleted/modified files, exact commands run with exit codes, and any blockers encountered.

## Verification

Run every check listed in [spec.md](spec.md) and confirm all pass before reporting completion.

## Change Log

Write `.tickets/001-scaffold-wails3-app/changelog.md` listing:
- Files created, modified, and deleted
- Verification commands executed and their results
- Any blockers or assumptions avoided
