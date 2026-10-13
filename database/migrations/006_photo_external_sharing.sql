-- ============================================================
-- UPB-CIENTIFICA
-- External Photo Sharing
-- ============================================================

CREATE TABLE photo_external_share (
    id_share UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    id_imagen UUID NOT NULL,

    id_creador UUID NOT NULL,

    token_hash VARCHAR(64) NOT NULL UNIQUE,

    fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    expira_en TIMESTAMPTZ NOT NULL,

    revocado_en TIMESTAMPTZ,

    CONSTRAINT fk_photo_external_share_imagen
        FOREIGN KEY (id_imagen)
        REFERENCES imagen(id_imagen)
        ON DELETE CASCADE,

    CONSTRAINT fk_photo_external_share_creador
        FOREIGN KEY (id_creador)
        REFERENCES usuario(id_usuario)
        ON DELETE CASCADE,

    CONSTRAINT chk_photo_external_share_hash
        CHECK (
            length(token_hash) = 64
        ),

    CONSTRAINT chk_photo_external_share_expiration
        CHECK (
            expira_en > fecha_creacion
        )
);

CREATE INDEX idx_photo_external_share_imagen
    ON photo_external_share(id_imagen);

CREATE INDEX idx_photo_external_share_expiration
    ON photo_external_share(expira_en);

CREATE INDEX idx_photo_external_share_active
    ON photo_external_share(
        id_imagen,
        expira_en
    )
    WHERE revocado_en IS NULL;
