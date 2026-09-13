$ErrorActionPreference = "Stop"

$BaseDir =
    Join-Path `
        $env:LOCALAPPDATA `
        "UPB-Cientifica"

$ConfigFile =
    Join-Path `
        $BaseDir `
        "sync-config.ps1"

$Client =
    Join-Path `
        $BaseDir `
        "sync-client.exe"

if (-not (Test-Path $ConfigFile)) {
    throw "No existe configuración local: $ConfigFile"
}

if (-not (Test-Path $Client)) {
    throw "No existe cliente Sync: $Client"
}

. $ConfigFile

$requiredVariables = @(
    "SYNC_SERVER",
    "SYNC_AUTH_URL",
    "SYNC_USERNAME",
    "SYNC_PASSWORD",
    "SYNC_DEVICE_ID",
    "SYNC_DIRECTORY",
    "SYNC_TLS_CA_FILE",
    "AUTH_TLS_CA_FILE"
)

foreach ($name in $requiredVariables) {

    $value =
        [Environment]::GetEnvironmentVariable(
            $name,
            "Process"
        )

    if ([string]::IsNullOrWhiteSpace($value)) {
        throw "Variable obligatoria ausente: $name"
    }
}

$env:SYNC_ONCE = "true"

Write-Host "========================================="
Write-Host " UPB-CIENTIFICA SCHEDULED FILE SYNC"
Write-Host "========================================="
Write-Host "Device: $env:SYNC_DEVICE_ID"
Write-Host "Directory: $env:SYNC_DIRECTORY"
Write-Host "Mode: SYNC_ONCE"
Write-Host "========================================="

& $Client

if ($LASTEXITCODE -ne 0) {
    throw "File Sync terminó con código $LASTEXITCODE"
}

Write-Host "SYNC_ONCE completado correctamente"
