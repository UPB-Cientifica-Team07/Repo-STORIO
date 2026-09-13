# UPB-CIENTIFICA
# File Sync Client - Windows

$env:SYNC_SERVER = "192.168.1.100:50055"
$env:SYNC_AUTH_URL = "https://192.168.10.13:8081"

$env:SYNC_USERNAME = "usuario"
$env:SYNC_PASSWORD = "CAMBIAR_LOCALMENTE"

$env:SYNC_DEVICE_ID = "$env:COMPUTERNAME-windows"

$env:SYNC_DIRECTORY = Join-Path `
    $env:USERPROFILE `
    "UPB-Cientifica"

# CA pública usada para validar Auth HTTPS
$env:AUTH_TLS_CA_FILE = "security/pki/upb_dev_ca.crt"

# CA pública para Sync gRPC TLS
$env:SYNC_TLS_CA_FILE = "C:\\RUTA\\upb_dev_ca.crt"
