[CmdletBinding()]
param()
$ErrorActionPreference = 'Stop'
$workspace = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$downloadDir = Join-Path $workspace '.local\downloads'
$installDir = Join-Path $workspace '.local\cockroach'
$zipPath = Join-Path $downloadDir 'cockroach-v26.2.7.zip'
$binary = Join-Path $installDir 'cockroach-v26.2.7.windows-6.2-amd64\cockroach.exe'
$sha256 = 'b6a0b253b1f7853d6786b1423b507c56878a11411afeebb0c5c5ed2a36098aef'
if ($env:PROCESSOR_ARCHITECTURE -ne 'AMD64') { throw 'Este instalador es para Windows x64.' }
if (Test-Path -LiteralPath $binary) {
    & $binary version
    if ($LASTEXITCODE -ne 0) { throw 'No se pudo ejecutar CockroachDB.' }
    return
}
New-Item -ItemType Directory -Force -Path $downloadDir,$installDir | Out-Null
Invoke-WebRequest -Uri 'https://binaries.cockroachdb.com/cockroach-v26.2.7.windows-6.2-amd64.zip' -OutFile $zipPath
if ((Get-FileHash -LiteralPath $zipPath -Algorithm SHA256).Hash.ToLowerInvariant() -ne $sha256) {
    throw 'SHA256 inesperado. No se instalará este archivo.'
}
Expand-Archive -LiteralPath $zipPath -DestinationPath $installDir -Force
& $binary version
if ($LASTEXITCODE -ne 0) { throw 'No se pudo ejecutar CockroachDB.' }
Write-Host 'Binario instalado en .local/cockroach. No se modificó el PATH del equipo.'
