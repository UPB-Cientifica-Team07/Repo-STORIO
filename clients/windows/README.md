# UPB-CIENTÍFICA - File Sync Client Windows

Cliente nativo Windows para sincronización automática y bidireccional
con la infraestructura de UPB-CIENTÍFICA.

El cliente comparte la implementación Go utilizada por el cliente Linux.

## Funcionalidades

- Autenticación mediante Auth Service.
- Comunicación gRPC con Sync Service.
- Detección automática de cambios del sistema de archivos.
- Creación, modificación y eliminación de archivos.
- Sincronización bidireccional multi-dispositivo.
- Reconexión automática.
- Reconciliación al iniciar.
- Estado local persistente.
- SHA-256 para detección de contenido.
- Versionado lógico.
- ACK explícito de cambios.
- Cursor persistente por dispositivo.
- Reentrega de cambios pendientes después de desconexión.
- Prevención de ciclos de sincronización.

## Arquitectura

    Directorio Windows
           |
           | fsnotify
           v
    sync-client.exe
           |
           | gRPC + Bearer token
           v
      Sync Service
           |
           +---- PostgreSQL
           |     metadata
           |     versiones
           |     cambios
           |     cursores
           |
           +---- File Service
                 Home central

## Compilación cruzada desde Linux

Desde la raíz del repositorio:

    ./clients/windows/build.sh

El ejecutable se genera en:

    clients/windows/bin/sync-client.exe

La carpeta `bin/` no se versiona.

## Compilación desde Windows

Con Go instalado, abrir PowerShell desde la raíz del repositorio:

    .\clients\windows\build.ps1

## Configuración

Ejemplo PowerShell:

    $env:SYNC_SERVER = "192.168.1.100:50055"
    $env:SYNC_AUTH_URL = "http://192.168.1.100:8081"

    $env:SYNC_USERNAME = "usuario"
    $env:SYNC_PASSWORD = "contraseña"

    $env:SYNC_DEVICE_ID = "$env:COMPUTERNAME-windows"

    $env:SYNC_DIRECTORY = Join-Path `
        $env:USERPROFILE `
        "UPB-Cientifica"

`SYNC_USERNAME`, `SYNC_PASSWORD` y `SYNC_DEVICE_ID`
son obligatorios.

La dirección configurada en `SYNC_SERVER` y `SYNC_AUTH_URL`
debe corresponder a la IP del servidor UPB-CIENTÍFICA accesible desde
el equipo Windows.

## Ejecución

    .\clients\windows\bin\sync-client.exe

El cliente permanece activo observando el directorio configurado
y propagando cambios automáticamente.

## Sincronización

Los cambios soportados son:

- `FILE_CREATED`
- `FILE_CHANGED`
- `FILE_DELETED`

Cada cambio recibido correctamente se confirma mediante ACK.
Un cambio no confirmado permanece pendiente para el dispositivo
y puede ser entregado nuevamente después de una reconexión.
