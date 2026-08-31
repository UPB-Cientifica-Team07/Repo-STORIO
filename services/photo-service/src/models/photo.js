class Photo {
  constructor({
    id,
    fileId,
    ownerId,
    name,
    mimeType,
    size,
    path,
    tags = [],
    albumId = null,
    createdAt = new Date().toISOString()
  }) {
    this.id = id;
    this.fileId = fileId;
    this.ownerId = ownerId;
    this.name = name;
    this.mimeType = mimeType;
    this.size = size;
    this.path = path;
    this.tags = tags;
    this.albumId = albumId;
    this.createdAt = createdAt;
  }
}

module.exports = Photo;
