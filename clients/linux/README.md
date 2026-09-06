# UPB-CIENTÍFICA - File Sync Client Linux

Cliente nativo Linux para la sincronización automática y bidireccional
de archivos de UPB-CIENTÍFICA.

## Funcionalidades

- Autenticación mediante Auth Service.
- Comunicación gRPC con Sync Service.
- Detección automática de cambios mediante fsnotify.
- Creación, modificación y eliminación de archivos.
- Recepción de cambios procedentes de otros dispositivos.
- Conservación del árbol de directorios.
- Estado local persistente.
- Huellas SHA-256 para detectar cambios reales.
- Debounce de eventos del sistema de archivos.
- Prevención de ciclos de sincronización.
- Reconciliación al iniciar el cliente.

## Compilación

Desde la raíz del repositorio:

    ./clients/linux/build.sh

El binario se genera en:

    clients/linux/bin/sync-client

La carpeta `bin/` no se versiona.

## Configuración

Variables requeridas:

    export SYNC_SERVER="127.0.0.1:50055"
    export SYNC_AUTH_URL="http://127.0.0.1:8081"
    export SYNC_USERNAME="usuario"
    export SYNC_PASSWORD="contraseña"
    export SYNC_DEVICE_ID="$(hostname)-linux"
    export SYNC_DIRECTORY="$HOME/UPB-Cientifica"

`SYNC_USERNAME`, `SYNC_PASSWORD` y `SYNC_DEVICE_ID`
deben definirse explícitamente.

## Ejecución

    ./clients/linux/bin/sync-client

## Arquitectura

    Directorio local
          |
          | fsnotify
          v
    File Sync Client Linux
          |
          | gRPC + Bearer token
          v
    Sync Service
          |
          +---- PostgreSQL
          |     metadata, versiones,
          |     cambios y cursores
          |
          +---- File Service
                contenido físico
                y Home central

## Sincronización multi-dispositivo

Los cambios registrados por un dispositivo son propagados mediante
`WatchChanges` a los demás dispositivos asociados al mismo usuario.

Tipos de cambios soportados:

- `FILE_CREATED`
- `FILE_CHANGED`
- `FILE_DELETED`
