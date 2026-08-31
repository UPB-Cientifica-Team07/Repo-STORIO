-- ============================================================
-- UPB-CIENTÍFICA
-- Esquema de base de datos v2
-- Motor: PostgreSQL
-- Arquitectura:
--   PostgreSQL = metadatos
--   Home/Repository = contenido físico
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- USUARIO
-- Identidad interna + referencia al servicio de directorio/Auth
-- ============================================================

CREATE TABLE usuario (
    id_usuario UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    directorio_id VARCHAR(100) NOT NULL UNIQUE,
    usuario VARCHAR(80) NOT NULL UNIQUE,
    correo VARCHAR(150),
    rol VARCHAR(30) NOT NULL,
    estado BOOLEAN NOT NULL DEFAULT TRUE,
    fecha_registro TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_usuario_rol
        CHECK (
            rol IN (
                'ADMIN',
                'INVESTIGADOR',
                'USUARIO'
            )
        )
);

-- ============================================================
-- HOME
-- Espacio lógico asignado a cada usuario
-- ============================================================

CREATE TABLE home (
    id_home UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_usuario UUID NOT NULL UNIQUE,
    ruta_base TEXT NOT NULL UNIQUE,
    cuota_maxima BIGINT NOT NULL,
    usado_bytes BIGINT NOT NULL DEFAULT 0,
    fecha_creacion TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_home_usuario
        FOREIGN KEY (id_usuario)
        REFERENCES usuario(id_usuario)
        ON DELETE CASCADE,

    CONSTRAINT chk_home_cuota
        CHECK (cuota_maxima >= 0),

    CONSTRAINT chk_home_usado
        CHECK (usado_bytes >= 0),

    CONSTRAINT chk_home_no_supera_cuota
        CHECK (usado_bytes <= cuota_maxima)
);

-- ============================================================
-- ARCHIVO
-- Metadatos de archivos almacenados físicamente en Home
-- ============================================================

CREATE TABLE archivo (
    id_archivo UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_home UUID NOT NULL,
    id_usuario UUID NOT NULL,
    nombre VARCHAR(255) NOT NULL,
    ruta_relativa TEXT NOT NULL,
    tamano BIGINT NOT NULL DEFAULT 0,
    mime_type VARCHAR(120),
    tipo_archivo VARCHAR(20) NOT NULL,
    permisos_unix VARCHAR(4) NOT NULL DEFAULT '0640',
    fecha_subida TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fecha_modificacion TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_archivo_home
        FOREIGN KEY (id_home)
        REFERENCES home(id_home)
        ON DELETE CASCADE,

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
        ),

    CONSTRAINT uq_archivo_home_ruta
        UNIQUE (
            id_home,
            ruta_relativa
        )
);

-- ============================================================
-- IMAGEN
-- Especialización de archivo
-- ============================================================

CREATE TABLE imagen (
    id_imagen UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_archivo UUID NOT NULL UNIQUE,
    ancho INTEGER,
    alto INTEGER,
    formato VARCHAR(20),
    fecha_registro TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_imagen_archivo
        FOREIGN KEY (id_archivo)
        REFERENCES archivo(id_archivo)
        ON DELETE CASCADE,

    CONSTRAINT chk_imagen_ancho
        CHECK (
            ancho IS NULL
            OR ancho > 0
        ),

    CONSTRAINT chk_imagen_alto
        CHECK (
            alto IS NULL
            OR alto > 0
        )
);

-- ============================================================
-- VIDEO
-- Especialización de archivo
-- ============================================================

CREATE TABLE video (
    id_video UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_archivo UUID NOT NULL UNIQUE,
    duracion_segundos INTEGER,
    calidad VARCHAR(30),
    formato VARCHAR(20),
    fecha_registro TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_video_archivo
        FOREIGN KEY (id_archivo)
        REFERENCES archivo(id_archivo)
        ON DELETE CASCADE,

    CONSTRAINT chk_video_duracion
        CHECK (
            duracion_segundos IS NULL
            OR duracion_segundos >= 0
        )
);

-- ============================================================
-- ALBUM
-- ============================================================

CREATE TABLE album (
    id_album UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_usuario UUID NOT NULL,
    nombre VARCHAR(150) NOT NULL,
    descripcion TEXT,
    fecha_creacion TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fecha_modificacion TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_album_usuario
        FOREIGN KEY (id_usuario)
        REFERENCES usuario(id_usuario)
        ON DELETE CASCADE
);

-- ============================================================
-- ALBUM_IMAGEN
-- Relación N:M
-- Una imagen puede estar en varios álbumes
-- ============================================================

CREATE TABLE album_imagen (
    id_album UUID NOT NULL,
    id_imagen UUID NOT NULL,
    fecha_agregada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (
        id_album,
        id_imagen
    ),

    CONSTRAINT fk_album_imagen_album
        FOREIGN KEY (id_album)
        REFERENCES album(id_album)
        ON DELETE CASCADE,

    CONSTRAINT fk_album_imagen_imagen
        FOREIGN KEY (id_imagen)
        REFERENCES imagen(id_imagen)
        ON DELETE CASCADE
);

-- ============================================================
-- TAG
-- ============================================================

CREATE TABLE tag (
    id_tag UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre VARCHAR(80) NOT NULL UNIQUE
);

-- ============================================================
-- IMAGEN_TAG
-- Relación N:M
-- ============================================================

CREATE TABLE imagen_tag (
    id_imagen UUID NOT NULL,
    id_tag UUID NOT NULL,

    PRIMARY KEY (
        id_imagen,
        id_tag
    ),

    CONSTRAINT fk_imagen_tag_imagen
        FOREIGN KEY (id_imagen)
        REFERENCES imagen(id_imagen)
        ON DELETE CASCADE,

    CONSTRAINT fk_imagen_tag_tag
        FOREIGN KEY (id_tag)
        REFERENCES tag(id_tag)
        ON DELETE CASCADE
);

-- ============================================================
-- PERMISO_RECURSO
-- Complementa permisos Unix para compartir recursos
-- ============================================================

CREATE TABLE permiso_recurso (
    id_permiso UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_archivo UUID NOT NULL,
    id_usuario UUID NOT NULL,
    puede_leer BOOLEAN NOT NULL DEFAULT TRUE,
    puede_escribir BOOLEAN NOT NULL DEFAULT FALSE,
    puede_compartir BOOLEAN NOT NULL DEFAULT FALSE,

    CONSTRAINT fk_permiso_archivo
        FOREIGN KEY (id_archivo)
        REFERENCES archivo(id_archivo)
        ON DELETE CASCADE,

    CONSTRAINT fk_permiso_usuario
        FOREIGN KEY (id_usuario)
        REFERENCES usuario(id_usuario)
        ON DELETE CASCADE,

    CONSTRAINT uq_permiso_archivo_usuario
        UNIQUE (
            id_archivo,
            id_usuario
        )
);

-- ============================================================
-- NODO HPC
-- ============================================================

CREATE TABLE nodo_hpc (
    id_nodo UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hostname VARCHAR(100) NOT NULL UNIQUE,
    cpu INTEGER NOT NULL,
    memoria_mb INTEGER NOT NULL,
    estado VARCHAR(20) NOT NULL,
    ip VARCHAR(45),
    ubicacion VARCHAR(120),

    CONSTRAINT chk_nodo_cpu
        CHECK (cpu > 0),

    CONSTRAINT chk_nodo_memoria
        CHECK (memoria_mb > 0),

    CONSTRAINT chk_nodo_estado
        CHECK (
            estado IN (
                'DISPONIBLE',
                'OCUPADO',
                'INACTIVO'
            )
        )
);

-- ============================================================
-- TRABAJO HPC
-- ============================================================

CREATE TABLE trabajo_hpc (
    id_job UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id_usuario UUID NOT NULL,
    id_nodo UUID,
    lenguaje VARCHAR(30) NOT NULL,
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

-- ============================================================
-- ÍNDICES
-- ============================================================

CREATE INDEX idx_usuario_directorio
    ON usuario(directorio_id);

CREATE INDEX idx_archivo_usuario
    ON archivo(id_usuario);

CREATE INDEX idx_archivo_home
    ON archivo(id_home);

CREATE INDEX idx_archivo_tipo
    ON archivo(tipo_archivo);

CREATE INDEX idx_imagen_archivo
    ON imagen(id_archivo);

CREATE INDEX idx_album_usuario
    ON album(id_usuario);

CREATE INDEX idx_album_imagen_album
    ON album_imagen(id_album);

CREATE INDEX idx_album_imagen_imagen
    ON album_imagen(id_imagen);

CREATE INDEX idx_imagen_tag_imagen
    ON imagen_tag(id_imagen);

CREATE INDEX idx_permiso_usuario
    ON permiso_recurso(id_usuario);

CREATE INDEX idx_trabajo_usuario
    ON trabajo_hpc(id_usuario);

CREATE INDEX idx_trabajo_estado
    ON trabajo_hpc(estado);