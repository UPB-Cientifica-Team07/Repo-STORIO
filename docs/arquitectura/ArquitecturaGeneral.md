# Arquitectura General

## 1. Descripción

UPB-CIENTÍFICA es una plataforma distribuida orientada a ofrecer servicios de almacenamiento, sincronización de archivos, gestión multimedia, monitoreo y ejecución de trabajos de computación de alto rendimiento.

El sistema se estructura mediante servicios con responsabilidades independientes que se comunican a través de protocolos definidos.

## 2. Estilo arquitectónico

El sistema adopta una arquitectura distribuida orientada a servicios.

Cada servicio tiene una responsabilidad específica y puede utilizar la tecnología más adecuada para su función.

## 3. Componentes principales

### Clientes

- Cliente Web
- Cliente Desktop
- Cliente Mobile

### Servicios

- API Gateway
- Auth Service
- File Service
- Sync Service
- Photo Service
- Streaming Service
- Monitoring Service
- HPC Service

### Infraestructura

- PostgreSQL
- Almacenamiento compartido
- Cluster HPC

## 4. Flujo general

Los clientes acceden al sistema a través del API Gateway.

El Gateway dirige las solicitudes hacia el servicio correspondiente.

Los servicios utilizan PostgreSQL para almacenar información estructurada, el almacenamiento compartido para archivos y el Cluster HPC para ejecutar trabajos de computación paralela.

## 5. Principios

- Separación de responsabilidades.
- Bajo acoplamiento entre servicios.
- Comunicación mediante contratos definidos.
- Independencia tecnológica cuando sea necesaria.
- Posibilidad de escalamiento de servicios.