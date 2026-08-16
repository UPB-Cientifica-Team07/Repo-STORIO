# Streaming Service - Contrato SOAP

## 1. Objetivo

Definir el contrato de comunicación del Streaming Service de UPB-CIENTÍFICA.

El Streaming Service es responsable de proporcionar operaciones relacionadas con la consulta y reproducción de contenido multimedia, principalmente videos almacenados dentro de la plataforma.

La comunicación se implementará mediante SOAP, utilizando mensajes XML y un contrato definido mediante WSDL.

El servicio permite inicialmente:

- Consultar los videos disponibles.
- Obtener información de un video.
- Obtener la información necesaria para iniciar su reproducción.
- Validar el acceso del usuario al contenido solicitado.

---

## 2. Alcance

Este documento define:

- La responsabilidad del Streaming Service.
- La comunicación mediante SOAP.
- Las operaciones remotas iniciales.
- La estructura general de solicitudes y respuestas.
- La relación con la entidad `video`.
- La integración con el API Gateway.
- La relación con File Service.
- El acceso al contenido multimedia.
- El manejo de errores.
- Las reglas generales del contrato.

Este documento no define todavía:

- Streaming adaptativo.
- HLS.
- MPEG-DASH.
- Transcodificación automática.
- Conversión de formatos.
- Subtítulos.
- Recomendaciones de contenido.
- Streaming en tiempo real.
- CDN o distribución geográfica del contenido.

Estas funcionalidades podrán incorporarse posteriormente si hacen parte del alcance del proyecto.

---

# 3. Modelo de comunicación

El cliente accede inicialmente mediante el API Gateway.

```text
Cliente
   │
   │ HTTPS
   ▼
API Gateway
   │
   │ SOAP
   ▼
Streaming Service
   │
   ├───────────────┐
   │               │
   ▼               ▼
PostgreSQL     File Service
   │               │
   │               ▼
   │          Shared Storage
   │
   ▼
video
archivo
```

El Streaming Service gestiona las operaciones relacionadas con el contenido multimedia, mientras que el almacenamiento físico continúa siendo responsabilidad del sistema de archivos.

---

# 4. Tecnología de comunicación

| Elemento | Tecnología |
|---|---|
| Protocolo externo | HTTPS |
| Comunicación interna | SOAP |
| Formato de mensajes | XML |
| Contrato | WSDL |
| Implementación | PHP |
| Persistencia | PostgreSQL |
| Almacenamiento | Shared Storage |

SOAP permite definir un contrato formal para las operaciones remotas del servicio.

---

# 5. Responsabilidades del Streaming Service

El Streaming Service será responsable de:

- Consultar videos disponibles.
- Obtener información específica de un video.
- Gestionar metadatos relacionados con videos.
- Validar el acceso al contenido solicitado.
- Proporcionar información necesaria para la reproducción.
- Coordinar la recuperación del contenido multimedia.

El Streaming Service no será responsable directamente de:

- La autenticación principal de usuarios.
- La generación de tokens.
- La sincronización de archivos.
- La carga general de archivos.
- La eliminación general de archivos.
- La ejecución de trabajos HPC.

---

# 6. Relación con File Service

El Streaming Service trabaja sobre archivos previamente registrados en el sistema.

La separación de responsabilidades es:

```text
File Service
    │
    ├── Carga de archivos
    ├── Gestión general
    ├── Registro de metadatos
    └── Almacenamiento físico
             │
             ▼
      Streaming Service
             │
             ├── Consulta de videos
             ├── Información multimedia
             ├── Duración
             ├── Calidad
             └── Reproducción
```

La relación principal se realiza mediante:

```text
id_archivo
```

---

# 7. Contrato SOAP

El Streaming Service expone sus operaciones mediante un servicio SOAP.

La estructura conceptual es:

```text
StreamingService
│
├── listVideos()
│
├── getVideo()
│
└── getStreamingInfo()
```

El contrato definitivo se describirá mediante un archivo WSDL.

La ubicación propuesta es:

```text
services/streaming-service/wsdl/streaming.wsdl
```

---

# 8. Operaciones principales

Las operaciones iniciales del servicio son:

| Operación | Descripción |
|---|---|
| `listVideos` | Consulta los videos disponibles |
| `getVideo` | Obtiene la información de un video |
| `getStreamingInfo` | Obtiene la información necesaria para reproducir un video |

---

# 9. Operación listVideos

## 9.1 Objetivo

Consultar los videos disponibles para el usuario autenticado.

Representación conceptual:

```text
listVideos(usuarioId)
```

La identidad del usuario puede ser transmitida desde el API Gateway después de validar el token.

---

## 9.2 Flujo

```text
API Gateway
      │
      │ listVideos()
      ▼
Streaming Service
      │
      │ Consultar videos
      ▼
PostgreSQL
      │
      ├── archivo
      │
      └── video
      │
      ▼
Lista de videos
```

---

## 9.3 Solicitud SOAP

Ejemplo conceptual:

```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
    <soap:Body>
        <listVideosRequest>
            <usuarioId>uuid</usuarioId>
        </listVideosRequest>
    </soap:Body>
</soap:Envelope>
```

---

## 9.4 Respuesta SOAP

```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
    <soap:Body>
        <listVideosResponse>

            <videos>
                <video>
                    <idVideo>uuid</idVideo>
                    <idArchivo>uuid</idArchivo>
                    <nombre>experimento.mp4</nombre>
                    <duracion>3600</duracion>
                    <calidad>1080p</calidad>
                    <formato>MP4</formato>
                </video>
            </videos>

        </listVideosResponse>
    </soap:Body>
</soap:Envelope>
```

La respuesta devuelve metadatos, no necesariamente el contenido binario completo del video.

---

# 10. Operación getVideo

## 10.1 Objetivo

Consultar la información detallada de un video específico.

Representación conceptual:

```text
getVideo(idVideo, usuarioId)
```

---

## 10.2 Solicitud SOAP

```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
    <soap:Body>
        <getVideoRequest>
            <idVideo>uuid</idVideo>
            <usuarioId>uuid</usuarioId>
        </getVideoRequest>
    </soap:Body>
</soap:Envelope>
```

---

## 10.3 Flujo

```text
API Gateway
      │
      │ getVideo()
      ▼
Streaming Service
      │
      ├── Validar acceso
      │
      ├── Consultar video
      │
      ▼
PostgreSQL
      │
      ├── video
      │
      └── archivo
      │
      ▼
Información del video
```

---

## 10.4 Respuesta SOAP

```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
    <soap:Body>
        <getVideoResponse>

            <video>
                <idVideo>uuid</idVideo>
                <idArchivo>uuid</idArchivo>
                <nombre>experimento.mp4</nombre>
                <duracion>3600</duracion>
                <calidad>1080p</calidad>
                <formato>MP4</formato>
                <tamano>104857600</tamano>
            </video>

        </getVideoResponse>
    </soap:Body>
</soap:Envelope>
```

---

# 11. Operación getStreamingInfo

## 11.1 Objetivo

Obtener la información necesaria para iniciar la reproducción de un video.

Representación conceptual:

```text
getStreamingInfo(idVideo, usuarioId)
```

Esta operación debe verificar:

- Que el video exista.
- Que el usuario tenga acceso.
- Que el archivo asociado esté disponible.

---

## 11.2 Flujo

```text
API Gateway
      │
      │ getStreamingInfo()
      ▼
Streaming Service
      │
      ├── Validar usuario
      │
      ├── Consultar video
      │
      ▼
PostgreSQL
      │
      ▼
Streaming Service
      │
      │ Consultar archivo
      ▼
File Service
      │
      ▼
Shared Storage
```

---

## 11.3 Solicitud SOAP

```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
    <soap:Body>
        <getStreamingInfoRequest>
            <idVideo>uuid</idVideo>
            <usuarioId>uuid</usuarioId>
        </getStreamingInfoRequest>
    </soap:Body>
</soap:Envelope>
```

---

## 11.4 Respuesta SOAP

```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
    <soap:Body>
        <getStreamingInfoResponse>

            <streaming>
                <idVideo>uuid</idVideo>
                <nombre>experimento.mp4</nombre>
                <formato>MP4</formato>
                <calidad>1080p</calidad>
                <duracion>3600</duracion>
                <estado>DISPONIBLE</estado>
            </streaming>

        </getStreamingInfoResponse>
    </soap:Body>
</soap:Envelope>
```

En una primera implementación, la respuesta puede proporcionar la información necesaria para que el cliente solicite el contenido mediante el mecanismo definido por la arquitectura.

La ubicación física interna del archivo no debe exponerse directamente.

---

# 12. Entidad video

El Streaming Service trabaja principalmente con la entidad:

```text
video
```

Los campos iniciales son:

| Campo | Tipo | Descripción |
|---|---|---|
| id_video | UUID | Identificador del video |
| id_archivo | UUID | Archivo asociado |
| duracion | INTEGER | Duración en segundos |
| calidad | VARCHAR | Calidad del video |

La información general se encuentra en la entidad:

```text
archivo
```

---

# 13. Relación entre archivo y video

El modelo conceptual es:

```text
usuario
   │
   ▼
archivo
   │
   ▼
video
```

La entidad `archivo` contiene información general:

```text
nombre
tipo_archivo
tamano
id_usuario
fecha_subida
```

La entidad `video` contiene información específica:

```text
duracion
calidad
```

---

# 14. Relación con PostgreSQL

El Streaming Service consulta principalmente:

```text
archivo
video
```

Modelo:

```text
Streaming Service
        │
        ▼
    PostgreSQL
        │
        ├── archivo
        │
        └── video
```

El contenido físico del video permanece separado de los metadatos.

---

# 15. Relación con Shared Storage

El almacenamiento físico se encuentra en Shared Storage.

```text
Streaming Service
        │
        ▼
     File Service
        │
        ▼
   Shared Storage
        │
        └── Video físico
```

La base de datos mantiene los metadatos necesarios para localizar y administrar el recurso.

---

# 16. Seguridad

El acceso a las operaciones debe estar controlado.

Flujo general:

```text
Cliente
   │
   │ Token
   ▼
API Gateway
   │
   │ Validación
   ▼
Auth Service
   │
   ▼
API Gateway
   │
   │ SOAP
   ▼
Streaming Service
```

El Streaming Service debe recibir la información necesaria para identificar al usuario autenticado.

Antes de entregar información sobre un video, el servicio debe validar:

1. Que el usuario esté identificado.
2. Que el video exista.
3. Que el usuario tenga permisos sobre el recurso.

---

# 17. Manejo de errores

Los errores deben utilizar una estructura SOAP controlada mediante `SOAP Fault`.

Ejemplo conceptual:

```xml
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
    <soap:Body>
        <soap:Fault>
            <faultcode>VIDEO_NOT_FOUND</faultcode>
            <faultstring>El video solicitado no existe</faultstring>
        </soap:Fault>
    </soap:Body>
</soap:Envelope>
```

Los errores iniciales pueden incluir:

| Código | Descripción |
|---|---|
| INVALID_REQUEST | Solicitud inválida |
| AUTH_REQUIRED | Usuario no autenticado |
| ACCESS_DENIED | Acceso no autorizado |
| VIDEO_NOT_FOUND | Video no encontrado |
| FILE_NOT_AVAILABLE | Archivo no disponible |
| INTERNAL_ERROR | Error interno |

---

# 18. Video no encontrado

Cuando el identificador solicitado no corresponda a un video existente:

```text
VIDEO_NOT_FOUND
```

Ejemplo:

```xml
<soap:Fault>
    <faultcode>VIDEO_NOT_FOUND</faultcode>
    <faultstring>El video solicitado no existe</faultstring>
</soap:Fault>
```

---

# 19. Acceso denegado

Cuando el usuario no tenga permisos sobre el contenido:

```text
ACCESS_DENIED
```

Ejemplo:

```xml
<soap:Fault>
    <faultcode>ACCESS_DENIED</faultcode>
    <faultstring>El usuario no tiene permisos sobre este video</faultstring>
</soap:Fault>
```

---

# 20. Flujo completo de consulta

```text
Cliente
   │
   │ HTTPS
   ▼
API Gateway
   │
   │ Validar token
   ▼
Auth Service
   │
   ▼
API Gateway
   │
   │ SOAP
   ▼
Streaming Service
   │
   │ Consultar
   ▼
PostgreSQL
   │
   ├── archivo
   │
   └── video
   │
   ▼
Streaming Service
   │
   │ SOAP Response
   ▼
API Gateway
   │
   ▼
Cliente
```

---

# 21. Flujo de reproducción

El flujo inicial de reproducción será:

```text
Cliente
   │
   │ Solicitar información
   ▼
API Gateway
   │
   ▼
Streaming Service
   │
   │ Validar acceso
   │
   ├── Consultar video
   ▼
PostgreSQL
   │
   ▼
Streaming Service
   │
   │ Verificar archivo
   ▼
File Service
   │
   ▼
Shared Storage
   │
   ▼
Contenido disponible
   │
   ▼
Streaming Service
   │
   ▼
Cliente
```

La implementación inicial puede realizar la reproducción mediante acceso progresivo al contenido disponible, sin implementar todavía protocolos especializados de streaming adaptativo.

---

# 22. Reglas generales del contrato

El Streaming Service debe cumplir las siguientes reglas:

1. La comunicación entre el API Gateway y el servicio utiliza SOAP.
2. Los mensajes SOAP utilizan XML.
3. Las operaciones deben estar definidas mediante un contrato WSDL.
4. El servicio gestiona operaciones relacionadas con videos.
5. Cada video debe estar asociado a un archivo.
6. Los metadatos se almacenan en PostgreSQL.
7. El contenido físico se mantiene separado en Shared Storage.
8. El usuario debe estar autenticado.
9. Se deben validar permisos antes de consultar información protegida.
10. La ubicación física interna del archivo no debe exponerse al cliente.
11. Los errores deben utilizar `SOAP Fault`.
12. El servicio no reemplaza las responsabilidades generales del File Service.

---

# 23. WSDL propuesto

La ubicación inicial del contrato será:

```text
services/
└── streaming-service/
    └── wsdl/
        └── streaming.wsdl
```

El WSDL definirá:

```text
Types
Messages
PortType
Binding
Service
```

Las operaciones iniciales serán:

```text
listVideos
getVideo
getStreamingInfo
```

---

# 24. Estructura inicial propuesta

```text
services/
└── streaming-service/
    │
    ├── src/
    │   ├── Service/
    │   │   └── StreamingService.php
    │   │
    │   ├── Model/
    │   │   └── Video.php
    │   │
    │   ├── Repository/
    │   │   └── VideoRepository.php
    │   │
    │   └── Client/
    │       └── FileServiceClient.php
    │
    ├── wsdl/
    │   └── streaming.wsdl
    │
    └── public/
        └── index.php
```

La estructura definitiva podrá ajustarse durante la implementación.

---

# 25. Resumen

El Streaming Service proporciona operaciones distribuidas relacionadas con la consulta y reproducción de videos dentro de UPB-CIENTÍFICA.

```text
Cliente
   │
   │ HTTPS
   ▼
API Gateway
   │
   │ SOAP
   ▼
Streaming Service
   │
   ├───────────────┐
   │               │
   ▼               ▼
PostgreSQL     File Service
   │               │
   ▼               ▼
 video       Shared Storage
```

Las operaciones iniciales son:

```text
listVideos()
getVideo()
getStreamingInfo()
```

El contrato utiliza SOAP y XML, y su definición formal se realizará mediante WSDL.

La entidad `video` almacena los metadatos específicos del contenido multimedia y mantiene una relación con la entidad `archivo` mediante `id_archivo`.

El Streaming Service complementa la gestión general realizada por el File Service, manteniendo la separación entre los metadatos multimedia, la lógica de reproducción y el almacenamiento físico del contenido.