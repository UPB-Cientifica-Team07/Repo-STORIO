# Directory Service - UPB-CIENTIFICA

Servicio de directorio centralizado basado en OpenLDAP.

## Propósito

El Directory Service centraliza:

- identidades de usuarios;
- credenciales;
- identificadores de directorio;
- grupos;
- roles de autorización.

Auth Service continúa siendo el punto de entrada de autenticación para el resto de la plataforma.

Arquitectura:

    Cliente
       |
       v
    Auth Service
    Java RMI + HTTP
       |
       v
    OpenLDAP
       |
       +-- ou=people
       +-- ou=groups
       |    +-- ADMIN
       |    +-- INVESTIGADOR
       |    +-- USUARIO
       +-- ou=services
       +-- ou=resources

Los demás servicios no consultan LDAP directamente.

    File --------\
    Photo --------\
    Streaming ----- > Auth Service -> OpenLDAP
    Sync ----------/
    HPC -----------/
    Web -----------/

## Dominio LDAP

Base DN:

    dc=upb-cientifica,dc=local

## Organización

    dc=upb-cientifica,dc=local
    |
    +-- ou=people
    |
    +-- ou=groups
    |   +-- cn=ADMIN
    |   +-- cn=INVESTIGADOR
    |   +-- cn=USUARIO
    |
    +-- ou=services
    |
    +-- ou=resources

## Identidad

Cada usuario utiliza:

- uid: nombre de login;
- employeeNumber: identificador estable usado por la plataforma;
- userPassword: credencial gestionada por OpenLDAP;
- membresía de grupo: rol de autorización.

Ejemplo:

    uid: prueba
    employeeNumber: user-002
    grupo: USUARIO

employeeNumber corresponde al directorio_id usado por la plataforma.

## Integración con Auth Service

Auth Service utiliza JNDI LDAP de Java.

Proceso:

    username + password
            |
            v
       Auth Service
            |
            v
         LDAP bind
            |
            +-- credenciales inválidas -> login rechazado
            |
            v
      employeeNumber
            |
            v
        grupo LDAP
            |
            v
           rol
            |
            v
        token UUID

Se mantienen los contratos existentes:

- Java RMI: puerto 1099;
- bridge HTTP interno: puerto 8081.

## Archivos LDIF

### 01-structure.ldif

Crea:

- ou=people
- ou=groups
- ou=services
- ou=resources

### 02-groups.ldif

Crea los grupos:

- ADMIN
- INVESTIGADOR
- USUARIO

### 03-users.template.ldif

Documenta la estructura de un usuario sin incluir credenciales.

### 03-users.local.ldif

Contiene los hashes generados localmente para los usuarios de laboratorio.

Está excluido mediante .gitignore y no debe versionarse.

### 04-group-members.ldif

Asigna usuarios iniciales a grupos de autorización.

## Aprovisionamiento

Las credenciales se suministran por variables de entorno.

Ejemplo:

    export LDAP_ADMIN_PASSWORD='<valor-local>'
    export LDAP_DEFAULT_USER_PASSWORD='<valor-local>'

Inicializar estructura:

    services/directory-service/scripts/bootstrap-directory.sh

Aprovisionar usuarios:

    services/directory-service/scripts/provision-users.sh

El script usa slappasswd para generar hashes antes de crear el LDIF local.

## Configuración de Auth Service

Variables soportadas:

    LDAP_URL=ldap://127.0.0.1:389
    LDAP_BASE_DN=dc=upb-cientifica,dc=local

Ejemplo:

    LDAP_URL="ldap://127.0.0.1:389" \
    LDAP_BASE_DN="dc=upb-cientifica,dc=local" \
    java -cp services/auth-services/bin auth.AuthServer

## Verificación

Estado de OpenLDAP:

    systemctl is-active slapd

Base DN:

    ldapsearch \
      -x \
      -H ldap://127.0.0.1:389 \
      -s base \
      -b "" \
      namingContexts

Usuarios:

    ldapsearch \
      -x \
      -H ldap://127.0.0.1:389 \
      -b "ou=people,dc=upb-cientifica,dc=local" \
      "(objectClass=inetOrgPerson)" \
      uid employeeNumber cn

Grupos:

    ldapsearch \
      -x \
      -H ldap://127.0.0.1:389 \
      -b "ou=groups,dc=upb-cientifica,dc=local" \
      "(objectClass=groupOfNames)" \
      cn member

## Seguridad

No se versionan:

- contraseñas LDAP;
- contraseñas administrativas;
- hashes locales de usuarios.

El entorno actual utiliza LDAP simple sobre localhost.

Antes de exponer LDAP por red en un despliegue distribuido se debe habilitar TLS/StartTLS o LDAPS.

## Pruebas realizadas

Se verificó:

- autenticación LDAP correcta;
- rechazo de contraseña incorrecta;
- dependencia de Auth Service respecto a OpenLDAP;
- login HTTP;
- autenticación Java RMI;
- validación de tokens;
- integración con Photo Service;
- integración con File Service;
- integración con Sync Service;
- integración con Streaming Service;
- rechazo de tokens inválidos.
