# Contratos Iniciales de Comunicación

## Objetivo

Definir de manera preliminar los principales puntos de comunicación entre los clientes y los servicios del sistema.

Los contratos podrán evolucionar durante la implementación.

---

# Auth Service

## Inicio de sesión

Solicitud:

POST /auth/login

```json
{
  "usuario": "nombre_usuario",
  "password": "contraseña"
}


File Service
Subir archivo

POST /files

Consultar archivos

GET /files

Descargar archivo

GET /files/{id}

Eliminar archivo

DELETE /files/{id}

HPC Service
Crear trabajo

POST /jobs

Consultar trabajo

GET /jobs/{id}

Listar trabajos

GET /jobs

Monitoring Service
Estado general

GET /monitoring/status

Estado de nodos

GET /monitoring/nodes

Métricas del sistema

GET /monitoring/metrics