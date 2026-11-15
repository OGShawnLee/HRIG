# 020 — Add Device Disconnection / Reconnection Simulation to Bot

| Field | Value |
| :--- | :--- |
| **ID** | `020-bot-device-disconnect-control` |
| **Status** | `completed` |
| **Scope** | `bot/main.go`, `bot/main_test.go` |

## Context

The telemetry simulator (`bot/main.go`) currently sends readings for all 7 logical sensor devices (`biomonitor_01`, `baumanometro_01`, `dosimeter_01`, `cabin_sensor_01`, `suit_sensor_01`, `lab_assay_01`, `ultrasound_01`) every batch interval. To allow testing real-world sensor dropouts (`StatusStale` when the server stops receiving data from a source), the bot needs manual controls to disconnect and reconnect individual devices (or all devices) via CLI stdin commands, HTTP control endpoints, and a built-in browser control page at `GET /`.

## Grounding Rules

- Modify only `bot/main.go`, `bot/main_test.go`, and `.tickets/020-bot-device-disconnect-control/`.
- Preserve existing `GET /status`, `POST /override`, `POST /cancel` JSON compatibility (`overrideState` top-level fields) and existing stdin commands (`abnormal`, `critical`, `cancel`, `normal`, `status`, `help`).
- Do NOT break nominal range compliance tests in `bot/main_test.go`.

## Steps

1. Update `bot/main.go`:
   - Track disconnected devices (`disconnected map[string]bool`) in `generator` with thread-safe `disconnect(device string)`, `reconnect(device string)`, `isDisconnected(device string)`, `disconnectedDevices() []string`, and `controlSnapshot()` methods. Support exact device names (`biomonitor_01`, `baumanometro_01`, `dosimeter_01`, `cabin_sensor_01`, `suit_sensor_01`, `lab_assay_01`, `ultrasound_01`) plus `"*"` / `"all"` (and empty target on `reconnect` to reconnect all).
   - Update `sendBatch` to skip any sensor `s` where `gen.isDisconnected(s.Device)` is `true`, so no `IngestRequest` payloads are posted for that apparatus while disconnected.
   - Extend `controlStatus` JSON payload returned by `/status`, `/override`, `/cancel`, `/disconnect`, and `/reconnect` to include `overrideState` at the top level alongside `disconnected_devices: []string` and `devices: []deviceControlInfo`.
   - Add `POST /disconnect` (`{"device": "..."}`), `POST /reconnect` (`{"device": "..."}`), and `GET /` (serving a clean embedded dark-HUD HTML control panel with both Anomaly Override and Device Disconnection/Reconnection controls) in `newControlHandler`.
   - Extend `handleCommandLine` for `disconnect` / `offline` and `reconnect` / `online` stdin commands, plus updated `status` and `help` output.
2. Add thorough tests in `bot/main_test.go` covering:
   - `sendBatch` skipping a disconnected device (`baumanometro_01`) while still transmitting the remaining devices, and resuming transmission after `reconnect`.
   - Invalid device error handling (`400 Bad Request`).
   - `disconnect all` + `reconnect all`, `POST /disconnect`, `POST /reconnect`, `GET /status`, `GET /` HTML UI, and stdin `disconnect` / `reconnect` commands.
3. Run `go test -v ./...` inside `bot/` and report verification results in `changelog.md`.
