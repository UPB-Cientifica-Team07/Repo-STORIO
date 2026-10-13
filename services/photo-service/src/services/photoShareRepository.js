const pool =
  require("../config/database");

class PhotoShareRepository {

  async canRead(
    fileId,
    userId
  ) {

    const result =
      await pool.query(
        `
        SELECT EXISTS (
          SELECT 1
          FROM permiso_recurso p
          JOIN usuario u
            ON u.id_usuario =
               p.id_usuario
          WHERE
            p.id_archivo =
              $1::uuid
            AND u.directorio_id =
              $2
            AND u.estado =
              true
            AND p.puede_leer =
              true
        ) AS allowed
        `,
        [
          fileId,
          userId
        ]
      );

    return Boolean(
      result.rows[0]
        ?.allowed
    );
  }

  async listByFile(
    fileId
  ) {

    const result =
      await pool.query(
        `
        SELECT
          u.directorio_id
            AS "userId",

          u.usuario
            AS username,

          p.puede_leer
            AS "canRead",

          p.puede_escribir
            AS "canWrite",

          p.puede_compartir
            AS "canShare"

        FROM permiso_recurso p

        JOIN usuario u
          ON u.id_usuario =
             p.id_usuario

        WHERE
          p.id_archivo =
            $1::uuid
          AND u.estado =
            true

        ORDER BY
          u.usuario
        `,
        [
          fileId
        ]
      );

    return result.rows;
  }

  async listSharedWithUser(
    userId
  ) {

    const result =
      await pool.query(
        `
        SELECT
          i.id_imagen::text
            AS id,

          a.id_archivo::text
            AS "fileId",

          owner.directorio_id
            AS "ownerId",

          owner.usuario
            AS "ownerUsername",

          a.nombre
            AS name,

          a.mime_type
            AS "mimeType",

          a.tamano
            AS size,

          a.ruta_relativa
            AS path,

          i.ancho
            AS width,

          i.alto
            AS height,

          i.formato
            AS format,

          i.fecha_registro
            AS "createdAt",

          p.puede_leer
            AS "canRead",

          p.puede_escribir
            AS "canWrite",

          p.puede_compartir
            AS "canShare",

          COALESCE(
            (
              SELECT
                ARRAY_AGG(
                  t.nombre
                  ORDER BY
                    t.nombre
                )

              FROM imagen_tag it

              JOIN tag t
                ON t.id_tag =
                   it.id_tag

              WHERE
                it.id_imagen =
                  i.id_imagen
            ),
            '{}'
          )
            AS tags

        FROM permiso_recurso p

        JOIN usuario grantee
          ON grantee.id_usuario =
             p.id_usuario

        JOIN archivo a
          ON a.id_archivo =
             p.id_archivo

        JOIN imagen i
          ON i.id_archivo =
             a.id_archivo

        JOIN usuario owner
          ON owner.id_usuario =
             a.id_usuario

        WHERE
          grantee.directorio_id =
            $1
          AND grantee.estado =
            true
          AND p.puede_leer =
            true

        ORDER BY
          i.fecha_registro DESC
        `,
        [
          userId
        ]
      );

    return result.rows;
  }
}

module.exports =
  new PhotoShareRepository();
