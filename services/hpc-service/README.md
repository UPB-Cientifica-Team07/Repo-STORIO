# HPC Service

Servicio de cómputo de alto rendimiento de UPB-CIENTÍFICA.

## Tecnologías

- Java 21
- Java RMI
- OpenMPI
- PostgreSQL
- Auth Service HTTP

## Arquitectura

Cliente HPC
    |
    | RMI
    v
HPC Coordinator
    |
    +-- Auth Service
    |
    +-- PostgreSQL
    |
    +-- HpcScheduler
           |
           +-- Worker node-01
           |
           +-- Worker node-02
                    |
                    v
                  OpenMPI

## Puerto

- RMI Registry HPC: 1100
- Auth RMI utiliza el puerto 1099

## Funcionalidades implementadas

- Registro de workers mediante Java RMI.
- Heartbeats periódicos.
- Detección de nodos inactivos.
- Recuperación automática.
- Scheduler best-fit según capacidad de CPU.
- Reserva de workers durante ejecución.
- Rechazo de trabajos sin capacidad suficiente.
- Ejecución paralela mediante OpenMPI.
- Ejecución concurrente en múltiples workers.
- Autenticación mediante token del Auth Service.
- Persistencia en `nodo_hpc`.
- Persistencia en `trabajo_hpc`.
- Estados PENDIENTE, EJECUTANDO, FINALIZADO y ERROR.
- Whitelist de programas MPI permitidos.

## Compilación

Desde la raíz del repositorio:

    services/hpc-service/scripts/compile.sh

## Iniciar Coordinator

    services/hpc-service/scripts/start-coordinator.sh

## Iniciar Worker

Ejemplo node-01 con 4 CPU:

    services/hpc-service/scripts/start-worker.sh \
      127.0.0.1 1100 node-01 node-01 4

Ejemplo node-02 con 8 CPU:

    services/hpc-service/scripts/start-worker.sh \
      127.0.0.1 1100 node-02 node-02 8

## Programas MPI

Actualmente se incluyen:

- `hello_mpi.c`
- `sleep_mpi.c`

Los binarios MPI se generan durante la compilación y no se versionan.

## Variables de entorno

- `HPC_DB_URL`
- `HPC_DB_USER`
- `HPC_DB_PASSWORD`
- `HPC_AUTH_SERVICE`
- `HPC_RMI_PORT`
- `HPC_MPI_DIR`
- `POSTGRES_JDBC_JAR`

## Nodos lógicos de desarrollo

Durante las pruebas locales se pueden ejecutar varios workers lógicos
sobre una misma máquina física.

Esto permite validar el scheduler, reserva de recursos, heartbeat y
concurrencia antes del despliegue sobre máquinas físicas independientes.

En producción cada worker deberá anunciar la identidad y dirección IP
real de su nodo.
