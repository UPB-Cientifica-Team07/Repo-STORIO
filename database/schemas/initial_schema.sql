-- ============================================================
-- UPB-CIENTÍFICA
-- Esquema inicial de base de datos
-- Motor: PostgreSQL
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- TABLA: usuario
-- ============================================================

CREATE TABLE usuario (
    id_usuario UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario VARCHAR(50) NOT NULL UNIQUE,
    correo VARCHAR(120) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    rol VARCHAR(30) NOT NULL,
    estado BOOLEAN NOT NULL DEFAULT TRUE,
    fecha_registro TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_usuario_rol
        CHECK (rol IN ('ADMINISTRADOR', 'INVESTIGADOR', 'USUARIO'))
);


-- ============================================================
-- TABLA: nodo_hpc
-- ============================================================

CREATE TABLE nodo_hpc (
    id_nodo UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hostname VARCHAR(100) NOT NULL UNIQUE,
    cpu INTEGER NOT NULL,
    memoria INTEGER NOT NULL,
    estado VARCHAR(20) NOT NULL,
    ip VARCHAR(45),
    ubicacion VARCHAR(100),

    CONSTRAINT chk_nodo_cpu
        CHECK (cpu > 0),

    CONSTRAINT chk_nodo_memoria
        CHECK (memoria > 0),

    CONSTRAINT chk_nodo_estado
        CHECK (estado IN ('DISPONIBLE', 'OCUPADO', 'INACTIVO'))
);


-- ============================================================
-- TABLA: archivo
-- ============================================================

CREATE TABLE archivo (
    id_archivo UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre VARCHAR(255) NOT NULL,
    ruta TEXT NOT NULL,
    tamano BIGINT NOT NULL,
    tipo_archivo VARCHAR(20) NOT NULL,
    id_usuario UUID NOT NULL,
    fecha_subida TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_archivo_usuario
        FOREIGN KEY (id_usuario)
        REFERENCES usuario(id_usuario)
        ON DELETE CASCADE,

    CONSTRAINT chk_archivo_tamano
        CHECK (tamano >= 0),

    CONSTRAINT chk_archivo_tipo
        CHECK (
            tipo_archivo IN (
                'DOCUMENTO',
                'IMAGEN',
                'VIDEO',
                'OTRO'
            )
        )
);


-- ============================================================
-- TABLA: imagen
-- ============================================================

CREATE TABLE imagen (
    id_imagen UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_archivo UUID NOT NULL UNIQUE,
    resolucion VARCHAR(30) NOT NULL,
    formato VARCHAR(10) NOT NULL,

    CONSTRAINT fk_imagen_archivo
        FOREIGN KEY (id_archivo)
        REFERENCES archivo(id_archivo)
        ON DELETE CASCADE,

    CONSTRAINT chk_imagen_formato
        CHECK (
            formato IN ('JPG', 'JPEG', 'PNG', 'WEBP')
        )
);


-- ============================================================
-- TABLA: video
-- ============================================================

CREATE TABLE video (
    id_video UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_archivo UUID NOT NULL UNIQUE,
    duracion INTEGER NOT NULL,
    calidad VARCHAR(20) NOT NULL,
    formato VARCHAR(10) NOT NULL,

    CONSTRAINT fk_video_archivo
        FOREIGN KEY (id_archivo)
        REFERENCES archivo(id_archivo)
        ON DELETE CASCADE,

    CONSTRAINT chk_video_duracion
        CHECK (duracion >= 0),

    CONSTRAINT chk_video_formato
        CHECK (
            formato IN ('MP4', 'WEBM', 'AVI', 'MKV')
        )
);


-- ============================================================
-- TABLA: token
-- ============================================================

CREATE TABLE token (
    id_token UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_usuario UUID NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expiracion TIMESTAMP NOT NULL,
    estado BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT fk_token_usuario
        FOREIGN KEY (id_usuario)
        REFERENCES usuario(id_usuario)
        ON DELETE CASCADE
);


-- ============================================================
-- TABLA: trabajo_hpc
-- ============================================================

CREATE TABLE trabajo_hpc (
    id_job UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_usuario UUID NOT NULL,
    id_nodo UUID,
    lenguaje VARCHAR(20) NOT NULL,
    estado VARCHAR(20) NOT NULL,
    inicio TIMESTAMP,
    fin TIMESTAMP,
    recursos TEXT,
    descripcion TEXT,

    CONSTRAINT fk_trabajo_usuario
        FOREIGN KEY (id_usuario)
        REFERENCES usuario(id_usuario)
        ON DELETE CASCADE,

    CONSTRAINT fk_trabajo_nodo
        FOREIGN KEY (id_nodo)
        REFERENCES nodo_hpc(id_nodo)
        ON DELETE SET NULL,

    CONSTRAINT chk_trabajo_estado
        CHECK (
            estado IN (
                'PENDIENTE',
                'EJECUTANDO',
                'FINALIZADO',
                'ERROR',
                'CANCELADO'
            )
        ),

    CONSTRAINT chk_trabajo_fechas
        CHECK (
            fin IS NULL
            OR inicio IS NULL
            OR fin >= inicio
        )
);