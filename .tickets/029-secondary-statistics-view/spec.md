# 029 Spec — Secondary Statistics View

## Acceptance Criteria

1. Route `/secondary-stats` (`name: "secondary-stats"`) resolves to `app/frontend/src/views/dashboard-secundario.vue` and appears in `ViewNavBar.vue` as `"Secondary Stats"`.
2. `useTelemetry.ts` exposes `getEnvSeries(metricCode)` for live environmental series (`cabin_ppco2`, `suit_pressure`, `suit_temp`, etc.).
3. `useTelemetryNavigation.ts` treats `"secondary-stats"` as a non-alert route in `NON_ALERT_ROUTE_NAMES` so alerts return properly to `/secondary-stats`.
4. `dashboard-secundario.vue` visually separates:
   - **Continuous Telemetry** (`resp_rate`, `hrv_rmssd`, `blood_glucose`, `cabin_ppco2`, `suit_pressure`, `radiation_dose_cum`) with live SVG waveforms.
   - **Periodic & Check-In Stats** (`sleep_hours`, `mood`, `hydration_intake`, `body_mass`, `rbc_count`/`wbc_count`, `ultrasound`) with discrete checkpoint bars / scale indicators / qualitative exam status and last-recorded timestamp (`MET` / UTC), clearly labeled as periodic (non-second-by-second).
5. Every card supports:
   - Crew switching and search/cadence filtering (`All`, `Continuous`, `Periodic`).
   - `abnormal` (solid yellow frame + yellow text), `critical` (blinking red frame + blinking red text), and `stale` (dashed `border-stale/75`, flat dashed baseline + `"Sensor Disconnected · No Signal"`, `"—"` value, `"NO SIGNAL"`).
   - `?` (`HelpCircle`) button navigating to `{ name: "search", query: { tab: "procedures", q: ... } }`.
6. Verification commands pass after all edits:
   - `bun run --cwd /home/sacred/bd/HRIG/app/frontend type-check` -> exit 0
   - `bun run --cwd /home/sacred/bd/HRIG/app/frontend build` -> exit 0
   - `rg -n "new WebSocket|fetch\(" /home/sacred/bd/HRIG/app/frontend/src/` -> 0 matches
   - `rg -n "secondary-stats" /home/sacred/bd/HRIG/app/frontend/src/` -> matches in `router/index.ts`, `ViewNavBar.vue`, `useTelemetryNavigation.ts`
