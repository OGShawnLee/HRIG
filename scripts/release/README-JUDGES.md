# HRIG — Resource Integration Gear (Judges' Quick Start)

Astronaut health monitoring prototype for long-duration missions. Everything is bundled: no installs and no admin rights. Internet is optional: it is only used to download the Satoshi font (offline, a system font is used instead).

## Run it

1. Unzip the whole folder anywhere (a short path such as `C:\HRIG` is best).
2. Double-click **`Start-HRIG.bat`**.
3. Wait a few seconds. The **first launch can take 1–2 minutes** while the local database is created and loaded with demo data (the console shows progress; please wait). Later launches are much faster.
4. The HRIG window opens. Close it to stop everything cleanly.

Requirements: Windows 10/11 (64-bit) with the Microsoft Edge WebView2 Runtime (already present on Windows 11 and up-to-date Windows 10).

## What is running

`Start-HRIG.bat` starts, in order: a private PostgreSQL database, the HRIG API server, a sensor simulator (bot) that continuously sends vitals, and finally the HRIG desktop app. Nothing is installed system-wide; data lives in `%LOCALAPPDATA%\HRIG`.

## Try the alerts

The console window prints the address of the **sensor simulator control panel** (default `http://127.0.0.1:8090`). Open it in a browser to force abnormal or critical values and to disconnect individual sensors, then watch HRIG react live (alert views, blinking cards, "No Signal" state). Use `Ctrl+K` inside HRIG to search crew, metrics and the Medical Manual.

## Troubleshooting

- **Windows SmartScreen warning** ("Windows protected your PC") — click *More info* → *Run anyway*. The files are unsigned.
- **The launcher prints an error, or the window never appears** — read the console text and the log files in `%LOCALAPPDATA%\HRIG\logs` (`postgres.log`, `server.err.log`, `bot.err.log`, `migrate.log.err`).
- **Missing `VCRUNTIME140.dll` / `MSVCP140.dll`** — install the Microsoft Visual C++ 2015–2022 Redistributable (x64) and run again.
- **Start fresh** — close HRIG, delete the folder `%LOCALAPPDATA%\HRIG`, and launch again.
- **Ports** — the launcher automatically picks free ports (database 54329+, API 18080+, simulator 8090+), so it will not clash with other software.
