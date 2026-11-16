# 028 — Changelog: Show Disconnected / No-Signal State Inside Vital Charts

## Files Modified

- `app/frontend/src/views/dashboard-principal.vue`
  - Imported `WifiOff` from `"lucide-vue-next"`.
  - Updated `cardSecondaryTextClass` to return `"text-stale"` for `ReadingStatus.StatusStale` and `cardFrameClass` to return `"border-2 border-dashed border-stale/75 opacity-85"` for `ReadingStatus.StatusStale`.
  - Updated `formatEnvLine` and `formatCrewSubLine` to output `"NO SIGNAL"` (e.g., `"CABIN: NO SIGNAL"`, `"HRV NO SIGNAL"`) when a subline reading status is `ReadingStatus.StatusStale`.
  - Added `isCrewMetricStale` and `isEnvMetricStale` helpers and extended `VitalRow` with `isDisconnected: boolean` and `isSublineDisconnected: boolean` across all five vital cards (`cardiac`, `oxygenation`, `pressure`, `radiation`, `temperature`).
  - When `row.isDisconnected` is `true`, sets `value` to `"—"` and `rangeLabel` to `"NO SIGNAL"`, replaces the active SVG waveform inside the `relative` chart container with a flat dashed baseline (`<line x1="0" y1="32" x2="300" y2="32" stroke-dasharray="6 5" stroke-width="2" class="stroke-stale/60" />`) plus a centered badge (`WifiOff` + `"Sensor Disconnected · No Signal"` in `text-stale`, `border border-dashed border-stale/75 bg-ground/90 rounded-md`), and displays `WifiOff` + `"NO SIGNAL"` in the subline.
  - When `!row.isDisconnected && row.isSublineDisconnected`, keeps the primary live waveform, numeric readout, and range intact while displaying `WifiOff` alongside the stale subline in `text-stale`.
- `.tickets/028-disconnected-sensor-chart-indicator/task.md`
  - Updated status from `in_progress` to `completed`.
- `.tickets/028-disconnected-sensor-chart-indicator/changelog.md`
  - Created this changelog.

## Verification Results

- `rg -n "Sensor Disconnected|isDisconnected|WifiOff" app/frontend/src/views/dashboard-principal.vue`: **PASS** (19 matches across import, `VitalRow`, `vitalRows`, chart overlay, and subline readout).
- `rg -n "new WebSocket|fetch\(" app/frontend/src/`: **PASS** (0 matches).
- `bun run --cwd app/frontend type-check` (`vue-tsc --noEmit`): **PASS** (exit code `0`).
- `bun run --cwd app/frontend build` (`vite build --mode production`): **PASS** (exit code `0`; 2429 modules transformed, built in 2.26s: `dist/index.html` 0.50 kB, `dist/assets/index-Bvg6CP1t.css` 22.31 kB, `dist/assets/index-CQkqDyp1.js` 316.44 kB).
- Scope check (`git status --short`): Only `app/frontend/src/views/dashboard-principal.vue` and `.tickets/028-disconnected-sensor-chart-indicator/` modified/untracked; no git commit or push performed.
