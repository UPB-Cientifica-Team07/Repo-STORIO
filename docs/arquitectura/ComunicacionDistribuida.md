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
1. El cliente realiza una solicitud.
           │
           ▼
2. La solicitud llega al API Gateway.
           │
           ▼
3. El Gateway identifica el recurso solicitado.
           │
           ▼
4. El Gateway dirige la solicitud al servicio correspondiente.
           │
           ▼
5. El servicio procesa la solicitud.
           │
           ▼
6. El servicio genera una respuesta.
           │
           ▼
7. La respuesta retorna al API Gateway.
           │
           ▼
8. El Gateway responde al cliente.
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







