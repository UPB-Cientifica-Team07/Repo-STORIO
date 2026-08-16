# ADR-001: Arquitectura Distribuida Basada en Servicios

## Estado

Aceptado

## Contexto

El proyecto UPB-CIENTÍFICA requiere implementar diferentes funcionalidades distribuidas, incluyendo autenticación, almacenamiento de archivos, sincronización, servicios multimedia, monitoreo y computación de alto rendimiento.

Estas funcionalidades presentan responsabilidades y requerimientos técnicos diferentes.

## Decisión

Se implementará una arquitectura distribuida basada en servicios independientes.

Cada servicio tendrá una responsabilidad específica y podrá utilizar la tecnología más adecuada para su implementación.

## Servicios identificados

- API Gateway
- Auth Service
- File Service
- Sync Service
- Photo Service
- Streaming Service
- Monitoring Service
- HPC Service

## Consecuencias positivas

- Separación de responsabilidades.
- Desarrollo paralelo.
- Posibilidad de utilizar diferentes tecnologías.
- Mayor facilidad para escalar componentes específicos.
- Menor acoplamiento funcional.

## Consecuencias negativas

- Mayor complejidad de integración.
- Necesidad de definir protocolos y contratos.
- Mayor dificultad en pruebas distribuidas.
- Necesidad de monitorear múltiples servicios.