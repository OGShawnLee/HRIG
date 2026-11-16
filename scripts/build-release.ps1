# Builds the judge-ready package into dist\HRIG (then ZIP that folder).
# Dev-machine only: requires Go, Bun and the Wails v3 CLI. Judges need none of them.
# Usage (from anywhere):  powershell -NoProfile -ExecutionPolicy Bypass -File scripts\build-release.ps1

$ErrorActionPreference = 'Stop'

$Repo = Split-Path -Parent $PSScriptRoot
$Out  = Join-Path $Repo 'dist\HRIG'
$PgSrc = Join-Path $Repo 'local\pgsql'

function Step($m) { Write-Host ("`n=== " + $m) -ForegroundColor Cyan }
function Check-Exit($what) { if ($LASTEXITCODE -ne 0) { throw "$what failed (exit $LASTEXITCODE)" } }

if (-not (Test-Path (Join-Path $PgSrc 'bin\postgres.exe'))) { throw "Portable PostgreSQL not found at $PgSrc" }

Step 'Clean output folder'
if (Test-Path $Out) {
    if (-not $Out.EndsWith('dist\HRIG')) { throw "Refusing to delete unexpected path: $Out" }
    Remove-Item -Recurse -Force $Out
}
New-Item -ItemType Directory -Force -Path "$Out\app", "$Out\services", "$Out\pgsql\bin", "$Out\db" | Out-Null

Step 'Build desktop app (frontend via Vite embedded in the Go binary)'
Push-Location (Join-Path $Repo 'app')
try {
    $appOut = ("$Out\app") -replace '\\', '/'
    & wails3 task build "BIN_DIR=$appOut"
    Check-Exit 'wails3 task build'
} finally { Pop-Location }

Step 'Build server and bot'
$ldflags = '-s -w'
& go -C (Join-Path $Repo 'server') build -trimpath -ldflags $ldflags -o "$Out\services\hrig-server.exe" ./cmd/server
Check-Exit 'go build server'
& go -C (Join-Path $Repo 'bot') build -trimpath -ldflags $ldflags -o "$Out\services\hrig-bot.exe" .
Check-Exit 'go build bot'

Step 'Copy slim PostgreSQL runtime (bin/lib/share only; no data, no pgAdmin)'
$exes = 'postgres.exe', 'initdb.exe', 'pg_ctl.exe', 'psql.exe', 'createdb.exe', 'dropdb.exe', 'pg_isready.exe'
foreach ($e in $exes) { Copy-Item (Join-Path $PgSrc "bin\$e") "$Out\pgsql\bin\" }
Copy-Item (Join-Path $PgSrc 'bin\*.dll') "$Out\pgsql\bin\"
Copy-Item (Join-Path $PgSrc 'lib')   "$Out\pgsql\lib"   -Recurse
Copy-Item (Join-Path $PgSrc 'share') "$Out\pgsql\share" -Recurse
$shareDoc = "$Out\pgsql\share\doc"
if (Test-Path $shareDoc) { Remove-Item -Recurse -Force $shareDoc }
Copy-Item (Join-Path $PgSrc 'server_license.txt') "$Out\pgsql\" -ErrorAction SilentlyContinue

Step 'Copy SQL and launcher files'
foreach ($sql in '001_schema.sql', '002_seed.sql', '003_demo_seed.sql', '004_medical_manual.sql') {
    Copy-Item (Join-Path $Repo "db\$sql") "$Out\db\"
}
Copy-Item (Join-Path $Repo 'scripts\release\*') $Out

Step 'Summary'
$bytes = (Get-ChildItem $Out -Recurse -File | Measure-Object Length -Sum).Sum
"{0:N1} MB in {1}" -f ($bytes / 1MB), $Out
Get-ChildItem $Out | Format-Table Name, Mode -AutoSize
Write-Host 'Done. ZIP the folder dist\HRIG and test it elsewhere.' -ForegroundColor Green
