[CmdletBinding()]
param([string]$Database = 'danber')
$ErrorActionPreference = 'Stop'
$workspace = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
if (-not $env:DANBER_DATABASE_URL) {
    $env:DANBER_DATABASE_URL = 'postgresql://root@127.0.0.1:26257/defaultdb?sslmode=disable'
}
& python (Join-Path $PSScriptRoot 'migrate.py') init --database $Database
if ($LASTEXITCODE -ne 0) { throw 'No se completó la inicialización. Revisa la salida antes de reintentar.' }
