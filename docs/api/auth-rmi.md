# Auth Service - Contrato Java RMI

## 1. Objetivo

Definir el contrato de comunicación entre el API Gateway y el Auth Service de UPB-CIENTÍFICA mediante Java RMI.

El Auth Service es responsable de las operaciones relacionadas con la autenticación, validación de usuarios, consulta de información básica del usuario y validación de tokens.

El objetivo de utilizar Java RMI es incorporar un modelo de objetos distribuidos dentro de la arquitectura del sistema.

---

## 2. Alcance

Este documento define:

- La comunicación entre API Gateway y Auth Service.
- La interfaz remota principal.
- Las operaciones iniciales disponibles.
- Los objetos de transferencia de datos.
- Las entradas y salidas de cada operación.
- El manejo general de errores.
- Las reglas de serialización.
- La relación con PostgreSQL.

Este documento no define todavía:

- La implementación concreta de las clases.
- El algoritmo definitivo de generación de tokens.
- La configuración del registro RMI.
- Las credenciales reales de la base de datos.
- Los detalles internos de cifrado de contraseñas.

Estos elementos serán definidos durante la fase de implementación.

---

# 3. Modelo de comunicación

El cliente no se comunica directamente con el Auth Service.

La comunicación se realiza a través del API Gateway.

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ Java RMI
   ▼
Auth Service
   │
   │ SQL
   ▼
PostgreSQL
```

El flujo general es:

```text
1. El cliente envía una solicitud al API Gateway.
2. El API Gateway recibe y valida la solicitud.
3. El Gateway invoca el objeto remoto correspondiente.
4. Auth Service procesa la operación.
5. Auth Service consulta o actualiza la información necesaria.
6. Auth Service retorna un objeto de respuesta.
7. El API Gateway transforma el resultado en una respuesta REST.
8. El cliente recibe la respuesta.
```

---

# 4. Tecnología de comunicación

La comunicación entre el API Gateway y el Auth Service utiliza Java RMI.

| Elemento | Tecnología |
|---|---|
| Modelo | Objetos distribuidos |
| Tecnología | Java RMI |
| Lenguaje principal | Java |
| Interfaz | `Remote` |
| Comunicación | Invocación de métodos remotos |
| Objetos transferidos | Objetos serializables |
| Servicio de persistencia | PostgreSQL |

La comunicación mediante RMI permite que el API Gateway invoque métodos definidos en el Auth Service sin conocer los detalles internos de su implementación.

---

# 5. Componentes principales

La arquitectura inicial del Auth Service está compuesta por:

```text
API Gateway
      │
      │ Cliente RMI
      ▼
┌───────────────────────────┐
│       Auth Service        │
│                           │
│  AuthServiceRemote        │
│            │              │
│            ▼              │
│  AuthServiceImpl          │
│            │              │
│            ▼              │
│     User Repository       │
└─────────────┬─────────────┘
              │
              │ SQL
              ▼
         PostgreSQL
```

Los componentes principales son:

- `AuthServiceRemote`: contrato remoto.
- `AuthServiceImpl`: implementación de las operaciones.
- Objetos DTO: intercambio de información.
- Capa de acceso a datos: comunicación con PostgreSQL.
- Cliente RMI: utilizado por el API Gateway.

---

# 6. Interfaz remota

La interfaz principal será:

```text
AuthServiceRemote
```

Esta interfaz define las operaciones que pueden ser invocadas remotamente.

Inicialmente se contemplan las siguientes operaciones:

```text
login(...)
validarToken(...)
obtenerUsuario(...)
```

Conceptualmente:

```java
public interface AuthServiceRemote extends Remote {

    AuthResponse login(LoginRequest request)
        throws RemoteException;

    TokenValidationResponse validarToken(String token)
        throws RemoteException;

    UsuarioDTO obtenerUsuario(UUID idUsuario)
        throws RemoteException;
}
```

La firma definitiva podrá ajustarse durante la implementación, pero el contrato funcional debe mantener las responsabilidades definidas en este documento.

---

# 7. Operación login

## 7.1 Objetivo

Permitir la autenticación de un usuario mediante sus credenciales.

La operación recibe las credenciales y verifica la información almacenada en el sistema.

```text
API Gateway
      │
      │ login(LoginRequest)
      ▼
Auth Service
      │
      ▼
PostgreSQL
```

---

## 7.2 Entrada

La operación recibe un objeto:

```text
LoginRequest
```

Este objeto contiene inicialmente:

| Campo | Tipo | Descripción |
|---|---|---|
| usuario | String | Nombre de usuario o identificador |
| password | String | Contraseña proporcionada |

Representación conceptual:

```text
LoginRequest
│
├── usuario
│
└── password
```

La contraseña únicamente se utiliza durante el proceso de autenticación.

No debe ser almacenada dentro de objetos de respuesta.

---

## 7.3 Proceso general

El flujo de autenticación será:

```text
LoginRequest
      │
      ▼
Validar datos
      │
      ▼
Buscar usuario
      │
      ├── No existe ──► Autenticación rechazada
      │
      ▼
Verificar estado
      │
      ├── Inactivo ──► Acceso rechazado
      │
      ▼
Verificar contraseña
      │
      ├── Incorrecta ──► Autenticación rechazada
      │
      ▼
Generar resultado
      │
      ▼
AuthResponse
```

---

## 7.4 Salida

La operación retorna:

```text
AuthResponse
```

El objeto puede contener:

| Campo | Tipo | Descripción |
|---|---|---|
| autenticado | boolean | Indica si la autenticación fue exitosa |
| mensaje | String | Resultado general |
| usuario | UsuarioDTO | Información básica del usuario |
| token | TokenDTO | Información del token generado |

Representación:

```text
AuthResponse
│
├── autenticado
├── mensaje
├── usuario
└── token
```

---

## 7.5 Autenticación exitosa

Ejemplo conceptual:

```text
autenticado = true

usuario:
    idUsuario
    usuario
    correo
    rol

token:
    valor
    expiracion
```

El resultado no debe incluir:

```text
password
password_hash
```

---

## 7.6 Autenticación rechazada

Ejemplo conceptual:

```text
autenticado = false
mensaje = "Credenciales inválidas"
usuario = null
token = null
```

El sistema no debe revelar información innecesaria sobre la causa específica de autenticación fallida.

---

# 8. Operación validarToken

## 8.1 Objetivo

Permitir al API Gateway verificar si un token proporcionado por un cliente es válido.

```text
Cliente
   │
   │ Bearer Token
   ▼
API Gateway
      │
      │ validarToken(token)
      ▼
Auth Service
```

---

## 8.2 Entrada

La operación recibe:

```text
token
```

Conceptualmente:

```java
validarToken(String token)
```

---

## 8.3 Salida

La operación retorna:

```text
TokenValidationResponse
```

Este objeto puede contener:

| Campo | Tipo | Descripción |
|---|---|---|
| valido | boolean | Indica si el token es válido |
| idUsuario | UUID | Usuario asociado |
| rol | String | Rol del usuario |
| expiracion | Timestamp | Fecha de expiración |
| mensaje | String | Resultado de la validación |

---

## 8.4 Resultado válido

```text
valido = true
```

El API Gateway puede continuar con la solicitud protegida.

---

## 8.5 Resultado inválido

```text
valido = false
```

El API Gateway debe rechazar la solicitud.

Conceptualmente:

```text
HTTP 401 Unauthorized
```

---

# 9. Operación obtenerUsuario

## 9.1 Objetivo

Obtener la información básica de un usuario identificado mediante su UUID.

---

## 9.2 Entrada

```text
idUsuario
```

Tipo:

```text
UUID
```

---

## 9.3 Flujo

```text
API Gateway
      │
      │ obtenerUsuario(idUsuario)
      ▼
Auth Service
      │
      │ SQL
      ▼
PostgreSQL
      │
      ▼
UsuarioDTO
```

---

## 9.4 Salida

La operación retorna:

```text
UsuarioDTO
```

Campos iniciales:

| Campo | Tipo |
|---|---|
| idUsuario | UUID |
| usuario | String |
| correo | String |
| rol | String |
| estado | boolean |

El objeto no debe contener información sensible como:

```text
password
password_hash
```

---

# 10. Objetos de transferencia

Los objetos intercambiados entre el API Gateway y el Auth Service deben ser serializables.

Inicialmente se definen los siguientes:

```text
LoginRequest
AuthResponse
UsuarioDTO
TokenDTO
TokenValidationResponse
```

La relación conceptual es:

```text
LoginRequest
      │
      ▼
Auth Service
      │
      ├── UsuarioDTO
      │
      └── TokenDTO
             │
             ▼
        AuthResponse
```

---

# 11. LoginRequest

Responsabilidad:

Representar las credenciales recibidas para realizar una solicitud de autenticación.

Estructura conceptual:

```text
LoginRequest
├── usuario : String
└── password : String
```

Este objeto debe implementar:

```java
Serializable
```

---

# 12. UsuarioDTO

Responsabilidad:

Representar la información básica de un usuario que puede ser transferida entre componentes.

Estructura conceptual:

```text
UsuarioDTO
├── idUsuario : UUID
├── usuario : String
├── correo : String
├── rol : String
└── estado : boolean
```

No debe contener:

```text
password
password_hash
```

---

# 13. TokenDTO

Responsabilidad:

Representar la información asociada a un token de autenticación.

Estructura conceptual:

```text
TokenDTO
├── token : String
└── expiracion : Timestamp
```

La estructura podrá ampliarse posteriormente si se requiere información adicional.

---

# 14. AuthResponse

Responsabilidad:

Representar el resultado de una operación de autenticación.

Estructura conceptual:

```text
AuthResponse
├── autenticado : boolean
├── mensaje : String
├── usuario : UsuarioDTO
└── token : TokenDTO
```

---

# 15. TokenValidationResponse

Responsabilidad:

Representar el resultado de validar un token.

Estructura conceptual:

```text
TokenValidationResponse
├── valido : boolean
├── idUsuario : UUID
├── rol : String
├── expiracion : Timestamp
└── mensaje : String
```

Cuando el token no sea válido, los campos relacionados con el usuario podrán ser nulos.

---

# 16. Relación con PostgreSQL

El Auth Service utiliza PostgreSQL para consultar la información necesaria durante la autenticación.

Las entidades principales relacionadas son:

```text
usuario
token
```

Modelo conceptual:

```text
Auth Service
      │
      ├── Consulta usuario
      │
      ▼
   usuario
      │
      └── Verifica información

Auth Service
      │
      ├── Consulta o registra token
      │
      ▼
    token
```

La estructura exacta de las consultas SQL corresponde a la implementación interna del Auth Service.

---

# 17. Manejo de errores

Los errores de comunicación RMI deben ser controlados mediante:

```text
RemoteException
```

Se contemplan inicialmente los siguientes escenarios:

| Situación | Resultado |
|---|---|
| Usuario inexistente | Autenticación rechazada |
| Contraseña incorrecta | Autenticación rechazada |
| Usuario inactivo | Acceso rechazado |
| Token inválido | Validación negativa |
| Token expirado | Validación negativa |
| Usuario no encontrado | Resultado controlado |
| Error de PostgreSQL | Error interno |
| Auth Service no disponible | Error de comunicación |
| Error RMI | Error interno |

El API Gateway debe transformar los errores de infraestructura en respuestas apropiadas para el cliente.

---

# 18. Serialización

Los objetos transferidos mediante Java RMI deben implementar:

```java
java.io.Serializable
```

Por ejemplo:

```text
LoginRequest
UsuarioDTO
TokenDTO
AuthResponse
TokenValidationResponse
```

Esto permite que los objetos sean transmitidos entre el API Gateway y el Auth Service.

---

# 19. Flujo completo de autenticación

```text
Cliente
   │
   │ POST /auth/login
   │ HTTPS / JSON
   ▼
API Gateway
   │
   │ Convierte la solicitud
   │ en LoginRequest
   ▼
Cliente RMI
   │
   │ login(LoginRequest)
   ▼
Auth Service
   │
   │ Consulta
   ▼
PostgreSQL
   │
   │ Usuario encontrado
   ▼
Auth Service
   │
   │ AuthResponse
   ▼
API Gateway
   │
   │ JSON / HTTP
   ▼
Cliente
```

---

# 20. Flujo de validación de token

```text
Cliente
   │
   │ Authorization: Bearer <token>
   ▼
API Gateway
   │
   │ validarToken(token)
   │ Java RMI
   ▼
Auth Service
   │
   ▼
TokenValidationResponse
   │
   ▼
API Gateway
   │
   ├── Token válido ──► Procesar solicitud
   │
   └── Token inválido ─► HTTP 401
```

---

# 21. Reglas del contrato

La comunicación entre el API Gateway y el Auth Service debe cumplir las siguientes reglas:

1. El cliente no se comunica directamente con el Auth Service.
2. El API Gateway utiliza Java RMI para las operaciones de autenticación.
3. La interfaz remota define explícitamente las operaciones disponibles.
4. Los objetos transferidos deben ser serializables.
5. Las contraseñas no deben ser retornadas.
6. Los hashes de contraseña no deben salir del Auth Service.
7. Los errores internos deben ser controlados.
8. `RemoteException` representa errores relacionados con la comunicación remota.
9. La lógica de autenticación permanece encapsulada dentro del Auth Service.
10. El API Gateway transforma los resultados RMI en respuestas HTTP para los clientes.

---

# 22. Estructura inicial propuesta

La implementación del servicio puede organizarse posteriormente de la siguiente manera:

```text
services/
└── auth-service/
    │
    ├── contract/
    │   ├── AuthServiceRemote.java
    │   ├── LoginRequest.java
    │   ├── AuthResponse.java
    │   ├── UsuarioDTO.java
    │   ├── TokenDTO.java
    │   └── TokenValidationResponse.java
    │
    ├── server/
    │   ├── AuthServiceImpl.java
    │   ├── RmiServer.java
    │   └── repository/
    │
    └── client/
        └── AuthRmiClient.java
```

Esta estructura puede ajustarse cuando iniciemos la implementación, pero mantiene separadas las responsabilidades del contrato remoto, la lógica del servidor y el cliente RMI.

---

# 23. Resumen

El Auth Service implementa uno de los mecanismos centrales de comunicación distribuida de UPB-CIENTÍFICA.

Su función es proporcionar autenticación y validación de identidad mediante un modelo de objetos distribuidos basado en Java RMI.

```text
Cliente
   │
   │ HTTPS / REST
   ▼
API Gateway
   │
   │ Java RMI
   ▼
Auth Service
   │
   │ SQL
   ▼
PostgreSQL
```

Las operaciones iniciales definidas son:

```text
login(...)
validarToken(...)
obtenerUsuario(...)
```

Los principales objetos transferidos son:

```text
LoginRequest
AuthResponse
UsuarioDTO
TokenDTO
TokenValidationResponse
```

Este documento define el contrato funcional inicial. La implementación concreta de Java RMI se realizará posteriormente respetando las interfaces, operaciones y reglas establecidas.