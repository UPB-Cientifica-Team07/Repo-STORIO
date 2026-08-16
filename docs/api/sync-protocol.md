# Sync Service - Protocolo TCP de Sincronización

## 1. Objetivo

Definir el protocolo de comunicación utilizado por el Sync Service de UPB-CIENTÍFICA para sincronizar archivos entre los clientes y el almacenamiento compartido.

A diferencia de los servicios REST, el Sync Service utiliza una conexión persistente basada en TCP Sockets. Este mecanismo permite mantener una comunicación directa entre el cliente y el servicio durante las operaciones de sincronización.

El protocolo define:

- El establecimiento de conexión.
- La identificación del cliente.
- El registro de dispositivos.
- La estructura de los mensajes.
- Las operaciones disponibles.
- La transferencia de archivos.
- La notificación de cambios.
- La sincronización inicial.
- El manejo de errores.
- El cierre de conexión.

---

# 2. Alcance

Este documento define el contrato inicial de comunicación entre:

```text
Cliente Desktop
        │
        │ TCP Socket
        ▼
   Sync Service
        │
        ▼
  Shared Storage
```

El protocolo puede ser utilizado principalmente por:

- Cliente Desktop.
- Cliente Mobile.

El Cliente Web utilizará principalmente comunicación HTTPS y REST mediante el API Gateway.

Este documento no define todavía:

- Un algoritmo definitivo de resolución de conflictos.
- Replicación geográfica.
- Sincronización entre múltiples servidores.
- Cifrado específico a nivel de aplicación.
- Transferencia optimizada de archivos de gran tamaño.
- Reanudación de transferencias interrumpidas.

Estos elementos podrán incorporarse en versiones posteriores del protocolo.

---

# 3. Modelo de comunicación

La comunicación se establece mediante una conexión TCP persistente.

```text
Cliente
   │
   │ TCP
   ▼
Sync Service
   │
   ├───────────────┐
   │               │
   ▼               ▼
PostgreSQL    Shared Storage
```

La conexión permite que tanto el cliente como el servidor participen activamente en la comunicación.

El cliente puede enviar operaciones como:

```text
SYNC
UPLOAD
DOWNLOAD
DELETE
LIST
PING
DISCONNECT
```

El servidor puede responder o enviar notificaciones relacionadas con:

```text
SUCCESS
ERROR
FILE_CHANGED
FILE_DELETED
SYNC_COMPLETE
PONG
```

---

# 4. Principio de funcionamiento

Cada cliente mantiene una conexión TCP con el Sync Service mientras se encuentra realizando operaciones de sincronización.

El flujo general es:

```text
1. El cliente establece conexión.
2. El cliente se identifica.
3. El servidor valida la sesión.
4. Se registra la conexión activa.
5. El cliente solicita sincronización.
6. El servidor compara el estado de los archivos.
7. Se transfieren o actualizan los cambios necesarios.
8. El servidor puede notificar cambios.
9. El cliente puede cerrar la conexión.
```

---

# 5. Arquitectura del Sync Service

El Sync Service puede estructurarse conceptualmente de la siguiente manera:

```text
┌─────────────────────────────┐
│         Cliente             │
└──────────────┬──────────────┘
               │
               │ TCP Socket
               ▼
┌─────────────────────────────┐
│        Sync Server          │
│                             │
│  Connection Manager         │
│             │               │
│             ▼               │
│      Protocol Handler       │
│             │               │
│             ├───────────────┐
│             │               │
│             ▼               ▼
│       Sync Manager     File Manager
└─────────────┬────────────────
              │
              ├───────────────┐
              │               │
              ▼               ▼
         PostgreSQL      Shared Storage
```

Los componentes principales son:

- Connection Manager.
- Protocol Handler.
- Sync Manager.
- File Manager.
- PostgreSQL.
- Shared Storage.

---

# 6. Establecimiento de conexión

El cliente inicia la conexión TCP con el Sync Service.

Modelo:

```text
Cliente
   │
   │ TCP CONNECT
   ▼
Sync Service
   │
   ▼
Conexión establecida
```

La conexión por sí sola no implica que el cliente esté autenticado.

Después de establecer la conexión, el cliente debe identificarse.

---

# 7. Identificación del cliente

Una vez establecida la conexión, el cliente debe enviar un mensaje de autenticación.

La operación inicial será:

```text
AUTH
```

Ejemplo conceptual:

```text
AUTH|token|device_id
```

Donde:

| Campo | Descripción |
|---|---|
| AUTH | Tipo de operación |
| token | Token de autenticación |
| device_id | Identificador del dispositivo |

Ejemplo:

```text
AUTH|eyJhbGciOiJIUzI1NiJ9...|desktop-001
```

---

# 8. Respuesta de autenticación

Si el cliente es aceptado:

```text
AUTH_OK|id_usuario
```

Ejemplo:

```text
AUTH_OK|550e8400-e29b-41d4-a716-446655440000
```

Si la autenticación falla:

```text
AUTH_ERROR|TOKEN_INVALID
```

El cliente no podrá ejecutar operaciones de sincronización hasta completar correctamente el proceso de autenticación.

---

# 9. Estructura general de mensajes

El protocolo utiliza inicialmente mensajes de texto delimitados por el carácter:

```text
|
```

La estructura general es:

```text
OPERACION|PARAMETRO_1|PARAMETRO_2|...|PARAMETRO_N
```

Ejemplo:

```text
LIST|/documentos
```

Otro ejemplo:

```text
DELETE|550e8400-e29b-41d4-a716-446655440000
```

La estructura permite implementar el protocolo inicialmente sin depender de formatos complejos.

---

# 10. Reglas de los mensajes

Los mensajes deben cumplir las siguientes reglas:

1. Cada mensaje representa una operación.
2. El primer campo identifica el comando.
3. Los campos adicionales representan parámetros.
4. El carácter `|` funciona como delimitador.
5. Los saltos de línea no forman parte de los parámetros.
6. Los mensajes deben utilizar UTF-8.
7. Los comandos se manejarán inicialmente en mayúsculas.
8. El servidor debe validar el número de parámetros.
9. Los mensajes inválidos deben generar una respuesta de error.
10. Los clientes autenticados solo pueden operar sobre sus recursos autorizados.

---

# 11. Operación AUTH

## 11.1 Objetivo

Identificar al usuario y asociar la conexión TCP con una sesión autenticada.

Formato:

```text
AUTH|token|device_id
```

Ejemplo:

```text
AUTH|token-del-usuario|desktop-001
```

Respuesta exitosa:

```text
AUTH_OK|id_usuario
```

Respuesta de error:

```text
AUTH_ERROR|TOKEN_INVALID
```

---

# 12. Operación SYNC

## 12.1 Objetivo

Solicitar la sincronización entre el estado local del cliente y el estado registrado en el servidor.

Formato inicial:

```text
SYNC
```

El servidor identifica al usuario utilizando la sesión autenticada.

Flujo:

```text
Cliente
   │
   │ SYNC
   ▼
Sync Service
   │
   ├── Consultar metadatos
   ▼
PostgreSQL
   │
   ▼
Comparar estado
   │
   ▼
Determinar cambios
```

---

## 12.2 Respuesta

El servidor puede enviar una lista de cambios:

```text
SYNC_BEGIN
FILE_CHANGED|id_archivo|nombre|fecha_modificacion
FILE_CHANGED|id_archivo|nombre|fecha_modificacion
SYNC_COMPLETE
```

Ejemplo:

```text
SYNC_BEGIN
FILE_CHANGED|550e8400-e29b-41d4-a716-446655440000|investigacion.pdf|2026-08-16T10:00:00
FILE_CHANGED|660e8400-e29b-41d4-a716-446655440000|microscopia.png|2026-08-16T11:00:00
SYNC_COMPLETE
```

---

# 13. Operación LIST

## 13.1 Objetivo

Consultar la lista de archivos disponibles para el usuario autenticado.

Formato:

```text
LIST
```

Respuesta:

```text
LIST_BEGIN
FILE|id_archivo|nombre|tipo|tamano|fecha_modificacion
FILE|id_archivo|nombre|tipo|tamano|fecha_modificacion
LIST_COMPLETE
```

Ejemplo:

```text
LIST_BEGIN
FILE|550e8400-e29b-41d4-a716-446655440000|investigacion.pdf|DOCUMENTO|2457600|2026-08-16T10:00:00
FILE|660e8400-e29b-41d4-a716-446655440000|microscopia.png|IMAGEN|5242880|2026-08-16T11:00:00
LIST_COMPLETE
```

---

# 14. Operación UPLOAD

## 14.1 Objetivo

Enviar un archivo desde el cliente hacia el Sync Service.

Debido a que un archivo puede contener información binaria, la transferencia no debe depender únicamente de una cadena delimitada por `|`.

Por esta razón, la operación se divide conceptualmente en dos etapas:

```text
1. Envío de metadatos.
2. Envío del contenido binario.
```

---

## 14.2 Inicio de carga

Formato:

```text
UPLOAD|nombre|tipo|tamano
```

Ejemplo:

```text
UPLOAD|investigacion.pdf|DOCUMENTO|2457600
```

El servidor valida la solicitud.

Respuesta:

```text
UPLOAD_READY|transfer_id
```

Ejemplo:

```text
UPLOAD_READY|abc123
```

---

## 14.3 Transferencia de contenido

Después de recibir:

```text
UPLOAD_READY
```

el cliente envía la cantidad de bytes especificada en:

```text
tamano
```

Flujo:

```text
Cliente
   │
   │ UPLOAD|nombre|tipo|tamano
   ▼
Sync Service
   │
   │ UPLOAD_READY|transfer_id
   ▼
Cliente
   │
   │ Datos binarios
   ▼
Sync Service
```

Una vez finalizada la transferencia:

```text
UPLOAD_COMPLETE|id_archivo
```

---

# 15. Operación DOWNLOAD

## 15.1 Objetivo

Solicitar un archivo almacenado en el servidor.

Formato:

```text
DOWNLOAD|id_archivo
```

Ejemplo:

```text
DOWNLOAD|550e8400-e29b-41d4-a716-446655440000
```

---

## 15.2 Respuesta inicial

El servidor responde con los metadatos:

```text
DOWNLOAD_READY|nombre|tipo|tamano
```

Ejemplo:

```text
DOWNLOAD_READY|investigacion.pdf|DOCUMENTO|2457600
```

Después de esta respuesta, el servidor transmite exactamente la cantidad de bytes indicada.

Flujo:

```text
Cliente
   │
   │ DOWNLOAD|id_archivo
   ▼
Sync Service
   │
   │ DOWNLOAD_READY|nombre|tipo|tamano
   ▼
Cliente
   │
   │ Preparado para recibir
   ▼
Sync Service
   │
   │ Datos binarios
   ▼
Cliente
```

Al finalizar:

```text
DOWNLOAD_COMPLETE|id_archivo
```

---

# 16. Operación DELETE

## 16.1 Objetivo

Eliminar un archivo asociado al usuario autenticado.

Formato:

```text
DELETE|id_archivo
```

Ejemplo:

```text
DELETE|550e8400-e29b-41d4-a716-446655440000
```

Flujo:

```text
Cliente
   │
   │ DELETE|id_archivo
   ▼
Sync Service
   │
   ├── Validar usuario
   │
   ├── Validar propiedad
   │
   ├── Eliminar archivo
   │
   ▼
Shared Storage
   │
   ▼
Sync Service
   │
   ├── Eliminar metadatos
   │
   ▼
PostgreSQL
```

Respuesta:

```text
DELETE_OK|id_archivo
```

---

# 17. Notificación FILE_CHANGED

El Sync Service puede informar a un cliente que existe una modificación relacionada con un archivo.

Formato:

```text
FILE_CHANGED|id_archivo|nombre|fecha_modificacion
```

Ejemplo:

```text
FILE_CHANGED|550e8400-e29b-41d4-a716-446655440000|investigacion.pdf|2026-08-16T12:00:00
```

Esta notificación permite que un cliente conozca que debe actualizar su estado local.

---

# 18. Notificación FILE_DELETED

Cuando un archivo sincronizado es eliminado, el servidor puede informar a los clientes conectados asociados al usuario.

Formato:

```text
FILE_DELETED|id_archivo
```

Ejemplo:

```text
FILE_DELETED|550e8400-e29b-41d4-a716-446655440000
```

---

# 19. Operación PING

## 19.1 Objetivo

Verificar que la conexión TCP continúe activa.

Formato:

```text
PING
```

Respuesta:

```text
PONG
```

Esta operación puede utilizarse como mecanismo básico de mantenimiento de la conexión.

---

# 20. Operación DISCONNECT

Permite al cliente finalizar correctamente la conexión.

Formato:

```text
DISCONNECT
```

Respuesta:

```text
BYE
```

Después de enviar:

```text
BYE
```

el servidor puede cerrar el socket asociado.

---

# 21. Manejo de errores

Todos los errores deben utilizar una estructura controlada.

Formato general:

```text
ERROR|CODIGO|MENSAJE
```

Ejemplo:

```text
ERROR|AUTH_REQUIRED|Cliente no autenticado
```

Los códigos iniciales pueden ser:

| Código | Descripción |
|---|---|
| INVALID_COMMAND | Comando no reconocido |
| INVALID_FORMAT | Formato incorrecto |
| AUTH_REQUIRED | Cliente no autenticado |
| TOKEN_INVALID | Token inválido |
| ACCESS_DENIED | Usuario sin permisos |
| FILE_NOT_FOUND | Archivo no encontrado |
| TRANSFER_ERROR | Error durante transferencia |
| STORAGE_ERROR | Error de almacenamiento |
| INTERNAL_ERROR | Error interno |

Ejemplo:

```text
ERROR|FILE_NOT_FOUND|El archivo solicitado no existe
```

---

# 22. Estados de conexión

Cada cliente conectado puede encontrarse en uno de los siguientes estados:

```text
CONNECTED
    │
    ▼
AUTHENTICATING
    │
    ├── Error
    │     │
    │     ▼
    │  DISCONNECTED
    │
    ▼
AUTHENTICATED
    │
    ├── SYNCING
    │
    ├── TRANSFERRING
    │
    └── IDLE
          │
          ▼
     DISCONNECTED
```

Un cliente no autenticado no puede ejecutar operaciones como:

```text
SYNC
LIST
UPLOAD
DOWNLOAD
DELETE
```

---

# 23. Gestión de conexiones

El Sync Service debe mantener una estructura interna para registrar las conexiones activas.

Conceptualmente:

```text
Conexiones activas
│
├── Usuario A
│     ├── Desktop
│     └── Mobile
│
├── Usuario B
│     └── Desktop
│
└── Usuario C
      ├── Desktop
      └── Mobile
```

Una misma cuenta puede tener múltiples conexiones activas desde diferentes dispositivos.

El identificador:

```text
device_id
```

permite distinguir las conexiones asociadas al mismo usuario.

---

# 24. Prevención de reenvío innecesario

Cuando un cliente realiza una modificación, el Sync Service debe identificar el origen del cambio.

Por ejemplo:

```text
Cliente Desktop A
       │
       │ UPLOAD
       ▼
Sync Service
       │
       ├── Registrar modificación
       │
       ├── Notificar otros dispositivos
       │
       ▼
Cliente Mobile A
```

El cliente que originó la modificación no necesita recibir nuevamente la misma notificación como si fuera un cambio externo.

Conceptualmente:

```text
origen_device_id != destino_device_id
```

Esto evita que el cliente que generó la operación procese innecesariamente su propio cambio.

---

# 25. Relación con PostgreSQL

El Sync Service utiliza PostgreSQL para consultar y actualizar los metadatos necesarios para la sincronización.

Principalmente puede relacionarse con:

```text
usuario
archivo
```

El modelo es:

```text
Sync Service
      │
      ▼
PostgreSQL
      │
      ├── usuario
      │
      └── archivo
```

Los datos físicos no se almacenan directamente en PostgreSQL.

---

# 26. Relación con Shared Storage

Shared Storage almacena el contenido físico de los archivos.

El Sync Service interactúa con este componente durante:

```text
UPLOAD
DOWNLOAD
DELETE
```

Modelo:

```text
Sync Service
      │
      ▼
Shared Storage
      │
      ├── documentos
      ├── imagenes
      └── videos
```

La implementación definitiva del almacenamiento podrá evolucionar posteriormente hacia un sistema de almacenamiento distribuido.

---

# 27. Flujo completo de sincronización

```text
Cliente
   │
   │ TCP CONNECT
   ▼
Sync Service
   │
   │ Conexión establecida
   ▼
Cliente
   │
   │ AUTH|token|device_id
   ▼
Sync Service
   │
   │ AUTH_OK|id_usuario
   ▼
Cliente
   │
   │ SYNC
   ▼
Sync Service
   │
   │ Consultar archivos
   ▼
PostgreSQL
   │
   ▼
Sync Service
   │
   │ Determinar cambios
   ▼
Cliente
   │
   │ FILE_CHANGED
   │ FILE_DELETED
   │ ...
   ▼
SYNC_COMPLETE
```

---

# 28. Flujo completo de carga

```text
Cliente
   │
   │ UPLOAD|nombre|tipo|tamano
   ▼
Sync Service
   │
   │ Validar operación
   ▼
UPLOAD_READY|transfer_id
   │
   ▼
Cliente
   │
   │ Datos binarios
   ▼
Sync Service
   │
   │ Guardar contenido
   ▼
Shared Storage
   │
   │ Confirmación
   ▼
Sync Service
   │
   │ Registrar metadatos
   ▼
PostgreSQL
   │
   ▼
UPLOAD_COMPLETE|id_archivo
```

---

# 29. Flujo completo de descarga

```text
Cliente
   │
   │ DOWNLOAD|id_archivo
   ▼
Sync Service
   │
   ├── Validar usuario
   │
   ├── Consultar archivo
   │
   ▼
PostgreSQL
   │
   ▼
Sync Service
   │
   │ Recuperar archivo
   ▼
Shared Storage
   │
   ▼
DOWNLOAD_READY|nombre|tipo|tamano
   │
   ▼
Datos binarios
   │
   ▼
DOWNLOAD_COMPLETE|id_archivo
```

---

# 30. Reglas generales del protocolo

El protocolo debe cumplir las siguientes reglas:

1. La comunicación utiliza TCP Sockets.
2. El cliente debe autenticarse antes de realizar operaciones protegidas.
3. Cada conexión está asociada a un usuario autenticado.
4. Un usuario puede tener múltiples dispositivos conectados.
5. Los mensajes de control utilizan texto UTF-8.
6. Los campos de control utilizan `|` como delimitador.
7. La transferencia de archivos utiliza datos binarios.
8. El tamaño del archivo debe definirse antes de iniciar la transferencia.
9. El servidor debe validar la cantidad de bytes recibidos.
10. El acceso a los archivos debe validar la propiedad o permisos.
11. Los errores deben utilizar una estructura controlada.
12. El servidor puede enviar notificaciones a clientes conectados.
13. El dispositivo que origina un cambio no debe recibir innecesariamente su propia notificación.
14. Una conexión debe cerrarse de forma controlada cuando sea posible.

---

# 31. Estructura inicial propuesta

La implementación puede organizarse posteriormente de la siguiente forma:

```text
services/
└── sync-service/
    │
    ├── server/
    │   ├── SyncServer.java
    │   ├── ClientHandler.java
    │   ├── ConnectionManager.java
    │   └── ProtocolHandler.java
    │
    ├── protocol/
    │   ├── Command.java
    │   ├── MessageParser.java
    │   ├── MessageBuilder.java
    │   └── ProtocolConstants.java
    │
    ├── sync/
    │   ├── SyncManager.java
    │   ├── FileTransferManager.java
    │   └── ChangeNotifier.java
    │
    └── storage/
        ├── FileRepository.java
        └── SharedStorageClient.java
```

La estructura definitiva podrá modificarse según las decisiones tomadas durante la implementación.

---

# 32. Resumen

El Sync Service implementa el mecanismo de comunicación persistente de UPB-CIENTÍFICA para la sincronización de archivos.

```text
Cliente Desktop / Mobile
          │
          │ TCP Socket
          ▼
      Sync Service
          │
     ┌────┴─────┐
     │          │
     ▼          ▼
PostgreSQL  Shared Storage
```

El protocolo inicial define las operaciones:

```text
AUTH
SYNC
LIST
UPLOAD
DOWNLOAD
DELETE
PING
DISCONNECT
```

Las notificaciones principales son:

```text
FILE_CHANGED
FILE_DELETED
SYNC_COMPLETE
```

La transferencia de archivos se realiza separando los mensajes de control de los datos binarios.

Los mensajes de control siguen la estructura:

```text
OPERACION|PARAMETRO_1|PARAMETRO_2|...
```

Este documento constituye el contrato inicial de comunicación TCP para el Sync Service y servirá como base para la implementación del servidor, los clientes y las pruebas de interoperabilidad.