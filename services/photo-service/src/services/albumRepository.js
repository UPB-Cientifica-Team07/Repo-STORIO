const pool = require("../config/database");

class AlbumRepository {
  async create({
    ownerId,
    name,
    description = ""
  }) {
    const result =
      await pool.query(
        `
        INSERT INTO album (
          id_usuario,
          nombre,
          descripcion
        )
        SELECT
          id_usuario,
          $2,
          $3
        FROM usuario
        WHERE directorio_id = $1
        RETURNING
          id_album AS id,
          nombre AS name,
          descripcion AS description,
          fecha_creacion AS "createdAt"
        `,
        [
          ownerId,
          name,
          description
        ]
      );

    if (result.rowCount === 0) {
      throw new Error(
        "Usuario no encontrado"
      );
    }

    return this.findById(
      result.rows[0].id
    );
  }

  async findById(id) {
    const result =
      await pool.query(
        `
        SELECT
          al.id_album AS id,
          u.directorio_id AS "ownerId",
          al.nombre AS name,
          al.descripcion AS description,
          al.fecha_creacion AS "createdAt",
          COALESCE(
            ARRAY_AGG(
              ai.id_imagen
            )
            FILTER (
              WHERE ai.id_imagen IS NOT NULL
            ),
            '{}'
          ) AS "photoIds"
        FROM album al
        JOIN usuario u
          ON u.id_usuario = al.id_usuario
        LEFT JOIN album_imagen ai
          ON ai.id_album = al.id_album
        WHERE al.id_album = $1
        GROUP BY
          al.id_album,
          u.directorio_id,
          al.nombre,
          al.descripcion,
          al.fecha_creacion
        `,
        [id]
      );

    if (result.rowCount === 0) {
      return null;
    }

    return result.rows[0];
  }

  async findAll() {
    const result =
      await pool.query(
        `
        SELECT
          al.id_album AS id
        FROM album al
        ORDER BY al.fecha_creacion
        `
      );

    const albums = [];

    for (const row of result.rows) {
      const album =
        await this.findById(
          row.id
        );

      if (album) {
        albums.push(album);
      }
    }

    return albums;
  }

  async findByOwner(ownerId) {
    const result =
      await pool.query(
        `
        SELECT
          al.id_album AS id
        FROM album al
        JOIN usuario u
          ON u.id_usuario = al.id_usuario
        WHERE u.directorio_id = $1
        ORDER BY al.fecha_creacion
        `,
        [ownerId]
      );

    const albums = [];

    for (const row of result.rows) {
      const album =
        await this.findById(
          row.id
        );

      if (album) {
        albums.push(album);
      }
    }

    return albums;
  }

  async addPhoto(
    albumId,
    photoId
  ) {
    await pool.query(
      `
      INSERT INTO album_imagen (
        id_album,
        id_imagen
      )
      VALUES ($1, $2)
      ON CONFLICT DO NOTHING
      `,
      [
        albumId,
        photoId
      ]
    );

    return this.findById(
      albumId
    );
  }

  async removePhoto(
    albumId,
    photoId
  ) {
    await pool.query(
      `
      DELETE FROM album_imagen
      WHERE id_album = $1
        AND id_imagen = $2
      `,
      [
        albumId,
        photoId
      ]
    );

    return this.findById(
      albumId
    );
  }

  async delete(id) {
    const result =
      await pool.query(
        `
        DELETE FROM album
        WHERE id_album = $1
        RETURNING id_album
        `,
        [id]
      );

    return result.rowCount > 0;
  }
}

module.exports =
  new AlbumRepository();
