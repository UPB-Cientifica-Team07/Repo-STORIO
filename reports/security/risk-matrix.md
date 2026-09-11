# Matriz de Riesgos de Seguridad — UPB-CIENTÍFICA

Fecha de evaluación: 2026-09-10

## Escala

- BAJO: impacto reducido o exposición principalmente informativa.
- MEDIO: puede facilitar ataques o comprometer parcialmente confidencialidad/integridad.
- ALTO: puede comprometer credenciales, autenticación, disponibilidad o datos.
- CRÍTICO: compromiso directo y amplio del sistema o de información sensible.

## Hallazgos

| ID | Hallazgo | Probabilidad | Impacto | Nivel | Estado |
|---|---|---:|---:|---|---|
| R-01 | Auth Service opera mediante HTTP sin TLS | Alta | Alta | ALTO | Abierto |
| R-02 | OpenLDAP opera mediante LDAP sin TLS en puerto 389 | Alta | Alta | ALTO | Abierto |
| R-03 | Servicios gRPC utilizan transporte plaintext | Alta | Alta | ALTO | Abierto |
| R-04 | Credenciales PostgreSQL embebidas como valores por defecto | Alta | Alta | ALTO | Mitigado |
| R-05 | Photo Service presenta vulnerabilidades conocidas en multer y qs | Alta | Alta | ALTO | Mitigado |
| R-06 | Auth Service sin rate limiting o bloqueo de intentos fallidos | Alta | Alta | ALTO | Mitigado |
| R-07 | Monitoring Service sin autenticación/autorización confirmada | Alta | Alta | ALTO | Mitigado |
| R-08 | Android permite tráfico cleartext | Alta | Alta | ALTO | Abierto |
| R-09 | Android almacena contraseña en SharedPreferences sin cifrado | Media | Alta | ALTO | Abierto |
| R-10 | LDAP permite enumeración anónima de estructura, usuarios y grupos | Media | Media | MEDIO | Abierto |
| R-11 | Servicios internos escuchan en todas las interfaces de red | Media | Alta | ALTO | Mitigado |
| R-12 | Web, Photo y Streaming carecen de headers HTTP defensivos | Media | Media | MEDIO | Abierto |
| R-13 | Express/PHP revelan tecnología y versión mediante headers | Media | Baja | BAJO | Abierto |
| R-14 | Java RMI expuesto en puertos 1099 y 1100 | Media | Alta | ALTO | Mitigado |
| R-15 | PostgreSQL restringido a loopback | Baja | Baja | BAJO | Mitigado |

## Evidencia principal

- Nmap detectó LDAP, Java RMI, HTTP, gRPC y servicios internos activos.
- PostgreSQL escucha únicamente en 127.0.0.1/::1.
- Auth, RMI, gRPC, Web, Photo y Streaming escuchan en interfaces no restringidas.
- LDAP permite búsqueda anónima y enumera usuarios y grupos.
- Se detectó uso de ldap://, http://, insecure.NewCredentials() y usePlaintext().
- Android tiene android:usesCleartextTraffic="true".
- npm audit detectó vulnerabilidades HIGH en multer y MODERATE en qs.
- No se encontraron vulnerabilidades npm en clients/web.
- No se encontró protección de rate limiting/lockout en Auth.
- No se encontró helmet ni express-rate-limit en Web/Photo.
- Android CredentialStore persiste username/password mediante SharedPreferences.
- Se detectaron passwords de PostgreSQL como defaults dentro del código fuente.

## Orden de remediación

1. Dependencias vulnerables de Photo Service. ✅ Mitigado
2. Eliminar credenciales embebidas. ✅ Mitigado ✅ Mitigado
3. Proteger Auth contra fuerza bruta. ✅ Mitigado ✅ Mitigado
4. Añadir autenticación/autorización a Monitoring. ✅ Mitigado
5. Restringir interfaces y puertos internos. ✅ Mitigado
6. TLS/LDAPS/gRPC TLS.
7. Endurecer Android.
8. Headers HTTP y eliminación de banners.
9. Restringir enumeración LDAP anónima.

## Remediaciones aplicadas

### R-05 — Dependencias vulnerables de Photo Service

- multer actualizado de 2.2.0 a 2.3.0.
- qs actualizado a 6.16.0.
- npm audit final: 0 vulnerabilidades.
- Estado: MITIGADO.


### R-04 — Credenciales PostgreSQL embebidas

Se eliminaron las contraseñas PostgreSQL predeterminadas del código fuente.

Variables utilizadas:

- File Service: `FILE_DB_PASSWORD` o `POSTGRES_PASSWORD`.
- Sync Service: `SYNC_DB_PASSWORD` o `POSTGRES_PASSWORD`.
- Photo Service: `PHOTO_DB_PASSWORD` o `DB_PASSWORD`.
- Streaming Service: `STREAM_DB_PASSWORD` o `DB_PASSWORD`.
- HPC Service: `HPC_DB_PASSWORD`.

Validación:

- Sin variable de contraseña, cada servicio rechaza iniciar explícitamente.
- Con la variable correspondiente, los servicios conectan correctamente.
- La búsqueda de `upb_dev_2026` dentro de `services/` devuelve cero resultados.
- File Service, Sync Service, Photo Service, Streaming y HPC fueron probados en runtime.

Estado: MITIGADO.


### R-06 — Protección contra fuerza bruta

Se implementó rate limiting para los dos canales de autenticación:

- HTTP: control por usuario + dirección IP.
- Java RMI: control por usuario.
- Máximo normal: 5 intentos.
- Ventana normal: 300 segundos.
- Bloqueo normal: 300 segundos.
- HTTP responde `429 Too Many Requests` durante bloqueo.
- Se incluye `Retry-After`.
- Un login exitoso limpia los intentos fallidos.
- Las contraseñas no se registran en logs.

Pruebas realizadas:

- Tres intentos fallidos con configuración de prueba.
- Bloqueo HTTP confirmado mediante código 429.
- Contraseña correcta rechazada mientras el bloqueo estaba activo.
- Login permitido después de terminar el bloqueo.
- La misma secuencia fue validada directamente mediante Java RMI.

Estado: MITIGADO.

### R-07 — Autenticación y autorización de Monitoring Service

- Se añadió interceptor gRPC obligatorio en Monitoring Service.
- `ReportMetrics` y `ReportStatus` requieren un token interno mediante metadata `x-monitoring-service-token`.
- El token interno se obtiene desde `MONITORING_SERVICE_TOKEN` y no se almacena en el repositorio.
- File Service, Sync Service, Analysis Service y Photo Service fueron migrados para enviar el token interno.
- Las consultas de Monitoring requieren `Authorization: Bearer <token>` validado contra Auth Service.
- Web Client reenvía el Bearer token hacia Monitoring mediante metadata gRPC.
- `CreateAlertRule`, `UpdateAlertRule` y `DeleteAlertRule` requieren rol `ADMIN`.
- El Web traduce `Unauthenticated` a HTTP 401 y `PermissionDenied` a HTTP 403.
- Prueba sin credenciales: HTTP 401.
- Prueba con token inválido: HTTP 401.
- Prueba con rol USUARIO sobre lectura de métricas: HTTP 200.
- Prueba con rol USUARIO sobre creación de regla: HTTP 403, `rol ADMIN requerido`.
- Prueba con rol ADMIN sobre creación de regla: operación permitida.
- File, Sync y Photo reportaron métricas correctamente después de activar el service token.
- Las reglas creadas durante las pruebas fueron eliminadas.
- Estado: MITIGADO.

### R-11 — Restricción de interfaces de servicios internos

Se redujo la superficie de exposición de los servicios que no requieren acceso directo desde otros equipos.

Bindings aplicados:

- Monitoring Service: `127.0.0.1:50051`.
- Photo Service: `127.0.0.1:50052`.
- File Service: `127.0.0.1:50053`.
- Analysis Service: `127.0.0.1:50054`.
- Streaming SOAP Service: `127.0.0.1:8082`.

Se mantienen accesibles por red los servicios que forman parte de la interfaz distribuida del sistema:

- Sync Service `50055`: requerido por clientes Windows, Linux y Android.
- Web `8080`: interfaz de usuario accesible por red.
- Auth HTTP `8081`: requerido por clientes distribuidos.

Los puertos Java RMI `1099` y `1100` se gestionan separadamente en R-14.

Validación runtime mediante `ss` confirmó que los servicios internos anteriores escuchan exclusivamente en loopback.

Después de la restricción:

- File Service continuó enviando métricas.
- Sync Service continuó enviando métricas.
- Photo Service continuó enviando métricas.
- Analysis Service registró estado correctamente.
- Web respondió HTTP 200.
- Streaming permaneció accesible en `127.0.0.1:8082`.

Estado: MITIGADO.

### R-14 — Endurecimiento de Java RMI

Se eliminó la exposición RMI mediante puertos dinámicos y listeners
sin restricción de interfaz.

Antes de la mitigación, Java RMI utilizaba:

- Auth Registry: `*:1099`.
- HPC Registry: `*:1100`.
- Auth Service: puerto efímero.
- HPC Coordinator: puerto efímero.
- Worker node-01: puerto efímero.
- Worker node-02: puerto efímero.

Los objetos que extendían `UnicastRemoteObject` utilizaban `super()`,
por lo que la JVM seleccionaba puertos dinámicos.

La mitigación implementó:

- puertos explícitos para todos los objetos RMI;
- `RMIServerSocketFactory` con bind sobre dirección configurable;
- dirección anunciada mediante `java.rmi.server.hostname`;
- separación entre dirección de bind y dirección publicada;
- configuración independiente para Coordinator y Workers.

Topología validada localmente:

- Auth Registry: `127.0.0.1:1099`.
- Auth remote object: `127.0.0.1:1101`.
- HPC Registry: `127.0.0.1:1100`.
- HPC Coordinator object: `127.0.0.1:1102`.
- HPC Worker node-01: `127.0.0.1:12001`.
- HPC Worker node-02: `127.0.0.1:12002`.

Los bindings HPC son configurables mediante variables de entorno para
permitir despliegues distribuidos en una interfaz LAN autorizada sin
volver a utilizar `0.0.0.0` ni puertos efímeros.

Validación funcional:

- Auth RMI lookup: OK.
- HPC Coordinator lookup: OK.
- HPC health: ACTIVE.
- Workers registrados: 2.
- RMI smoke test: OK.
- Código de salida: 0.

Los puertos efímeros observados previamente dejaron de existir.

Estado: MITIGADO.
