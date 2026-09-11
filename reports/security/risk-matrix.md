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
| R-06 | Auth Service sin rate limiting o bloqueo de intentos fallidos | Alta | Alta | ALTO | Abierto |
| R-07 | Monitoring Service sin autenticación/autorización confirmada | Alta | Alta | ALTO | Abierto |
| R-08 | Android permite tráfico cleartext | Alta | Alta | ALTO | Abierto |
| R-09 | Android almacena contraseña en SharedPreferences sin cifrado | Media | Alta | ALTO | Abierto |
| R-10 | LDAP permite enumeración anónima de estructura, usuarios y grupos | Media | Media | MEDIO | Abierto |
| R-11 | Servicios internos escuchan en todas las interfaces de red | Media | Alta | ALTO | Abierto |
| R-12 | Web, Photo y Streaming carecen de headers HTTP defensivos | Media | Media | MEDIO | Abierto |
| R-13 | Express/PHP revelan tecnología y versión mediante headers | Media | Baja | BAJO | Abierto |
| R-14 | Java RMI expuesto en puertos 1099 y 1100 | Media | Alta | ALTO | Abierto |
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
2. Eliminar credenciales embebidas. ✅ Mitigado
3. Proteger Auth contra fuerza bruta.
4. Añadir autenticación/autorización a Monitoring.
5. Restringir interfaces y puertos internos.
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
