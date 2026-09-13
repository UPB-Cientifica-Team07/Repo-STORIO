BEGIN;

-- ============================================================
-- GRID RESOURCE
-- ============================================================
--
-- Registro de recursos heterogéneos disponibles para la
-- infraestructura Grid de UPB-CIENTÍFICA.
--
-- COMPUTE:
--   estaciones Linux/Windows capaces de aportar cómputo.
--
-- MOBILE:
--   teléfonos/tabletas integrados como recursos heterogéneos.
--
-- IOT:
--   dispositivos embebidos/sensores.
--
-- mpi_capable únicamente puede ser TRUE para COMPUTE.
-- ============================================================

CREATE TABLE IF NOT EXISTS grid_resource (
    id_recurso UUID PRIMARY KEY
        DEFAULT gen_random_uuid(),

    resource_id VARCHAR(120)
        NOT NULL UNIQUE,

    nombre VARCHAR(160)
        NOT NULL,

    tipo VARCHAR(20)
        NOT NULL,

    ip VARCHAR(45),

    cpu_cores INTEGER
        NOT NULL DEFAULT 1,

    memoria_mb BIGINT
        NOT NULL DEFAULT 0,

    mpi_capable BOOLEAN
        NOT NULL DEFAULT FALSE,

    capabilities JSONB
        NOT NULL DEFAULT '[]'::JSONB,

    estado VARCHAR(20)
        NOT NULL DEFAULT 'AVAILABLE',

    last_heartbeat TIMESTAMP WITHOUT TIME ZONE
        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    created_at TIMESTAMP WITHOUT TIME ZONE
        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP WITHOUT TIME ZONE
        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_grid_tipo
        CHECK (
            tipo IN (
                'COMPUTE',
                'MOBILE',
                'IOT'
            )
        ),

    CONSTRAINT chk_grid_estado
        CHECK (
            estado IN (
                'AVAILABLE',
                'BUSY',
                'INACTIVE'
            )
        ),

    CONSTRAINT chk_grid_cpu
        CHECK (
            cpu_cores > 0
        ),

    CONSTRAINT chk_grid_memoria
        CHECK (
            memoria_mb >= 0
        ),

    CONSTRAINT chk_grid_mpi_compute
        CHECK (
            mpi_capable = FALSE
            OR tipo = 'COMPUTE'
        ),

    CONSTRAINT chk_grid_capabilities_array
        CHECK (
            jsonb_typeof(capabilities) = 'array'
        )
);

CREATE INDEX IF NOT EXISTS idx_grid_resource_tipo
ON grid_resource(tipo);

CREATE INDEX IF NOT EXISTS idx_grid_resource_estado
ON grid_resource(estado);

CREATE INDEX IF NOT EXISTS idx_grid_resource_heartbeat
ON grid_resource(last_heartbeat);

CREATE INDEX IF NOT EXISTS idx_grid_compute_available
ON grid_resource(
    tipo,
    mpi_capable,
    estado
)
WHERE
    tipo = 'COMPUTE'
    AND mpi_capable = TRUE
    AND estado = 'AVAILABLE';

COMMIT;
