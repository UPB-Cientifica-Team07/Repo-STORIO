# Arquitectura de Comunicación Distribuida

## 1. Objetivo

Definir los mecanismos de comunicación entre los clientes, el API Gateway y los servicios distribuidos que componen UPB-CIENTÍFICA.

El sistema utiliza diferentes tecnologías de comunicación según las necesidades funcionales de cada componente.

---

## 2. Alcance

Este documento define:

- Los componentes que participan en la comunicación distribuida.
- Los protocolos utilizados entre componentes.
- La dirección de las comunicaciones.
- La responsabilidad de cada mecanismo de comunicación.
- La separación entre comunicación externa e interna.
- Las reglas generales de integración entre servicios.

Este documento no define todavía el detalle completo de cada endpoint, interfaz RMI, contrato SOAP o archivo `.proto`. Dichos contratos se documentarán progresivamente dentro de `docs/api/`.

---

## 3. Vista general de comunicación

```text
                    Cliente Web
                         │
                    Cliente Desktop
                         │
                    Cliente Mobile
                         │
                    HTTPS / REST
                         │
                         ▼
                    API Gateway
                         │
        ┌────────────────┼────────────────┐
        │                │                │
        ▼                ▼                ▼
   File Service     Photo Service     HPC Service
        │
        ▼
   PostgreSQL / Storage


API Gateway ── Java RMI ── Auth Service

API Gateway ── SOAP ────── Streaming Service

API Gateway ── REST ────── Monitoring Service
                                     ▲
                                     │ gRPC
                              Servicios / Nodos


Cliente Desktop / Mobile
           │
           │ TCP Sockets
           ▼
      Sync Service


HPC Service
     │
     │ OpenMPI
     ▼
 Cluster HPC