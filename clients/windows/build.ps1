$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Resolve-Path (Join-Path $ScriptDir "..\..")
$SyncService = Join-Path $RootDir "services\sync-service"
$OutputDir = Join-Path $RootDir "clients\windows\bin"
$OutputFile = Join-Path $OutputDir "sync-client.exe"

New-Item `
    -ItemType Directory `
    -Force `
    -Path $OutputDir | Out-Null

Set-Location $SyncService

Write-Host "Compilando File Sync Client para Windows amd64..."

$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

go build `
    -trimpath `
    -o $OutputFile `
    ./cmd/watcher

if ($LASTEXITCODE -ne 0) {
    throw "La compilacion del cliente Windows fallo."
}

Write-Host ""
Write-Host "Cliente generado:"
Write-Host $OutputFile

Get-Item $OutputFile |
    Select-Object Name, Length, LastWriteTime
