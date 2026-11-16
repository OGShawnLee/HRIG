# 029 — Changelog: Secondary Statistics View (Continuous Telemetry + Periodic Check-Ins)

## Files Created & Modified

- `app/frontend/src/composables/useTelemetry.ts`
  - Added and exported `getEnvSeries(metricCode: string): TelemetrySeriesPoint[]` returning `seriesByKey.value[readingKey(null, metricCode)] ?? []`, and exposed `getEnvSeries` in `useTelemetry()`.
- `app/frontend/src/composables/useTelemetryNavigation.ts`
  - Added `"secondary-stats"` to `NON_ALERT_ROUTE_NAMES` so alert auto-recovery and `Close Alert` return to `/secondary-stats`.
  - Added `openSecondaryStatsWithRouter`, exported `openSecondaryStats(astronaut?)`, and exposed `openSecondaryStats` from `useTelemetryNavigation()`.
- `app/frontend/src/router/index.ts`
  - Imported `DashboardSecundario` from `@/views/dashboard-secundario.vue` and added route `{ path: "/secondary-stats", name: "secondary-stats", component: DashboardSecundario }`.
- `app/frontend/src/components/ViewNavBar.vue`
  - Added `<RouterLink :to="{ name: 'secondary-stats' }">` labeled `"Secondary Stats"` styled identically to `"Primary Vitals"`.
- `app/frontend/src/views/dashboard-secundario.vue`
  - Implemented the dark-HUD Secondary Statistics dashboard matching `dashboard-principal.vue`'s UnoCSS tokens, `rem` sizing, solid yellow `abnormal` / blinking red `critical` alert frames & text animations, `StatusStale` dashed border + flatline `"Sensor Disconnected · No Signal"` / `"—"` / `"NO SIGNAL"` state, and `?` (`HelpCircle`) Medical Manual navigation buttons.
  - Added top controls bar with `Home` button, `"SECONDARY STATS"` title, crew member selector (`setSelectedAstronaut` + route query sync), metric search input, and cadence filter (`All`, `Continuous`, `Periodic`).
  - **Section 1 (`CONTINUOUS MONITORING · LIVE STREAM`)**: Live SVG waveforms, numeric readouts, and paired sublines for `resp_rate`, `hrv_rmssd`, `blood_glucose`, `cabin_ppco2` (paired with `cabin_o2`), `suit_pressure` (paired with `suit_temp`), and `radiation_dose_cum` (paired with `radiation_rate`).
  - **Section 2 (`PERIODIC & CHECK-IN STATS · NON-SECOND CADENCE`)**: Discrete checkpoint bar charts, cadence pills (`DAILY LOG`, `DAILY CHECK-IN`, `LAB ASSAY`, `SCHEDULED IMAGING`), target ranges, and last-recorded `MET` / UTC timestamps for `sleep_hours`, `mood` (1–5 scale blocks), `hydration_intake`, `body_mass`, `rbc_count` + `wbc_count` (dual hematology summary), and `ultrasound` (qualitative finding card).
  - Loaded crew and environmental trends (`loadMetricTrends(cleanId, metricCode, 40)` and `loadMetricTrends("", metricCode, 40)`) on mount and astronaut change.
- `.tickets/029-secondary-statistics-view/task.md`
  - Updated status from `in_progress` to `completed`.
- `.tickets/029-secondary-statistics-view/changelog.md`
  - Created this changelog.

## Verification Results

- `bun run --cwd /home/sacred/bd/HRIG/app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`).
- `bun run --cwd /home/sacred/bd/HRIG/app/frontend build` (`vite build --mode production`): **PASS** (exit code `0`; 2432 modules transformed, built in 3.43s: `dist/index.html` 0.50 kB, `dist/assets/index-B2RbnIvT.css` 24.29 kB, `dist/assets/index-Cdt38v09.js` 348.03 kB).
- `rg -n "new WebSocket|fetch\(" /home/sacred/bd/HRIG/app/frontend/src/`: **PASS** (0 matches).
- `rg -n "secondary-stats" /home/sacred/bd/HRIG/app/frontend/src/`: **PASS** (matches in `useTelemetryNavigation.ts`, `router/index.ts`, `ViewNavBar.vue`, and `dashboard-secundario.vue`).
- Scope check: `dashboard-principal.vue` untouched; no git commit or push performed.
