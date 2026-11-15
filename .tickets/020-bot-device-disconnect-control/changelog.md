# 020 — Changelog: Bot Device Disconnection / Reconnection Control

## Files Touched

- `bot/main.go`
- `bot/main_test.go`
- `.tickets/020-bot-device-disconnect-control/changelog.md` (new)

## Summary of Changes

1. **Thread-Safe Device Disconnection State (`bot/main.go`)**
   - Added `disconnected map[string]bool` to `generator` (initialized in `newGenerator`), plus `isDisconnected(device string)`, `disconnectedDevices()`, `controlSnapshot()`, `disconnect(device string)`, and `reconnect(device string)` supporting all 7 configured device IDs (`biomonitor_01`, `baumanometro_01`, `dosimeter_01`, `cabin_sensor_01`, `suit_sensor_01`, `lab_assay_01`, `ultrasound_01`), `"all"`, `"*"`, and empty target on `reconnect` to reconnect all devices.
   - Updated `sendBatch` to skip posting `IngestRequest` payloads for any sensor `s` where `gen.isDisconnected(s.Device)` is `true`.

2. **Extended HTTP Control API & Embedded Dark-HUD Control Panel (`bot/main.go`)**
   - Added `deviceControlInfo` and `controlStatus` (embedding `overrideState` at top level alongside `disconnected_devices` and `devices` metadata), returned by `GET /status`, `POST /override`, `POST /cancel`, `POST /disconnect`, and `POST /reconnect`.
   - Added `POST /disconnect` (`{"device": "<name|all|*>"}`, returning `400 Bad Request` on invalid JSON or unknown/empty device), `POST /reconnect` (`{"device": "<name|all|*|''>"}`, triggering `onDispatch` for immediate batch emission), and `GET /` serving `controlPanelHTML` (`text/html; charset=utf-8`), an interactive dark-HUD panel with Anomaly Override and Device Connection Control actions.

3. **Stdin Command Extensions (`bot/main.go`)**
   - Extended `handleCommandLine` with `disconnect` / `offline <device|all>` and `reconnect` / `online [device|all]` (triggering `onDispatch` on reconnect), and updated `status` and `help` output.

4. **Tests (`bot/main_test.go`)**
   - Added `TestSendBatchSkipsDisconnectedDeviceAndResumesOnReconnect` and `TestDeviceDisconnectAllReconnectAllAndHTTPControls` covering `sendBatch` skipping disconnected devices and resuming on reconnect, `400 Bad Request` handling, `disconnect all` / `reconnect all` (`all`, `*`, `""`), `GET /`, `GET /status`, `POST /disconnect`, `POST /reconnect`, and stdin `disconnect`/`offline`/`reconnect`/`online`/`status`/`help` commands.

## Verification Results

- `go -C /home/sacred/bd/HRIG/bot test -v ./...` -> exit `0` (5/5 tests `PASS`: `TestNominalRangesComplyWithSeedThresholds`, `TestOverrideAndCancelTransitions`, `TestControlHTTPEndpointsAndStdinCommands`, `TestSendBatchSkipsDisconnectedDeviceAndResumesOnReconnect`, `TestDeviceDisconnectAllReconnectAllAndHTTPControls`).
- `go -C /home/sacred/bd/HRIG/bot test -race -v ./...` -> exit `0` (5/5 tests `PASS` under race detector).
- `go -C /home/sacred/bd/HRIG/bot vet ./...` -> exit `0`.
