# 028 — Show Disconnected / No-Signal State Inside Vital Charts

| Field | Value |
| :--- | :--- |
| **ID** | `028-disconnected-sensor-chart-indicator` |
| **Status** | `completed` |
| **Scope** | `app/frontend/src/views/dashboard-principal.vue` |

## Context

When a simulated apparatus is disconnected in the bot (e.g., `baumanometro_01`, `biomonitor_01`, `dosimeter_01`, `cabin_sensor_01`), the backend marks its metrics `StatusStale` (`isStale = true`) after `HRIG_STALE_AFTER`. In `dashboard-principal.vue`, stale cards still render animated-looking waveforms without telling the user that no data is arriving or that the sensor is disconnected.

## Grounding Rules

- Modify only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/028-disconnected-sensor-chart-indicator/`.
- Keep all Ticket `018` (alert frame/text), `021`/`022` (help button), `024`/`025` (`rem` scaling), and `026` (silhouette alert states) behavior intact.
- Keep zero `new WebSocket` or `fetch(` calls in `app/frontend/src/`.

## Steps

1. Update `app/frontend/src/views/dashboard-principal.vue`:
   - Import `WifiOff` from `"lucide-vue-next"`.
   - Extend `VitalRow` with `isDisconnected: boolean` (`true` when the card's primary reading status is `ReadingStatus.StatusStale`) and `isSublineDisconnected: boolean` (`true` when the paired crew/environmental subline metric is `ReadingStatus.StatusStale`).
   - Update `formatEnvLine` / `formatCrewSubLine` so a `StatusStale` subline displays `"NO SIGNAL"` cleanly.
   - Update `cardSecondaryTextClass` to return `"text-stale"` for `ReadingStatus.StatusStale`, and update `cardFrameClass` to return `"border-2 border-dashed border-stale/75 opacity-85"` for `ReadingStatus.StatusStale`.
   - In the card template (`<article v-for="row in vitalRows">`):
     - When `row.isDisconnected` is `true`:
       - Set header `rangeLabel` text to `"NO SIGNAL"`.
       - Inside the SVG chart container (`relative`), replace the active waveform with a flat dashed disconnected baseline (`<line x1="0" y1="32" x2="300" y2="32" stroke-dasharray="6 5" stroke-width="2" class="stroke-stale/60" />`) and an overlay status badge (`WifiOff` icon + `"Sensor Disconnected · No Signal"` in `text-stale`, `border border-dashed border-stale/75 bg-ground/90 rounded-md`).
       - In the right numeric readout column, render `"—"` for `row.value` (instead of frozen stale numbers) and show `"NO SIGNAL"` / `WifiOff` in the subline so it is unmistakable that no new telemetry is arriving.
     - When `!row.isDisconnected && row.isSublineDisconnected`:
       - Display a `WifiOff` icon alongside the stale subline (`CABIN: NO SIGNAL` / `HRV NO SIGNAL`).
2. Run `bun run --cwd app/frontend type-check`, `bun run --cwd app/frontend build`, and `rg` checks, write `changelog.md`, and report back.
