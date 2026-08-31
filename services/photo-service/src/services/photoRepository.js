const pool =
  require("../config/database");

class PhotoRepository {

  async create({
    fileId,
    directorioId,
    tags = []
  }) {

    if (!fileId) {
      throw new Error(
        "fileId es obligatorio"
      );
    }

    if (!directorioId) {
      throw new Error(
        "directorioId es obligatorio"
      );
    }

    const client =
      await pool.connect();

    try {

      await client.query(
        "BEGIN"
      );

      // =====================================
      // VALIDAR ARCHIVO CREADO POR FILE SERVICE
      // =====================================

      const fileResult =
        await client.query(
          `
          SELECT
            a.id_archivo,
            a.id_home,
            a.id_usuario,
            a.nombre,
            a.ruta_relativa,
            a.tamano,
            a.mime_type,
            a.tipo_archivo,

            u.directorio_id

          FROM archivo a

          JOIN usuario u
            ON u.id_usuario =
               a.id_usuario

          WHERE
            a.id_archivo = $1
            AND
            u.directorio_id = $2

          FOR UPDATE
          `,
          [
            fileId,
            directorioId
          ]
        );

      if (
        fileResult.rowCount === 0
      ) {
        throw new Error(
          "El archivo no existe en File Service o no pertenece al usuario autenticado"
        );
      }

      const file =
        fileResult.rows[0];

      if (
        !file.mime_type ||
        !file.mime_type.startsWith(
          "image/"
        )
      ) {
        throw new Error(
          "El archivo registrado en File Service no es una imagen"
        );
      }

      // =====================================
      // NORMALIZAR CLASIFICACIÓN DEL ARCHIVO
      // =====================================

      if (
        file.tipo_archivo !==
        "IMAGEN"
      ) {

        await client.query(
          `
          UPDATE archivo
          SET
            tipo_archivo =
              'IMAGEN',
            fecha_modificacion =
              CURRENT_TIMESTAMP
          WHERE
            id_archivo = $1
          `,
          [
            fileId
          ]
        );
      }

      // =====================================
      // IMAGEN
      // =====================================

      const imageResult =
        await client.query(
          `
          INSERT INTO imagen (
            id_archivo,
            formato
          )
          VALUES (
            $1,
            $2
          )
          ON CONFLICT (
            id_archivo
          )
          DO UPDATE
          SET
            formato =
              EXCLUDED.formato
          RETURNING
            id_imagen
          `,
          [
            fileId,
            this.extractFormat(
              file.mime_type
            )
          ]
        );

      const imageId =
        imageResult.rows[0]
          .id_imagen;

      // =====================================
      // TAGS
      // =====================================

      for (
        const tagName
        of tags
      ) {

        const normalizedTag =
          String(
            tagName
          )
            .trim()
            .toLowerCase();

        if (!normalizedTag) {
          continue;
        }

        const tagResult =
          await client.query(
            `
            INSERT INTO tag (
              nombre
            )
            VALUES (
              $1
            )
            ON CONFLICT (
              nombre
            )
            DO UPDATE
            SET
              nombre =
                EXCLUDED.nombre
            RETURNING
              id_tag
            `,
            [
              normalizedTag
            ]
          );

        const tagId =
          tagResult.rows[0]
            .id_tag;

        await client.query(
          `
          INSERT INTO imagen_tag (
            id_imagen,
            id_tag
          )
          VALUES (
            $1,
            $2
          )
          ON CONFLICT
          DO NOTHING
          `,
          [
            imageId,
            tagId
          ]
        );
      }

      await client.query(
        "COMMIT"
      );

      return this.findById(
        imageId
      );

    } catch (error) {

      await client.query(
        "ROLLBACK"
      );

      throw error;

    } finally {

      client.release();
    }
  }

  async findById(id) {

    const result =
      await pool.query(
        `
        SELECT
          i.id_imagen
            AS id,

          a.id_archivo
            AS "fileId",

          u.directorio_id
            AS "ownerId",

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

          COALESCE(
            (
              SELECT
                ARRAY_AGG(
                  ai.id_album
                  ORDER BY
                    ai.fecha_agregada
                )

              FROM album_imagen ai

              WHERE
                ai.id_imagen =
                i.id_imagen
            ),
            '{}'
          )
            AS "albumIds",

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

        FROM imagen i

        JOIN archivo a
          ON a.id_archivo =
             i.id_archivo

        JOIN usuario u
          ON u.id_usuario =
             a.id_usuario

        WHERE
          i.id_imagen = $1
        `,
        [
          id
        ]
      );

    if (
      result.rowCount === 0
    ) {
      return null;
    }

    return result.rows[0];
  }

  async findAll() {

    const result =
      await pool.query(
        `
        SELECT
          i.id_imagen
            AS id

        FROM imagen i

        ORDER BY
          i.fecha_registro
        `
      );

    const photos = [];

    for (
      const row
      of result.rows
    ) {

      const photo =
        await this.findById(
          row.id
        );

      if (photo) {

        photos.push(
          photo
        );
      }
    }

    return photos;
  }

  async search({
    ownerId,
    name,
    tag
  }) {

    const result =
      await pool.query(
        `
        SELECT DISTINCT
          i.id_imagen
            AS id

        FROM imagen i

        JOIN archivo a
          ON a.id_archivo =
             i.id_archivo

        JOIN usuario u
          ON u.id_usuario =
             a.id_usuario

        LEFT JOIN imagen_tag it
          ON it.id_imagen =
             i.id_imagen

        LEFT JOIN tag t
          ON t.id_tag =
             it.id_tag

        WHERE
          (
            $1::text IS NULL
            OR
            u.directorio_id = $1
          )

          AND

          (
            $2::text IS NULL
            OR
            LOWER(
              a.nombre
            )
            LIKE
            LOWER(
              '%' || $2 || '%'
            )
          )

          AND

          (
            $3::text IS NULL
            OR
            LOWER(
              t.nombre
            )
            =
            LOWER(
              $3
            )
          )

        ORDER BY
          id
        `,
        [
          ownerId || null,
          name || null,
          tag || null
        ]
      );

    const photos = [];

    for (
      const row
      of result.rows
    ) {

      const photo =
        await this.findById(
          row.id
        );

      if (photo) {

        photos.push(
          photo
        );
      }
    }

    return photos;
  }

  async delete(id) {

    const result =
      await pool.query(
        `
        DELETE FROM imagen

        WHERE
          id_imagen = $1

        RETURNING
          id_archivo
        `,
        [
          id
        ]
      );

    if (
      result.rowCount === 0
    ) {
      return null;
    }

    return result.rows[0]
      .id_archivo;
  }

  extractFormat(
    mimeType
  ) {

    if (!mimeType) {
      return null;
    }

    const parts =
      mimeType.split(
        "/"
      );

    if (
      parts.length !== 2
    ) {
      return null;
    }

    return parts[1]
      .toUpperCase()
      .replace(
        "JPEG",
        "JPG"
      );
  }
}

module.exports =
  new PhotoRepository();
