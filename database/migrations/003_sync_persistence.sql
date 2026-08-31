BEGIN;

-- ============================================================
-- SYNC FILE
-- ============================================================
--
-- Estado lógico de los archivos administrados por Sync.
--
-- id_archivo:
--   mismo UUID generado por File Service.
--
-- relative_path:
--   ruta lógica observada por el cliente.
--
-- Ejemplo:
--   Universidad/SistemasDistribuidos/proyecto/main.go
--
-- No contiene la ruta física con UUID usada por File Service.
--
-- ============================================================

CREATE TABLE IF NOT EXISTS sync_file (
    id_archivo UUID PRIMARY KEY,

    id_usuario UUID NOT NULL
        REFERENCES usuario(id_usuario)
        ON DELETE CASCADE,

    nombre VARCHAR(255) NOT NULL,

    mime_type VARCHAR(255),

    relative_path TEXT,

    tamano BIGINT NOT NULL
        CHECK (tamano >= 0),

    version BIGINT NOT NULL DEFAULT 1
        CHECK (version >= 1),

    eliminado BOOLEAN NOT NULL DEFAULT FALSE,

    fecha_creacion TIMESTAMP WITHOUT TIME ZONE
        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    fecha_modificacion TIMESTAMP WITHOUT TIME ZONE
        NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Un usuario no debe tener dos archivos activos
-- ocupando exactamente la misma ruta lógica.
CREATE UNIQUE INDEX IF NOT EXISTS uq_sync_file_user_path_active
ON sync_file (
    id_usuario,
    relative_path
)
WHERE
    eliminado = FALSE
    AND relative_path IS NOT NULL
    AND relative_path <> '';

CREATE INDEX IF NOT EXISTS idx_sync_file_usuario
ON sync_file(id_usuario);

CREATE INDEX IF NOT EXISTS idx_sync_file_usuario_activo
ON sync_file(id_usuario, eliminado);


-- ============================================================
-- SYNC CHANGE
-- ============================================================
--
-- Log persistente y ordenado de cambios.
--
-- id_cambio BIGSERIAL funciona como secuencia monotónica.
--
-- No colocamos FK desde id_archivo hacia archivo porque un
-- FILE_DELETED debe continuar existiendo después de eliminar
-- físicamente/metadata el archivo desde File Service.
--
-- ============================================================

CREATE TABLE IF NOT EXISTS sync_change (
    id_cambio BIGSERIAL PRIMARY KEY,

    id_usuario UUID NOT NULL
        REFERENCES usuario(id_usuario)
        ON DELETE CASCADE,

    id_archivo UUID NOT NULL,

    tipo VARCHAR(32) NOT NULL
        CHECK (
            tipo IN (
                'FILE_CREATED',
                'FILE_CHANGED',
                'FILE_DELETED'
            )
        ),

    nombre VARCHAR(255) NOT NULL,

    relative_path TEXT,

    version BIGINT NOT NULL
        CHECK (version >= 1),

    origin_device_id VARCHAR(255) NOT NULL,

    fecha_cambio TIMESTAMP WITHOUT TIME ZONE
        NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sync_change_usuario_cambio
ON sync_change(
    id_usuario,
    id_cambio
);

CREATE INDEX IF NOT EXISTS idx_sync_change_archivo
ON sync_change(id_archivo);


-- ============================================================
-- SYNC DEVICE CURSOR
-- ============================================================
--
-- Guarda hasta qué cambio persistente ha consumido
-- cada dispositivo de cada usuario.
--
-- Ejemplo:
--
-- user-003 / desktop-001 -> cambio 35
-- user-003 / mobile-001  -> cambio 29
--
-- De esta manera cada dispositivo puede avanzar
-- independientemente.
--
-- ============================================================

CREATE TABLE IF NOT EXISTS sync_device_cursor (
    id_usuario UUID NOT NULL
        REFERENCES usuario(id_usuario)
        ON DELETE CASCADE,

    device_id VARCHAR(255) NOT NULL,

    last_change_id BIGINT NOT NULL DEFAULT 0
        CHECK (last_change_id >= 0),

    fecha_actualizacion TIMESTAMP WITHOUT TIME ZONE
        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (
        id_usuario,
        device_id
    )
);

CREATE INDEX IF NOT EXISTS idx_sync_device_cursor_usuario
ON sync_device_cursor(id_usuario);

COMMIT;
