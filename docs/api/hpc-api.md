# HPC Service - Contrato REST

## 1. Objetivo

Definir el contrato de comunicación del HPC Service de UPB-CIENTÍFICA.

El HPC Service es responsable de gestionar los trabajos de computación de alto rendimiento enviados por los usuarios de la plataforma. El servicio permite registrar, consultar, ejecutar y supervisar trabajos que requieren procesamiento paralelo dentro del Cluster HPC.

La comunicación principal con el servicio se realiza mediante REST, mientras que la ejecución distribuida de los procesos dentro del cluster utiliza OpenMPI.

---

# 2. Alcance

Este documento define:

- La responsabilidad del HPC Service.
- La comunicación mediante REST.
- La creación de trabajos HPC.
- La consulta de trabajos.
- La consulta de nodos HPC.
- La cancelación de trabajos.
- La relación con PostgreSQL.
- La comunicación con el Cluster HPC.
- El uso de OpenMPI.
- Los estados de ejecución.
- El manejo de errores.
- Las reglas generales del contrato.

Este documento no define todavía:

- Un planificador avanzado de trabajos.
- Priorización automática.
- Balanceo dinámico avanzado.
- Contenedores por trabajo.
- Ejecución en múltiples clusters geográficamente distribuidos.
- Recuperación automática ante fallos de nodos.
- Cuotas avanzadas de recursos.
- Integración con sistemas como SLURM o Kubernetes.

Estas funcionalidades podrán incorporarse posteriormente.

---

# 3. Modelo de comunicación

La comunicación principal sigue la siguiente arquitectura:

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ REST
   ▼
HPC Service
   │
   ├───────────────┐
   │               │
   ▼               ▼
PostgreSQL      Cluster HPC
                    │
                    │ OpenMPI
                    ▼
              Nodos HPC
```

El API Gateway funciona como punto principal de acceso para los clientes.

El HPC Service administra la lógica relacionada con los trabajos y coordina su ejecución dentro del Cluster HPC.

---

# 4. Tecnología de comunicación

| Componente | Tecnología |
|---|---|
| Cliente → API Gateway | HTTPS / REST |
| API Gateway → HPC Service | REST |
| HPC Service → PostgreSQL | Acceso a datos |
| HPC Service → Cluster HPC | Ejecución de procesos |
| Cluster HPC | OpenMPI |
| Nodos del cluster | Comunicación MPI |

El procesamiento paralelo se realizará mediante OpenMPI.

---

# 5. Responsabilidades del HPC Service

El HPC Service será responsable de:

- Registrar trabajos HPC.
- Asociar trabajos a usuarios.
- Consultar el estado de los trabajos.
- Coordinar la ejecución.
- Seleccionar o asignar recursos disponibles.
- Consultar nodos del Cluster HPC.
- Actualizar el estado de ejecución.
- Registrar el inicio y finalización de los trabajos.
- Cancelar trabajos cuando sea posible.

El HPC Service no será responsable directamente de:

- La autenticación principal de usuarios.
- La generación de tokens.
- La gestión general de archivos.
- La sincronización mediante TCP.
- La reproducción de contenido multimedia.
- El monitoreo general de toda la infraestructura.

---

# 6. Recursos principales

Los recursos REST iniciales son:

```text
POST   /hpc/jobs
GET    /hpc/jobs
GET    /hpc/jobs/{id}
DELETE /hpc/jobs/{id}

GET    /hpc/nodes
GET    /hpc/nodes/{id}
```

Estos endpoints representan el contrato inicial del servicio.

---

# 7. Autenticación y autorización

Las operaciones del HPC Service requieren un usuario autenticado.

El flujo general es:

```text
Cliente
   │
   │ Authorization: Bearer <token>
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
   │ REST
   ▼
HPC Service
```

El HPC Service debe recibir la información necesaria para identificar al usuario que solicita la operación.

Un usuario solo podrá consultar y administrar los trabajos sobre los cuales tenga autorización.

---

# 8. Operación: crear trabajo HPC

## 8.1 Endpoint

```text
POST /hpc/jobs
```

Permite registrar un nuevo trabajo para su ejecución dentro del Cluster HPC.

---

## 8.2 Solicitud

Ejemplo:

```json
{
  "descripcion": "Simulación de procesamiento distribuido",
  "lenguaje": "MPI"
}
```

Inicialmente, el campo `lenguaje` representa el mecanismo o entorno principal asociado al trabajo.

---

## 8.3 Flujo

```text
Cliente
   │
   │ POST /hpc/jobs
   ▼
API Gateway
   │
   ▼
HPC Service
   │
   ├── Validar usuario
   │
   ├── Crear trabajo
   ▼
PostgreSQL
   │
   │ estado = PENDIENTE
   ▼
HPC Service
```

El trabajo inicialmente queda registrado con el estado:

```text
PENDIENTE
```

---

## 8.4 Respuesta

Código:

```text
201 Created
```

Ejemplo:

```json
{
  "success": true,
  "message": "Trabajo HPC creado correctamente",
  "data": {
    "idJob": "uuid",
    "estado": "PENDIENTE",
    "descripcion": "Simulación de procesamiento distribuido",
    "lenguaje": "MPI"
  }
}
```

---

# 9. Operación: listar trabajos HPC

## 9.1 Endpoint

```text
GET /hpc/jobs
```

Permite consultar los trabajos asociados al usuario autenticado.

---

## 9.2 Flujo

```text
Cliente
   │
   │ GET /hpc/jobs
   ▼
API Gateway
   │
   ▼
HPC Service
   │
   │ Consultar trabajos
   ▼
PostgreSQL
   │
   └── trabajo_hpc
```

---

## 9.3 Respuesta

```json
{
  "success": true,
  "message": "Trabajos obtenidos correctamente",
  "data": [
    {
      "idJob": "uuid",
      "descripcion": "Simulación inicial de prueba",
      "lenguaje": "MPI",
      "estado": "FINALIZADO",
      "inicio": "2026-08-16T10:00:00",
      "fin": "2026-08-16T10:15:00"
    },
    {
      "idJob": "uuid",
      "descripcion": "Procesamiento distribuido en ejecución",
      "lenguaje": "MPI",
      "estado": "EJECUTANDO",
      "inicio": "2026-08-16T11:00:00",
      "fin": null
    }
  ]
}
```

---

# 10. Operación: consultar trabajo

## 10.1 Endpoint

```text
GET /hpc/jobs/{id}
```

Permite consultar la información detallada de un trabajo HPC.

---

## 10.2 Parámetro

| Parámetro | Tipo | Descripción |
|---|---|---|
| id | UUID | Identificador del trabajo |

---

## 10.3 Flujo

```text
Cliente
   │
   │ GET /hpc/jobs/{id}
   ▼
API Gateway
   │
   ▼
HPC Service
   │
   ├── Validar usuario
   │
   ├── Consultar trabajo
   ▼
PostgreSQL
   │
   ▼
Información del trabajo
```

---

## 10.4 Respuesta

```json
{
  "success": true,
  "message": "Trabajo encontrado",
  "data": {
    "idJob": "uuid",
    "descripcion": "Procesamiento distribuido",
    "lenguaje": "MPI",
    "estado": "EJECUTANDO",
    "idNodo": "uuid",
    "inicio": "2026-08-16T11:00:00",
    "fin": null
  }
}
```

---

# 11. Operación: cancelar trabajo

## 11.1 Endpoint

```text
DELETE /hpc/jobs/{id}
```

Permite solicitar la cancelación de un trabajo.

La cancelación depende del estado actual del trabajo.

---

## 11.2 Flujo

```text
Cliente
   │
   │ DELETE /hpc/jobs/{id}
   ▼
API Gateway
   │
   ▼
HPC Service
   │
   ├── Validar usuario
   │
   ├── Consultar estado
   │
   ├── Cancelar ejecución
   │
   ▼
Cluster HPC
   │
   ▼
Actualizar estado
   │
   ▼
PostgreSQL
```

---

## 11.3 Respuesta

```json
{
  "success": true,
  "message": "Trabajo cancelado correctamente",
  "data": {
    "idJob": "uuid",
    "estado": "CANCELADO"
  }
}
```

---

# 12. Operación: listar nodos HPC

## 12.1 Endpoint

```text
GET /hpc/nodes
```

Permite consultar los nodos registrados en el Cluster HPC.

---

## 12.2 Flujo

```text
Cliente
   │
   │ GET /hpc/nodes
   ▼
API Gateway
   │
   ▼
HPC Service
   │
   │ Consultar nodos
   ▼
PostgreSQL
   │
   └── nodo_hpc
```

---

## 12.3 Respuesta

```json
{
  "success": true,
  "message": "Nodos obtenidos correctamente",
  "data": [
    {
      "idNodo": "uuid",
      "hostname": "hpc-node-01",
      "cpu": 16,
      "memoria": 32768,
      "estado": "DISPONIBLE"
    },
    {
      "idNodo": "uuid",
      "hostname": "hpc-node-02",
      "cpu": 32,
      "memoria": 65536,
      "estado": "OCUPADO"
    }
  ]
}
```

---

# 13. Operación: consultar nodo

## 13.1 Endpoint

```text
GET /hpc/nodes/{id}
```

Permite consultar la información de un nodo específico.

---

## 13.2 Respuesta

```json
{
  "success": true,
  "message": "Nodo encontrado",
  "data": {
    "idNodo": "uuid",
    "hostname": "hpc-node-01",
    "cpu": 16,
    "memoria": 32768,
    "estado": "DISPONIBLE"
  }
}
```

---

# 14. Estados de un trabajo HPC

Los trabajos pueden tener los siguientes estados iniciales:

```text
PENDIENTE
EJECUTANDO
FINALIZADO
ERROR
CANCELADO
```

El flujo general es:

```text
             ┌─────────────┐
             │  PENDIENTE  │
             └──────┬──────┘
                    │
                    ▼
             ┌─────────────┐
             │ EJECUTANDO  │
             └──────┬──────┘
                    │
          ┌─────────┼──────────┐
          │         │          │
          ▼         ▼          ▼
     FINALIZADO    ERROR   CANCELADO
```

Un trabajo también puede ser cancelado antes de iniciar su ejecución.

---

# 15. Entidad trabajo_hpc

El HPC Service trabaja principalmente con la entidad:

```text
trabajo_hpc
```

Los campos principales son:

| Campo | Tipo | Descripción |
|---|---|---|
| id_job | UUID | Identificador del trabajo |
| id_usuario | UUID | Usuario propietario |
| id_nodo | UUID | Nodo asignado |
| descripcion | TEXT | Descripción del trabajo |
| lenguaje | VARCHAR | Entorno o tecnología |
| estado | VARCHAR | Estado actual |
| inicio | TIMESTAMP | Inicio de ejecución |
| fin | TIMESTAMP | Finalización |

---

# 16. Entidad nodo_hpc

El Cluster HPC está representado inicialmente mediante:

```text
nodo_hpc
```

Campos:

| Campo | Tipo | Descripción |
|---|---|---|
| id_nodo | UUID | Identificador |
| hostname | VARCHAR | Nombre del nodo |
| cpu | INTEGER | Número de núcleos |
| memoria | INTEGER | Memoria RAM |
| estado | VARCHAR | Estado del nodo |

Estados iniciales:

```text
DISPONIBLE
OCUPADO
FUERA_DE_SERVICIO
```

---

# 17. Relación entre usuario, trabajo y nodo

El modelo conceptual es:

```text
usuario
   │
   │ 1
   ▼
trabajo_hpc
   │
   │
   ▼
nodo_hpc
```

Un usuario puede registrar múltiples trabajos.

Un nodo puede ejecutar diferentes trabajos en distintos momentos, según su disponibilidad.

---

# 18. Comunicación con el Cluster HPC

Una vez que un trabajo ha sido preparado para su ejecución, el HPC Service coordina su envío al Cluster HPC.

Modelo conceptual:

```text
HPC Service
     │
     │ Preparar trabajo
     ▼
Cluster HPC
     │
     │ mpirun
     ▼
┌─────────────┬─────────────┐
│             │             │
▼             ▼             ▼
Nodo 1       Nodo 2       Nodo N
│             │             │
└─────────────┴─────────────┘
              │
              ▼
          Resultado
```

OpenMPI permite distribuir la ejecución entre los procesos participantes.

---

# 19. Ejecución con OpenMPI

El modelo conceptual de ejecución es:

```text
HPC Service
     │
     │ Registrar trabajo
     ▼
PostgreSQL
     │
     │ PENDIENTE
     ▼
HPC Service
     │
     │ Asignar recursos
     ▼
Cluster HPC
     │
     │ mpirun
     ▼
Procesos MPI
     │
     ├── Proceso 0
     ├── Proceso 1
     ├── Proceso 2
     └── Proceso N
     │
     ▼
Resultado
```

Durante la ejecución, el estado del trabajo debe actualizarse.

---

# 20. Actualización de estados

El ciclo básico de un trabajo es:

```text
POST /hpc/jobs
       │
       ▼
   PENDIENTE
       │
       ▼
  EJECUTANDO
       │
       ├──────────────► ERROR
       │
       ├──────────────► CANCELADO
       │
       ▼
   FINALIZADO
```

Cada cambio importante debe registrarse en la base de datos.

---

# 21. Relación con PostgreSQL

El HPC Service utiliza PostgreSQL principalmente para:

```text
trabajo_hpc
nodo_hpc
usuario
```

Modelo:

```text
HPC Service
     │
     ▼
PostgreSQL
     │
     ├── usuario
     │
     ├── trabajo_hpc
     │
     └── nodo_hpc
```

La base de datos almacena el estado administrativo de los trabajos y los nodos.

La ejecución real ocurre fuera de PostgreSQL, dentro del Cluster HPC.

---

# 22. Manejo de errores

El servicio debe responder de forma controlada ante errores.

| Situación | Código HTTP |
|---|---|
| Solicitud inválida | 400 |
| Usuario no autenticado | 401 |
| Acceso denegado | 403 |
| Trabajo no encontrado | 404 |
| Nodo no encontrado | 404 |
| Estado inválido | 409 |
| Error interno | 500 |
| Cluster no disponible | 503 |

Ejemplo:

```json
{
  "success": false,
  "message": "Trabajo HPC no encontrado",
  "data": null
}
```

---

# 23. Validación de operaciones

Antes de ejecutar una operación, el HPC Service debe validar:

1. La identidad del usuario.
2. Los permisos sobre el trabajo.
3. La existencia del trabajo.
4. El estado actual.
5. La disponibilidad de recursos cuando sea necesaria.
6. La existencia del nodo asignado.
7. La disponibilidad del Cluster HPC.

Por ejemplo, un trabajo finalizado no debe volver a ejecutarse automáticamente mediante una operación de cancelación.

---

# 24. Flujo completo de creación y ejecución

```text
Cliente
   │
   │ POST /hpc/jobs
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
   │ REST
   ▼
HPC Service
   │
   │ Crear trabajo
   ▼
PostgreSQL
   │
   │ PENDIENTE
   ▼
HPC Service
   │
   │ Asignar recursos
   ▼
Cluster HPC
   │
   │ OpenMPI
   ▼
Nodos HPC
   │
   │ Ejecutar procesos
   ▼
Resultado
   │
   ▼
HPC Service
   │
   ▼
PostgreSQL
   │
   │ FINALIZADO
   ▼
Cliente
```

---

# 25. Flujo de consulta de estado

```text
Cliente
   │
   │ GET /hpc/jobs/{id}
   ▼
API Gateway
   │
   ▼
HPC Service
   │
   │ Consultar trabajo
   ▼
PostgreSQL
   │
   ▼
HPC Service
   │
   ▼
API Gateway
   │
   │ JSON
   ▼
Cliente
```

---

# 26. Seguridad

Las reglas principales son:

1. El usuario debe estar autenticado.
2. Un usuario solo puede administrar sus propios trabajos, salvo roles autorizados.
3. La ejecución debe validar los parámetros recibidos.
4. No se deben ejecutar comandos arbitrarios enviados directamente por el cliente.
5. La información interna del Cluster HPC no debe exponerse innecesariamente.
6. Los errores internos deben manejarse de forma controlada.
7. La comunicación externa debe realizarse mediante HTTPS.
8. Los cambios de estado deben registrarse de forma consistente.

---

# 27. Estructura inicial propuesta

```text
services/
└── hpc-service/
    │
    ├── api/
    │   ├── JobController.java
    │   └── NodeController.java
    │
    ├── service/
    │   ├── JobService.java
    │   ├── NodeService.java
    │   └── ClusterService.java
    │
    ├── repository/
    │   ├── JobRepository.java
    │   └── NodeRepository.java
    │
    ├── model/
    │   ├── TrabajoHpc.java
    │   └── NodoHpc.java
    │
    └── cluster/
        └── MpiExecutor.java
```

La estructura definitiva puede ajustarse durante la implementación, manteniendo la separación entre:

```text
API
Lógica de negocio
Persistencia
Comunicación con el Cluster HPC
```

---

# 28. Reglas generales del contrato

El HPC Service debe cumplir las siguientes reglas:

1. El acceso externo se realiza mediante el API Gateway.
2. La comunicación principal utiliza REST.
3. Las solicitudes y respuestas utilizan JSON.
4. Los trabajos pertenecen a un usuario.
5. Cada trabajo tiene un estado definido.
6. La información administrativa se almacena en PostgreSQL.
7. La ejecución paralela utiliza OpenMPI.
8. El Cluster HPC se mantiene separado de la API pública.
9. Los permisos deben validarse antes de consultar o modificar un trabajo.
10. Las operaciones inválidas deben generar respuestas controladas.
11. La ejecución de trabajos no debe permitir comandos arbitrarios desde el cliente.
12. El servicio debe verificar la disponibilidad de recursos antes de iniciar una ejecución.

---

# 29. Resumen

El HPC Service proporciona la interfaz distribuida para gestionar trabajos de computación de alto rendimiento dentro de UPB-CIENTÍFICA.

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ REST
   ▼
HPC Service
   │
   ├──────────────┐
   │              │
   ▼              ▼
PostgreSQL     Cluster HPC
                   │
                   │ OpenMPI
                   ▼
                Nodos
```

Los endpoints iniciales son:

```text
POST   /hpc/jobs
GET    /hpc/jobs
GET    /hpc/jobs/{id}
DELETE /hpc/jobs/{id}

GET    /hpc/nodes
GET    /hpc/nodes/{id}
```

El servicio administra los estados:

```text
PENDIENTE
EJECUTANDO
FINALIZADO
ERROR
CANCELADO
```

La ejecución distribuida se realiza dentro del Cluster HPC utilizando OpenMPI, mientras que PostgreSQL mantiene la información administrativa de los trabajos y nodos.

Este contrato constituye la base inicial para implementar el servicio HPC y su integración con el resto de la arquitectura distribuida de UPB-CIENTÍFICA.