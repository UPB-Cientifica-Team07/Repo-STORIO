# API Gateway - Contrato de Comunicación

## 1. Objetivo

Definir el contrato de comunicación del API Gateway de UPB-CIENTÍFICA.

El API Gateway constituye el punto principal de entrada para los clientes de la plataforma y centraliza el acceso a los servicios distribuidos.

Su responsabilidad es recibir las solicitudes externas, identificar el servicio correspondiente y dirigir la comunicación hacia dicho servicio utilizando el protocolo definido por la arquitectura.

El API Gateway no concentra la lógica de negocio de los servicios.

---

## 2. Alcance

Este documento define:

- La comunicación entre los clientes y el API Gateway.
- Los recursos principales expuestos inicialmente.
- El formato general de solicitudes y respuestas.
- El mecanismo inicial de autenticación.
- El direccionamiento hacia los servicios internos.
- Los códigos de respuesta HTTP.
- Las reglas generales de comunicación.

Este documento no define completamente los contratos internos de cada servicio. Estos se encuentran documentados en archivos independientes.

---

## 3. Modelo de comunicación

El API Gateway funciona como intermediario entre los clientes y los servicios internos.

```text
                    Cliente Web
                         │
                         │
                    Cliente Desktop
                         │
                         │
                    Cliente Mobile
                         │
                         │ HTTPS / REST
                         ▼
                    API Gateway
                         │
        ┌────────────────┼────────────────────┐
        │                │                    │
        ▼                ▼                    ▼
   Auth Service     File Service        Photo Service
        │                │                    │
      RMI              REST              REST


        ┌────────────────┼────────────────────┐
        │                │                    │
        ▼                ▼                    ▼
Streaming Service    HPC Service    Monitoring Service
        │                │                    │
      SOAP              REST              REST
```

El flujo general será:

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ Comunicación interna
   ▼
Servicio correspondiente
```

---

## 4. Clientes soportados

Inicialmente, el API Gateway será utilizado por los siguientes clientes:

- Cliente Web.
- Cliente Desktop.
- Cliente Mobile.

Todos los clientes utilizarán HTTPS y una interfaz basada en REST para las operaciones generales de la plataforma.

```text
Cliente Web ────────┐
                    │
Cliente Desktop ────┼── HTTPS / REST ──► API Gateway
                    │
Cliente Mobile ─────┘
```

La sincronización de archivos constituye una excepción, ya que utiliza una conexión TCP directa con el Sync Service.

---

## 5. Protocolo de comunicación externa

La comunicación entre los clientes y el API Gateway utilizará:

| Elemento | Tecnología |
|---|---|
| Protocolo de transporte | HTTPS |
| Protocolo de aplicación | HTTP |
| Estilo de API | REST |
| Formato principal | JSON |
| Codificación | UTF-8 |

Las solicitudes relacionadas con carga de archivos podrán utilizar formatos compatibles con transferencia de archivos, como:

```text
multipart/form-data
```

---

## 6. Estructura general de la API

Los recursos principales estarán organizados inicialmente bajo las siguientes rutas:

```text
/auth
/users
/files
/photos
/streaming
/hpc
/monitoring
```

La estructura general será:

| Recurso | Servicio responsable |
|---|---|
| `/auth` | Auth Service |
| `/users` | Auth Service |
| `/files` | File Service |
| `/photos` | Photo Service |
| `/streaming` | Streaming Service |
| `/hpc` | HPC Service |
| `/monitoring` | Monitoring Service |

El API Gateway utilizará estas rutas para identificar el servicio responsable de procesar cada solicitud.

---

## 7. Direccionamiento de solicitudes

El flujo de una solicitud general será:

```text
1. El cliente realiza una solicitud HTTPS.
           │
           ▼
2. El API Gateway recibe la solicitud.
           │
           ▼
3. El Gateway identifica el recurso solicitado.
           │
           ▼
4. Se determina el servicio responsable.
           │
           ▼
5. El Gateway establece la comunicación interna.
           │
           ▼
6. El servicio procesa la solicitud.
           │
           ▼
7. El resultado retorna al Gateway.
           │
           ▼
8. El Gateway responde al cliente.
```

Representación:

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │
   ▼
Servicio
   │
   │ Resultado
   ▼
API Gateway
   │
   │ HTTPS
   ▼
Cliente
```

---

## 8. Rutas iniciales

### 8.1 Autenticación

```text
POST /auth/login
POST /auth/validate
```

Estas solicitudes serán procesadas mediante el Auth Service.

La comunicación interna utilizará Java RMI.

```text
API Gateway
      │
      │ Java RMI
      ▼
Auth Service
```

---

### 8.2 Usuarios

```text
GET /users/{id}
```

Inicialmente, la información relacionada con usuarios será gestionada por el Auth Service.

---

### 8.3 Archivos

```text
GET    /files
POST   /files
GET    /files/{id}
GET    /files/{id}/download
DELETE /files/{id}
```

Estas solicitudes serán dirigidas al File Service.

```text
API Gateway
      │
      │ REST
      ▼
File Service
```

---

### 8.4 Imágenes

```text
GET /photos
GET /photos/{id}
GET /albums
GET /albums/{id}
```

Estas solicitudes serán dirigidas al Photo Service.

```text
API Gateway
      │
      │ REST
      ▼
Photo Service
```

---

### 8.5 Streaming

```text
GET /streaming
GET /streaming/{id}
```

El API Gateway realizará la comunicación interna con el Streaming Service mediante SOAP.

```text
API Gateway
      │
      │ SOAP
      ▼
Streaming Service
```

---

### 8.6 Trabajos HPC

```text
GET    /hpc/jobs
POST   /hpc/jobs
GET    /hpc/jobs/{id}
DELETE /hpc/jobs/{id}
```

Estas solicitudes serán dirigidas al HPC Service.

```text
API Gateway
      │
      │ REST
      ▼
HPC Service
```

---

### 8.7 Monitoreo

```text
GET /monitoring
GET /monitoring/services
GET /monitoring/nodes
```

Estas solicitudes serán dirigidas al Monitoring Service.

```text
API Gateway
      │
      │ REST
      ▼
Monitoring Service
```

---

## 9. Autenticación de solicitudes

Las operaciones protegidas deberán incluir un token de acceso.

Inicialmente, el formato será:

```text
Authorization: Bearer <token>
```

El API Gateway recibirá el token desde el cliente y realizará la validación utilizando el mecanismo definido con el Auth Service.

El flujo conceptual será:

```text
Cliente
   │
   │ Authorization: Bearer <token>
   ▼
API Gateway
   │
   │ validarToken(...)
   │ Java RMI
   ▼
Auth Service
```

El token no debe ser almacenado ni expuesto innecesariamente en registros del sistema.

---

## 10. Formato general de respuestas

Las respuestas REST del API Gateway utilizarán una estructura consistente.

### 10.1 Respuesta exitosa

```json
{
  "success": true,
  "message": "Operación realizada correctamente",
  "data": {}
}
```

### 10.2 Respuesta con error

```json
{
  "success": false,
  "message": "Descripción del error",
  "data": null
}
```

El contenido específico del campo `data` dependerá de cada operación.

---

## 11. Códigos HTTP

El API Gateway utilizará códigos HTTP para representar el resultado general de las operaciones.

| Código | Nombre | Uso |
|---|---|---|
| 200 | OK | Operación realizada correctamente |
| 201 | Created | Recurso creado |
| 204 | No Content | Operación realizada sin contenido de respuesta |
| 400 | Bad Request | Solicitud inválida |
| 401 | Unauthorized | Usuario no autenticado |
| 403 | Forbidden | Usuario sin permisos |
| 404 | Not Found | Recurso no encontrado |
| 409 | Conflict | Conflicto con el estado actual |
| 500 | Internal Server Error | Error interno |
| 503 | Service Unavailable | Servicio no disponible |

---

## 12. Manejo de errores internos

Los clientes no deben depender directamente de las excepciones o errores internos generados por los servicios distribuidos.

El API Gateway será responsable de transformar estos errores en respuestas HTTP apropiadas.

Ejemplo:

```text
Cliente
   │
   │ Solicitud
   ▼
API Gateway
   │
   ▼
Servicio interno
   │
   │ Error interno
   ▼
API Gateway
   │
   │ HTTP 500 / 503
   ▼
Cliente
```

Los detalles técnicos internos no deben exponerse directamente al cliente.

---

## 13. Disponibilidad de servicios

El API Gateway depende de la disponibilidad de los servicios internos.

Inicialmente, se contemplan los siguientes escenarios:

| Situación | Respuesta esperada |
|---|---|
| Servicio disponible | Procesar solicitud |
| Servicio temporalmente no disponible | 503 Service Unavailable |
| Error de comunicación | 503 o 500 según el contexto |
| Tiempo de espera | Error controlado |
| Error interno del servicio | Respuesta controlada |

La implementación de mecanismos como reintentos, circuit breakers o balanceo de carga podrá incorporarse posteriormente.

---

## 14. Separación entre comunicación externa e interna

La arquitectura mantiene dos niveles de comunicación.

### Comunicación externa

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
```

### Comunicación interna

```text
API Gateway
   │
   ├── Java RMI ──► Auth Service
   │
   ├── REST ──────► File Service
   │
   ├── REST ──────► Photo Service
   │
   ├── SOAP ──────► Streaming Service
   │
   ├── REST ──────► HPC Service
   │
   └── REST ──────► Monitoring Service
```

Esta separación evita que los clientes dependan directamente de la ubicación o implementación de los servicios internos.

---

## 15. Excepción: Sync Service

El Sync Service no forma parte del flujo general de solicitudes REST.

Los clientes Desktop y Mobile podrán establecer una conexión directa con este servicio mediante TCP Sockets para realizar operaciones de sincronización.

```text
Cliente Desktop / Mobile
           │
           │ TCP Socket
           ▼
      Sync Service
```

Esta comunicación estará limitada a las operaciones relacionadas con sincronización de archivos.

El protocolo específico se encuentra definido en:

```text
docs/api/sync-protocol.md
```

---

## 16. Responsabilidades del API Gateway

El API Gateway será responsable de:

- Recibir solicitudes externas.
- Centralizar el acceso a la plataforma.
- Identificar el servicio responsable.
- Direccionar solicitudes.
- Gestionar la autenticación de solicitudes protegidas.
- Transformar errores internos en respuestas HTTP.
- Retornar las respuestas a los clientes.

El API Gateway no será responsable de:

- Gestionar directamente usuarios.
- Almacenar archivos.
- Procesar imágenes.
- Ejecutar trabajos HPC.
- Realizar procesamiento paralelo.
- Mantener métricas internas.
- Implementar la lógica de negocio específica de los servicios.

---

## 17. Reglas del contrato

La comunicación a través del API Gateway deberá cumplir las siguientes reglas:

1. Las operaciones generales ingresan mediante HTTPS.
2. Los clientes utilizan la API REST expuesta por el Gateway.
3. Los servicios internos no se exponen directamente como punto principal de acceso.
4. Cada recurso se dirige al servicio responsable.
5. Los errores internos deben ser controlados.
6. Las respuestas deben mantener una estructura consistente.
7. Las operaciones protegidas requieren autenticación.
8. Los secretos y credenciales no deben ser expuestos en las respuestas.
9. La comunicación interna utiliza el protocolo definido para cada servicio.
10. El Sync Service mantiene su comunicación TCP independiente.

---

## 18. Resumen

El API Gateway representa la capa de entrada principal de UPB-CIENTÍFICA.

```text
                    CLIENTES
                        │
                        │ HTTPS / REST
                        ▼
                   API GATEWAY
                        │
        ┌───────────────┼────────────────┐
        │               │                │
        ▼               ▼                ▼
   Auth Service    File Service     Photo Service
      Java RMI         REST              REST

        │               │                │
        ▼               ▼                ▼
Streaming Service   HPC Service   Monitoring Service
      SOAP              REST              REST
```

El Gateway centraliza la comunicación externa y permite integrar diferentes tecnologías de comunicación interna sin trasladar esa complejidad a los clientes.

Los contratos específicos de cada servicio se documentan de forma independiente dentro de:

```text
docs/api/
```