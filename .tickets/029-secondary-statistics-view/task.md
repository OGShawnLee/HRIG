# 029 — Secondary Statistics View (Continuous Telemetry + Periodic Check-Ins)

| Field | Value |
| :--- | :--- |
| **ID** | `029-secondary-statistics-view` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/views/dashboard-secundario.vue`, `app/frontend/src/router/index.ts`, `app/frontend/src/components/ViewNavBar.vue`, `app/frontend/src/composables/useTelemetry.ts`, `app/frontend/src/composables/useTelemetryNavigation.ts` |

## Context

`dashboard-principal.vue` displays only the 5 core primary vital sign waveforms (`heart_rate`, `spo2`, `bp_systolic`/`bp_diastolic`, `radiation_rate`, `core_temp`). The remaining secondary/subsystem telemetry (`resp_rate`, `hrv_rmssd`, `cabin_ppco2`, `suit_pressure`/`suit_temp`, `radiation_dose_cum`) and non-second-by-second periodic check-ins (`sleep_hours`, `mood`, `hydration_intake`, `body_mass`, `blood_glucose`, `rbc_count`, `wbc_count`, `ultrasound`) have no dedicated dashboard view.

The user requested a dedicated Secondary Statistics view that clearly distinguishes between metrics that require continuous monitoring and periodic metrics that do not need per-second updates (such as sleep hours, mood, hydration, lab panels, and scheduled ultrasound).

## Grounding Rules

- Do NOT modify `app/frontend/src/views/dashboard-principal.vue`, `bot/`, `server/`, or `db/`.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`. Use existing `useTelemetry()` bindings/composables (`loadMetricTrends`, `getCrewReading`, `getEnvReading`, `getCrewSeries`, `getEnvSeries`, `formatMetricValue`, `resolveReadingStatus`).
- Preserve existing alert behavior: solid non-blinking yellow for `abnormal`, blinking red for `critical`, dashed `StatusStale` disconnected / `NO SIGNAL` state, `?` (`HelpCircle`) Medical Manual navigation, and `rem`-based sizing compatible with `App.vue` viewport scaling.

## Steps

1. Extend `app/frontend/src/composables/useTelemetry.ts`:
   - Export `getEnvSeries(metricCode: string): TelemetrySeriesPoint[]` (`return seriesByKey.value[readingKey(null, metricCode)] ?? []`) and include it in `useTelemetry()`.
2. Extend `app/frontend/src/composables/useTelemetryNavigation.ts`:
   - Add `"secondary-stats"` to `NON_ALERT_ROUTE_NAMES` so alert auto-recovery and `Close Alert` return cleanly to `/secondary-stats`.
   - Add and export `openSecondaryStats(astronaut?)` alongside `openPrimaryVitals`.
3. Update `app/frontend/src/router/index.ts`:
   - Import `DashboardSecundario` from `@/views/dashboard-secundario.vue` and add route `{ path: "/secondary-stats", name: "secondary-stats", component: DashboardSecundario }`.
4. Update `app/frontend/src/components/ViewNavBar.vue`:
   - Add `<RouterLink :to="{ name: 'secondary-stats' }">` labeled `"Secondary Stats"` styled identically to `"Primary Vitals"`.
5. Create `app/frontend/src/views/dashboard-secundario.vue`:
   - Render a clean dark-HUD Secondary Statistics dashboard with:
     - Top controls bar: `Home` button, `"SECONDARY STATS"` title, crew member selector (switching active astronaut smoothly), metric search input, and cadence filter (`All`, `Continuous`, `Periodic`).
     - **Section 1 — Continuous Telemetry (`CONTINUOUS MONITORING · LIVE STREAM`)**: live trend SVG waveforms + numeric readouts + sublines + `StatusStale` disconnected flatline overlay (`"Sensor Disconnected · No Signal"`) + `?` Medical Manual button for `resp_rate`, `hrv_rmssd`, `blood_glucose`, `cabin_ppco2` (paired with `cabin_o2`), `suit_pressure` (paired with `suit_temp`), and `radiation_dose_cum` (paired with `radiation_rate`).
     - **Section 2 — Periodic & Check-In Stats (`PERIODIC & CHECK-IN STATS · NON-SECOND CADENCE`)**: discrete checkpoint bar/step SVG visualizations (instead of live pulse waves), cadence pills (`DAILY LOG`, `DAILY CHECK-IN`, `LAB ASSAY`, `SCHEDULED IMAGING`), last recorded `MET` / timestamp, target range, scale visualization for `mood` (`1–5`), dual hematology summary for `rbc_count` + `wbc_count`, and qualitative finding card for `ultrasound`, plus `sleep_hours`, `hydration_intake`, and `body_mass` — all with alert status contours (`warn` solid yellow / `crit` blinking red), `StatusStale` disconnected / `NO SIGNAL` handling, and `?` Medical Manual buttons.
   - Load both crew trends (`loadMetricTrends(cleanId, metricCode, 40)`) and environmental trends (`loadMetricTrends("", metricCode, 40)`) on mount and astronaut change.
   - Use pure `rem` / standard Tailwind classes so `App.vue` viewport scaling never clips readouts.
6. Run `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks, mark `task.md` `completed`, and write `changelog.md`.
