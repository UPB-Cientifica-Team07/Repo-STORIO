# UPB-CIENTIFICA
# File Sync Client - Windows

$env:SYNC_SERVER = "192.168.1.100:50055"
$env:SYNC_AUTH_URL = "http://192.168.1.100:8081"

$env:SYNC_USERNAME = "usuario"
$env:SYNC_PASSWORD = "CAMBIAR_LOCALMENTE"

$env:SYNC_DEVICE_ID = "$env:COMPUTERNAME-windows"

$env:SYNC_DIRECTORY = Join-Path `
    $env:USERPROFILE `
    "UPB-Cientifica"
