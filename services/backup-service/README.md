# UPB-CIENTÍFICA - GPG Backup Service

Servicio de respaldo cifrado para el almacenamiento compartido de UPB-CIENTÍFICA.

## Objetivo

Proteger las copias de seguridad mediante cifrado asimétrico GPG.

La clave pública puede permanecer en el repositorio y se utiliza para cifrar.

La clave privada permanece fuera del repositorio y se utiliza únicamente para restaurar.

## Directorio respaldado

Por defecto:

services/file-service/data/shared-storage

Puede modificarse mediante:

BACKUP_SOURCE_DIR

## Flujo de backup

1. Se empaqueta Shared Storage en un archivo temporal `.tar.gz`.
2. Se calcula SHA-256 del archivo comprimido.
3. Se importa la clave pública del proyecto en un GNUPGHOME temporal.
4. Se cifra el archivo mediante GPG.
5. Se guarda únicamente la copia `.tar.gz.gpg`.
6. El archivo `.tar.gz` sin cifrar se elimina junto con el directorio temporal.
7. Se conserva el manifiesto SHA-256 para verificar la restauración.

## Flujo de restore

1. Se recibe una copia `.tar.gz.gpg`.
2. Se utiliza el GNUPGHOME externo que contiene la clave privada.
3. GPG descifra el archivo temporal.
4. Se calcula SHA-256.
5. Se compara con el manifiesto original.
6. Si la integridad es correcta, se extrae el contenido.
7. El archivo temporal descifrado se elimina.

## Clave pública

La clave pública se encuentra en:

services/backup-service/keys/upb-backup-public.asc

Fingerprint:

17FD906D1542B32F6622341C28652D1F56237981

Identidad:

UPB-CIENTIFICA Backup <backup@upb-cientifica.local>

## Clave privada

La clave privada NO pertenece al repositorio.

En el ambiente de desarrollo utilizado para las pruebas se encuentra en:

~/.local/share/upb-cientifica/gpg

En producción debe almacenarse fuera del árbol de la aplicación con permisos restringidos.

## Crear un backup

Desde la raíz del repositorio:

./services/backup-service/scripts/backup.sh

Opcionalmente:

BACKUP_SOURCE_DIR=/ruta/origen \
BACKUP_OUTPUT_DIR=/ruta/backups \
./services/backup-service/scripts/backup.sh

## Restaurar

Ejemplo:

BACKUP_GNUPGHOME="$HOME/.local/share/upb-cientifica/gpg" \
./services/backup-service/scripts/restore.sh \
  services/backup-service/backups/upb-backup-AAAAmmdd-HHMMSS.tar.gz.gpg \
  /tmp/upb-restore

## Verificación

Ejecutar:

./services/backup-service/verify.sh

La verificación comprueba:

- disponibilidad de GPG;
- validez de la clave pública;
- scripts ejecutables;
- ausencia de clave privada en el servicio;
- sintaxis Bash.

## Prueba realizada

Se ejecutó un ciclo completo:

Shared Storage
→ TAR temporal
→ SHA-256
→ GPG
→ archivo cifrado
→ descifrado
→ verificación SHA-256
→ restauración

Resultado:

- 24 archivos originales;
- 24 archivos restaurados;
- contenido idéntico mediante `diff -qr`;
- integridad SHA-256 correcta.

## Seguridad

Nunca deben versionarse:

- claves privadas;
- passphrases;
- archivos temporales descifrados;
- copias `.tar.gz` sin cifrar.

El `.gitignore` del servicio excluye material sensible y backups generados.
