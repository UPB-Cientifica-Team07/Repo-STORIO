const albumRepository =
  require("./albumRepository");

const photoService =
  require("./photoService");

class AlbumService {
  async createAlbum(data) {
    if (!data.ownerId) {
      throw new Error(
        "ownerId es obligatorio"
      );
    }

    if (!data.name) {
      throw new Error(
        "name es obligatorio"
      );
    }

    return albumRepository.create({
      ownerId: data.ownerId,
      name: data.name,
      description:
        data.description || ""
    });
  }

  async getAlbum(id) {
    return albumRepository.findById(
      id
    );
  }

  async listAlbums() {
    return albumRepository.findAll();
  }

  async listAlbumsByOwner(
    ownerId
  ) {
    return albumRepository.findByOwner(
      ownerId
    );
  }

  async addPhoto(
    albumId,
    photoId
  ) {
    const album =
      await albumRepository.findById(
        albumId
      );

    if (!album) {
      throw new Error(
        "Álbum no encontrado"
      );
    }

    const photo =
      await photoService.getPhoto(
        photoId
      );

    if (!photo) {
      throw new Error(
        "Foto no encontrada"
      );
    }

    return albumRepository.addPhoto(
      albumId,
      photoId
    );
  }

  async removePhoto(
    albumId,
    photoId
  ) {
    const album =
      await albumRepository.findById(
        albumId
      );

    if (!album) {
      throw new Error(
        "Álbum no encontrado"
      );
    }

    return albumRepository.removePhoto(
      albumId,
      photoId
    );
  }

  async deleteAlbum(id) {
    return albumRepository.delete(
      id
    );
  }
}

module.exports =
  new AlbumService();
