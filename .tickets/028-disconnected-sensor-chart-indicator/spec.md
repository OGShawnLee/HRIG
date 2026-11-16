# Spec — 028 Disconnected Sensor Chart Indicator

## Functional Requirements

1. When a vital card's primary telemetry enters `ReadingStatus.StatusStale` (apparatus disconnected / no data arriving), the card displays:
   - A dashed stale border (`border-dashed border-stale/75`),
   - A flat dashed baseline across the SVG chart (`M 0 32 H 300` / `<line ... stroke-dasharray="6 5" />`) instead of an active waveform,
   - A clear centered chart overlay with `WifiOff` and `"Sensor Disconnected · No Signal"`,
   - `"—"` in the primary numeric readout and `"NO SIGNAL"` on range/subline indicators.
2. When a paired subline sensor (e.g. `cabin_pressure`, `cabin_o2`, `cabin_temp`, `hrv_rmssd`, `radiation_dose_cum`) is `StatusStale` while the primary sensor remains live, the subline reflects `"NO SIGNAL"` with a `WifiOff` indicator without hiding the primary live graph.
3. Reconnecting the apparatus immediately restores normal live waveforms, numeric readouts, and ranges.

## Acceptance Criteria

- `rg -n "Sensor Disconnected|isDisconnected|WifiOff" app/frontend/src/views/dashboard-principal.vue` confirms the disconnected chart indicator and state wiring.
- `rg -n "new WebSocket|fetch\(" app/frontend/src/` returns zero matches.
- `bun run --cwd app/frontend type-check` exits `0`.
- `bun run --cwd app/frontend build` exits `0`.
