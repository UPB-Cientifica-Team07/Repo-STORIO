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

        name:
          file.originalname,

        mimeType:
          file.mimetype,

        size:
          file.size,

        path:
          `files/${fileId}.bin`,

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
    // 1. LEER METADATA POSTGRESQL
    // =====================================

    const photo =
      await photoRepository.findById(
        id
      );

    if (!photo) {
      return false;
    }

    // =====================================
    // 2. LEER ARCHIVO DESDE FILE SERVICE
    // =====================================

    const physicalFile =
      await fileClient.getFile(
        photo.fileId,
        token
      );

    // =====================================
    // 3. BORRAR ARCHIVO FÍSICO
    // =====================================

    await fileClient.deleteFile(
      photo.fileId,
      token
    );

    try {

      // =====================================
      // 4. BORRAR POSTGRESQL
      // =====================================

      const deletedFileId =
        await photoRepository.delete(
          id
        );

      if (!deletedFileId) {

        throw new Error(
          "No se pudo eliminar la metadata de PostgreSQL"
        );
      }

      if (
        deletedFileId !==
        photo.fileId
      ) {

        throw new Error(
          "El fileId eliminado en PostgreSQL no coincide con File Service"
        );
      }

      console.log(
        `[Photo] DELETE completado | Photo: ${id} | File: ${photo.fileId}`
      );

      return true;

    } catch (error) {

      // =====================================
      // 5. COMPENSACIÓN
      // =====================================

      try {

        const restored =
          await fileClient.restoreFile({
            token,

            fileId:
              photo.fileId,

            userId:
              physicalFile.userId,

            fileName:
              physicalFile.fileName,

            fileType:
              physicalFile.fileType,

            content:
              physicalFile.content
          });

        console.warn(
          `[Photo] DELETE compensado | Archivo restaurado: ${restored.fileId}`
        );

      } catch (
        compensationError
      ) {

        console.error(
          "[Photo] ERROR CRÍTICO: falló PostgreSQL y también RestoreFile:",
          compensationError.message
        );
      }

      throw error;
    }
  }
}

module.exports =
  new PhotoService();