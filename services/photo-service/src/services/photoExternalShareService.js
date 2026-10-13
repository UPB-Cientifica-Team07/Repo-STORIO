const crypto =
  require("crypto");

const fs =
  require("fs/promises");

const path =
  require("path");

const config =
  require("../config/config");

const repository =
  require("./photoExternalShareRepository");

function httpError(
  status,
  message
) {

  const error =
    new Error(
      message
    );

  error.status =
    status;

  return error;
}

function hashToken(
  token
) {

  return crypto
    .createHash(
      "sha256"
    )
    .update(
      token,
      "utf8"
    )
    .digest(
      "hex"
    );
}

function validateRawToken(
  token
) {

  if (
    typeof token !==
      "string" ||
    token.length < 40 ||
    token.length > 128 ||
    !/^[A-Za-z0-9_-]+$/.test(
      token
    )
  ) {

    throw httpError(
      404,
      "Enlace externo no encontrado"
    );
  }
}

function publicMetadata(
  share
) {

  return {
    id:
      share.photoId,

    name:
      share.name,

    mimeType:
      share.mimeType,

    size:
      Number(
        share.size || 0
      ),

    createdAt:
      share.photoCreatedAt,

    expiresAt:
      share.expiresAt
  };
}

class PhotoExternalShareService {

  async create({
    photo,
    user,
    expiresInHours = 24
  }) {

    if (
      !photo ||
      !user
    ) {

      throw new Error(
        "Foto y usuario son obligatorios"
      );
    }

    const hours =
      Number(
        expiresInHours
      );

    if (
      !Number.isFinite(
        hours
      ) ||
      hours <= 0 ||
      hours > 720
    ) {

      throw httpError(
        400,
        "expiresInHours debe estar entre 0 y 720"
      );
    }

    const token =
      crypto
        .randomBytes(
          32
        )
        .toString(
          "base64url"
        );

    const tokenHash =
      hashToken(
        token
      );

    const expiresAt =
      new Date(
        Date.now() +
        hours *
        60 *
        60 *
        1000
      );

    const share =
      await repository
        .create({
          photoId:
            photo.id,

          creatorUserId:
            user.userId,

          tokenHash,

          expiresAt:
            expiresAt.toISOString()
        });

    return {
      token,
      share: {
        id:
          share.id,

        createdAt:
          share.createdAt,

        expiresAt:
          share.expiresAt
      }
    };
  }

  async resolve(
    token
  ) {

    validateRawToken(
      token
    );

    const share =
      await repository
        .findByTokenHash(
          hashToken(
            token
          )
        );

    if (!share) {

      throw httpError(
        404,
        "Enlace externo no encontrado"
      );
    }

    if (share.revokedAt) {

      throw httpError(
        410,
        "Enlace externo revocado"
      );
    }

    if (
      new Date(
        share.expiresAt
      ).getTime() <=
      Date.now()
    ) {

      throw httpError(
        410,
        "Enlace externo expirado"
      );
    }

    return share;
  }

  async getMetadata(
    token
  ) {

    const share =
      await this.resolve(
        token
      );

    return publicMetadata(
      share
    );
  }

  async getContent(
    token
  ) {

    const share =
      await this.resolve(
        token
      );

    const root =
      path.resolve(
        config.sharedStorageRoot
      );

    const absolutePath =
      path.resolve(
        root,
        share.path
      );

    if (
      absolutePath !== root &&
      !absolutePath.startsWith(
        root +
        path.sep
      )
    ) {

      throw httpError(
        500,
        "Ruta física de fotografía inválida"
      );
    }

    let content;

    try {

      content =
        await fs.readFile(
          absolutePath
        );

    } catch (error) {

      if (
        error &&
        error.code ===
          "ENOENT"
      ) {

        throw httpError(
          404,
          "Contenido físico no encontrado"
        );
      }

      throw error;
    }

    return {
      content,

      name:
        share.name,

      mimeType:
        share.mimeType ||
        "application/octet-stream",

      expiresAt:
        share.expiresAt
    };
  }

  async list(
    photoId
  ) {

    return repository
      .listByPhoto(
        photoId
      );
  }

  async revoke(
    photoId,
    shareId
  ) {

    return repository
      .revoke(
        photoId,
        shareId
      );
  }
}

module.exports =
  new PhotoExternalShareService();
