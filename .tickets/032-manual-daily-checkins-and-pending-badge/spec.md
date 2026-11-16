# 032 Spec — Astronaut Manual Daily Check-Ins & Pending Badge

## Acceptance Criteria

1. `bot/main.go` no longer includes `sleep_hours`, `mood`, `hydration_intake`, or `body_mass` in `biomonitor_01` `Metrics`.
2. `db/002_seed.sql` seeds device `manual_checkin` (`kind = 'manual'`), and `server/internal/realtime/hub.go` uses a `24h` staleness threshold for daily check-in readings (`sleep_hours`, `mood`, `hydration_intake`, `body_mass`, or `device_code == "manual_checkin"` / `source == "manual"`).
3. `app/client.go`, `app/app.go`, and `app/frontend/bindings/hrig/app/app.ts` expose `SubmitCheckin(req)`.
4. `useTelemetry.ts` exposes `isDailyMetricLoggedToday`, `getPendingDailyMetrics`, `hasPendingDailyCheckin`, and `submitCrewDailyMetric`.
5. `CrewCard.vue` renders a small yellow circle (`bg-warn`) next to the astronaut's name when `hasPendingDailyCheckin(astronaut.id)` is `true`.
6. `dashboard-secundario.vue` (`OTHER STATS`) shows `"PENDING TODAY"` with a yellow dot for unlogged daily metrics (`sleep_hours`, `mood`, `hydration_intake`, `body_mass`) and provides inline inputs/buttons to log each metric directly.
7. Verification commands pass after final edits:
   - `go -C /home/sacred/bd/HRIG/bot test ./...` -> exit 0
   - `go -C /home/sacred/bd/HRIG/server test ./internal/realtime/... ./internal/http/...` -> exit 0
   - `go -C /home/sacred/bd/HRIG/app test ./...` -> exit 0
   - `bun run --cwd /home/sacred/bd/HRIG/app/frontend type-check` -> exit 0
   - `bun run --cwd /home/sacred/bd/HRIG/app/frontend build` -> exit 0
   - `rg -n "new WebSocket|fetch\(" /home/sacred/bd/HRIG/app/frontend/src/` -> 0 matches
