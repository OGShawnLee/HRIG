@echo off
title HRIG Launcher
cd /d "%~dp0"
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0start.ps1"
if errorlevel 1 (
  echo.
  echo HRIG failed to start. See the messages above.
  pause
)
