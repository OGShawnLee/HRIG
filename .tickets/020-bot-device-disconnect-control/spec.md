# Spec — 020 Bot Device Disconnection Control

## Functional Requirements

1. Disconnecting a valid device (`biomonitor_01`, `baumanometro_01`, `dosimeter_01`, `cabin_sensor_01`, `suit_sensor_01`, `lab_assay_01`, `ultrasound_01`) causes `sendBatch` to stop posting readings for that device until reconnected.
2. HTTP control API on `BOT_CONTROL_ADDR` (`127.0.0.1:8090`):
   - `GET /` returns `200 OK` (`text/html`) with an interactive panel controlling both anomaly overrides and device connect/disconnect state.
   - `GET /status` returns top-level `mode`, `metric_code`, `astronaut`, `updated_at_utc`, plus `disconnected_devices` (`[]string`) and `devices` metadata.
   - `POST /disconnect` accepts `{"device": "<name|all|*>"}`, validates the device, updates state, returns `200 OK` JSON status, or `400 Bad Request` on invalid input.
   - `POST /reconnect` accepts `{"device": "<name|all|*|''>"}`, reconnects device(s), triggers an immediate dispatch batch (`onDispatch`), and returns `200 OK` JSON status.
3. Stdin loop supports `disconnect <device|all>` (alias `offline`) and `reconnect [device|all]` (alias `online`).

## Acceptance Criteria

- `go test -v ./...` in `bot/` passes all existing and new tests (`exit code 0`).
- `go vet ./...` in `bot/` passes with `exit code 0`.
