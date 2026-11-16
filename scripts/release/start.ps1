# HRIG launcher: DB -> Server -> Bot -> App, sequentially. Cleans everything up on exit.
# Windows PowerShell 5.1 compatible. Keep this file ASCII-only.

$ErrorActionPreference = 'Stop'

$Root     = $PSScriptRoot
$PgBin    = Join-Path $Root 'pgsql\bin'
$Server   = Join-Path $Root 'services\hrig-server.exe'
$Bot      = Join-Path $Root 'services\hrig-bot.exe'
$App      = Join-Path $Root 'app\hrig.exe'
$DbDir    = Join-Path $Root 'db'

if ($env:HRIG_HOME) { $HomeDir = $env:HRIG_HOME } else { $HomeDir = Join-Path $env:LOCALAPPDATA 'HRIG' }
$PgData   = Join-Path $HomeDir 'pgdata'
$LogDir   = Join-Path $HomeDir 'logs'
$RunDir   = Join-Path $HomeDir 'run'
New-Item -ItemType Directory -Force -Path $HomeDir, $LogDir, $RunDir | Out-Null

function Say($msg)  { Write-Host ("[HRIG] " + $msg) -ForegroundColor Cyan }
function Fail($msg) { Write-Host ("[HRIG] ERROR: " + $msg) -ForegroundColor Red; Write-Host ("[HRIG] Logs: " + $LogDir) -ForegroundColor Yellow }

function Get-FreePort([int]$start) {
    for ($p = $start; $p -lt ($start + 50); $p++) {
        $busy = Get-NetTCPConnection -State Listen -LocalPort $p -ErrorAction SilentlyContinue
        if (-not $busy) { return $p }
    }
    throw "No free TCP port found near $start"
}

function Invoke-Native([string]$exe, [string[]]$arguments, [string]$logName) {
    $log = Join-Path $LogDir $logName
    $outFile = [System.IO.Path]::GetTempFileName()
    try {
        $p = Start-Process -FilePath $exe -ArgumentList $arguments -NoNewWindow -Wait -PassThru `
            -RedirectStandardOutput $outFile -RedirectStandardError ($log + '.err')
        Get-Content $outFile -ErrorAction SilentlyContinue | Out-File -FilePath $log -Append -Encoding utf8
        return $p.ExitCode
    } finally { Remove-Item $outFile -Force -ErrorAction SilentlyContinue }
}

function Q([string]$s) { return '"' + $s + '"' }   # quote a path for Start-Process -ArgumentList

$serverProc = $null
$botProc    = $null
$pgStarted  = $false

function Stop-Everything {
    Say 'Shutting down...'
    foreach ($p in @($script:botProc, $script:serverProc)) {
        if ($p -and -not $p.HasExited) { try { Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue } catch {} }
    }
    if ($script:pgStarted -or (Test-Path (Join-Path $PgData 'postmaster.pid'))) {
        try { Invoke-Native (Join-Path $PgBin 'pg_ctl.exe') @('-D', (Q $PgData), 'stop', '-m', 'fast', '-w') 'pg_ctl_stop.log' | Out-Null } catch {}
    }
}

try {
    foreach ($f in @($Server, $Bot, $App, (Join-Path $PgBin 'postgres.exe'), (Join-Path $PgBin 'initdb.exe'), (Join-Path $PgBin 'psql.exe'))) {
        if (-not (Test-Path $f)) { throw "Missing file: $f  (did the ZIP extract completely?)" }
    }

    # Files from a downloaded ZIP carry a "Mark of the Web" that makes Windows show a security prompt
    # when the unsigned app starts. Remove it from everything in this folder (harmless if absent).
    Say 'Preparing files...'
    try { Get-ChildItem -LiteralPath $Root -Recurse -File -ErrorAction SilentlyContinue | Unblock-File -ErrorAction SilentlyContinue } catch {}

    Say 'Cleaning leftovers from a previous run (if any)...'
    Get-Process -Name 'hrig-server', 'hrig-bot', 'hrig' -ErrorAction SilentlyContinue |
        Where-Object { $_.Path -and $_.Path.StartsWith($Root, [System.StringComparison]::OrdinalIgnoreCase) } |
        ForEach-Object { Stop-Process -Id $_.Id -Force -ErrorAction SilentlyContinue }
    if (Test-Path (Join-Path $PgData 'postmaster.pid')) {
        Invoke-Native (Join-Path $PgBin 'pg_ctl.exe') @('-D', (Q $PgData), 'stop', '-m', 'fast', '-w') 'pg_ctl_stop.log' | Out-Null
    }

    $PgPort  = Get-FreePort 54329
    $ApiPort = Get-FreePort 18080
    $BotPort = Get-FreePort 8090

    # ---- 1. Database -------------------------------------------------------------
    if (-not (Test-Path (Join-Path $PgData 'PG_VERSION'))) {
        Say 'First run: initializing database cluster (about 10-30 seconds)...'
        $rc = Invoke-Native (Join-Path $PgBin 'initdb.exe') @('-D', (Q $PgData), '-U', 'postgres', '-E', 'UTF8', '--locale=C', '-A', 'trust') 'initdb.log'
        if ($rc -ne 0) { throw 'initdb failed (see initdb.log). If you ran this as Administrator, try a normal user.' }
    }

    Say ("Starting PostgreSQL on 127.0.0.1:" + $PgPort + " ...")
    # NOTE: no output redirection here. postgres.exe is a long-lived child that would inherit the
    # redirected handles and make Start-Process -Wait hang forever. Postgres logs to postgres.log (-l).
    # Also: Windows PowerShell 5.1 "Start-Process -Wait" waits for the whole process TREE (it would wait
    # for the postgres daemon forever), so wait on the pg_ctl process only.
    $pgctl = Start-Process -FilePath (Join-Path $PgBin 'pg_ctl.exe') -WindowStyle Hidden -PassThru `
        -ArgumentList @('-D', (Q $PgData), '-l', (Q (Join-Path $LogDir 'postgres.log')), '-o', (Q "-h 127.0.0.1 -p $PgPort"), 'start', '-w', '-t', '90')
    $null = $pgctl.Handle
    if (-not $pgctl.WaitForExit(120000)) { throw 'pg_ctl start timed out (see postgres.log).' }
    if ($pgctl.ExitCode -ne 0) { throw 'PostgreSQL did not start (see postgres.log).' }
    $pgStarted = $true

    $psql = Join-Path $PgBin 'psql.exe'
    $psqlBase = @('-X', '-q', '-h', '127.0.0.1', '-p', "$PgPort", '-U', 'postgres')

    $outFile = [System.IO.Path]::GetTempFileName()
    $p = Start-Process -FilePath $psql -ArgumentList ($psqlBase + @('-d', 'postgres', '-tAc', (Q "SELECT 1 FROM pg_database WHERE datname='hrig'"))) -NoNewWindow -Wait -PassThru -RedirectStandardOutput $outFile -RedirectStandardError (Join-Path $LogDir 'psql_check.err')
    $dbExists = ((Get-Content $outFile -ErrorAction SilentlyContinue) -join '').Trim() -eq '1'
    Remove-Item $outFile -Force -ErrorAction SilentlyContinue

    if (-not $dbExists) {
        Say 'Creating database "hrig"...'
        $rc = Invoke-Native (Join-Path $PgBin 'createdb.exe') @('-h', '127.0.0.1', '-p', "$PgPort", '-U', 'postgres', 'hrig') 'createdb.log'
        if ($rc -ne 0) { throw 'createdb failed (see createdb.log.err).' }
    }

    $outFile = [System.IO.Path]::GetTempFileName()
    $p = Start-Process -FilePath $psql -ArgumentList ($psqlBase + @('-d', 'hrig', '-tAc', (Q "SELECT to_regclass('public.astronaut') IS NOT NULL"))) -NoNewWindow -Wait -PassThru -RedirectStandardOutput $outFile -RedirectStandardError (Join-Path $LogDir 'psql_check.err')
    $schemaReady = ((Get-Content $outFile -ErrorAction SilentlyContinue) -join '').Trim() -eq 't'
    Remove-Item $outFile -Force -ErrorAction SilentlyContinue

    if (-not $schemaReady) {
        Say 'Loading schema and demo data...'
        foreach ($sql in '001_schema.sql', '002_seed.sql', '003_demo_seed.sql', '004_medical_manual.sql') {
            $rc = Invoke-Native $psql ($psqlBase + @('-d', 'hrig', '-v', 'ON_ERROR_STOP=1', '-f', (Q (Join-Path $DbDir $sql)))) 'migrate.log'
            if ($rc -ne 0) {
                # Leave no half-applied schema behind: next launch starts clean.
                Invoke-Native (Join-Path $PgBin 'dropdb.exe') @('-h', '127.0.0.1', '-p', "$PgPort", '-U', 'postgres', '--if-exists', 'hrig') 'dropdb.log' | Out-Null
                throw "Loading $sql failed (see migrate.log.err)."
            }
        }
    }

    # ---- 2. Server ---------------------------------------------------------------
    $env:DATABASE_URL    = "postgres://postgres@127.0.0.1:$PgPort/hrig?sslmode=disable"
    $env:HRIG_API_ADDR   = "127.0.0.1:$ApiPort"
    Say ("Starting server on 127.0.0.1:" + $ApiPort + " ...")
    $serverProc = Start-Process -FilePath $Server -WorkingDirectory $RunDir -WindowStyle Hidden -PassThru `
        -RedirectStandardOutput (Join-Path $LogDir 'server.log') -RedirectStandardError (Join-Path $LogDir 'server.err.log')

    $ready = $false
    for ($i = 0; $i -lt 60; $i++) {
        if ($serverProc.HasExited) { throw 'Server exited during startup (see server.log / server.err.log).' }
        try {
            $r = Invoke-WebRequest -Uri "http://127.0.0.1:$ApiPort/health" -UseBasicParsing -TimeoutSec 2
            if ($r.StatusCode -eq 200) { $ready = $true; break }
        } catch {}
        Start-Sleep -Milliseconds 500
    }
    if (-not $ready) { throw 'Server did not become healthy in 30 seconds.' }

    # ---- 3. Bot ------------------------------------------------------------------
    $env:SENSOR_ENDPOINT  = "http://127.0.0.1:$ApiPort/"
    $env:BOT_CONTROL_ADDR = "127.0.0.1:$BotPort"
    Say 'Starting sensor simulator (bot)...'
    $botProc = Start-Process -FilePath $Bot -WorkingDirectory $RunDir -WindowStyle Hidden -PassThru `
        -RedirectStandardOutput (Join-Path $LogDir 'bot.log') -RedirectStandardError (Join-Path $LogDir 'bot.err.log')
    Start-Sleep -Seconds 2
    if ($botProc.HasExited) { throw 'Bot exited during startup (see bot.log / bot.err.log).' }

    # ---- 4. App ------------------------------------------------------------------
    $env:HRIG_API_URL = "http://127.0.0.1:$ApiPort"
    Say 'All services are up. Launching HRIG...'
    Write-Host ''
    Write-Host ("  Sensor simulator control panel (inject anomalies / disconnect sensors):  http://127.0.0.1:" + $BotPort) -ForegroundColor Green
    Write-Host '  Close the HRIG window to stop everything.' -ForegroundColor Green
    Write-Host ''
    # CreateProcess (UseShellExecute=false) instead of the shell launch path: no security-prompt dialog.
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = $App
    $psi.WorkingDirectory = $RunDir
    $psi.UseShellExecute = $false
    $appProc = [System.Diagnostics.Process]::Start($psi)
    $appProc.WaitForExit()
    Say 'HRIG window closed.'
}
catch {
    Fail $_.Exception.Message
    Stop-Everything
    exit 1
}
Stop-Everything
Say 'Done.'
exit 0
