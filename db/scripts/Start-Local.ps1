[CmdletBinding()]
param()
$ErrorActionPreference = 'Stop'
$workspace = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$binary = Join-Path $workspace '.local\cockroach\cockroach-v26.2.7.windows-6.2-amd64\cockroach.exe'
if (-not (Test-Path -LiteralPath $binary)) {
    throw 'Ejecuta primero .\db\scripts\Install-Cockroach.ps1'
}
$dataPath = Join-Path $workspace '.local\cockroach-data-dev'
$logPath = Join-Path $workspace '.local\cockroach-logs-dev'
New-Item -ItemType Directory -Force -Path $dataPath,$logPath | Out-Null
Write-Host 'Desarrollo local con datos ficticios. Deja esta terminal abierta; Ctrl+C detiene el nodo.'
Write-Host 'SQL: 127.0.0.1:26257 / Consola: http://127.0.0.1:8080'
& $binary start-single-node --insecure --listen-addr=127.0.0.1:26258 --sql-addr=127.0.0.1:26257 --http-addr=127.0.0.1:8080 "--store=path=$dataPath" "--log-dir=$logPath" --cache=128MiB --max-sql-memory=128MiB
if ($LASTEXITCODE -ne 0) { throw 'CockroachDB no pudo iniciar; revisa puertos y logs.' }
