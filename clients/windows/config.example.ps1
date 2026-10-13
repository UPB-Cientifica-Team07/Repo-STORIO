# UPB-CIENTIFICA
# File Sync Client - Windows

$ServerAddress = "10.152.172.91"

$env:SYNC_SERVER = "${ServerAddress}:50055"
$env:SYNC_AUTH_URL = "https://${ServerAddress}:8081"

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
