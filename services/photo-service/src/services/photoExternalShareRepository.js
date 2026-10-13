const pool =
  require("../config/database");

class PhotoExternalShareRepository {

  async create({
    photoId,
    creatorUserId,
    tokenHash,
    expiresAt
  }) {

    const result =
      await pool.query(
        `
        INSERT INTO photo_external_share (
          id_imagen,
          id_creador,
          token_hash,
          expira_en
        )
        SELECT
          $1::uuid,
          u.id_usuario,
          $3,
          $4::timestamptz
        FROM usuario u
        WHERE
          u.directorio_id = $2
          AND u.estado = true
        RETURNING
          id_share::text
            AS id,
          fecha_creacion
            AS "createdAt",
          expira_en
            AS "expiresAt"
        `,
        [
          photoId,
          creatorUserId,
          tokenHash,
          expiresAt
        ]
      );

    if (result.rowCount !== 1) {
      throw new Error(
        "Usuario creador no encontrado o inactivo"
      );
    }

    return result.rows[0];
  }

  async findByTokenHash(
    tokenHash
  ) {

    const result =
      await pool.query(
        `
        SELECT
          s.id_share::text
            AS "shareId",

          s.id_imagen::text
            AS "photoId",

          i.id_archivo::text
            AS "fileId",

          a.nombre
            AS name,

          a.mime_type
            AS "mimeType",

          a.tamano
            AS size,

          a.ruta_relativa
            AS path,

          i.fecha_registro
            AS "photoCreatedAt",

          s.fecha_creacion
            AS "shareCreatedAt",

          s.expira_en
            AS "expiresAt",

          s.revocado_en
            AS "revokedAt"

        FROM photo_external_share s

        JOIN imagen i
          ON i.id_imagen =
             s.id_imagen

        JOIN archivo a
          ON a.id_archivo =
             i.id_archivo

        WHERE
          s.token_hash =
            $1

        LIMIT 1
        `,
        [
          tokenHash
        ]
      );

    return result.rows[0] || null;
  }

  async listByPhoto(
    photoId
  ) {

    const result =
      await pool.query(
        `
        SELECT
          s.id_share::text
            AS id,

          creator.directorio_id
            AS "createdBy",

          s.fecha_creacion
            AS "createdAt",

          s.expira_en
            AS "expiresAt",

          s.revocado_en
            AS "revokedAt",

          (
            s.revocado_en IS NULL
            AND s.expira_en > CURRENT_TIMESTAMP
          )
            AS active

        FROM photo_external_share s

        JOIN usuario creator
          ON creator.id_usuario =
             s.id_creador

        WHERE
          s.id_imagen =
            $1::uuid

        ORDER BY
          s.fecha_creacion DESC
        `,
        [
          photoId
        ]
      );

    return result.rows;
  }

  async revoke(
    photoId,
    shareId
  ) {

    const result =
      await pool.query(
        `
        UPDATE photo_external_share
        SET
          revocado_en =
            CURRENT_TIMESTAMP
        WHERE
          id_imagen =
            $1::uuid
          AND id_share =
            $2::uuid
          AND revocado_en
            IS NULL
        `,
        [
          photoId,
          shareId
        ]
      );

    return result.rowCount === 1;
  }
}

module.exports =
  new PhotoExternalShareRepository();
