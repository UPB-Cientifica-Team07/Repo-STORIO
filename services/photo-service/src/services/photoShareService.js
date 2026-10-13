const photoShareRepository =
  require("./photoShareRepository");

const fileClient =
  require("./fileClient");

class PhotoShareService {

  async canReadPhoto(
    photo,
    user
  ) {

    if (
      !photo ||
      !user
    ) {
      return false;
    }

    if (
      photo.ownerId ===
        user.userId ||
      user.role ===
        "ADMIN"
    ) {
      return true;
    }

    return photoShareRepository
      .canRead(
        photo.fileId,
        user.userId
      );
  }

  async sharePhoto({
    photo,
    token,
    targetUserId,
    canWrite = false,
    canShare = false
  }) {

    if (!photo) {
      throw new Error(
        "Foto requerida"
      );
    }

    return fileClient.shareFile({
      token,
      fileId:
        photo.fileId,
      targetUserId,
      canRead:
        true,
      canWrite:
        Boolean(canWrite),
      canShare:
        Boolean(canShare)
    });
  }

  async revokePhotoShare({
    photo,
    token,
    targetUserId
  }) {

    if (!photo) {
      throw new Error(
        "Foto requerida"
      );
    }

    return fileClient
      .revokeFileAccess({
        token,
        fileId:
          photo.fileId,
        targetUserId
      });
  }

  async listPhotoShares(
    photo
  ) {

    if (!photo) {
      return [];
    }

    return photoShareRepository
      .listByFile(
        photo.fileId
      );
  }

  async listSharedWithUser(
    userId
  ) {

    return photoShareRepository
      .listSharedWithUser(
        userId
      );
  }
}

module.exports =
  new PhotoShareService();
