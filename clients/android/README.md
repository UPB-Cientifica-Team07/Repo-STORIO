# UPB-CIENTÍFICA - Android File Sync Client

Cliente Android nativo para el servicio File Sync de UPB-CIENTÍFICA.

## Funcionalidad implementada

El cliente incluye:

- autenticación mediante Auth Service;
- comunicación con Sync Service mediante gRPC;
- identificación persistente del dispositivo;
- sincronización de archivos;
- descarga de archivos creados o modificados;
- eliminación de archivos removidos;
- `change_id` para identificación monotónica de cambios;
- confirmación explícita mediante `AcknowledgeChanges`;
- redelivery cuando un cambio no ha sido confirmado;
- almacenamiento local dentro del directorio privado de la aplicación;
- sincronización automática;
- WorkManager periódico;
- WorkManager al iniciar la aplicación;
- restricción a conectividad Wi-Fi;
- validación de la subred local `192.168.10.0/24`;
- comprobación de alcance del Sync Service;
- configuración local de usuario y contraseña;
- generación de APK Android nativo.

## Servicios

Configuración actual del laboratorio:

- Auth Service: `192.168.10.13:8081`
- Sync Service: `192.168.10.13:50055`
- LAN autorizada: `192.168.10.0/24`

Esta configuración deberá parametrizarse para el despliegue definitivo.

## Confiabilidad y ACK

El servidor entrega un `change_id` por cada cambio pendiente.

El cliente aplica el cambio y solamente después de procesarlo
correctamente ejecuta:

AcknowledgeChanges(change_id)

Si falla la descarga, escritura o eliminación, no se envía ACK.
El cambio permanece pendiente para una entrega posterior.

## Sincronización automática

La aplicación utiliza:

- ConnectivityManager;
- NetworkCapabilities.TRANSPORT_WIFI;
- política de red local;
- WorkManager.

Existe un trabajo periódico y otro al iniciar la aplicación.

## Credenciales

No existen usuarios ni contraseñas de prueba embebidos en el código activo.

Las credenciales son introducidas por el usuario y almacenadas en el
espacio privado de la aplicación para que SyncWorker pueda ejecutar
sincronización en segundo plano.

El endurecimiento posterior mediante token/MFA y almacenamiento seguro
de secretos forma parte de la fase general de seguridad.

## Compilación

Desde:

clients/android

Ejecutar:

./gradlew clean assembleDebug

El APK generado queda en:

app/build/outputs/apk/debug/app-debug.apk

## Verificación

Desde la raíz del repositorio:

./clients/android/verify.sh

El script comprueba:

- equivalencia del protobuf con el servidor;
- presencia de AcknowledgeChanges;
- ausencia de credenciales de prueba;
- CredentialStore;
- WorkManager;
- política Wi-Fi/LAN;
- ausencia de backups obsoletos;
- pruebas Gradle;
- generación del APK.

## Estado de validación

La implementación ha sido validada mediante compilación, pruebas Gradle,
equivalencia de protocolo y verificaciones estáticas de integración.

No se realizó una prueba end-to-end sobre hardware Android físico ni
emulador Android. Actualmente no se dispone de dicho entorno.

Por lo tanto, el cliente Android se considera implementado y compilado,
pero la validación runtime en Android queda documentada como pendiente.
