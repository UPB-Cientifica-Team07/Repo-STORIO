const pool =
  require("../config/database");

class PhotoRepository {

  async create({
    fileId,
    directorioId,
    name,
    mimeType,
    size,
    path,
    tags = []
  }) {

    const client =
      await pool.connect();

    try {

      await client.query(
        "BEGIN"
      );

      // =====================================
      // USUARIO + HOME
      // =====================================

      const userResult =
        await client.query(
          `
          SELECT
            u.id_usuario,
            h.id_home
          FROM usuario u
          JOIN home h
            ON h.id_usuario =
               u.id_usuario
          WHERE u.directorio_id = $1
          `,
          [
            directorioId
          ]
        );

      if (
        userResult.rowCount === 0
      ) {
        throw new Error(
          "Usuario o Home no encontrado"
        );
      }

      const {
        id_usuario: userId,
        id_home: homeId
      } =
        userResult.rows[0];

      // =====================================
      // ARCHIVO
      // =====================================

      await client.query(
        `
        INSERT INTO archivo (
          id_archivo,
          id_home,
          id_usuario,
          nombre,
          ruta_relativa,
          tamano,
          mime_type,
          tipo_archivo
        )
        VALUES (
          $1,
          $2,
          $3,
          $4,
          $5,
          $6,
          $7,
          'IMAGEN'
        )
        `,
        [
          fileId,
          homeId,
          userId,
          name,
          path,
          size,
          mimeType
        ]
      );

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
          RETURNING id_imagen
          `,
          [
            fileId,
            this.extractFormat(
              mimeType
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
            VALUES ($1)
            ON CONFLICT (nombre)
            DO UPDATE
            SET nombre =
                EXCLUDED.nombre
            RETURNING id_tag
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
          ON CONFLICT DO NOTHING
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
          i.id_imagen AS id,

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
          ) AS "albumIds",

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
          ) AS tags

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
          i.id_imagen AS id
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
          i.id_imagen AS id

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
            LOWER(a.nombre)
            LIKE LOWER(
              '%' || $2 || '%'
            )
          )
          AND
          (
            $3::text IS NULL
            OR
            LOWER(t.nombre)
            =
            LOWER($3)
          )
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
        DELETE FROM archivo
        WHERE id_archivo = (
          SELECT
            id_archivo
          FROM imagen
          WHERE
            id_imagen = $1
        )
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
      mimeType.split("/");

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
