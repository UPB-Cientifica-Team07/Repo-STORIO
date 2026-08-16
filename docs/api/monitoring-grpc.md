# Monitoring Service - Contrato gRPC

## 1. Objetivo

Definir el contrato de comunicación del Monitoring Service de UPB-CIENTÍFICA.

El Monitoring Service es responsable de recopilar, consultar y centralizar información relacionada con el estado y rendimiento de los servicios distribuidos y los nodos que componen la plataforma.

La comunicación interna para el monitoreo se realizará mediante gRPC, utilizando Protocol Buffers como mecanismo de definición y serialización de los mensajes.

---

# 2. Alcance

Este documento define:

- La responsabilidad del Monitoring Service.
- La comunicación mediante gRPC.
- El uso de Protocol Buffers.
- El registro de métricas.
- La consulta del estado de servicios.
- La consulta del estado de nodos.
- La comunicación entre componentes.
- La estructura conceptual del archivo `.proto`.
- Los mensajes principales.
- El manejo básico de errores.
- Las reglas generales del contrato.

Este documento no define todavía:

- Un sistema completo de observabilidad.
- Monitoreo distribuido geográfico.
- Análisis predictivo de fallos.
- Trazabilidad distribuida completa.
- Integración con Prometheus.
- Integración con Grafana.
- Alertas automáticas externas.
- Recuperación automática de servicios.

Estas funcionalidades podrán agregarse posteriormente.

---

# 3. Modelo de comunicación

El Monitoring Service recibe información desde los diferentes servicios y nodos de la arquitectura.

```text
                  API Gateway
                       │
                       │
                       ▼
                 Monitoring Service
                       ▲
                       │
                      gRPC
                       │
        ┌──────────────┼──────────────┐
        │              │              │
        ▼              ▼              ▼
   Auth Service    File Service    Sync Service
        │              │              │
        └──────────────┼──────────────┘
                       │
                       ▼
                  HPC Service
                       │
                       ▼
                   Cluster HPC
                       │
                       ▼
                    Nodos HPC
```

Cada componente puede enviar información sobre su estado al Monitoring Service mediante gRPC.

---

# 4. Tecnología de comunicación

| Elemento | Tecnología |
|---|---|
| Comunicación interna | gRPC |
| Definición de contrato | Protocol Buffers |
| Serialización | Protobuf |
| Implementación principal | Go |
| Persistencia inicial | PostgreSQL |
| Comunicación externa | REST mediante API Gateway |

gRPC se utiliza principalmente para la comunicación interna entre servicios debido a su modelo basado en contratos y a la serialización eficiente mediante Protocol Buffers.

---

# 5. Responsabilidades del Monitoring Service

El Monitoring Service será responsable de:

- Recibir métricas de los servicios.
- Registrar el estado de los componentes.
- Consultar información de los nodos.
- Identificar servicios disponibles o no disponibles.
- Centralizar información básica de monitoreo.
- Exponer información de monitoreo a través del API Gateway.
- Mantener información actualizada sobre el estado general del sistema.

El Monitoring Service no será responsable directamente de:

- Autenticar usuarios.
- Ejecutar trabajos HPC.
- Gestionar archivos.
- Sincronizar archivos.
- Almacenar contenido multimedia.
- Reiniciar automáticamente servicios.
- Ejecutar mecanismos de recuperación.

---

# 6. Componentes monitoreados

Inicialmente, el sistema puede monitorear los siguientes componentes:

```text
API Gateway
Auth Service
File Service
Sync Service
Photo Service
Streaming Service
HPC Service
Monitoring Service
Cluster HPC
Nodos HPC
PostgreSQL
```

Cada componente puede reportar su estado y métricas básicas.

---

# 7. Modelo de métricas

Las métricas iniciales pueden incluir:

- Estado del servicio.
- Uso de CPU.
- Uso de memoria.
- Número de solicitudes.
- Tiempo de respuesta.
- Número de conexiones activas.
- Estado de disponibilidad.
- Número de trabajos activos.
- Estado de nodos HPC.

Modelo conceptual:

```text
Componente
    │
    ├── nombre
    ├── estado
    ├── cpu
    ├── memoria
    ├── solicitudes
    └── timestamp
```

---

# 8. Comunicación mediante gRPC

El flujo general será:

```text
Servicio o Nodo
       │
       │ gRPC
       ▼
Monitoring Service
       │
       │ Registrar información
       ▼
PostgreSQL
```

Cada servicio puede enviar periódicamente información sobre su estado.

El modelo inicial se basa en llamadas gRPC directas hacia el Monitoring Service.

---

# 9. Servicio gRPC propuesto

El contrato inicial puede definirse mediante el servicio:

```text
MonitoringService
```

Las operaciones principales serán:

```text
ReportMetrics
ReportStatus
GetServiceStatus
GetNodeStatus
```

---

# 10. Estructura conceptual del archivo .proto

El archivo principal se ubicará inicialmente en:

```text
services/
└── monitoring-service/
    └── proto/
        └── monitoring.proto
```

La estructura conceptual será:

```proto
syntax = "proto3";

package monitoring;

service MonitoringService {

    rpc ReportMetrics (MetricsRequest)
        returns (MetricsResponse);

    rpc ReportStatus (StatusRequest)
        returns (StatusResponse);

    rpc GetServiceStatus (ServiceRequest)
        returns (ServiceResponse);

    rpc GetNodeStatus (NodeRequest)
        returns (NodeResponse);
}
```

Esta estructura representa el contrato inicial del servicio.

---

# 11. Mensaje MetricsRequest

El mensaje `MetricsRequest` permite que un servicio o nodo envíe sus métricas al Monitoring Service.

Modelo conceptual:

```proto
message MetricsRequest {

    string component_id = 1;

    string component_name = 2;

    double cpu_usage = 3;

    double memory_usage = 4;

    int64 active_connections = 5;

    int64 total_requests = 6;

    int64 timestamp = 7;
}
```

---

# 12. Mensaje MetricsResponse

La respuesta confirma la recepción de las métricas.

```proto
message MetricsResponse {

    bool success = 1;

    string message = 2;
}
```

Ejemplo conceptual:

```text
success = true

message = "Metrics registered successfully"
```

---

# 13. Operación ReportMetrics

La operación permite registrar métricas de un componente.

Modelo:

```text
Servicio
   │
   │ ReportMetrics()
   ▼
Monitoring Service
   │
   │ Registrar métricas
   ▼
PostgreSQL
```

Los servicios pueden enviar información periódicamente.

Ejemplos de componentes:

```text
gateway
auth-service
file-service
sync-service
photo-service
streaming-service
hpc-service
hpc-node-01
```

---

# 14. Mensaje StatusRequest

El mensaje `StatusRequest` permite reportar el estado de un servicio o nodo.

```proto
message StatusRequest {

    string component_id = 1;

    string component_name = 2;

    string status = 3;

    string message = 4;

    int64 timestamp = 5;
}
```

Los estados iniciales pueden ser:

```text
AVAILABLE
UNAVAILABLE
DEGRADED
```

---

# 15. Mensaje StatusResponse

```proto
message StatusResponse {

    bool success = 1;

    string message = 2;
}
```

Ejemplo:

```text
success = true

message = "Status registered successfully"
```

---

# 16. Operación ReportStatus

Esta operación permite que un componente informe cambios en su estado.

Flujo:

```text
Servicio
   │
   │ ReportStatus()
   ▼
Monitoring Service
   │
   │ Registrar estado
   ▼
PostgreSQL
```

Ejemplo:

```text
File Service
       │
       │ status = AVAILABLE
       ▼
Monitoring Service
```

Si un servicio detecta una condición que afecta su disponibilidad:

```text
HPC Service
       │
       │ status = DEGRADED
       ▼
Monitoring Service
```

---

# 17. Mensaje ServiceRequest

Este mensaje permite consultar el estado de un servicio.

```proto
message ServiceRequest {

    string component_name = 1;
}
```

Ejemplo:

```text
component_name = "file-service"
```

---

# 18. Mensaje ServiceResponse

```proto
message ServiceResponse {

    string component_id = 1;

    string component_name = 2;

    string status = 3;

    string message = 4;

    int64 last_updated = 5;
}
```

Ejemplo conceptual:

```text
component_name = file-service

status = AVAILABLE
```

---

# 19. Operación GetServiceStatus

Permite consultar el estado actual de un servicio.

```text
Componente
      │
      │ GetServiceStatus()
      ▼
Monitoring Service
      │
      │ Consultar información
      ▼
PostgreSQL
      │
      ▼
Estado del servicio
```

Esta operación también puede ser utilizada internamente por otros componentes autorizados.

---

# 20. Mensaje NodeRequest

Permite solicitar el estado de un nodo.

```proto
message NodeRequest {

    string node_id = 1;
}
```

---

# 21. Mensaje NodeResponse

La respuesta contiene información básica del nodo.

```proto
message NodeResponse {

    string node_id = 1;

    string hostname = 2;

    string status = 3;

    int32 cpu_cores = 4;

    int64 memory = 5;

    double cpu_usage = 6;

    double memory_usage = 7;

    int64 last_updated = 8;
}
```

---

# 22. Operación GetNodeStatus

Permite consultar el estado de un nodo HPC.

Flujo:

```text
HPC Service
      │
      │ GetNodeStatus()
      ▼
Monitoring Service
      │
      │ Consultar información
      ▼
PostgreSQL
      │
      ▼
Información del nodo
```

Ejemplo de respuesta:

```text
hostname = hpc-node-01

status = AVAILABLE

cpu_cores = 16

memory = 32768
```

---

# 23. Estados de componentes

Los estados iniciales definidos para los servicios son:

```text
AVAILABLE
UNAVAILABLE
DEGRADED
```

El significado de cada estado es:

| Estado | Descripción |
|---|---|
| AVAILABLE | El componente está funcionando correctamente |
| UNAVAILABLE | El componente no está disponible |
| DEGRADED | El componente funciona con limitaciones |

Para los nodos HPC también pueden mantenerse los estados:

```text
DISPONIBLE
OCUPADO
FUERA_DE_SERVICIO
```

La conversión entre el estado operativo del nodo y el estado general de disponibilidad puede realizarse dentro del Monitoring Service.

---

# 24. Registro de métricas

El flujo general de registro será:

```text
Servicio
   │
   │ Obtener métricas locales
   ▼
Métricas
   │
   ├── CPU
   ├── Memoria
   ├── Solicitudes
   └── Conexiones
   │
   │ gRPC
   ▼
Monitoring Service
   │
   ▼
Persistencia
```

La frecuencia de envío se definirá durante la implementación.

Inicialmente, cada servicio puede enviar métricas de forma periódica.

---

# 25. Relación con PostgreSQL

El Monitoring Service requiere almacenar información relacionada con el monitoreo.

Inicialmente, pueden agregarse entidades específicas para:

```text
servicio_estado
metrica_servicio
estado_nodo
```

Modelo conceptual:

```text
Monitoring Service
        │
        ▼
    PostgreSQL
        │
        ├── servicio_estado
        │
        ├── metrica_servicio
        │
        └── estado_nodo
```

Estas tablas no forman parte todavía del esquema inicial y deberán agregarse mediante una migración cuando se implemente el servicio.

---

# 26. Comunicación con el API Gateway

Los clientes no deben comunicarse directamente con el servidor gRPC.

El acceso externo debe seguir el flujo:

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ REST
   ▼
Monitoring Service
```

La comunicación gRPC se reserva principalmente para los servicios internos.

Ejemplo conceptual:

```text
GET /monitoring/services

GET /monitoring/services/{name}

GET /monitoring/nodes

GET /monitoring/nodes/{id}
```

Estos endpoints serán definidos posteriormente en el contrato REST correspondiente.

---

# 27. Flujo completo de reporte

```text
File Service
      │
      │ Obtener métricas
      ▼
    CPU
    Memoria
    Conexiones
      │
      │ gRPC
      ▼
Monitoring Service
      │
      │ Validar mensaje
      │
      ▼
PostgreSQL
      │
      ▼
Métricas almacenadas
```

El mismo modelo puede aplicarse a los demás servicios.

---

# 28. Flujo de monitoreo del Cluster HPC

```text
Nodo HPC
   │
   │ Métricas locales
   │
   ├── CPU
   ├── Memoria
   └── Estado
   │
   │ gRPC
   ▼
Monitoring Service
   │
   ▼
PostgreSQL
   │
   ▼
Información centralizada
```

El HPC Service también puede consultar esta información para conocer el estado general de los nodos.

---

# 29. Manejo de errores

Los errores en gRPC se gestionarán mediante códigos de estado.

Los casos iniciales pueden incluir:

| Situación | Código gRPC |
|---|---|
| Solicitud inválida | INVALID_ARGUMENT |
| Componente no encontrado | NOT_FOUND |
| Acceso no autorizado | UNAUTHENTICATED |
| Permiso insuficiente | PERMISSION_DENIED |
| Servicio no disponible | UNAVAILABLE |
| Error interno | INTERNAL |

Ejemplo conceptual:

```text
StatusCode = NOT_FOUND

Message = "Component not found"
```

---

# 30. Validación de mensajes

Antes de procesar una solicitud, el Monitoring Service debe validar:

1. Que el componente esté identificado.
2. Que los campos requeridos estén presentes.
3. Que los valores numéricos sean válidos.
4. Que el estado corresponda a uno de los valores permitidos.
5. Que la marca de tiempo sea válida.
6. Que el componente tenga autorización para reportar información.

---

# 31. Seguridad

Las reglas iniciales son:

1. gRPC se utiliza principalmente para comunicación interna.
2. Los clientes externos no acceden directamente al puerto gRPC.
3. Los servicios deben estar identificados.
4. Los mensajes deben validarse antes de almacenarse.
5. Las métricas no deben contener información sensible.
6. El acceso externo al monitoreo debe pasar por el API Gateway.
7. Los componentes no autorizados no deben poder registrar métricas.
8. La exposición de información interna debe controlarse según el tipo de usuario.

En una etapa posterior puede incorporarse:

```text
mTLS
Service Authentication
API Keys internas
Certificados
```

---

# 32. Archivo monitoring.proto propuesto

La ubicación será:

```text
services/
└── monitoring-service/
    └── proto/
        └── monitoring.proto
```

Contenido conceptual completo:

```proto
syntax = "proto3";

package monitoring;

service MonitoringService {

    rpc ReportMetrics (MetricsRequest)
        returns (MetricsResponse);

    rpc ReportStatus (StatusRequest)
        returns (StatusResponse);

    rpc GetServiceStatus (ServiceRequest)
        returns (ServiceResponse);

    rpc GetNodeStatus (NodeRequest)
        returns (NodeResponse);
}

message MetricsRequest {

    string component_id = 1;
    string component_name = 2;

    double cpu_usage = 3;
    double memory_usage = 4;

    int64 active_connections = 5;
    int64 total_requests = 6;

    int64 timestamp = 7;
}

message MetricsResponse {

    bool success = 1;
    string message = 2;
}

message StatusRequest {

    string component_id = 1;
    string component_name = 2;

    string status = 3;
    string message = 4;

    int64 timestamp = 5;
}

message StatusResponse {

    bool success = 1;
    string message = 2;
}

message ServiceRequest {

    string component_name = 1;
}

message ServiceResponse {

    string component_id = 1;
    string component_name = 2;

    string status = 3;
    string message = 4;

    int64 last_updated = 5;
}

message NodeRequest {

    string node_id = 1;
}

message NodeResponse {

    string node_id = 1;
    string hostname = 2;

    string status = 3;

    int32 cpu_cores = 4;
    int64 memory = 5;

    double cpu_usage = 6;
    double memory_usage = 7;

    int64 last_updated = 8;
}
```

Este archivo representa el contrato inicial que posteriormente podrá implementarse directamente en Go.

---

# 33. Estructura inicial propuesta

```text
services/
└── monitoring-service/
    │
    ├── cmd/
    │   └── server/
    │
    ├── internal/
    │   ├── grpc/
    │   ├── service/
    │   ├── repository/
    │   └── model/
    │
    ├── proto/
    │   └── monitoring.proto
    │
    └── generated/
```

La separación inicial será:

```text
cmd/
    Punto de entrada del servicio.

grpc/
    Implementación de los métodos gRPC.

service/
    Lógica de negocio.

repository/
    Acceso a datos.

model/
    Estructuras del dominio.

proto/
    Contrato Protocol Buffers.

generated/
    Código generado desde el archivo .proto.
```

---

# 34. Flujo general de comunicación

La arquitectura completa del monitoreo puede representarse así:

```text
                    ┌─────────────────┐
                    │   API Gateway   │
                    └────────┬────────┘
                             │
                             │ REST
                             ▼
                    ┌─────────────────┐
                    │ Monitoring      │
                    │ Service         │
                    └────────▲────────┘
                             │
                             │ gRPC
          ┌──────────────────┼──────────────────┐
          │                  │                  │
          ▼                  ▼                  ▼
     Auth Service       File Service       Sync Service

          │                  │                  │
          └──────────────────┼──────────────────┘
                             │
                             ▼
                        HPC Service
                             │
                             ▼
                         Nodos HPC
```

---

# 35. Reglas generales del contrato

El Monitoring Service debe cumplir las siguientes reglas:

1. La comunicación interna utiliza gRPC.
2. Los contratos se definen mediante Protocol Buffers.
3. Los mensajes se serializan utilizando Protobuf.
4. Los servicios pueden reportar métricas.
5. Los servicios pueden reportar cambios de estado.
6. El Monitoring Service centraliza la información recibida.
7. Los clientes externos no acceden directamente al puerto gRPC.
8. El acceso externo pasa por el API Gateway.
9. Los mensajes deben validarse antes de ser procesados.
10. Los componentes deben estar identificados.
11. Los errores deben utilizar códigos estándar de gRPC.
12. Las métricas sensibles no deben exponerse directamente.
13. La información persistente puede almacenarse en PostgreSQL.
14. El contrato `.proto` debe mantenerse versionado dentro del repositorio.
15. Cualquier cambio incompatible en Protocol Buffers debe gestionarse mediante versionado del contrato.

---

# 36. Resumen

El Monitoring Service centraliza la información básica de monitoreo de UPB-CIENTÍFICA.

La comunicación interna utiliza gRPC y Protocol Buffers:

```text
Servicios / Nodos
       │
       │ gRPC
       ▼
Monitoring Service
       │
       ▼
PostgreSQL
```

Las operaciones principales son:

```text
ReportMetrics()
ReportStatus()
GetServiceStatus()
GetNodeStatus()
```

El acceso de los usuarios a la información de monitoreo se realizará mediante el API Gateway:

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ REST
   ▼
Monitoring Service
```

El contrato principal se definirá en:

```text
services/monitoring-service/proto/monitoring.proto
```

Con este documento quedan definidos los principales contratos de comunicación de la arquitectura inicial de UPB-CIENTÍFICA.