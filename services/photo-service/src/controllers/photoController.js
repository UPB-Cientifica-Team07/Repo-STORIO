const photoService =
  require("../services/photoService");

function parseTags(
  value
) {

  if (!value) {
    return [];
  }

  if (
    Array.isArray(value)
  ) {
    return value;
  }

  try {

    const parsed =
      JSON.parse(
        value
      );

    if (
      Array.isArray(parsed)
    ) {
      return parsed;
    }

  } catch (error) {
  }

  return String(
    value
  )
    .split(",")
    .map(
      tag =>
        tag.trim()
    )
    .filter(Boolean);
}

// =====================================
// CREATE
// =====================================

async function createPhoto(
  req,
  res
) {

  try {

    if (!req.file) {

      return res
        .status(400)
        .json({
          success: false,
          message:
            "Debe enviar una imagen en el campo 'file'"
        });
    }

    const photo =
      await photoService.createPhoto({
        ownerId:
          req.user.userId,

        token:
          req.user.token,

        file:
          req.file,

        tags:
          parseTags(
            req.body.tags
          )
      });

    return res
      .status(201)
      .json({
        success: true,
        message:
          "Foto almacenada correctamente",
        photo
      });

  } catch (error) {

    return res
      .status(400)
      .json({
        success: false,
        message:
          error.message
      });
  }
}

// =====================================
// LIST
// =====================================

async function listPhotos(
  req,
  res
) {

  try {

    const filters = {

      ownerId:
        req.user.role ===
        "ADMIN"
          ? req.query.ownerId
          : req.user.userId,

      name:
        req.query.name,

      tag:
        req.query.tag
    };

    const photos =
      await photoService.searchPhotos(
        filters
      );

    return res
      .status(200)
      .json({
        success: true,
        total:
          photos.length,
        photos
      });

  } catch (error) {

    return res
      .status(500)
      .json({
        success: false,
        message:
          "Error consultando fotos",
        detail:
          error.message
      });
  }
}

// =====================================
// GET
// =====================================

async function getPhoto(
  req,
  res
) {

  try {

    const photo =
      await photoService.getPhoto(
        req.params.id
      );

    if (!photo) {

      return res
        .status(404)
        .json({
          success: false,
          message:
            "Foto no encontrada"
        });
    }

    if (
      photo.ownerId !==
        req.user.userId &&
      req.user.role !==
        "ADMIN"
    ) {

      return res
        .status(403)
        .json({
          success: false,
          message:
            "No tiene permisos para acceder a esta foto"
        });
    }

    return res
      .status(200)
      .json({
        success: true,
        photo
      });

  } catch (error) {

    return res
      .status(500)
      .json({
        success: false,
        message:
          "Error consultando foto",
        detail:
          error.message
      });
  }
}

// =====================================
// CONTENT
// =====================================

async function getPhotoContent(
  req,
  res
) {

  try {

    const photo =
      await photoService.getPhoto(
        req.params.id
      );

    if (!photo) {

      return res
        .status(404)
        .json({
          success: false,
          message:
            "Foto no encontrada"
        });
    }

    if (
      photo.ownerId !==
        req.user.userId &&
      req.user.role !==
        "ADMIN"
    ) {

      return res
        .status(403)
        .json({
          success: false,
          message:
            "No tiene permisos para acceder a esta foto"
        });
    }

    const result =
      await photoService.getPhotoContent(
        req.params.id,
        req.user.token
      );

    if (!result) {

      return res
        .status(404)
        .json({
          success: false,
          message:
            "Contenido de foto no encontrado"
        });
    }

    res.setHeader(
      "Content-Type",
      result.mimeType ||
      "application/octet-stream"
    );

    res.setHeader(
      "Content-Length",
      result.content.length
    );

    res.setHeader(
      "Content-Disposition",
      `inline; filename="${result.fileName}"`
    );

    return res
      .status(200)
      .send(
        result.content
      );

  } catch (error) {

    return res
      .status(500)
      .json({
        success: false,
        message:
          "Error obteniendo contenido de foto",
        detail:
          error.message
      });
  }
}

// =====================================
// DELETE
// =====================================

async function deletePhoto(
  req,
  res
) {

  try {

    const photo =
      await photoService.getPhoto(
        req.params.id
      );

    if (!photo) {

      return res
        .status(404)
        .json({
          success: false,
          message:
            "Foto no encontrada"
        });
    }

    if (
      photo.ownerId !==
        req.user.userId &&
      req.user.role !==
        "ADMIN"
    ) {

      return res
        .status(403)
        .json({
          success: false,
          message:
            "No tiene permisos para eliminar esta foto"
        });
    }

    const deleted =
      await photoService.deletePhoto(
        req.params.id,
        req.user.token
      );

    if (!deleted) {

      return res
        .status(404)
        .json({
          success: false,
          message:
            "Foto no encontrada"
        });
    }

    return res
      .status(200)
      .json({
        success: true,
        message:
          "Foto eliminada correctamente"
      });

  } catch (error) {

    return res
      .status(500)
      .json({
        success: false,
        message:
          error.message
      });
  }
}

module.exports = {
  createPhoto,
  listPhotos,
  getPhoto,
  getPhotoContent,
  deletePhoto
};