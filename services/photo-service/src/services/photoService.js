const photoRepository =
  require("./photoRepository");

const fileClient =
  require("./fileClient");

class PhotoService {

  // =====================================
  // CREAR FOTO
  // =====================================

  async createPhoto({
    ownerId,
    token,
    file,
    tags = []
  }) {

    if (!ownerId) {
      throw new Error(
        "ownerId es obligatorio"
      );
    }

    if (!token) {
      throw new Error(
        "token es obligatorio"
      );
    }

    if (!file) {
      throw new Error(
        "La imagen es obligatoria"
      );
    }

    if (
      !file.mimetype ||
      !file.mimetype.startsWith(
        "image/"
      )
    ) {
      throw new Error(
        "El archivo debe ser una imagen"
      );
    }

    // =====================================
    // 1. FILE SERVICE
    // =====================================

    const uploadResult =
      await fileClient.uploadFile({
        token,

        // Este userId se envía por compatibilidad
        // con el proto, pero File Service toma
        // la identidad real desde el token.
        userId:
          ownerId,

        fileName:
          file.originalname,

        fileType:
          file.mimetype,

        content:
          file.buffer
      });

    const fileId =
      uploadResult.fileId;

    try {

      // =====================================
      // 2. POSTGRESQL
      // =====================================

      return await photoRepository.create({
        fileId,

        directorioId:
          ownerId,

        tags
      });

    } catch (error) {

      // =====================================
      // COMPENSACIÓN CREATE
      // =====================================

      try {

        await fileClient.deleteFile(
          fileId,
          token
        );

        console.warn(
          `[Photo] CREATE compensado: archivo ${fileId} eliminado de File Service`
        );

      } catch (
        compensationError
      ) {

        console.error(
          "[Photo] ERROR CRÍTICO en compensación CREATE:",
          compensationError.message
        );
      }

      throw error;
    }
  }

  // =====================================
  // GET
  // =====================================

  async getPhoto(
    id
  ) {

    return photoRepository.findById(
      id
    );
  }

  // =====================================
  // GET CONTENT
  // =====================================

  async getPhotoContent(
    id,
    token
  ) {

    if (!token) {
      throw new Error(
        "token es obligatorio"
      );
    }

    const photo =
      await photoRepository.findById(
        id
      );

    if (!photo) {
      return null;
    }

    const file =
      await fileClient.getFile(
        photo.fileId,
        token
      );

    return {
      photo,
      content:
        file.content,
      mimeType:
        file.fileType ||
        photo.mimeType,
      fileName:
        file.fileName ||
        photo.name
    };
  }

  // =====================================
  // LIST
  // =====================================

  async listPhotos() {

    return photoRepository.findAll();
  }

  // =====================================
  // SEARCH
  // =====================================

  async searchPhotos(
    filters
  ) {

    return photoRepository.search(
      filters
    );
  }

  // =====================================
  // DELETE DISTRIBUIDO
  // =====================================

  async deletePhoto(
    id,
    token
  ) {

    if (!token) {
      throw new Error(
        "token es obligatorio"
      );
    }

    // =====================================
    // 1. LEER METADATA DE LA FOTO
    // =====================================

    const photo =
      await photoRepository.findById(
        id
      );

    if (!photo) {
      return false;
    }

    // =====================================
    // 2. FILE SERVICE ES LA AUTORIDAD
    //
    // DeleteFile elimina:
    // - archivo físico
    // - metadata archivo
    // - uso de cuota
    //
    // PostgreSQL elimina imagen,
    // imagen_tag y album_imagen mediante
    // ON DELETE CASCADE.
    // =====================================

    await fileClient.deleteFile(
      photo.fileId,
      token
    );

    // =====================================
    // 3. VERIFICACIÓN POST-DELETE
    // =====================================

    const remainingPhoto =
      await photoRepository.findById(
        id
      );

    if (remainingPhoto) {
      throw new Error(
        "La metadata de imagen permaneció después de DeleteFile"
      );
    }

    console.log(
      `[Photo] DELETE completado | Photo: ${id} | File: ${photo.fileId}`
    );

    return true;
  }
}

module.exports =
  new PhotoService();