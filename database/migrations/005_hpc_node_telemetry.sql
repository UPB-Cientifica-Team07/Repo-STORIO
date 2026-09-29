BEGIN;

-- ============================================================
-- HPC NODE TELEMETRY
-- ============================================================
--
-- cpu_usage:
--   porcentaje de utilización de CPU del sistema operativo.
--
-- memory_usage:
--   porcentaje de memoria física utilizada.
--
-- storage_usage:
--   bytes utilizados en el filesystem observado por el worker.
--
-- last_heartbeat:
--   timestamp registrado por el Coordinator al recibir heartbeat.
--
-- No recrea nodo_hpc ni trabajo_hpc.
-- ============================================================

ALTER TABLE nodo_hpc
    ADD COLUMN IF NOT EXISTS cpu_usage
        DOUBLE PRECISION
        NOT NULL
        DEFAULT 0,

    ADD COLUMN IF NOT EXISTS memory_usage
        DOUBLE PRECISION
        NOT NULL
        DEFAULT 0,

    ADD COLUMN IF NOT EXISTS storage_usage
        BIGINT
        NOT NULL
        DEFAULT 0,

    ADD COLUMN IF NOT EXISTS last_heartbeat
        TIMESTAMP WITH TIME ZONE;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_nodo_cpu_usage'
    ) THEN
        ALTER TABLE nodo_hpc
            ADD CONSTRAINT chk_nodo_cpu_usage
            CHECK (
                cpu_usage >= 0
                AND cpu_usage <= 100
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_nodo_memory_usage'
    ) THEN
        ALTER TABLE nodo_hpc
            ADD CONSTRAINT chk_nodo_memory_usage
            CHECK (
                memory_usage >= 0
                AND memory_usage <= 100
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_nodo_storage_usage'
    ) THEN
        ALTER TABLE nodo_hpc
            ADD CONSTRAINT chk_nodo_storage_usage
            CHECK (
                storage_usage >= 0
            );
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS
    idx_nodo_hpc_last_heartbeat
ON nodo_hpc(last_heartbeat);

COMMIT;
