const photoService =
  require("../services/photoService");

const externalShareService =
  require("../services/photoExternalShareService");

function canManage(
  photo,
  user
) {

  return Boolean(
    photo &&
    user &&
    (
      photo.ownerId ===
        user.userId ||
      user.role ===
        "ADMIN"
    )
  );
}

function sendError(
  res,
  error,
  fallback
) {

  return res
    .status(
      Number(
        error?.status ||
        500
      )
    )
    .json({
      success: false,
      message:
        error?.message ||
        fallback
    });
}

async function createExternalShare(
  req,
  res
) {

  try {

    const photo =
      await photoService
        .getPhoto(
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
      !canManage(
        photo,
        req.user
      )
    ) {

      return res
        .status(403)
        .json({
          success: false,
          message:
            "No tiene permisos para crear enlaces externos"
        });
    }

    const result =
      await externalShareService
        .create({
          photo,
          user:
            req.user,
          expiresInHours:
            req.body
              ?.expiresInHours ??
            24
        });

    return res
      .status(201)
      .json({
        success: true,

        message:
          "Enlace externo creado correctamente",

        token:
          result.token,

        share:
          result.share,

        publicPath:
          `/photos/public/${result.token}`,

        contentPath:
          `/photos/public/${result.token}/content`
      });

  } catch (error) {

    return sendError(
      res,
      error,
      "Error creando enlace externo"
    );
  }
}

async function listExternalShares(
  req,
  res
) {

  try {

    const photo =
      await photoService
        .getPhoto(
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
      !canManage(
        photo,
        req.user
      )
    ) {

      return res
        .status(403)
        .json({
          success: false,
          message:
            "No tiene permisos para consultar enlaces externos"
        });
    }

    const shares =
      await externalShareService
        .list(
          photo.id
        );

    return res
      .status(200)
      .json({
        success: true,
        total:
          shares.length,
        shares
      });

  } catch (error) {

    return sendError(
      res,
      error,
      "Error consultando enlaces externos"
    );
  }
}

async function revokeExternalShare(
  req,
  res
) {

  try {

    const photo =
      await photoService
        .getPhoto(
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
      !canManage(
        photo,
        req.user
      )
    ) {

      return res
        .status(403)
        .json({
          success: false,
          message:
            "No tiene permisos para revocar enlaces externos"
        });
    }

    const revoked =
      await externalShareService
        .revoke(
          photo.id,
          req.params.shareId
        );

    if (!revoked) {

      return res
        .status(404)
        .json({
          success: false,
          message:
            "Enlace externo no encontrado o ya revocado"
        });
    }

    return res
      .status(200)
      .json({
        success: true,
        message:
          "Enlace externo revocado correctamente"
      });

  } catch (error) {

    return sendError(
      res,
      error,
      "Error revocando enlace externo"
    );
  }
}

async function getExternalPhoto(
  req,
  res
) {

  try {

    const photo =
      await externalShareService
        .getMetadata(
          req.params.token
        );

    return res
      .status(200)
      .json({
        success: true,
        photo
      });

  } catch (error) {

    return sendError(
      res,
      error,
      "Error consultando enlace externo"
    );
  }
}

async function getExternalPhotoContent(
  req,
  res
) {

  try {

    const result =
      await externalShareService
        .getContent(
          req.params.token
        );

    res.setHeader(
      "Content-Type",
      result.mimeType
    );

    res.setHeader(
      "Content-Length",
      result.content.length
    );

    res.setHeader(
      "Content-Disposition",
      "inline; filename*=UTF-8''" +
      encodeURIComponent(
        result.name
      )
    );

    res.setHeader(
      "Cache-Control",
      "private, no-store"
    );

    return res
      .status(200)
      .send(
        result.content
      );

  } catch (error) {

    return sendError(
      res,
      error,
      "Error obteniendo fotografía externa"
    );
  }
}

module.exports = {
  createExternalShare,
  listExternalShares,
  revokeExternalShare,
  getExternalPhoto,
  getExternalPhotoContent
};
