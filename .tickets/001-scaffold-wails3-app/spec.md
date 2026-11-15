# 001 — Verification Spec

All checks must pass with binary Yes/No evidence.

## Directory & Template Hygiene
- [x] `app/go.mod` exists directly at `app/go.mod` (no `app/app/` directory exists) and contains `module hrig/app` and `github.com/wailsapp/wails/v3 v3.0.0`.
- [x] `go.work` does not exist at repo root.
- [x] `app/greetservice.go` does not exist.
- [x] `app/build/ios/`, `app/build/android/`, and `app/build/docker/` do not exist.
- [x] `app/build/config.yml` exists, contains `productName: "HRIG"` and `productIdentifier: "com.hrig.app"`.
- [x] Template frontend artifacts (`app/frontend/public/style.css`, `app/frontend/public/bg-desktop.jpg`, `app/frontend/public/bg-mobile.jpg`, `app/frontend/public/vue.svg`, `app/frontend/public/Inter-Medium.ttf`, `app/frontend/Inter Font License.txt`, `app/frontend/README.md`) do not exist; `app/frontend/public/wails.png` exists.
- [x] `.gitignore` exists at repo root and contains `build/` followed by `!app/build/` and `node_modules/`.

## Go & Taskfile Configuration
- [x] `app/app.go` defines `type App struct`, `NewApp()`, `SetContext(ctx context.Context)`, and `Ping() string` returning `"pong"`.
- [x] `app/app_window.go` defines `OnMinimize()`, `OnMaximize()`, and `OnClose()` on `*App`.
- [x] `app/main.go` embeds `all:frontend/dist`, sets `Name: "HRIG"`, `Title: "HRIG — Resource Integration Gear"`, `Frameless: true`, and registers `appService`.
- [x] `app/Taskfile.yml` exists with `APP_NAME: "hrig"` and tasks `build`, `package`, `run`, `start`, `dev`, `bindings`, `check`.
- [x] Root `Taskfile.yml` exists and delegates `client:check`, `client:build`, `client:start`, `client:dev`, `client:bindings` to `dir: app` using `wails3 task`.

## Frontend Configuration & Source
- [x] `app/frontend/.npmrc` exists and contains `minimum-release-age=10080`.
- [x] `app/frontend/package.json` exists with `"name": "hrig-app-frontend"`, `"@wailsio/runtime": "3.0.0"`, and `"typescript": "^5.7.3"`.
- [x] `app/frontend/tsconfig.json`, `app/frontend/vite.config.ts`, `app/frontend/uno.config.ts`, and `app/frontend/index.html` exist with `<title>HRIG</title>` and `@bindings/*` alias pointing to `bindings/*`.
- [x] `app/frontend/src/main.ts` imports `"virtual:uno.css"` before `"@unocss/reset/tailwind-compat.css"`.
- [x] `app/frontend/src/router/index.ts` uses `createWebHashHistory()` and routes `/` to `HomeView`.
- [x] `app/frontend/src/components/WindowBar.vue` imports `OnClose`, `OnMaximize`, `OnMinimize` from `@bindings/hrig/app/app.js`.
- [x] `app/frontend/src/views/HomeView.vue` imports `Ping` from `@bindings/hrig/app/app.js` and renders the **Test Go Bridge** button.

## Build, Bindings & Type-Check Pipeline
- [x] `app/frontend/bindings/hrig/app/app.ts` exists and exports `Ping`, `OnMinimize`, `OnMaximize`, and `OnClose`.
- [x] `bun run --cwd app/frontend type-check` exits `0`.
- [x] `bun run --cwd app/frontend build` creates `app/frontend/dist/index.html` and exits `0`.
- [x] `wails3 task check` inside `app/` exits `0`.
- [x] `wails3 task client:check` at repo root exits `0`.

## Scope & Ground Rules
- [x] `git status --short` shows zero modifications in `.routing/`, `.claude/`, `docs/`, `README.md`, `PROJECT-REVIEW.md`, or `wails3_barebones_scaffold_guide.md`.
- [x] All code, config, and `.tickets/001-scaffold-wails3-app/changelog.md` are written in English.
- [x] No unapproved git commits or pushes were executed.
- [x] `.tickets/001-scaffold-wails3-app/changelog.md` exists and records files touched and command outputs.
