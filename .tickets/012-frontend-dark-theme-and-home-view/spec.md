# 012 — Verification Spec

Every check is binary (`PASS` / `FAIL`). All checks must pass before ticket `012` is accepted.

## 1. Dark Theme & Window Shell (`app/main.go`, `app/frontend/uno.config.ts`, `WindowBar.vue`)

- [x] `app/main.go` sets `BackgroundColour: application.NewRGB(0, 0, 0)` without altering `Options.Services` or window title/size.
- [x] `app/frontend/uno.config.ts` defines `#000000` `ground`, `#FFFFFF` `main`/`line`, high-contrast secondary/dull colors (`>= 6:1`), clinical status colors (`ok`, `warn`, `crit`, `stale`), and dark `hud-card` / `hud-pill` / `field` / `button*` shortcuts while keeping `presetWind3`, `presetWebFonts`, and `transformerDirectives`.
- [x] `app/frontend/src/components/WindowBar.vue` uses dark chrome (`bg-ground`, `border-lineDim`) and preserves `OnClose`, `OnMinimize`, `OnMaximize`, plus a compact live connection status indicator.

## 2. Telemetry Composable & Figma `HomeView` (`app/frontend/src/`)

- [x] `app/frontend/src/composables/useTelemetry.ts` imports bound RPCs from `@bindings/hrig/app/app.js` and subscribes to `"telemetry:envelope"` and `"telemetry:connection"` via `@wailsio/runtime` `Events.On`.
- [x] `HomeView.vue` + `CrewCard.vue` render active crew dynamically (3 seeded astronauts from DB) with top Search bar (`Ctrl+K` focus + filter) and the exact 6 DB-backed physiological rows: `Oxygenation:` (`spo2`), `Blood Pressure:` (`bp_systolic`/`bp_diastolic`), `Rad. Exposure:` (`radiation_rate` in `µSv/h`), `Acc. Radiation:` (`radiation_dose_cum` in `mSv`), `Heart Beat:` (`heart_rate` in `bpm`), `Temperature:` (`core_temp` in `°C`).
- [x] `MedicalManualButton.vue` renders bottom-left `Medical Manual` rounded card button, and `EnvironmentCard.vue` renders bottom-right `Environment` card with `O₂:` (`cabin_o2`), `Cabin Pressure:` (`cabin_pressure`), `Cabin ppCO₂:` (`cabin_ppco2`), and `Temperature:` (`cabin_temp`).
- [x] All numeric telemetry uses `tabular-nums`, status is never communicated by color alone (paired with Lucide icon + label/title for `abnormal`/`critical`/`stale`), and no hardcoded clinical threshold numbers appear in Vue/TS files.

## 3. Build, Type-Check & Live Smoke Verification

- [x] `wails3 task client:check` exits with `0` (`bindings` -> `bun run --cwd frontend type-check` -> `bun run --cwd frontend build` -> `go build .`).
- [x] Live check with `postgres`, `server`, and `bot` verifies `GET /api/astronauts`, `GET /api/metrics`, `GET /api/snapshot`, and `/ws/realtime` feed live readings without errors, and all background processes are stopped afterward.

## 4. Scope & Ground Rules

- [x] `git diff --name-only` and untracked files show changes only inside `May create or edit` paths listed in `task.md`.
- [x] `server/**`, `bot/**`, `shared/**`, `db/**`, `app/app.go`, `app/client.go`, `app/go.mod`, `app/frontend/package.json`, and `app/frontend/bindings/**` remain untouched by ticket `012`.
- [x] Code, `.tickets/`, and `changelog.md` are written in English; `.tickets/012-frontend-dark-theme-and-home-view/changelog.md` records all changed files and exact outputs.
