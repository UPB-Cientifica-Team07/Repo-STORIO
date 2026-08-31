class Album {
  constructor({
    id,
    ownerId,
    name,
    description = "",
    photoIds = [],
    createdAt = new Date().toISOString()
  }) {
    this.id = id;
    this.ownerId = ownerId;
    this.name = name;
    this.description = description;
    this.photoIds = photoIds;
    this.createdAt = createdAt;
  }
}

module.exports = Album;
