const photoService =
  require("../services/photoService");

const photoShareService =
  require("../services/photoShareService");

function errorStatus(
  error
) {

  switch (
    error?.code
  ) {

    case 3:
      return 400;

    case 5:
      return 404;

    case 7:
      return 403;

    case 16:
      return 401;

    default:
      return 400;
  }
}

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

async function sharePhoto(
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
            "No tiene permisos para compartir esta foto"
        });
    }

    const targetUserId =
      String(
        req.body
          ?.targetUserId ||
        ""
      ).trim();

    if (!targetUserId) {

      return res
        .status(400)
        .json({
          success: false,
          message:
            "targetUserId es obligatorio"
        });
    }

    if (
      targetUserId ===
        photo.ownerId
    ) {

      return res
        .status(400)
        .json({
          success: false,
          message:
            "El propietario ya posee acceso completo"
        });
    }

    const canWrite =
      req.body?.canWrite ===
        true;

    const canShare =
      req.body?.canShare ===
        true;

    const response =
      await photoShareService
        .sharePhoto({
          photo,
          token:
            req.user.token,
          targetUserId,
          canWrite,
          canShare
        });

    return res
      .status(200)
      .json({
        success: true,
        message:
          response.message ||
          "Fotografía compartida correctamente",

        share: {
          targetUserId,
          canRead: true,
          canWrite,
          canShare
        }
      });

  } catch (error) {

    return res
      .status(
        errorStatus(
          error
        )
      )
      .json({
        success: false,
        message:
          error.details ||
          error.message ||
          "Error compartiendo fotografía"
      });
  }
}

async function revokePhotoShare(
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
            "No tiene permisos para revocar esta foto"
        });
    }

    const targetUserId =
      String(
        req.params
          .targetUserId ||
        ""
      ).trim();

    if (!targetUserId) {

      return res
        .status(400)
        .json({
          success: false,
          message:
            "targetUserId es obligatorio"
        });
    }

    const response =
      await photoShareService
        .revokePhotoShare({
          photo,
          token:
            req.user.token,
          targetUserId
        });

    return res
      .status(200)
      .json({
        success: true,
        message:
          response.message ||
          "Acceso a fotografía revocado"
      });

  } catch (error) {

    return res
      .status(
        errorStatus(
          error
        )
      )
      .json({
        success: false,
        message:
          error.details ||
          error.message ||
          "Error revocando fotografía"
      });
  }
}

async function listPhotoShares(
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
            "No tiene permisos para consultar comparticiones"
        });
    }

    const shares =
      await photoShareService
        .listPhotoShares(
          photo
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

    return res
      .status(500)
      .json({
        success: false,
        message:
          "Error consultando comparticiones",
        detail:
          error.message
      });
  }
}

async function listSharedWithMe(
  req,
  res
) {

  try {

    const photos =
      await photoShareService
        .listSharedWithUser(
          req.user.userId
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
          "Error consultando fotografías compartidas",
        detail:
          error.message
      });
  }
}

module.exports = {
  sharePhoto,
  revokePhotoShare,
  listPhotoShares,
  listSharedWithMe
};
