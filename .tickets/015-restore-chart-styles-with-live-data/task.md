# 015 — Restore Iconic Chart Waveforms & Colors Driven by Real-Time Telemetry

| Field | Value |
| :--- | :--- |
| **ID** | `015-restore-chart-styles-with-live-data` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/views/dashboard-principal.vue` |

## Context

Ticket `014` connected `dashboard-principal.vue` to live `GetTrends` + WebSocket series data, but replaced the original clinical chart waveforms and colors (`CARDIAC RHYTHM` green ECG, `OXYGENATION` green plethysmograph, `PRESSURE` white arterial wave, `RADIATION` yellow sawtooth with translucent yellow fill, `TEMPERATURE` white thermal trace) with uniform green polylines.

## Grounding Rules

- Modify only `app/frontend/src/views/dashboard-principal.vue` and ticket `015` files.
- Keep `GetTrends` (`loadMetricTrends`) and live `getCrewSeries` data binding active; zero hardcoded static-only charts or mock timers.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. In `app/frontend/src/views/dashboard-principal.vue`:
   - Preserve the original color palette per vital card on nominal readings using `uno.config.ts` classes:
     - `CARDIAC RHYTHM`: `stroke-ok` (green)
     - `OXYGENATION`: `stroke-ok` (green)
     - `PRESSURE`: `stroke-main` (white)
     - `RADIATION`: `stroke-warn` + `fill-warn/20` (yellow sawtooth + translucent fill)
     - `TEMPERATURE`: `stroke-main` (white)
     - Shift stroke/fill to `stroke-warn`/`fill-warn/20` on `abnormal`, `stroke-crit`/`fill-crit/20` on `critical`, and `stroke-stale` on `stale` (while still showing clear nominal colors when normal).
   - Implement real-time data-driven waveform path generators (`svg <path>`) for each of the 5 charts fed by `getCrewSeries` (`GetTrends` + live WebSocket samples):
     - **Cardiac (`heart_rate`)**: ECG P-QRS-T waveform whose beat frequency/spacing tracks live BPM, R-wave heights modulate from recent `hrSeries` samples, and phase advances on new samples.
     - **Oxygenation (`spo2`)**: Smooth plethysmograph wave whose amplitude/level follows `spo2Series` samples and advances on new samples.
     - **Pressure (`bp_systolic`, `bp_diastolic`)**: White arterial pulse wave whose peak-to-trough height scales with live systolic/diastolic values and advances on new samples.
     - **Radiation (`radiation_rate`)**: Yellow sawtooth pulse train + closed area `fillPath` whose tooth heights scale with live `radSeries` values and advance on new samples.
     - **Temperature (`core_temp`)**: White thermal trace whose vertical height/contour responds smoothly to live `tempSeries` values (sitting near the baseline `y≈54–56` at nominal `36.5–37.2 °C`, rising when elevated).
   - Keep `rangeLabel` (`min–max` in `tabular-nums`) in the header so real incoming sample bounds remain visible.
   - Style the enclosing white card frame (`<article class="hud-card ...">`) so that when `row.status` is `abnormal` it changes to yellow (`border-warn`) and blinks smoothly, and when `row.status` is `critical` it changes to red (`border-crit`) and blinks faster (keeping inner text/values solid and legible via border/box-shadow keyframes).
2. Run `bun run --cwd app/frontend type-check` and `bun run --cwd app/frontend build` (exit `0`), verify untouched files, write `changelog.md`, and report back.
