$ErrorActionPreference = "Stop"

$ScriptDir =
    Split-Path `
        -Parent `
        $MyInvocation.MyCommand.Path

$RootDir =
    Resolve-Path (
        Join-Path `
            $ScriptDir `
            "..\.."
    )

$SourceClient =
    Join-Path `
        $RootDir `
        "clients\windows\bin\sync-client.exe"

$SourceRunner =
    Join-Path `
        $RootDir `
        "clients\windows\run-scheduled-sync.ps1"

$BaseDir =
    Join-Path `
        $env:LOCALAPPDATA `
        "UPB-Cientifica"

$TargetClient =
    Join-Path `
        $BaseDir `
        "sync-client.exe"

$TargetRunner =
    Join-Path `
        $BaseDir `
        "run-scheduled-sync.ps1"

$ConfigFile =
    Join-Path `
        $BaseDir `
        "sync-config.ps1"

$SyncDirectory =
    Join-Path `
        $env:USERPROFILE `
        "UPB-Cientifica"

$TaskName =
    "UPB-CIENTIFICA File Sync"

New-Item `
    -ItemType Directory `
    -Force `
    -Path $BaseDir |
    Out-Null

New-Item `
    -ItemType Directory `
    -Force `
    -Path $SyncDirectory |
    Out-Null

Copy-Item `
    -Force `
    $SourceClient `
    $TargetClient

Copy-Item `
    -Force `
    $SourceRunner `
    $TargetRunner

if (-not (Test-Path $ConfigFile)) {

    @"
`$env:SYNC_SERVER = "192.168.10.13:50055"
`$env:SYNC_AUTH_URL = "https://192.168.10.13:8081"

`$env:SYNC_USERNAME = "CAMBIAR_LOCALMENTE"
`$env:SYNC_PASSWORD = "CAMBIAR_LOCALMENTE"

`$env:SYNC_DEVICE_ID = "`$env:COMPUTERNAME-windows-scheduled"

`$env:SYNC_DIRECTORY = "$SyncDirectory"

`$env:SYNC_TLS_CA_FILE = "C:\RUTA\upb_dev_ca.crt"
`$env:AUTH_TLS_CA_FILE = "C:\RUTA\upb_dev_ca.crt"
"@ |
        Set-Content `
            -Encoding UTF8 `
            $ConfigFile

    Write-Host ""
    Write-Host "IMPORTANTE:"
    Write-Host "Edite localmente:"
    Write-Host $ConfigFile
    Write-Host ""
    Write-Host "No copie credenciales al repositorio."
}

$PowerShell =
    (Get-Command powershell.exe).Source

$Arguments =
    "-NoProfile -ExecutionPolicy Bypass -File `"$TargetRunner`""

$Action =
    New-ScheduledTaskAction `
        -Execute $PowerShell `
        -Argument $Arguments

$Trigger =
    New-ScheduledTaskTrigger `
        -Once `
        -At (Get-Date).AddMinutes(2) `
        -RepetitionInterval (
            New-TimeSpan `
                -Minutes 15
        ) `
        -RepetitionDuration (
            New-TimeSpan `
                -Days 3650
        )

$Settings =
    New-ScheduledTaskSettingsSet `
        -StartWhenAvailable `
        -ExecutionTimeLimit (
            New-TimeSpan `
                -Minutes 5
        ) `
        -MultipleInstances IgnoreNew

$Principal =
    New-ScheduledTaskPrincipal `
        -UserId $env:USERNAME `
        -LogonType Interactive `
        -RunLevel Limited

Register-ScheduledTask `
    -TaskName $TaskName `
    -Action $Action `
    -Trigger $Trigger `
    -Settings $Settings `
    -Principal $Principal `
    -Force |
    Out-Null

Write-Host ""
Write-Host "========================================="
Write-Host " PROGRAMACIÓN INSTALADA"
Write-Host "========================================="
Write-Host "Task: $TaskName"
Write-Host "Intervalo: 15 minutos"
Write-Host "Modo: SYNC_ONCE"
Write-Host "Config local:"
Write-Host $ConfigFile
Write-Host "========================================="
