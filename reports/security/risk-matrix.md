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
| R-01 | Auth Service opera mediante HTTP sin TLS | Alta | Alta | ALTO | Mitigado |
| R-02 | OpenLDAP opera mediante LDAP sin TLS en puerto 389 | Alta | Alta | ALTO | Mitigado |
| R-03 | Servicios gRPC utilizan transporte plaintext | Alta | Alta | ALTO | Mitigado |
| R-04 | Credenciales PostgreSQL embebidas como valores por defecto | Alta | Alta | ALTO | Mitigado |
| R-05 | Photo Service presenta vulnerabilidades conocidas en multer y qs | Alta | Alta | ALTO | Mitigado |
| R-06 | Auth Service sin rate limiting o bloqueo de intentos fallidos | Alta | Alta | ALTO | Mitigado |
| R-07 | Monitoring Service sin autenticación/autorización confirmada | Alta | Alta | ALTO | Mitigado |
| R-08 | Android permite tráfico cleartext | Alta | Alta | ALTO | Mitigado |
| R-09 | Android almacena contraseña en SharedPreferences sin cifrado | Media | Alta | ALTO | Mitigado |
| R-10 | LDAP permite enumeración anónima de estructura, usuarios y grupos | Media | Media | MEDIO | Mitigado |
| R-11 | Servicios internos escuchan en todas las interfaces de red | Media | Alta | ALTO | Mitigado |
| R-12 | Web, Photo y Streaming carecen de headers HTTP defensivos | Media | Media | MEDIO | Mitigado |
| R-13 | Express/PHP revelan tecnología y versión mediante headers | Media | Baja | BAJO | Mitigado |
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

### R-03 — Cifrado TLS para Sync gRPC

El servicio Sync expone `50055` a la red local para atender clientes
Linux, Windows y Android. Antes de la mitigación, el transporte gRPC
utilizaba credenciales inseguras y los clientes Android utilizaban
`usePlaintext()`.

La mitigación implementó:

- TLS obligatorio en Sync Service mediante credenciales gRPC de servidor;
- certificado X.509 firmado por una CA local del proyecto;
- SAN para `localhost`, `127.0.0.1` y la dirección LAN autorizada;
- CA configurable mediante `SYNC_TLS_CA_FILE` en clientes Go;
- certificado y clave configurables mediante `SYNC_TLS_CERT_FILE` y
  `SYNC_TLS_KEY_FILE`;
- validación TLS en el watcher Linux y cliente CLI;
- validación explícita de la CA en el cliente Android;
- eliminación de `usePlaintext()` en Android;
- exclusión del material criptográfico privado mediante `.gitignore`.

Las claves privadas y artefactos PKI runtime no se versionan en Git.

Validación runtime:

- Sync Service inició con `TLS gRPC: ACTIVO`;
- handshake OpenSSL: `Verification: OK`;
- `Verify return code: 0 (ok)`;
- el certificado servido contiene los SAN esperados;
- una conexión gRPC plaintext fue rechazada;
- una conexión gRPC TLS alcanzó la capa de aplicación;
- Sync continuó enviando métricas a Monitoring;
- `go test ./services/sync-service/...`: código 0;
- Android `:app:compileDebugKotlin`: código 0.

Los servicios gRPC internos `50051`, `50053` y `50054` permanecen
restringidos a loopback y no atraviesan la red LAN.

Estado: MITIGADO.

### R-01 — Auth Service protegido mediante HTTPS

El Auth Service exponía el bridge de autenticación en el puerto `8081`
mediante HTTP plano. Esto afectaba login, validación y revocación de
tokens utilizados por los componentes distribuidos.

La mitigación implementó:

- `HttpsServer` en el Auth Service Java;
- certificado X.509 independiente para Auth;
- reutilización de la CA del proyecto;
- SAN para `localhost`, `127.0.0.1` y la dirección LAN autorizada;
- PKCS#12 local para el servidor Java;
- TLS mínimo 1.2 en los clientes que lo soportan;
- validación explícita de la CA en Go, Java, Node.js, PHP y Android;
- migración de File, Sync, Monitoring, Photo, Streaming, HPC, Web,
  Linux, Windows y Android a `https://...:8081`;
- rechazo de HTTP plaintext en el puerto `8081`;
- claves privadas excluidas de Git.

Validación runtime:

- handshake TLS: `Verification: OK`;
- `Verify return code: 0 (ok)`;
- `/internal/auth/validate` sobre HTTPS respondió `401` sin token;
- HTTP plaintext respondió `000` / conexión cerrada;
- login HTTPS real exitoso;
- token de 43 caracteres generado;
- Web realizó login correctamente a través de Auth HTTPS;
- File, Sync y Photo continuaron reportando métricas;
- HPC Coordinator se reinició con Auth HTTPS;
- workers `node-01` y `node-02` se registraron correctamente;
- job MPI autenticado terminó con `Success: true` y `Exit code: 0`.

Estado: MITIGADO.

### R-08 — Android sin transporte cleartext

El cliente Android utilizaba HTTP para Auth y permitía tráfico cleartext.
El canal gRPC Sync también utilizaba anteriormente transporte plaintext.

La mitigación implementó:

- Auth URL migrada a HTTPS;
- `HttpsURLConnection` con CA del proyecto;
- Sync gRPC protegido mediante TLS;
- CA pública embebida en `res/raw`;
- `android:usesCleartextTraffic="false"`;
- eliminación de `usePlaintext()`.

Validación:

- `:app:compileDebugKotlin`: código 0;
- no quedan URLs `http://` en el código productivo Android;
- no queda `usePlaintext()` en Android;
- configuración de cleartext deshabilitada.

No se realizó validación sobre dispositivo físico en esta sesión.

Estado: MITIGADO.

### R-02 — OpenLDAP protegido mediante LDAPS

OpenLDAP exponía originalmente LDAP plaintext sobre el puerto `389`.
El Auth Service realizaba binds mediante `ldap://127.0.0.1:389`.

La mitigación implementó:

- certificado X.509 específico para OpenLDAP;
- certificado firmado por la CA privada de UPB-CIENTIFICA;
- `ldaps:///` sobre el puerto `636`;
- configuración TLS mediante `cn=config`;
- validación explícita de la CA en el cliente Java JNDI;
- migración de Auth a `ldaps://127.0.0.1:636`;
- migración de los scripts administrativos a LDAPS;
- eliminación de `ldap:///` de `SLAPD_SERVICES`;
- cierre completo del listener TCP `389`.

Validación:

- `openssl s_client`: `Verification: OK`;
- `Verify return code: 0 (ok)`;
- consulta LDAPS sobre `636`: exit code `0`;
- consulta LDAP plaintext sobre `389`: exit code `255`;
- login real `HTTPS -> Auth -> LDAPS -> OpenLDAP`: exitoso;
- usuario `user-002`;
- rol `USUARIO`;
- token de 43 caracteres emitido correctamente.

La enumeración anónima del directorio se aborda de forma separada
en `R-10`.

Estado: MITIGADO.

### R-10 — Enumeración LDAP anónima restringida

La configuración anterior permitía lectura general del directorio mediante
`by * read`, lo que permitía enumerar usuarios, identificadores, grupos y
membresías sin autenticación.

La mitigación implementa una cuenta técnica `auth-reader` con permisos de
solo lectura. Auth Service utiliza el bind del usuario exclusivamente para
validar la contraseña y emplea `auth-reader` para consultar atributos y
resolver roles.

Las ACL finales conservan únicamente `anonymous auth` sobre `userPassword`,
necesario para verificar un bind simple, y niegan lectura general a
conexiones anónimas.

La administración local mediante `ldapi:///` y SASL/EXTERNAL permanece
disponible exclusivamente para `root`.

Las pruebas confirmaron:

- bind de `auth-reader`: exitoso;
- lectura de usuarios mediante `auth-reader`: exit code 0;
- login `prueba`: exitoso;
- usuario: `user-002`;
- rol: `USUARIO`;
- token emitido correctamente;
- enumeración anónima de usuarios: bloqueada;
- enumeración anónima de grupos: bloqueada.

Estado: MITIGADO.

### R-09 — Credenciales Android protegidas con Android Keystore

`CredentialStore` almacenaba anteriormente `username/password`
directamente mediante `SharedPreferences`.

La mitigación reemplaza la persistencia plaintext de la contraseña por
cifrado autenticado `AES-256-GCM`. La clave AES se genera y conserva
mediante `AndroidKeyStore` y no se almacena en el repositorio ni en
`SharedPreferences`.

La aplicación persiste únicamente:

- `username`;
- `password_ciphertext`;
- `password_iv`.

Las instalaciones que todavía contengan la clave legacy `password` son
migradas automáticamente: la contraseña se cifra y la entrada plaintext
se elimina.

La aplicación también mantiene `android:allowBackup="false"`.

La implementación fue validada mediante revisión estática y compilación
Android. No se detectó almacenamiento directo de password ni claves AES
embebidas.

La ejecución del Android Keystore sobre hardware real queda pendiente de
validación posterior debido a que actualmente no se dispone de dispositivo
Android físico y no se utiliza emulador.

Estado: MITIGADO.

### R-12 / R-13 — Endurecimiento HTTP y reducción de disclosure

Web, Photo y Streaming carecían de headers HTTP defensivos. Además,
Express y PHP exponían su tecnología mediante `X-Powered-By`.

Web y Photo ahora:

- deshabilitan `x-powered-by`;
- envían `X-Content-Type-Options: nosniff`;
- envían `X-Frame-Options: DENY`;
- aplican `Referrer-Policy: no-referrer`;
- restringen capacidades mediante `Permissions-Policy`;
- aplican `Content-Security-Policy`.

Streaming elimina `X-Powered-By`, configura los mismos controles
defensivos y dispone de un arranque reproducible con `expose_php=0`.

Las pruebas runtime confirmaron que los tres servicios responden con los
headers defensivos y ya no presentan `X-Powered-By`.

El WSDL, `/health` y los servicios HTTP continuaron operativos después
del endurecimiento.

HSTS no se configura en estos listeners porque actualmente operan mediante
HTTP interno. Debe configurarse en la terminación HTTPS correspondiente
durante el despliegue final.

Estado R-12: MITIGADO.
Estado R-13: MITIGADO.
