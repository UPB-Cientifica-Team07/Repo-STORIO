# Arquitectura de Comunicación Distribuida

## 1. Objetivo

Definir los mecanismos de comunicación entre los clientes, el API Gateway y los servicios distribuidos que componen UPB-CIENTÍFICA.

La arquitectura integra diferentes modelos de comunicación distribuida, incluyendo REST, Java RMI, TCP Sockets, SOAP, gRPC y OpenMPI.

---

## 2. Alcance

Este documento define:

- Los componentes que participan en la comunicación.
- Los protocolos utilizados entre los componentes.
- La dirección de los principales flujos de comunicación.
- La separación entre comunicación externa e interna.
- Las responsabilidades generales de cada servicio.

Los detalles específicos de endpoints, interfaces RMI, contratos SOAP, protocolos de sockets y archivos `.proto` se documentarán durante la implementación de cada componente.

---

## 3. Vista general de comunicación

El API Gateway funciona como punto principal de acceso para los clientes. Desde este componente, las solicitudes son dirigidas hacia el servicio correspondiente.

Los mecanismos definidos inicialmente son:

- HTTPS / REST para la comunicación general.
- Java RMI para autenticación.
- TCP Sockets para sincronización de archivos.
- SOAP para streaming.
- gRPC para monitoreo interno.
- OpenMPI para procesamiento paralelo en el Cluster HPC.

```text
Cliente Web
Cliente Desktop
Cliente Mobile
        │
        │ HTTPS / REST
        ▼
   API Gateway
        │
        ├── REST ─────────────── File Service
        │
        ├── REST ─────────────── Photo Service
        │
        ├── REST ─────────────── HPC Service
        │
        ├── REST ─────────────── Monitoring Service
        │
        ├── Java RMI ─────────── Auth Service
        │
        └── SOAP ─────────────── Streaming Service


Cliente Desktop / Mobile
        │
        │ TCP Socket
        ▼
    Sync Service


Servicios / Nodos
        │
        │ gRPC
        ▼
Monitoring Service


HPC Service
        │
        │ OpenMPI
        ▼
   Cluster HPC
```

---

## 4. Componentes de comunicación

Los componentes principales de la arquitectura son:

- Cliente Web.
- Cliente Desktop.
- Cliente Mobile.
- API Gateway.
- Auth Service.
- File Service.
- Sync Service.
- Photo Service.
- Streaming Service.
- HPC Service.
- Monitoring Service.
- PostgreSQL.
- Shared Storage.
- Cluster HPC.

Cada componente utiliza el mecanismo de comunicación definido según su responsabilidad dentro del sistema.

---

### 4.1 Clientes

El sistema contará con tres tipos de clientes:

- Cliente Web.
- Cliente Desktop.
- Cliente Mobile.

Los clientes representan la capa de interacción con los usuarios.

La comunicación principal con el sistema se realizará mediante HTTPS y REST a través del API Gateway.

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
```

Mediante este mecanismo, los clientes podrán acceder a funcionalidades relacionadas con:

- Autenticación.
- Gestión de archivos.
- Gestión de imágenes.
- Contenido multimedia.
- Trabajos HPC.
- Monitoreo del sistema.

El Cliente Desktop y el Cliente Mobile utilizarán adicionalmente TCP Sockets para los procesos de sincronización de archivos.

```text
Cliente Desktop / Mobile
           │
           │ TCP Socket
           ▼
      Sync Service
```

Por tanto, la comunicación de los clientes se divide en:

| Comunicación | Tecnología | Propósito |
|---|---|---|
| General | HTTPS / REST | Acceso a los servicios |
| Sincronización | TCP Sockets | Sincronización de archivos |

---

### 4.2 API Gateway

El API Gateway es el punto principal de entrada a la plataforma.

Su función es recibir las solicitudes de los clientes y dirigirlas hacia el servicio responsable de procesarlas.

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   ▼
Servicio correspondiente
```

Sus responsabilidades principales son:

- Centralizar el acceso a los servicios.
- Recibir solicitudes externas.
- Identificar el servicio correspondiente.
- Redirigir la comunicación.
- Separar la comunicación externa de la arquitectura interna.

El API Gateway no contiene la lógica específica de cada servicio.

Los servicios accesibles mediante el Gateway son:

- Auth Service.
- File Service.
- Photo Service.
- Streaming Service.
- HPC Service.
- Monitoring Service.

El Sync Service utiliza una comunicación independiente mediante TCP Sockets debido a la necesidad de mantener una conexión durante el proceso de sincronización.

---

### 4.3 Servicios distribuidos

La plataforma está compuesta por servicios con responsabilidades específicas.

#### Auth Service

Responsable de:

- Validar credenciales.
- Autenticar usuarios.
- Gestionar identidad.
- Validar tokens o sesiones.

La comunicación con este servicio se realizará mediante Java RMI.

```text
API Gateway
     │
     │ Java RMI
     ▼
Auth Service
```

---

#### File Service

Responsable de la gestión de archivos y sus metadatos.

Sus funciones principales son:

- Registrar archivos.
- Consultar archivos.
- Cargar archivos.
- Descargar archivos.
- Eliminar archivos.

El servicio utiliza PostgreSQL para los metadatos y Shared Storage para el almacenamiento físico.

```text
API Gateway
     │
     │ REST
     ▼
File Service
     │
     ├── PostgreSQL
     │
     └── Shared Storage
```

---

#### Sync Service

Responsable de la sincronización de archivos entre los clientes y el sistema.

Sus funciones principales son:

- Mantener la conexión durante la sincronización.
- Recibir cambios desde los clientes.
- Enviar actualizaciones.
- Coordinar la información con el almacenamiento.

La comunicación se realizará mediante TCP Sockets.

```text
Cliente Desktop / Mobile
           │
           │ TCP Socket
           ▼
      Sync Service
```

---

#### Photo Service

Responsable de las funcionalidades relacionadas con imágenes.

Sus funciones incluyen:

- Consultar imágenes.
- Gestionar metadatos.
- Organizar fotografías y álbumes.

La comunicación se realizará mediante REST.

```text
API Gateway
     │
     │ REST
     ▼
Photo Service
```

El servicio utilizará PostgreSQL para los metadatos y podrá acceder al almacenamiento de archivos cuando sea necesario.

---

#### Streaming Service

Responsable de la gestión del contenido multimedia.

Sus funciones incluyen:

- Consultar videos.
- Obtener información multimedia.
- Gestionar solicitudes relacionadas con reproducción.

La comunicación se realizará mediante SOAP.

```text
API Gateway
     │
     │ SOAP
     ▼
Streaming Service
```

La implementación propuesta será realizada en PHP.

---

#### HPC Service

Responsable de gestionar los trabajos de computación de alto rendimiento.

Sus funciones incluyen:

- Recibir trabajos HPC.
- Registrar su estado.
- Coordinar recursos.
- Consultar resultados.
- Gestionar la ejecución en el Cluster HPC.

La comunicación con el API Gateway se realizará mediante REST.

```text
API Gateway
     │
     │ REST
     ▼
HPC Service
     │
     ▼
Cluster HPC
```

La ejecución paralela dentro del clúster utilizará OpenMPI.

---

#### Monitoring Service

Responsable de recopilar información sobre el estado de los servicios y nodos.

Sus funciones incluyen:

- Consultar disponibilidad de servicios.
- Obtener métricas de CPU y memoria.
- Supervisar nodos HPC.
- Consolidar información del sistema.

La comunicación externa se realizará mediante REST a través del API Gateway.

La comunicación interna con servicios y nodos utilizará gRPC.

```text
Servicios / Nodos
       │
       │ gRPC
       ▼
Monitoring Service
       │
       │ REST
       ▼
API Gateway
```

La implementación propuesta para este servicio será realizada en Go.

---

### Resumen de servicios

| Servicio | Responsabilidad | Comunicación |
|---|---|---|
| Auth Service | Autenticación | Java RMI |
| File Service | Gestión de archivos | REST |
| Sync Service | Sincronización | TCP Sockets |
| Photo Service | Gestión de imágenes | REST |
| Streaming Service | Gestión multimedia | SOAP |
| HPC Service | Gestión de trabajos HPC | REST / OpenMPI |
| Monitoring Service | Monitoreo | REST / gRPC |


---

## 5. Comunicación Cliente → API Gateway

La comunicación principal entre los clientes y UPB-CIENTÍFICA se realiza a través del API Gateway.

Los clientes no se comunican directamente con los servicios internos. Todas las solicitudes generales ingresan inicialmente al Gateway, el cual identifica el servicio responsable y dirige la solicitud utilizando el mecanismo de comunicación correspondiente.

### 5.1 Protocolo

La comunicación entre los clientes y el API Gateway utiliza:

- HTTPS como protocolo de transporte seguro.
- HTTP como protocolo de comunicación.
- REST como estilo arquitectónico para la exposición de los recursos.

El flujo general es:

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
```

Este mecanismo será utilizado por los siguientes clientes:

- Cliente Web.
- Cliente Desktop.
- Cliente Mobile.

---

### 5.2 Solicitudes al Gateway

Las solicitudes realizadas por los clientes estarán relacionadas con las funcionalidades principales del sistema.

Inicialmente, se contemplan las siguientes categorías:

| Funcionalidad | Ejemplo de ruta |
|---|---|
| Autenticación | `/auth/login` |
| Usuarios | `/users` |
| Archivos | `/files` |
| Imágenes | `/photos` |
| Streaming | `/streaming` |
| Trabajos HPC | `/hpc/jobs` |
| Monitoreo | `/monitoring` |

Estas rutas representan una organización inicial de la API. La definición completa de endpoints, métodos HTTP, parámetros y estructuras de respuesta será documentada posteriormente.

---

### 5.3 Flujo de una solicitud

El flujo general de una solicitud es el siguiente:

```text
1\. El cliente realiza una solicitud.
           │
           ▼
2\. La solicitud llega al API Gateway.
           │
           ▼
3\. El Gateway identifica el recurso solicitado.
           │
           ▼
4\. El Gateway dirige la solicitud al servicio correspondiente.
           │
           ▼
5\. El servicio procesa la solicitud.
           │
           ▼
6\. El servicio genera una respuesta.
           │
           ▼
7\. La respuesta retorna al API Gateway.
           │
           ▼
8\. El Gateway responde al cliente.
```

Representación simplificada:

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ Comunicación interna
   ▼
Servicio
   │
   │ Respuesta
   ▼
API Gateway
   │
   ▼
Cliente
```

---

### 5.4 Ejemplo de comunicación

Por ejemplo, un cliente puede solicitar la consulta de sus archivos:

```text
GET /files
```

El flujo sería:

```text
Cliente
   │
   │ GET /files
   ▼
API Gateway
   │
   │ REST
   ▼
File Service
   │
   ▼
PostgreSQL / Shared Storage
```

Una vez procesada la solicitud, la respuesta sigue el camino inverso:

```text
File Service
      │
      │ Respuesta
      ▼
API Gateway
      │
      │ HTTPS
      ▼
Cliente
```

---

### 5.5 Responsabilidad del API Gateway

Dentro de este flujo, el API Gateway es responsable de:

- Recibir las solicitudes externas.
- Identificar el servicio responsable.
- Redirigir la solicitud.
- Retornar la respuesta al cliente.
- Mantener un punto centralizado de acceso.

El Gateway no debe implementar la lógica de negocio específica de servicios como autenticación, archivos, HPC o monitoreo.

Por ejemplo:

```text
Solicitud: POST /hpc/jobs

Cliente
   │
   ▼
API Gateway
   │
   │ REST
   ▼
HPC Service
```

La lógica para crear, gestionar y ejecutar el trabajo corresponde al HPC Service.

---

### 5.6 Excepción: Sync Service

La comunicación mediante API Gateway corresponde al flujo general del sistema.

El Sync Service constituye una excepción, debido a que los procesos de sincronización requieren una comunicación persistente.

Por esta razón, el Cliente Desktop y el Cliente Mobile podrán establecer una conexión directa mediante TCP Socket:

```text
Cliente Desktop / Mobile
           │
           │ TCP Socket
           ▼
      Sync Service
```

Esta conexión estará limitada a las operaciones de sincronización y utilizará un protocolo de mensajes definido específicamente para este servicio.

---

### 5.7 Resumen de la comunicación externa

| Origen | Destino | Protocolo | Propósito |
|---|---|---|---|
| Cliente Web | API Gateway | HTTPS / REST | Acceso a la plataforma |
| Cliente Desktop | API Gateway | HTTPS / REST | Acceso a servicios |
| Cliente Mobile | API Gateway | HTTPS / REST | Acceso a servicios |
| Cliente Desktop | Sync Service | TCP Socket | Sincronización |
| Cliente Mobile | Sync Service | TCP Socket | Sincronización |

La comunicación mediante HTTPS y REST constituye el mecanismo principal de acceso a UPB-CIENTÍFICA, mientras que TCP Sockets se utiliza únicamente para los procesos que requieren una conexión persistente durante la sincronización de archivos.


---

## 6. Comunicación API Gateway → Auth Service

El Auth Service es responsable de las operaciones relacionadas con la autenticación e identidad de los usuarios.

A diferencia de los demás servicios principales, la comunicación entre el API Gateway y el Auth Service se realizará mediante **\*\*Java RMI (Remote Method Invocation)\*\***.

Este mecanismo permite que el API Gateway invoque métodos implementados en un objeto remoto ubicado en el Auth Service.

### 6.1 Modelo de comunicación

La comunicación sigue el siguiente flujo:

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ Java RMI
   ▼
Auth Service
```

El cliente no interactúa directamente con el objeto remoto. La comunicación mediante RMI permanece dentro de la arquitectura interna.

---

### 6.2 Funcionamiento general

El proceso de autenticación puede seguir el siguiente flujo:

```text
1\. El cliente envía sus credenciales al API Gateway.
           │
           ▼
2\. El API Gateway recibe la solicitud.
           │
           ▼
3\. El Gateway invoca un método remoto del Auth Service.
           │
           ▼
4\. El Auth Service procesa la solicitud.
           │
           ▼
5\. El Auth Service retorna el resultado.
           │
           ▼
6\. El API Gateway genera la respuesta para el cliente.
```

Representación simplificada:

```text
Cliente
   │
   │ POST /auth/login
   ▼
API Gateway
   │
   │ login(...)
   │ Java RMI
   ▼
Auth Service
   │
   ▼
PostgreSQL
```

---

### 6.3 Operaciones iniciales

Inicialmente, el Auth Service deberá soportar operaciones relacionadas con:

- Autenticación de usuarios.
- Validación de credenciales.
- Consulta de información básica del usuario.
- Validación de tokens.
- Gestión de acceso.

Estas operaciones serán definidas mediante una interfaz remota de Java.

De forma conceptual:

```text
AuthService
│
├── login(...)
├── validarToken(...)
├── obtenerUsuario(...)
└── validarCredenciales(...)
```

La definición definitiva de los métodos dependerá del contrato que se establezca durante la implementación.

---

### 6.4 Interfaz remota

La comunicación mediante RMI requiere una interfaz compartida que defina las operaciones disponibles en el servicio remoto.

Conceptualmente, la estructura será:

```text
API Gateway
      │
      │ Utiliza
      ▼
Interfaz Remota
      │
      │ Implementada por
      ▼
Auth Service
```

La interfaz remota actuará como contrato entre el API Gateway y el Auth Service.

Esto permite que el Gateway conozca únicamente las operaciones disponibles, sin depender directamente de la implementación interna del servicio.

La interfaz y los modelos compartidos deberán mantenerse separados de la lógica de implementación.

Una estructura inicial podría ser:

```text
services/
└── auth-service/
    ├── contract/
    ├── server/
    └── client/
```

El directorio `contract/` contendrá las interfaces y objetos necesarios para la comunicación remota.

---

### 6.5 Acceso a la base de datos

El Auth Service será el componente encargado de consultar la información relacionada con los usuarios.

El flujo será:

```text
API Gateway
      │
      │ Java RMI
      ▼
Auth Service
      │
      │ SQL
      ▼
PostgreSQL
```

La tabla principal involucrada será:

```text
usuario
```

El servicio podrá consultar información como:

- Identificador del usuario.
- Nombre de usuario.
- Correo electrónico.
- Rol.
- Estado.
- `password_hash`.

El campo `password_hash` no debe ser expuesto a otros servicios ni a los clientes como parte de una respuesta.

---

### 6.6 Resultado de autenticación

Una operación de autenticación debe generar un resultado que permita al API Gateway determinar si el usuario puede acceder al sistema.

Conceptualmente:

```text
Credenciales
     │
     ▼
Auth Service
     │
     ├── Válidas ──────► Usuario autenticado
     │
     └── Inválidas ────► Error de autenticación
```

El resultado podrá incluir información como:

- Estado de autenticación.
- Identificador del usuario.
- Nombre de usuario.
- Rol.
- Token de acceso, cuando corresponda.

La estructura definitiva del resultado será definida posteriormente como parte del contrato de autenticación.

---

### 6.7 Manejo general de errores

Los errores internos de la comunicación RMI deben ser controlados por el API Gateway antes de generar una respuesta hacia el cliente.

Se consideran inicialmente los siguientes escenarios:

| Situación | Resultado esperado |
|---|---|
| Usuario inexistente | Error de autenticación |
| Credenciales inválidas | Error de autenticación |
| Usuario inactivo | Acceso denegado |
| Auth Service no disponible | Error interno del sistema |
| Error de comunicación RMI | Error interno del sistema |

El cliente recibirá una respuesta mediante HTTPS, sin depender directamente de las excepciones internas generadas por Java RMI.

---

### 6.8 Regla de comunicación

La comunicación de autenticación seguirá esta separación:

```text
Comunicación externa

Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway


Comunicación interna

API Gateway
   │
   │ Java RMI
   ▼
Auth Service
```

De esta manera, REST se utiliza como mecanismo de acceso externo, mientras que Java RMI implementa la comunicación interna basada en objetos distribuidos.


---

## 7. Comunicación API Gateway → File Service

El File Service es responsable de gestionar las operaciones relacionadas con los archivos almacenados en UPB-CIENTÍFICA.

La comunicación entre el API Gateway y este servicio se realizará mediante REST sobre HTTP.

### 7.1 Modelo de comunicación

El flujo general será:

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ REST / HTTP
   ▼
File Service
```

El API Gateway recibe la solicitud externa y la dirige al File Service, que se encarga de procesar la operación correspondiente.

---

### 7.2 Operaciones principales

Inicialmente, el File Service deberá gestionar operaciones relacionadas con:

- Carga de archivos.
- Consulta de archivos.
- Descarga de archivos.
- Eliminación de archivos.
- Consulta de metadatos.

De forma conceptual, las operaciones estarán relacionadas con recursos como:

```text
/files
/files/{id}
/files/{id}/download
```

La definición final de rutas, métodos HTTP y estructuras de solicitud y respuesta se documentará posteriormente como parte del contrato del servicio.

---

### 7.3 Gestión de metadatos y almacenamiento

La gestión de un archivo se divide en dos partes:

1\. Metadatos almacenados en PostgreSQL.
2\. Contenido físico almacenado en Shared Storage.

El flujo será:

```text
File Service
     │
     ├── PostgreSQL
     │       │
     │       └── Metadatos
     │
     └── Shared Storage
             │
             └── Archivo físico
```

Los metadatos principales están relacionados con la entidad `archivo` definida en el modelo de datos.

Entre ellos se encuentran:

- Identificador del archivo.
- Nombre.
- Tipo de archivo.
- Tamaño.
- Propietario.
- Fecha de creación o carga.
- Ubicación o referencia del archivo.

El contenido físico no debe almacenarse directamente dentro de la base de datos.

---

### 7.4 Flujo de carga de archivos

El proceso general para cargar un archivo será:

```text
1\. El cliente selecciona un archivo.
           │
           ▼
2\. El cliente realiza una solicitud al API Gateway.
           │
           ▼
3\. El Gateway dirige la solicitud al File Service.
           │
           ▼
4\. El File Service valida la solicitud.
           │
           ▼
5\. El archivo se almacena en Shared Storage.
           │
           ▼
6\. Se registran los metadatos en PostgreSQL.
           │
           ▼
7\. Se genera la respuesta.
           │
           ▼
8\. La respuesta retorna al cliente.
```

Representación simplificada:

```text
Cliente
   │
   │ POST /files
   ▼
API Gateway
   │
   │ REST
   ▼
File Service
   │
   ├──────────► Shared Storage
   │
   └──────────► PostgreSQL
```

---

### 7.5 Flujo de consulta y descarga

Para consultar un archivo, el servicio deberá utilizar inicialmente los metadatos almacenados en PostgreSQL.

```text
Cliente
   │
   │ GET /files/{id}
   ▼
API Gateway
   │
   ▼
File Service
   │
   ▼
PostgreSQL
```

Cuando se solicite la descarga del contenido:

```text
Cliente
   │
   │ Solicitud de descarga
   ▼
API Gateway
   │
   ▼
File Service
   │
   ├── Consulta metadatos ──► PostgreSQL
   │
   └── Obtiene archivo ─────► Shared Storage
```

El File Service será responsable de localizar el recurso solicitado y generar la respuesta correspondiente.

---

### 7.6 Relación con otros servicios

El File Service puede ser utilizado por otros componentes que requieran acceder a archivos almacenados.

Inicialmente, los servicios relacionados son:

- Sync Service.
- Photo Service.
- Streaming Service.

Estos servicios podrán consultar o utilizar información relacionada con los archivos sin asumir directamente la responsabilidad sobre la gestión general del almacenamiento.

La comunicación específica entre estos servicios será definida durante la implementación.

---

### 7.7 Manejo general de errores

Se consideran inicialmente los siguientes escenarios:

| Situación | Resultado esperado |
|---|---|
| Archivo inexistente | Recurso no encontrado |
| Solicitud inválida | Error de validación |
| Archivo demasiado grande | Solicitud rechazada |
| Error de almacenamiento | Error interno |
| Error de base de datos | Error interno |
| Usuario sin permisos | Acceso denegado |

Cuando corresponda, estos resultados serán representados mediante códigos de estado HTTP.

Ejemplos:

```text
200 OK
201 Created
400 Bad Request
403 Forbidden
404 Not Found
500 Internal Server Error
```

---

### 7.8 Regla de comunicación

La gestión de archivos seguirá la siguiente separación:

```text
Comunicación externa

Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway


Comunicación interna

API Gateway
   │
   │ REST / HTTP
   ▼
File Service
   │
   ├── PostgreSQL
   │
   └── Shared Storage
```

El API Gateway actúa como punto de entrada, mientras que el File Service concentra la lógica relacionada con la gestión de archivos, sus metadatos y su almacenamiento.


---

## 8. Comunicación Cliente → Sync Service

El Sync Service es responsable de gestionar la sincronización de archivos entre los clientes y la plataforma.

A diferencia de los servicios accesibles mediante solicitudes REST, este componente utiliza una comunicación basada en TCP Sockets, permitiendo mantener una conexión durante el proceso de sincronización.

### 8.1 Modelo de comunicación

El flujo general será:

```text
Cliente Desktop / Mobile
           │
           │ TCP Socket
           ▼
      Sync Service
           │
           ▼
   Shared Storage
```

El Cliente Web no participa inicialmente en este mecanismo de sincronización.

---

### 8.2 Conexión persistente

La comunicación mediante TCP permite establecer una conexión entre el cliente y el Sync Service.

Esta conexión puede utilizarse para:

- Iniciar un proceso de sincronización.
- Enviar información sobre archivos modificados.
- Transferir cambios.
- Recibir actualizaciones.
- Confirmar el resultado de una operación.

El flujo general de conexión será:

```text
Cliente
   │
   │ Solicitud de conexión TCP
   ▼
Sync Service
   │
   │ Conexión establecida
   ▼
Proceso de sincronización
```

La conexión permanece activa mientras se desarrolla la operación y puede cerrarse cuando el proceso finalice.

---

### 8.3 Flujo general de sincronización

El proceso inicial de sincronización puede seguir la siguiente secuencia:

```text
1\. El cliente establece conexión con el Sync Service.
           │
           ▼
2\. El cliente identifica los archivos o cambios locales.
           │
           ▼
3\. El cliente envía la información de sincronización.
           │
           ▼
4\. El Sync Service procesa la solicitud.
           │
           ▼
5\. Se determina si existen archivos nuevos o modificados.
           │
           ▼
6\. Se actualiza la información correspondiente.
           │
           ▼
7\. El servidor responde con el resultado.
           │
           ▼
8\. El cliente confirma la sincronización.
```

Representación simplificada:

```text
Cliente
   │
   │ TCP
   ▼
Sync Service
   │
   ├── Actualización
   │
   ├── Consulta
   │
   ▼
Shared Storage / File Service
```

---

### 8.4 Protocolo de mensajes

La comunicación mediante TCP requiere definir un protocolo de mensajes propio.

Inicialmente, los mensajes podrán representar operaciones como:

```text
SYNC
UPLOAD
DOWNLOAD
UPDATE
DELETE
STATUS
```

Cada mensaje deberá contener la información necesaria para identificar la operación y el recurso involucrado.

Una estructura conceptual podría ser:

```text
OPERACION|IDENTIFICADOR|DATOS
```

Ejemplos:

```text
SYNC|archivo-001|...
UPLOAD|archivo-002|...
DELETE|archivo-003
STATUS|archivo-001
```

La estructura definitiva deberá definirse antes de implementar la comunicación completa.

El contrato del protocolo se documentará en:

```text
common/protocol/sync-protocol.md
```

---

### 8.5 Relación con File Service

El Sync Service es responsable del proceso de sincronización, pero la gestión centralizada de los metadatos continúa siendo responsabilidad del File Service.

Por esta razón, durante una operación de sincronización pueden intervenir ambos servicios:

```text
Cliente
   │
   │ TCP Socket
   ▼
Sync Service
   │
   ├── Actualización física
   │
   ▼
Shared Storage

Sync Service
   │
   │ Comunicación interna
   ▼
File Service
   │
   ▼
PostgreSQL
```

El mecanismo específico de comunicación entre Sync Service y File Service se definirá durante la implementación.

---

### 8.6 Manejo general de errores

El protocolo deberá contemplar respuestas para situaciones como:

| Situación | Resultado esperado |
|---|---|
| Archivo inexistente | Error de recurso |
| Operación no válida | Error de protocolo |
| Datos incompletos | Solicitud rechazada |
| Error de almacenamiento | Error interno |
| Conexión interrumpida | Sincronización cancelada o reintentada |
| Usuario no autorizado | Acceso denegado |

Las respuestas deberán seguir un formato consistente.

Ejemplo conceptual:

```text
OK|SYNC_COMPLETED
OK|FILE_UPDATED
ERROR|FILE_NOT_FOUND
ERROR|INVALID_OPERATION
ERROR|UNAUTHORIZED
```

Los códigos y mensajes definitivos serán definidos dentro del contrato del protocolo.

---

### 8.7 Regla de comunicación

La sincronización utiliza un mecanismo independiente del flujo REST general.

```text
Comunicación general

Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway


Comunicación de sincronización

Cliente Desktop / Mobile
           │
           │ TCP Socket
           ▼
      Sync Service
```

El API Gateway continúa siendo el punto principal de acceso para las funcionalidades generales del sistema, mientras que el Sync Service utiliza TCP Sockets exclusivamente para las operaciones de sincronización de archivos.

---

## 9. Comunicación API Gateway → Photo Service

El Photo Service es responsable de gestionar las operaciones relacionadas con imágenes y álbumes dentro de UPB-CIENTÍFICA.

La comunicación entre el API Gateway y el Photo Service se realizará mediante REST sobre HTTP.

### 9.1 Modelo de comunicación

El flujo general será:

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ REST / HTTP
   ▼
Photo Service
```

El API Gateway recibe la solicitud del cliente y la dirige al Photo Service para su procesamiento.

---

### 9.2 Operaciones principales

Inicialmente, el Photo Service gestionará operaciones relacionadas con:

- Consulta de imágenes.
- Consulta de información de una imagen.
- Organización de imágenes.
- Gestión de álbumes.
- Consulta de metadatos.

De forma conceptual, las operaciones estarán asociadas a recursos como:

```text
/photos
/photos/{id}
/albums
/albums/{id}
```

La definición definitiva de las rutas, métodos HTTP y estructuras de datos será documentada posteriormente dentro del contrato del servicio.

---

### 9.3 Gestión de imágenes

La gestión de imágenes se divide entre la información registrada en la base de datos y el archivo físico almacenado en el sistema.

El flujo conceptual será:

```text
Photo Service
      │
      ├── PostgreSQL
      │       │
      │       └── Metadatos
      │
      └── File Service / Shared Storage
              │
              └── Imagen física
```

La información relacionada con una imagen puede incluir:

- Identificador.
- Archivo asociado.
- Resolución.
- Formato.
- Propietario.
- Fecha de registro.

El contenido físico de la imagen se mantiene separado de sus metadatos.

---

### 9.4 Flujo de consulta

Cuando un cliente solicita información sobre una imagen, el flujo será:

```text
Cliente
   │
   │ GET /photos/{id}
   ▼
API Gateway
   │
   │ REST
   ▼
Photo Service
   │
   ▼
PostgreSQL
```

El Photo Service consulta los metadatos correspondientes y genera una respuesta que retorna a través del API Gateway.

```text
Photo Service
      │
      ▼
API Gateway
      │
      ▼
Cliente
```

---

### 9.5 Acceso al archivo físico

Cuando una operación requiera acceder al contenido físico de una imagen, el Photo Service deberá utilizar la información asociada al archivo.

El flujo conceptual será:

```text
Photo Service
      │
      │ Consulta información del archivo
      ▼
File Service
      │
      ▼
Shared Storage
```

De esta forma, el Photo Service mantiene la responsabilidad sobre las funcionalidades relacionadas con imágenes, mientras que el File Service continúa siendo responsable de la gestión general de los archivos.

---

### 9.6 Relación con PostgreSQL

El Photo Service utilizará PostgreSQL para consultar y gestionar los metadatos asociados a las imágenes.

La entidad principal relacionada inicialmente es:

```text
imagen
```

Esta entidad se encuentra asociada a la entidad:

```text
archivo
```

La relación conceptual es:

```text
Archivo
   │
   │ 1
   │
   │ 1
   ▼
Imagen
```

Esto permite mantener separados los datos generales del archivo de la información específica correspondiente a una imagen.

---

### 9.7 Manejo general de errores

Se consideran inicialmente los siguientes escenarios:

| Situación | Resultado esperado |
|---|---|
| Imagen inexistente | Recurso no encontrado |
| Solicitud inválida | Error de validación |
| Archivo asociado inexistente | Error interno o inconsistencia |
| Usuario sin permisos | Acceso denegado |
| Error de base de datos | Error interno |

Cuando corresponda, estos resultados serán representados mediante códigos de estado HTTP.

Ejemplos:

```text
200 OK
400 Bad Request
403 Forbidden
404 Not Found
500 Internal Server Error
```

---

### 9.8 Regla de comunicación

La comunicación seguirá la siguiente estructura:

```text
Comunicación externa

Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway


Comunicación interna

API Gateway
   │
   │ REST / HTTP
   ▼
Photo Service
   │
   ├── PostgreSQL
   │
   └── File Service / Shared Storage
```

El API Gateway centraliza el acceso externo, mientras que el Photo Service gestiona la información específica de las imágenes y utiliza los componentes de almacenamiento definidos por la arquitectura.
---

## 10. Comunicación API Gateway → Streaming Service

El Streaming Service gestiona las operaciones relacionadas con el contenido de video. La integración interna propuesta utiliza SOAP, mientras que el cliente continúa accediendo a la plataforma mediante HTTPS/REST a través del API Gateway.

### 10.1 Modelo de comunicación

```text
Cliente
   │ HTTPS / REST
   ▼
API Gateway
   │ SOAP
   ▼
Streaming Service
```

### 10.2 Operaciones principales

Inicialmente se contemplan:

- Consulta de videos.
- Consulta de metadatos.
- Obtención de información para reproducción.
- Consulta de disponibilidad del recurso.

La interfaz SOAP y sus operaciones se definirán en el contrato del servicio.

### 10.3 Acceso al contenido

El Streaming Service podrá consultar los metadatos del video y acceder al recurso físico mediante los componentes de almacenamiento definidos.

```text
Streaming Service
      │
      ├── PostgreSQL ──► Metadatos
      │
      └── Shared Storage ──► Video
```

### 10.4 Regla de comunicación

SOAP se utiliza únicamente en la comunicación interna definida entre el API Gateway y el Streaming Service. El cliente no consume directamente el servicio SOAP.

---

## 11. Comunicación API Gateway → HPC Service

El HPC Service recibe y administra solicitudes relacionadas con trabajos de computación de alto rendimiento.

### 11.1 Modelo de comunicación

```text
Cliente
   │ HTTPS / REST
   ▼
API Gateway
   │ REST / HTTP
   ▼
HPC Service
   │
   ▼
Cluster HPC
```

### 11.2 Operaciones principales

- Crear un trabajo HPC.
- Consultar su estado.
- Consultar resultados.
- Cancelar un trabajo cuando la operación lo permita.
- Asignar o registrar el nodo asociado.

### 11.3 Flujo de ejecución

```text
Cliente
   │ POST /hpc/jobs
   ▼
API Gateway
   ▼
HPC Service
   │
   ├── Registra el trabajo ──► PostgreSQL
   │
   └── Envía ejecución ──────► Cluster HPC
```

El estado del trabajo puede pasar, inicialmente, por los valores `PENDIENTE`, `EJECUTANDO`, `FINALIZADO` o `ERROR`.

### 11.4 Regla de comunicación

El API Gateway no ejecuta procesos paralelos. La gestión de la solicitud corresponde al HPC Service y la ejecución distribuida corresponde al Cluster HPC.

---

## 12. Comunicación HPC Service → Cluster HPC

La comunicación entre el HPC Service y el Cluster HPC coordina la ejecución de trabajos distribuidos.

### 12.1 Responsabilidades

El HPC Service debe:

- Recibir el trabajo.
- Validar la solicitud.
- Registrar su estado.
- Coordinar su envío al entorno de ejecución.
- Recibir o registrar el resultado final.

### 12.2 Flujo general

```text
HPC Service
     │
     │ Trabajo HPC
     ▼
Cluster HPC
     │
     ▼
Proceso paralelo
```

La implementación concreta del mecanismo de envío y planificación se definirá durante la construcción del servicio.

---

## 13. Comunicación dentro del Cluster HPC

Dentro del Cluster HPC, los procesos paralelos se comunican utilizando OpenMPI.

### 13.1 Modelo

```text
Nodo coordinador
      │
      ├──── MPI ────► Nodo HPC 1
      │
      ├──── MPI ────► Nodo HPC 2
      │
      └──── MPI ────► Nodo HPC N
```

### 13.2 Responsabilidad

OpenMPI se utiliza para la comunicación y coordinación entre procesos participantes en un trabajo paralelo.

Este mecanismo no reemplaza REST, gRPC o RMI; su ámbito está limitado a la ejecución distribuida dentro del entorno HPC.

---

## 14. Comunicación API Gateway → Monitoring Service

El Monitoring Service expone información consolidada sobre el estado de los servicios y nodos.

### 14.1 Modelo de comunicación

```text
Cliente
   │ HTTPS / REST
   ▼
API Gateway
   │ REST / HTTP
   ▼
Monitoring Service
```

### 14.2 Información inicial

El servicio podrá proporcionar información relacionada con:

- Disponibilidad.
- Estado de servicios.
- Estado de nodos HPC.
- CPU.
- Memoria.
- Métricas definidas durante la implementación.

### 14.3 Regla de comunicación

El acceso externo al monitoreo se centraliza en el API Gateway. La recolección interna de métricas se documenta mediante gRPC en la siguiente sección.

---

## 15. Comunicación Monitoring Service → Servicios y Nodos

La comunicación interna para obtener métricas utilizará gRPC.

### 15.1 Modelo

```text
Servicios / Nodos
       │
       │ gRPC
       ▼
Monitoring Service
```

### 15.2 Responsabilidad

Cada servicio o nodo que participe en el monitoreo deberá exponer la información definida por el contrato gRPC correspondiente.

El contrato deberá definir:

- Servicios RPC disponibles.
- Mensajes de solicitud.
- Mensajes de respuesta.
- Tipos de métricas.
- Manejo de errores.

### 15.3 Archivo de contrato

La definición se almacenará inicialmente en una ubicación equivalente a:

```text
common/protocol/monitoring.proto
```

Las estructuras definitivas dependerán de la implementación en Go y de los servicios que publiquen métricas.

---

## 16. Comunicación con PostgreSQL

PostgreSQL almacena la información persistente definida en el modelo de datos.

### 16.1 Servicios relacionados

Inicialmente, los siguientes servicios requieren acceso a información persistente:

- Auth Service.
- File Service.
- Photo Service.
- Streaming Service.
- HPC Service.
- Monitoring Service, cuando requiera persistir información histórica.

### 16.2 Regla de acceso

Cada servicio debe acceder únicamente a los datos necesarios para su responsabilidad.

```text
Servicio
   │
   │ SQL
   ▼
PostgreSQL
```

Las credenciales de conexión no deben almacenarse directamente en el código fuente.

### 16.3 Entidades iniciales

El esquema actual contempla, entre otras, las entidades:

- `usuario`
- `archivo`
- `imagen`
- `video`
- `trabajo_hpc`
- `nodo_hpc`
- `token`

---

## 17. Comunicación con Shared Storage

El Shared Storage contiene los archivos físicos utilizados por la plataforma.

### 17.1 Separación de responsabilidades

```text
PostgreSQL
    │
    └── Metadatos

Shared Storage
    │
    └── Contenido físico
```

El File Service mantiene la responsabilidad principal sobre la gestión de archivos.

Photo Service, Streaming Service y Sync Service pueden requerir acceso controlado al almacenamiento según su funcionalidad.

### 17.2 Regla general

La ubicación física de un archivo no debe ser tratada como un detalle expuesto directamente a los clientes. Los servicios son responsables de localizar y gestionar el recurso.

---

## 18. Contratos de comunicación

Cada mecanismo distribuido requiere un contrato explícito.

| Comunicación | Contrato principal |
|---|---|
| REST | Endpoints, métodos y modelos |
| Java RMI | Interfaz remota y objetos transferidos |
| TCP Socket | Formato y tipos de mensajes |
| SOAP | Operaciones y contrato del servicio |
| gRPC | Archivo `.proto` |
| OpenMPI | Protocolo interno del programa paralelo |

Los contratos deben mantenerse versionados junto con el proyecto.

### 18.1 Ubicación inicial

```text
docs/
└── api/

common/
└── protocol/
```

La documentación pública y los contratos compartidos deben permanecer separados de la lógica específica de cada servicio.

---

## 19. Manejo general de errores

Los errores internos deben transformarse en respuestas adecuadas para el mecanismo de comunicación utilizado.

### 19.1 Comunicación REST

Se utilizarán códigos HTTP cuando corresponda:

```text
200 OK
201 Created
400 Bad Request
401 Unauthorized
403 Forbidden
404 Not Found
500 Internal Server Error
503 Service Unavailable
```

### 19.2 Comunicación distribuida interna

RMI, SOAP, gRPC y TCP Sockets deberán manejar errores propios de:

- Servicio no disponible.
- Error de comunicación.
- Solicitud inválida.
- Tiempo de espera.
- Error interno.

Los detalles internos no deben exponerse directamente al cliente.

---

## 20. Seguridad de la comunicación

La seguridad debe aplicarse según el tipo de comunicación.

### 20.1 Comunicación externa

La comunicación entre clientes y API Gateway utilizará HTTPS.

### 20.2 Autenticación

El Auth Service será responsable de validar las credenciales y de proporcionar el resultado de autenticación al API Gateway.

Los mecanismos concretos para emisión y validación de tokens se definirán en el contrato de autenticación.

### 20.3 Comunicación interna

Las credenciales, cadenas de conexión y configuraciones sensibles deberán manejarse mediante variables de entorno o mecanismos equivalentes, evitando incluir secretos en el repositorio.

---

## 21. Reglas generales de integración

La integración entre componentes seguirá estas reglas:

1. Los clientes utilizan el API Gateway para las operaciones generales.
2. El Sync Service utiliza TCP Sockets exclusivamente para sincronización.
3. Cada servicio mantiene su responsabilidad funcional.
4. Los contratos de comunicación deben definirse antes de depender de una implementación externa.
5. Los detalles internos de un servicio no deben ser expuestos innecesariamente.
6. PostgreSQL almacena información persistente y metadatos según el modelo de datos.
7. Shared Storage almacena el contenido físico de los archivos.
8. OpenMPI se limita al procesamiento paralelo dentro del Cluster HPC.
9. gRPC se utiliza para la comunicación interna de monitoreo definida por contrato.
10. Los errores internos deben ser controlados antes de retornar una respuesta al cliente.

---

## 22. Resumen de la arquitectura de comunicación

La plataforma utiliza diferentes mecanismos de comunicación según la naturaleza de cada componente.

```text
Clientes
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   ├── Java RMI ───► Auth Service
   │
   ├── REST ───────► File Service
   │                    ├── PostgreSQL
   │                    └── Shared Storage
   │
   ├── REST ───────► Photo Service
   │
   ├── SOAP ───────► Streaming Service
   │
   ├── REST ───────► HPC Service
   │                    │
   │                    └── OpenMPI ───► Cluster HPC
   │
   └── REST ───────► Monitoring Service
                         ▲
                         │ gRPC
                         │
                    Servicios / Nodos


Cliente Desktop / Mobile
           │
           │ TCP Socket
           ▼
      Sync Service
           │
           └──────────────► Shared Storage
```

| Componente | Comunicación principal |
|---|---|
| Cliente → API Gateway | HTTPS / REST |
| API Gateway → Auth Service | Java RMI |
| API Gateway → File Service | REST |
| Cliente → Sync Service | TCP Sockets |
| API Gateway → Photo Service | REST |
| API Gateway → Streaming Service | SOAP |
| API Gateway → HPC Service | REST |
| HPC Cluster interno | OpenMPI |
| API Gateway → Monitoring Service | REST |
| Servicios/Nodos → Monitoring | gRPC |

Este documento establece la arquitectura base de comunicación distribuida de UPB-CIENTÍFICA. Los contratos concretos de cada mecanismo deberán desarrollarse progresivamente durante la implementación de los servicios.
