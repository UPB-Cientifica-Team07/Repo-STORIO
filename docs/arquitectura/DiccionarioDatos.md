# Diccionario de Datos

## Usuario

| Campo | Tipo | Descripción |
|---|---|---|
| id_usuario | UUID | Identificador único del usuario |
| usuario | VARCHAR(50) | Nombre de usuario |
| correo | VARCHAR(120) | Correo electrónico |
| password_hash | VARCHAR(255) | Contraseña almacenada como hash |
| rol | VARCHAR(30) | Rol asignado al usuario |
| estado | BOOLEAN | Estado de la cuenta |

---

## Archivo

| Campo | Tipo | Descripción |
|---|---|---|
| id_archivo | UUID | Identificador único |
| nombre | VARCHAR(255) | Nombre del archivo |
| ruta | TEXT | Ubicación del archivo |
| tamaño | BIGINT | Tamaño en bytes |
| propietario | UUID | Usuario propietario |
| fecha_subida | TIMESTAMP | Fecha de carga |

---

## Imagen

| Campo | Tipo | Descripción |
|---|---|---|
| id_imagen | UUID | Identificador |
| archivo | UUID | Archivo asociado |
| resolución | VARCHAR(30) | Resolución de la imagen |
| formato | VARCHAR(10) | Formato del archivo |

---

## Video

| Campo | Tipo | Descripción |
|---|---|---|
| id_video | UUID | Identificador |
| archivo | UUID | Archivo asociado |
| duración | INTEGER | Duración en segundos |
| calidad | VARCHAR(20) | Calidad del video |

---

## Trabajo HPC

| Campo | Tipo | Descripción |
|---|---|---|
| id_job | UUID | Identificador del trabajo |
| usuario | UUID | Usuario propietario |
| lenguaje | VARCHAR(20) | Tecnología utilizada |
| estado | VARCHAR(20) | Estado del trabajo |
| inicio | TIMESTAMP | Fecha de inicio |
| fin | TIMESTAMP | Fecha de finalización |

---

## Nodo HPC

| Campo | Tipo | Descripción |
|---|---|---|
| id_nodo | UUID | Identificador del nodo |
| hostname | VARCHAR(100) | Nombre del nodo |
| cpu | INTEGER | Número de núcleos |
| memoria | INTEGER | Memoria disponible |
| estado | VARCHAR(20) | Estado actual |

---

## Token

| Campo | Tipo | Descripción |
|---|---|---|
| id_token | UUID | Identificador |
| usuario | UUID | Usuario asociado |
| token | TEXT | Token de autenticación |
| expiración | TIMESTAMP | Fecha de expiración |