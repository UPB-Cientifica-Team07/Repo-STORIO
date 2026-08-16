## 1. Nombres de carpetas

Las carpetas utilizarán:

kebab-case

Ejemplo:

auth-service
file-service
hpc-service

---

## 2. Git

### Ramas

- main
- develop
- feature/nombre-funcionalidad
- fix/nombre-error
- docs/nombre-documento

Ejemplos:

feature/auth-login
feature/file-upload
fix/login-error
docs/arquitectura

---

## 3. Commits

Formato:

tipo: descripción

Tipos principales:

- feat: nueva funcionalidad
- fix: corrección de error
- docs: documentación
- refactor: reorganización de código
- test: pruebas
- chore: configuración o mantenimiento

Ejemplos:

feat: agregar autenticacion de usuarios

docs: agregar diagrama de contenedores

fix: corregir conexion con base de datos

chore: crear estructura inicial del repositorio

---

## 4. Código

### Java

- Clases: PascalCase
- Métodos: camelCase
- Variables: camelCase
- Constantes: UPPER_SNAKE_CASE

### Base de datos

- Tablas: snake_case
- Columnas: snake_case
- Claves primarias: id_nombre