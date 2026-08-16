# Photo Service - Contrato REST

## 1. Objetivo

Definir el contrato de comunicación del Photo Service de UPB-CIENTÍFICA.

El Photo Service es responsable de las operaciones especializadas relacionadas con imágenes dentro de la plataforma. Su función es complementar la gestión general realizada por el File Service, proporcionando acceso a información específica de imágenes y operaciones orientadas a su consulta.

El File Service mantiene la gestión general del archivo físico, mientras que el Photo Service trabaja con los recursos cuyo tipo corresponde a:

```text
IMAGEN
```

---

# 2. Alcance

Este documento define:

- La responsabilidad del Photo Service.
- La comunicación con el API Gateway.
- Los recursos REST iniciales.
- La consulta de imágenes.
- La consulta de información específica.
- La relación con File Service.
- La relación con PostgreSQL.
- La relación con Shared Storage.
- La autenticación y autorización.
- Las respuestas y errores principales.

Este documento no define todavía:

- Edición avanzada de imágenes.
- Reconocimiento automático de imágenes.
- Inteligencia artificial aplicada a imágenes.
- Álbumes colaborativos.
- Compartición entre usuarios.
- Procesamiento distribuido de imágenes.
- Generación de miniaturas avanzada.
- Compresión automática.

Estas funcionalidades podrán incorporarse posteriormente según el alcance final del proyecto.

---

# 3. Modelo de comunicación

El acceso principal al Photo Service se realiza mediante el API Gateway.

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ REST
   ▼
Photo Service
   │
   ├───────────────┐
   │               │
   ▼               ▼
File Service   PostgreSQL
   │               │
   ▼               ▼
Shared Storage    imagen
```

El cliente no debe acceder directamente al almacenamiento interno.

---

# 4. Tecnología de comunicación

La comunicación utiliza una arquitectura REST.

| Elemento | Tecnología |
|---|---|
| Protocolo externo | HTTPS |
| Comunicación principal | REST |
| Formato | JSON |
| Codificación | UTF-8 |
| Persistencia | PostgreSQL |
| Archivo físico | Shared Storage |
| Servicio relacionado | File Service |

---

# 5. Responsabilidades del Photo Service

El Photo Service será responsable de:

- Consultar imágenes disponibles para un usuario.
- Consultar información específica de una imagen.
- Gestionar metadatos relacionados con imágenes.
- Proporcionar acceso especializado a recursos de tipo imagen.
- Verificar la relación entre una imagen y su archivo principal.
- Validar permisos de acceso.
- Coordinar la recuperación del contenido físico cuando sea necesario.

El Photo Service no será responsable directamente de:

- La autenticación de usuarios.
- La generación de tokens.
- El almacenamiento general de todos los tipos de archivos.
- La sincronización mediante TCP Sockets.
- La ejecución de trabajos HPC.

---

# 6. Relación con File Service

El Photo Service trabaja sobre archivos administrados por el File Service.

La separación de responsabilidades es:

```text
File Service
    │
    ├── Gestión general del archivo
    ├── Registro de metadatos principales
    ├── Carga
    ├── Descarga
    └── Eliminación
             │
             ▼
        Photo Service
             │
             ├── Información específica
             ├── Resolución
             ├── Formato
             └── Operaciones relacionadas con imágenes
```

La relación principal se establece mediante el identificador del archivo:

```text
id_archivo
```

---

# 7. Recurso principal

El recurso principal será:

```text
/photos
```

Las rutas iniciales son:

```text
GET /photos
GET /photos/{id}
GET /photos/{id}/download
```

Posteriormente podrán agregarse operaciones adicionales según las necesidades del proyecto.

---

# 8. Autenticación y autorización

Las operaciones requieren un usuario autenticado.

El flujo general será:

```text
Cliente
   │
   │ Authorization: Bearer <token>
   ▼
API Gateway
   │
   │ Validar acceso
   ▼
Photo Service
```

El API Gateway realiza la validación principal del token mediante el mecanismo definido con el Auth Service.

El Photo Service debe recibir la identidad del usuario autenticado para verificar que tenga permisos sobre la imagen solicitada.

---

# 9. Operación: listar imágenes

## 9.1 Endpoint

```text
GET /photos
```

Esta operación permite consultar las imágenes disponibles para el usuario autenticado.

---

## 9.2 Flujo

```text
Cliente
   │
   │ GET /photos
   ▼
API Gateway
   │
   ▼
Photo Service
   │
   │ Consultar información
   ▼
PostgreSQL
   │
   ├── archivo
   │
   └── imagen
   │
   ▼
Lista de imágenes
```

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
  "message": "Imágenes obtenidas correctamente",
  "data": [
    {
      "idImagen": "uuid",
      "idArchivo": "uuid",
      "nombre": "microscopia.png",
      "resolucion": "1920x1080",
      "formato": "PNG",
      "tamano": 5242880,
      "fechaSubida": "2026-08-16T10:00:00"
    }
  ]
}
```

La respuesta contiene metadatos y no el contenido binario de la imagen.

---

# 10. Operación: consultar imagen

## 10.1 Endpoint

```text
GET /photos/{id}
```

Permite consultar la información de una imagen específica.

---

## 10.2 Parámetro

| Parámetro | Tipo | Descripción |
|---|---|---|
| id | UUID | Identificador de la imagen |

---

## 10.3 Flujo

```text
Cliente
   │
   │ GET /photos/{id}
   ▼
API Gateway
   │
   ▼
Photo Service
   │
   ├── Validar usuario
   │
   ├── Consultar imagen
   │
   ▼
PostgreSQL
   │
   ├── imagen
   │
   └── archivo
   │
   ▼
Información de imagen
```

---

## 10.4 Respuesta exitosa

```json
{
  "success": true,
  "message": "Imagen encontrada",
  "data": {
    "idImagen": "uuid",
    "idArchivo": "uuid",
    "nombre": "microscopia.png",
    "resolucion": "1920x1080",
    "formato": "PNG",
    "tamano": 5242880,
    "fechaSubida": "2026-08-16T10:00:00"
  }
}
```

---

# 11. Operación: descargar imagen

## 11.1 Endpoint

```text
GET /photos/{id}/download
```

Permite recuperar el contenido físico de una imagen.

---

## 11.2 Flujo

```text
Cliente
   │
   │ GET /photos/{id}/download
   ▼
API Gateway
   │
   ▼
Photo Service
   │
   ├── Validar usuario
   │
   ├── Consultar imagen
   │
   ▼
PostgreSQL
   │
   ▼
Photo Service
   │
   │ Obtener id_archivo
   ▼
File Service / Shared Storage
   │
   ▼
Contenido de imagen
   │
   ▼
Photo Service
   │
   ▼
API Gateway
   │
   ▼
Cliente
```

El mecanismo concreto de recuperación puede ajustarse durante la implementación, pero el cliente no debe conocer la ubicación física interna del archivo.

---

# 12. Metadatos de imagen

La entidad especializada es:

```text
imagen
```

Su relación conceptual es:

```text
archivo
   │
   │ 1
   ▼
imagen
```

Los campos iniciales son:

| Campo | Tipo | Descripción |
|---|---|---|
| id_imagen | UUID | Identificador de la imagen |
| id_archivo | UUID | Archivo asociado |
| resolucion | VARCHAR | Resolución de la imagen |
| formato | VARCHAR | Formato de imagen |

El archivo general contiene información como:

```text
nombre
tipo_archivo
tamano
id_usuario
fecha_subida
```

---

# 13. Relación entre archivo e imagen

La estructura conceptual es:

```text
usuario
   │
   │ 1
   ▼
archivo
   │
   │ 1
   ▼
imagen
```

Cada registro de imagen está asociado a un archivo.

La relación permite mantener separada:

```text
Información general
        │
        ▼
      archivo


Información específica
        │
        ▼
      imagen
```

---

# 14. Formatos iniciales soportados

Inicialmente se consideran formatos comunes:

```text
JPG
JPEG
PNG
```

El sistema podrá incorporar posteriormente:

```text
WEBP
GIF
TIFF
BMP
```

La disponibilidad definitiva dependerá de las capacidades implementadas en el servicio.

---

# 15. Resolución

La resolución representa las dimensiones de la imagen.

Ejemplo:

```text
1920x1080
```

Otros ejemplos:

```text
1280x720
3840x2160
800x600
```

Este valor se almacena como metadato asociado a la entidad:

```text
imagen
```

---

# 16. Relación con PostgreSQL

El Photo Service utiliza PostgreSQL principalmente para consultar:

```text
archivo
imagen
```

Modelo:

```text
Photo Service
      │
      ▼
PostgreSQL
      │
      ├── archivo
      │
      └── imagen
```

La consulta debe permitir relacionar la información general del archivo con los datos específicos de la imagen.

Conceptualmente:

```text
archivo
   │
   └── id_archivo
           │
           ▼
        imagen
```

---

# 17. Relación con Shared Storage

El contenido físico de las imágenes se encuentra en Shared Storage.

```text
Photo Service
      │
      ▼
File Service
      │
      ▼
Shared Storage
```

La separación es:

```text
PostgreSQL
    │
    └── Metadatos


Shared Storage
    │
    └── Imagen física
```

Esto evita almacenar directamente grandes contenidos binarios como parte del modelo principal de datos.

---

# 18. Manejo de errores

El servicio debe retornar respuestas controladas ante errores.

Los códigos iniciales son:

| Situación | Código |
|---|---|
| Solicitud inválida | 400 |
| Usuario no autenticado | 401 |
| Usuario sin permisos | 403 |
| Imagen no encontrada | 404 |
| Error de almacenamiento | 500 |
| Error interno | 500 |
| Servicio no disponible | 503 |

---

# 19. Imagen no encontrada

Cuando no exista una imagen asociada al identificador solicitado:

```text
404 Not Found
```

Ejemplo:

```json
{
  "success": false,
  "message": "Imagen no encontrada",
  "data": null
}
```

---

# 20. Acceso denegado

Cuando el usuario no tenga permisos sobre una imagen:

```text
403 Forbidden
```

Ejemplo:

```json
{
  "success": false,
  "message": "Acceso denegado",
  "data": null
}
```

---

# 21. Seguridad

Las principales reglas de seguridad son:

1. El usuario debe estar autenticado.
2. El acceso debe validar la propiedad o permisos.
3. La ubicación física de la imagen no debe exponerse.
4. Los metadatos internos innecesarios no deben enviarse al cliente.
5. Los errores internos deben manejarse de forma controlada.
6. La autenticación se centraliza mediante el API Gateway y Auth Service.

---

# 22. Flujo completo de consulta

```text
Cliente
   │
   │ GET /photos/{id}
   ▼
API Gateway
   │
   │ Validar token
   ▼
Photo Service
   │
   │ Consultar imagen
   ▼
PostgreSQL
   │
   ├── imagen
   │
   └── archivo
   │
   ▼
Photo Service
   │
   ▼
API Gateway
   │
   │ JSON
   ▼
Cliente
```

---

# 23. Flujo completo de descarga

```text
Cliente
   │
   │ GET /photos/{id}/download
   ▼
API Gateway
   │
   ▼
Photo Service
   │
   ├── Validar permisos
   │
   ├── Consultar metadatos
   │
   ▼
PostgreSQL
   │
   ▼
Photo Service
   │
   │ Solicitar contenido
   ▼
File Service
   │
   ▼
Shared Storage
   │
   │
   ▼
Contenido físico
   │
   ▼
Photo Service
   │
   ▼
API Gateway
   │
   ▼
Cliente
```

---

# 24. Integración con File Service

La interacción entre ambos servicios debe respetar la separación de responsabilidades.

```text
                 Archivo
                    │
                    ▼
              File Service
                    │
       ┌────────────┴────────────┐
       │                         │
       ▼                         ▼
PostgreSQL                  Shared Storage
       │
       ▼
   Photo Service
```

El File Service mantiene el ciclo general de vida del archivo.

El Photo Service trabaja con la información especializada de los archivos clasificados como imágenes.

---

# 25. Reglas del contrato

El Photo Service debe cumplir las siguientes reglas:

1. Los clientes acceden mediante el API Gateway.
2. La comunicación utiliza REST.
3. Las respuestas de consulta utilizan JSON.
4. El contenido físico se encuentra separado de los metadatos.
5. PostgreSQL almacena la información relacionada con la imagen.
6. Shared Storage almacena el contenido físico.
7. Toda imagen debe estar asociada a un archivo.
8. El usuario debe estar autenticado.
9. Se deben validar permisos antes de acceder a una imagen.
10. La ubicación interna del archivo no debe exponerse.
11. Los errores deben manejarse de forma controlada.
12. El Photo Service no reemplaza las responsabilidades generales del File Service.

---

# 26. Estructura inicial propuesta

La implementación puede organizarse posteriormente de la siguiente manera:

```text
services/
└── photo-service/
    │
    ├── api/
    │   ├── PhotoController.java
    │   └── PhotoRoutes.java
    │
    ├── service/
    │   └── PhotoService.java
    │
    ├── repository/
    │   └── PhotoRepository.java
    │
    ├── model/
    │   ├── Imagen.java
    │   └── PhotoMetadata.java
    │
    └── client/
        └── FileServiceClient.java
```

La estructura definitiva podrá modificarse durante la implementación, manteniendo la separación entre API, lógica de negocio, persistencia y comunicación con otros servicios.

---

# 27. Resumen

El Photo Service proporciona funcionalidades especializadas para la gestión y consulta de imágenes dentro de UPB-CIENTÍFICA.

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ REST
   ▼
Photo Service
   │
   ├──────────────┐
   │              │
   ▼              ▼
PostgreSQL    File Service
                   │
                   ▼
             Shared Storage
```

Las operaciones iniciales son:

```text
GET /photos
GET /photos/{id}
GET /photos/{id}/download
```

El servicio utiliza la entidad `imagen` para gestionar metadatos específicos y se relaciona con `archivo` mediante `id_archivo`.

La separación entre File Service y Photo Service permite mantener una responsabilidad clara: el primero administra el archivo como recurso general y el segundo proporciona funcionalidades especializadas para los recursos de tipo imagen.