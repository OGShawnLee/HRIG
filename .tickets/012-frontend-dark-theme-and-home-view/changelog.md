# 012 — Changelog

## Summary

Configured the minimal dark HUD UnoCSS theme, updated the Wails window background and shell chrome to `#000000` high-contrast dark mode, created the shared real-time telemetry composable (`useTelemetry.ts`), and implemented the Figma-aligned `HomeView` dashboard grounded in PostgreSQL seeded astronauts and `db/002_seed.sql` metric codes.

## Files Created & Modified

### Modified
- `app/main.go` — Changed `BackgroundColour` from `application.NewRGB(255, 255, 255)` to `application.NewRGB(0, 0, 0)`.
- `app/frontend/index.html` — Updated `<body>` class from `text-side` to `text-main` on `bg-ground`.
- `app/frontend/uno.config.ts` — Configured minimal dark HUD palette (`ground: "#000000"`, `surface: "#0A0A0A"`, `main: "#FFFFFF"`, `side: "#D4D4D8"`, `dull: "#9CA3AF"`, `line: "#FFFFFF"`, `lineDim: "#262626"`), clinical status colors (`ok`, `warn`, `crit`, `stale`), and dark `hud-card`, `hud-pill`, `hud-value`, `button`, `button-main`, `button-side`, and `field` shortcuts while preserving `presetWind3()`, `presetWebFonts()`, and `transformerDirectives()`.
- `app/frontend/src/App.vue` — Applied `bg-ground text-main` to the root shell container.
- `app/frontend/src/components/WindowBar.vue` — Updated header bar to `bg-ground` with `border-lineDim`, preserved `OnClose` / `OnMinimize` / `OnMaximize`, and added a right-aligned live WebSocket status badge (`Live` / `Reconnecting`) driven by `useTelemetry()`.
- `app/frontend/src/views/HomeView.vue` — Implemented the Figma-aligned Home dashboard with top centered crew search bar (`Ctrl+K` focus, clear button, `Escape` handling), responsive crew cards grid, offline/empty/no-match states with `Retry`, bottom-left `MedicalManualButton` with inline accessible stub notice (`"Medical Manual view will be added in an upcoming UI step"`), and bottom-right `EnvironmentCard`.

### Created
- `app/frontend/src/composables/useTelemetry.ts` — Shared reactive telemetry store loading initial `GetAstronauts()`, `GetMetrics()`, `GetSnapshot()`, and `GetConnectionStatus()` from `@bindings/hrig/app/app.js`, subscribing to `"telemetry:envelope"` and `"telemetry:connection"` via `@wailsio/runtime` `Events.On`, and providing `getCrewReading`, `getEnvReading`, `resolveReadingStatus`, `formatMetricValue`, and `formatBloodPressure` (strictly using server-classified `reading.status` and DB unit decimals/symbols).
- `app/frontend/src/components/CrewCard.vue` — Figma-aligned crew card with offline-first SVG NASA-inspired portrait badge, split centered full name, and the 6 primary `metric_type = 1` physiological rows (`spo2`, `bp_systolic`/`bp_diastolic`, `radiation_rate`, `radiation_dose_cum`, `heart_rate`, `core_temp`) using `tabular-nums` and Lucide icon + accessible label redundancy for non-normal states.
- `app/frontend/src/components/EnvironmentCard.vue` — Figma-aligned `Environment` card displaying the 4 `metric_type = 2` cabin metrics (`cabin_o2`, `cabin_pressure`, `cabin_ppco2`, `cabin_temp`) using `tabular-nums` and status icon redundancy.
- `app/frontend/src/components/MedicalManualButton.vue` — Figma-aligned `Medical Manual` card button emitting `click`.
- `.tickets/012-frontend-dark-theme-and-home-view/changelog.md` — This changelog.

## Verification Outputs

### 1. `wails3 task client:check` (Exit Code: `0`)

Note: Temporary `app/app.exe` created by `go build .` was removed after verification.

```text
task: [client:check] wails3 task check
task: [bindings] wails3 generate bindings -clean=true -ts
Wails v3.0.0 › Generate Bindings
 INFO  Processed: 267 Packages, 1 Service, 13 Methods, 5 Enums, 13 Models, 2 Events in 4.3355144s.
 INFO  Output directory: C:\Users\OGSha\OneDrive\Escritorio\Developer\Web Development\HRIG\app\frontend\bindings
task: [check] bun run --cwd frontend type-check
$ vue-tsc --noEmit
task: [check] bun run --cwd frontend build
$ vite build --mode production
vite v8.3.8 building client environment for production...
transforming...
✓ 1784 modules transformed.
rendering chunks...
computing gzip size...
dist/index.html                   0.50 kB │ gzip:  0.34 kB
dist/assets/index-BF9nLAPA.css   14.27 kB │ gzip:  3.86 kB
dist/assets/index-CZjZsX4H.js   143.92 kB │ gzip: 51.70 kB

✓ built in 4.23s
task: [check] go build .
```

### 2. Live Backend Smoke Check (`postgres` + `bin/hrig-server.exe` + `bot` with `SENSOR_INTERVAL=1s`) (Exit Code: `0`)

Verified `GET /health`, `GET /api/astronauts`, `GET /api/metrics`, `GET /api/snapshot`, and WebSocket `ws://127.0.0.1:80/ws/realtime` covering all 11 HomeView metrics (`spo2`, `bp_systolic`, `bp_diastolic`, `radiation_rate`, `radiation_dose_cum`, `heart_rate`, `core_temp`, `cabin_o2`, `cabin_pressure`, `cabin_ppco2`, `cabin_temp`), followed by clean shutdown of `bot`, `hrig-server`, and `postgres`:

```text
=== 1. Starting PostgreSQL ===
task: [db:start] ./local/pgsql/bin/pg_ctl.exe -D ./local/pgsql/data -l ./local/pgsql/logfile -o "-h 127.0.0.1 -p 5432" start
esperando que el servidor se inicie.... listo
servidor iniciado
=== 2. Starting bin/hrig-server.exe ===
Health OK: store=postgres status=ok
=== 3. Building temp bot binary and starting bot (SENSOR_INTERVAL=1s) ===
=== 4. Checking REST endpoints ===
Astronauts count=3 names=Ana Torres, Liam Okafor, Sofia Petrov
Metrics count=23 missingHomeMetrics=
Snapshot readings=57 uniqueMetrics=23 missingFromSnapshot=
=== 5. Checking WS /ws/realtime ===
WS snapshot readings=57 alerts=0 homeMetricsCovered=11/11
WS live reading metric=spo2 astronautId=11111111-1111-4111-8111-111111111111 value=98.6739 status=normal
=== 6. Stopping bot, hrig-server, and postgres ===
task: [db:stop] ./local/pgsql/bin/pg_ctl.exe -D ./local/pgsql/data stop -m fast
esperando que el servidor se detenga.... listo
servidor detenido
=== 7. Checking remaining processes ===
Remaining target processes=0
```

## Notes & Assumptions Avoided
- No clinical threshold values were hardcoded in Vue or TypeScript; all status styling and icon redundancy (`AlertTriangle`, `AlertCircle`, `WifiOff`) derive strictly from server-classified `reading.status` and `reading.isStale`.
- `EnvironmentCard.vue` maps strictly to `db/002_seed.sql` `metric_type = 2` cabin codes (`cabin_o2`, `cabin_pressure`, `cabin_ppco2`, `cabin_temp`) rather than inventing an environmental radiation metric.
- `CrewCard.vue` uses an offline-first SVG portrait badge so no external network image requests are made.
