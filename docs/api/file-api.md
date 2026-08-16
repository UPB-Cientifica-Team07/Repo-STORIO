# File Service - Contrato REST

## 1. Objetivo

Definir el contrato de comunicación del File Service de UPB-CIENTÍFICA.

El File Service es responsable de gestionar los archivos de los usuarios dentro de la plataforma, incluyendo su carga, consulta, descarga y eliminación.

El servicio administra dos tipos principales de información:

- Los metadatos del archivo.
- El contenido físico del archivo.

Los metadatos se almacenan en PostgreSQL, mientras que el archivo físico se almacena en Shared Storage.

---

# 2. Alcance

Este documento define:

- La comunicación entre el API Gateway y el File Service.
- Las operaciones principales relacionadas con archivos.
- Los recursos REST iniciales.
- La estructura general de solicitudes y respuestas.
- El manejo de metadatos.
- La interacción con PostgreSQL.
- La interacción con Shared Storage.
- La autenticación y validación de permisos.
- El manejo general de errores.

Este documento no define todavía:

- El sistema físico definitivo de almacenamiento distribuido.
- Los límites finales de tamaño de archivos.
- Las políticas definitivas de versionado.
- Los mecanismos de replicación.
- El sistema de recuperación ante fallos del almacenamiento.

Estos elementos podrán definirse durante las siguientes etapas de implementación.

---

# 3. Modelo de comunicación

El File Service recibe solicitudes a través del API Gateway.

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
   │
   ├──────────────────┐
   │                  │
   ▼                  ▼
PostgreSQL       Shared Storage
   │                  │
   │                  │
Metadatos       Archivo físico
```

El cliente no debe acceder directamente a PostgreSQL ni al almacenamiento interno.

---

# 4. Tecnología de comunicación

La comunicación entre el API Gateway y el File Service utiliza REST.

| Elemento | Tecnología |
|---|---|
| Protocolo externo | HTTPS |
| Protocolo interno | HTTP/REST |
| Formato principal | JSON |
| Transferencia de archivos | multipart/form-data |
| Codificación | UTF-8 |
| Persistencia | PostgreSQL |
| Almacenamiento físico | Shared Storage |

---

# 5. Responsabilidades del File Service

El File Service será responsable de:

- Recibir solicitudes relacionadas con archivos.
- Validar la información recibida.
- Gestionar la carga de archivos.
- Registrar los metadatos.
- Consultar archivos de un usuario.
- Consultar información de un archivo específico.
- Recuperar archivos desde Shared Storage.
- Gestionar la eliminación de archivos.
- Validar que el usuario tenga acceso al archivo solicitado.

El File Service no debe encargarse directamente de:

- La autenticación de usuarios.
- La generación de tokens.
- La interfaz gráfica de los clientes.
- El monitoreo global de la plataforma.
- La sincronización persistente mediante sockets.

---

# 6. Recurso principal

El recurso principal del servicio será:

```text
/files
```

Las rutas iniciales son:

```text
GET    /files
POST   /files
GET    /files/{id}
GET    /files/{id}/download
DELETE /files/{id}
```

---

# 7. Autenticación y autorización

Las operaciones del File Service requieren que el usuario haya sido autenticado.

El flujo general es:

```text
Cliente
   │
   │ Authorization: Bearer <token>
   ▼
API Gateway
   │
   │ Validación de autenticación
   ▼
File Service
```

El API Gateway será responsable de realizar la validación principal del token mediante el Auth Service.

El File Service recibirá la información necesaria del usuario autenticado para determinar si tiene permisos sobre el recurso solicitado.

Un usuario no debe poder acceder, descargar o eliminar archivos pertenecientes a otro usuario, salvo que posteriormente se definan mecanismos explícitos de compartición.

---

# 8. Operación: cargar archivo

## 8.1 Endpoint

```text
POST /files
```

Esta operación permite registrar un nuevo archivo en la plataforma.

---

## 8.2 Tipo de contenido

La carga del archivo utilizará:

```text
multipart/form-data
```

La solicitud contendrá inicialmente:

```text
file
```

Opcionalmente, en futuras versiones, se podrán incluir campos adicionales relacionados con:

```text
descripcion
categoria
metadatos adicionales
```

---

## 8.3 Flujo de carga

```text
Cliente
   │
   │ POST /files
   │ Archivo
   ▼
API Gateway
   │
   ▼
File Service
   │
   ├── Validar solicitud
   │
   ├── Validar usuario
   │
   ├── Almacenar archivo
   │
   ▼
Shared Storage
   │
   │ Ubicación del archivo
   ▼
File Service
   │
   │ Registrar metadatos
   ▼
PostgreSQL
   │
   ▼
Respuesta
```

---

## 8.4 Validaciones iniciales

El File Service deberá validar:

- Que exista un archivo.
- Que el archivo tenga un nombre válido.
- Que el usuario esté autorizado.
- Que el archivo pueda almacenarse correctamente.
- Que los metadatos puedan registrarse.

La validación específica del tipo y tamaño máximo podrá ampliarse posteriormente.

---

## 8.5 Respuesta exitosa

Código:

```text
201 Created
```

Ejemplo:

```json
{
  "success": true,
  "message": "Archivo cargado correctamente",
  "data": {
    "idArchivo": "uuid",
    "nombre": "investigacion.pdf",
    "tipoArchivo": "DOCUMENTO",
    "tamano": 2457600,
    "fechaSubida": "2026-08-16T10:00:00"
  }
}
```

---

# 9. Operación: listar archivos

## 9.1 Endpoint

```text
GET /files
```

Esta operación permite consultar los archivos asociados al usuario autenticado.

---

## 9.2 Flujo

```text
Cliente
   │
   │ GET /files
   ▼
API Gateway
   │
   ▼
File Service
   │
   │ Consultar
   ▼
PostgreSQL
   │
   ▼
Lista de archivos
```

Inicialmente, esta operación consulta los metadatos.

No debe transferir automáticamente el contenido físico de los archivos.

---

## 9.3 Respuesta exitosa

Código:

```text
200 OK
```

Ejemplo:

```json
{
  "success": true,
  "message": "Archivos obtenidos correctamente",
  "data": [
    {
      "idArchivo": "uuid",
      "nombre": "investigacion.pdf",
      "tipoArchivo": "DOCUMENTO",
      "tamano": 2457600,
      "fechaSubida": "2026-08-16T10:00:00"
    },
    {
      "idArchivo": "uuid",
      "nombre": "microscopia.png",
      "tipoArchivo": "IMAGEN",
      "tamano": 5242880,
      "fechaSubida": "2026-08-16T10:05:00"
    }
  ]
}
```

---

# 10. Operación: consultar archivo

## 10.1 Endpoint

```text
GET /files/{id}
```

Esta operación permite consultar los metadatos de un archivo específico.

---

## 10.2 Parámetro

| Parámetro | Tipo | Descripción |
|---|---|---|
| id | UUID | Identificador del archivo |

---

## 10.3 Flujo

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
   ├── Verificar usuario
   │
   ├── Buscar archivo
   │
   ▼
PostgreSQL
   │
   ▼
Metadatos
```

---

## 10.4 Respuesta exitosa

```json
{
  "success": true,
  "message": "Archivo encontrado",
  "data": {
    "idArchivo": "uuid",
    "nombre": "investigacion.pdf",
    "tipoArchivo": "DOCUMENTO",
    "tamano": 2457600,
    "fechaSubida": "2026-08-16T10:00:00"
  }
}
```

---

# 11. Operación: descargar archivo

## 11.1 Endpoint

```text
GET /files/{id}/download
```

Esta operación permite recuperar el contenido físico de un archivo.

---

## 11.2 Flujo

```text
Cliente
   │
   │ GET /files/{id}/download
   ▼
API Gateway
   │
   ▼
File Service
   │
   ├── Validar existencia
   │
   ├── Validar permisos
   │
   ├── Consultar metadatos
   │
   ▼
PostgreSQL
   │
   │
   ▼
File Service
   │
   │ Recuperar archivo
   ▼
Shared Storage
   │
   ▼
File Service
   │
   ▼
API Gateway
   │
   ▼
Cliente
```

---

## 11.3 Validaciones

Antes de recuperar el archivo, el servicio debe verificar:

1. Que el archivo exista.
2. Que el usuario tenga permisos.
3. Que el registro de metadatos exista.
4. Que el archivo físico esté disponible.

---

## 11.4 Respuesta

Si la operación es exitosa, el contenido del archivo será retornado al cliente.

El formato dependerá del tipo de archivo.

Por ejemplo:

```text
application/pdf
image/png
image/jpeg
video/mp4
application/octet-stream
```

---

# 12. Operación: eliminar archivo

## 12.1 Endpoint

```text
DELETE /files/{id}
```

Esta operación permite eliminar un archivo asociado al usuario.

---

## 12.2 Flujo

```text
Cliente
   │
   │ DELETE /files/{id}
   ▼
API Gateway
   │
   ▼
File Service
   │
   ├── Verificar existencia
   │
   ├── Validar permisos
   │
   ├── Eliminar contenido físico
   │
   ▼
Shared Storage
   │
   ▼
File Service
   │
   ├── Eliminar metadatos
   │
   ▼
PostgreSQL
```

---

## 12.3 Respuesta

Código:

```text
204 No Content
```

También podrá utilizarse una respuesta estructurada si la implementación requiere confirmar la operación.

---

# 13. Metadatos del archivo

La entidad principal relacionada con el File Service es:

```text
archivo
```

Los campos iniciales definidos en el modelo de datos son:

| Campo | Tipo | Descripción |
|---|---|---|
| id_archivo | UUID | Identificador único |
| nombre | VARCHAR | Nombre del archivo |
| tipo_archivo | VARCHAR | Tipo general del archivo |
| tamano | BIGINT | Tamaño en bytes |
| id_usuario | UUID | Propietario |
| fecha_subida | TIMESTAMP | Fecha de carga |

Dependiendo de la implementación del almacenamiento, podrá existir información adicional relacionada con la ubicación interna del archivo.

Esta información no debe exponerse directamente al cliente.

---

# 14. Clasificación inicial de archivos

El sistema maneja inicialmente las siguientes categorías:

```text
DOCUMENTO
IMAGEN
VIDEO
```

Estas categorías permiten relacionar los archivos con otros servicios.

Por ejemplo:

```text
DOCUMENTO
    │
    ▼
File Service


IMAGEN
    │
    ├── File Service
    │
    ▼
Photo Service


VIDEO
    │
    ├── File Service
    │
    ▼
Streaming Service
```

El File Service mantiene la gestión general del recurso.

Los servicios especializados procesan los archivos según su tipo.

---

# 15. Relación con PostgreSQL

PostgreSQL almacena principalmente los metadatos de los archivos.

```text
File Service
      │
      │ SQL
      ▼
PostgreSQL
      │
      ▼
   archivo
```

El contenido binario del archivo no forma parte del almacenamiento principal de la tabla de metadatos.

La separación es:

```text
PostgreSQL
    │
    └── Metadatos

Shared Storage
    │
    └── Archivo físico
```

---

# 16. Relación con Shared Storage

Shared Storage almacena el contenido físico de los archivos.

El File Service interactúa con este componente durante:

- Carga.
- Descarga.
- Eliminación.

Modelo conceptual:

```text
                File Service
                     │
          ┌──────────┴──────────┐
          │                     │
          ▼                     ▼
     PostgreSQL            Shared Storage
          │                     │
          │                     │
       Metadatos          Archivo físico
```

El mecanismo concreto de almacenamiento distribuido será definido durante la implementación.

---

# 17. Relación con otros servicios

El File Service puede proporcionar información o acceso controlado a otros servicios.

## 17.1 Photo Service

Puede utilizar archivos clasificados como:

```text
IMAGEN
```

Flujo conceptual:

```text
Photo Service
      │
      │ Información / Archivo
      ▼
File Service
      │
      ▼
Shared Storage
```

---

## 17.2 Streaming Service

Puede utilizar archivos clasificados como:

```text
VIDEO
```

Flujo conceptual:

```text
Streaming Service
       │
       ▼
File Service
       │
       ▼
Shared Storage
```

---

## 17.3 Sync Service

El Sync Service puede interactuar con el almacenamiento para mantener los archivos sincronizados.

```text
Sync Service
      │
      ▼
Shared Storage
```

La comunicación específica del Sync Service se define en:

```text
docs/api/sync-protocol.md
```

---

# 18. Manejo de errores

El File Service debe manejar los errores internos y retornar resultados controlados.

Los escenarios iniciales son:

| Situación | Código |
|---|---|
| Solicitud inválida | 400 |
| Usuario no autenticado | 401 |
| Usuario sin permisos | 403 |
| Archivo no encontrado | 404 |
| Conflicto durante operación | 409 |
| Error interno | 500 |
| Servicio no disponible | 503 |

---

# 19. Errores durante la carga

Durante la carga pueden ocurrir los siguientes casos:

```text
Archivo no enviado
Archivo inválido
Error al almacenar
Error al registrar metadatos
Almacenamiento no disponible
Error interno
```

El servicio debe evitar que una operación parcialmente completada deje información inconsistente.

Por ejemplo:

```text
Archivo almacenado
       │
       ▼
Error al registrar metadatos
       │
       ▼
Aplicar limpieza o compensación
```

La estrategia concreta de compensación será definida durante la implementación.

---

# 20. Errores durante la descarga

Los principales escenarios son:

```text
Archivo no existe en PostgreSQL
Archivo físico no disponible
Usuario sin permisos
Shared Storage no disponible
Error interno
```

La información técnica interna no debe exponerse directamente al cliente.

---

# 21. Seguridad y permisos

El acceso a un archivo debe considerar:

```text
Usuario autenticado
        │
        ▼
Identidad del usuario
        │
        ▼
Propiedad o permiso sobre el archivo
        │
        ├── Permitido
        │
        └── Denegado
```

Inicialmente, el propietario del archivo será identificado mediante:

```text
id_usuario
```

La implementación de archivos compartidos o permisos entre múltiples usuarios podrá agregarse posteriormente.

---

# 22. Flujo completo de carga

```text
Cliente
   │
   │ POST /files
   │ multipart/form-data
   ▼
API Gateway
   │
   │ Validación de acceso
   ▼
File Service
   │
   ├── Validar archivo
   │
   ├── Almacenar contenido
   │
   ▼
Shared Storage
   │
   │ Confirmación
   ▼
File Service
   │
   ├── Registrar metadatos
   │
   ▼
PostgreSQL
   │
   │ Confirmación
   ▼
File Service
   │
   ▼
API Gateway
   │
   │ 201 Created
   ▼
Cliente
```

---

# 23. Flujo completo de descarga

```text
Cliente
   │
   │ GET /files/{id}/download
   ▼
API Gateway
   │
   ▼
File Service
   │
   ├── Validar permisos
   │
   ├── Consultar metadatos
   │
   ▼
PostgreSQL
   │
   │ Información del archivo
   ▼
File Service
   │
   ├── Recuperar contenido
   │
   ▼
Shared Storage
   │
   │ Archivo físico
   ▼
File Service
   │
   ▼
API Gateway
   │
   ▼
Cliente
```

---

# 24. Flujo completo de eliminación

```text
Cliente
   │
   │ DELETE /files/{id}
   ▼
API Gateway
   │
   ▼
File Service
   │
   ├── Validar usuario
   │
   ├── Consultar archivo
   │
   ▼
PostgreSQL
   │
   │
   ▼
File Service
   │
   ├── Eliminar contenido
   │
   ▼
Shared Storage
   │
   │
   ▼
File Service
   │
   ├── Eliminar metadatos
   │
   ▼
PostgreSQL
   │
   ▼
204 No Content
```

---

# 25. Reglas del contrato

El File Service debe cumplir las siguientes reglas:

1. Los clientes acceden al servicio a través del API Gateway.
2. Las operaciones utilizan REST.
3. Los archivos físicos se almacenan fuera de PostgreSQL.
4. PostgreSQL almacena principalmente los metadatos.
5. El usuario debe estar autenticado para realizar operaciones.
6. El acceso a un archivo debe validar permisos.
7. La ubicación interna del archivo no debe exponerse al cliente.
8. Los errores internos deben ser controlados.
9. Las operaciones deben evitar inconsistencias entre metadatos y almacenamiento físico.
10. Los servicios especializados deben utilizar mecanismos controlados para acceder a los archivos necesarios.

---

# 26. Resumen

El File Service centraliza la gestión de archivos dentro de UPB-CIENTÍFICA.

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ REST
   ▼
File Service
   │
   ├──────────────────┐
   │                  │
   ▼                  ▼
PostgreSQL       Shared Storage
   │                  │
   │                  │
Metadatos       Archivo físico
```

Las operaciones iniciales son:

```text
POST   /files
GET    /files
GET    /files/{id}
GET    /files/{id}/download
DELETE /files/{id}
```

El File Service separa la gestión de los metadatos del almacenamiento físico, permitiendo que la arquitectura evolucione posteriormente hacia mecanismos de almacenamiento distribuido sin modificar el contrato principal expuesto a los clientes.