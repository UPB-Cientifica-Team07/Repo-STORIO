const express = require("express");

const authMiddleware =
  require("../middleware/authMiddleware");

const {
  createAlbum,
  listAlbums,
  getAlbum,
  addPhotoToAlbum,
  removePhotoFromAlbum
} = require("../controllers/albumController");

const router =
  express.Router();

router.use(authMiddleware);

router.post(
  "/",
  createAlbum
);

router.get(
  "/",
  listAlbums
);

router.get(
  "/:id",
  getAlbum
);

router.post(
  "/:id/photos/:photoId",
  addPhotoToAlbum
);

router.delete(
  "/:id/photos/:photoId",
  removePhotoFromAlbum
);

module.exports = router;
