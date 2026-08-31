const albumService =
  require("../services/albumService");

const photoService =
  require("../services/photoService");

async function createAlbum(req, res) {
  try {
    const album =
      await albumService.createAlbum({
        ...req.body,
        ownerId: req.user.userId
      });

    return res.status(201).json({
      success: true,
      message:
        "Álbum creado correctamente",
      album
    });
  } catch (error) {
    return res.status(400).json({
      success: false,
      message: error.message
    });
  }
}

async function listAlbums(req, res) {
  try {
    let albums;

    if (req.user.role === "ADMIN") {
      albums =
        await albumService.listAlbums();
    } else {
      albums =
        await albumService.listAlbumsByOwner(
          req.user.userId
        );
    }

    return res.status(200).json({
      success: true,
      total: albums.length,
      albums
    });
  } catch (error) {
    return res.status(500).json({
      success: false,
      message:
        "Error consultando álbumes",
      detail:
        error.message
    });
  }
}

async function getAlbum(req, res) {
  try {
    const album =
      await albumService.getAlbum(
        req.params.id
      );

    if (!album) {
      return res.status(404).json({
        success: false,
        message: "Álbum no encontrado"
      });
    }

    if (
      album.ownerId !== req.user.userId &&
      req.user.role !== "ADMIN"
    ) {
      return res.status(403).json({
        success: false,
        message:
          "No tiene permisos para acceder a este álbum"
      });
    }

    return res.status(200).json({
      success: true,
      album
    });
  } catch (error) {
    return res.status(500).json({
      success: false,
      message:
        "Error consultando álbum",
      detail:
        error.message
    });
  }
}

async function addPhotoToAlbum(req, res) {
  try {
    const album =
      await albumService.getAlbum(
        req.params.id
      );

    if (!album) {
      return res.status(404).json({
        success: false,
        message: "Álbum no encontrado"
      });
    }

    if (
      album.ownerId !== req.user.userId &&
      req.user.role !== "ADMIN"
    ) {
      return res.status(403).json({
        success: false,
        message:
          "No tiene permisos para modificar este álbum"
      });
    }

    const photo =
      await photoService.getPhoto(
        req.params.photoId
      );

    if (!photo) {
      return res.status(404).json({
        success: false,
        message: "Foto no encontrada"
      });
    }

    if (
      photo.ownerId !== req.user.userId &&
      req.user.role !== "ADMIN"
    ) {
      return res.status(403).json({
        success: false,
        message:
          "No tiene permisos para usar esta foto"
      });
    }

    const updatedAlbum =
      await albumService.addPhoto(
        req.params.id,
        req.params.photoId
      );

    return res.status(200).json({
      success: true,
      message:
        "Foto agregada al álbum",
      album: updatedAlbum
    });
  } catch (error) {
    return res.status(400).json({
      success: false,
      message: error.message
    });
  }
}

async function removePhotoFromAlbum(req, res) {
  try {
    const album =
      await albumService.getAlbum(
        req.params.id
      );

    if (!album) {
      return res.status(404).json({
        success: false,
        message: "Álbum no encontrado"
      });
    }

    if (
      album.ownerId !== req.user.userId &&
      req.user.role !== "ADMIN"
    ) {
      return res.status(403).json({
        success: false,
        message:
          "No tiene permisos para modificar este álbum"
      });
    }

    const updatedAlbum =
      await albumService.removePhoto(
        req.params.id,
        req.params.photoId
      );

    return res.status(200).json({
      success: true,
      message:
        "Foto eliminada del álbum",
      album: updatedAlbum
    });
  } catch (error) {
    return res.status(400).json({
      success: false,
      message: error.message
    });
  }
}

module.exports = {
  createAlbum,
  listAlbums,
  getAlbum,
  addPhotoToAlbum,
  removePhotoFromAlbum
};
