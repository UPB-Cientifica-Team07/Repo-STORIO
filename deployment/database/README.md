# PostgreSQL nativo - UPB-CIENTIFICA

La base de datos de UPB-CIENTIFICA se despliega directamente sobre
Linux utilizando PostgreSQL nativo.

No se utilizan contenedores para la base de datos del proyecto.

## Runtime oficial

Host: 127.0.0.1
Puerto: 5434
Base de datos: upb_cientifica
Rol de aplicacion: upb_app

Las credenciales no se almacenan en el repositorio.

Los servicios reciben la configuracion mediante variables de entorno:

POSTGRES_HOST
POSTGRES_PORT
POSTGRES_DB
POSTGRES_USER
POSTGRES_PASSWORD

## Instalacion

PostgreSQL debe instalarse mediante el gestor de paquetes del sistema
operativo Linux utilizado por la infraestructura.

El cluster PostgreSQL utilizado por UPB-CIENTIFICA debe escuchar en
el puerto 5434.

La base de datos upb_cientifica y el rol upb_app deben crearse
administrativamente en el servidor.

La contrasena del rol debe proporcionarse fuera del repositorio.

## Esquema

Para preparar una base nueva se aplica primero:

database/schemas/schema_v2.sql

Luego se aplican en orden las migraciones disponibles en:

database/migrations/

Entre ellas se encuentran las migraciones para:

- persistencia de sincronizacion;
- recursos Grid;
- telemetria de nodos HPC.

Los archivos de database/seeds corresponden a datos de prueba y no
son obligatorios para produccion.

## Verificacion

Desde la raiz del repositorio puede ejecutarse:

POSTGRES_PASSWORD="<valor-seguro>" deployment/database/verify-native.sh

El script verifica:

- disponibilidad de PostgreSQL;
- puerto 5434;
- base upb_cientifica;
- usuario upb_app;
- tablas principales requeridas por la plataforma.
