const express =
  require("express");

const authMiddleware =
  require("../middleware/authMiddleware");

const upload =
  require("../middleware/uploadMiddleware");

const {
  createPhoto,
  listPhotos,
  getPhoto,
  getPhotoContent,
  deletePhoto
} =
  require("../controllers/photoController");

const router =
  express.Router();

router.use(
  authMiddleware
);

router.post(
  "/",
  upload.single(
    "file"
  ),
  createPhoto
);

router.get(
  "/",
  listPhotos
);

// IMPORTANTE:
// /:id/content debe ir antes de /:id

router.get(
  "/:id/content",
  getPhotoContent
);

router.get(
  "/:id",
  getPhoto
);

router.delete(
  "/:id",
  deletePhoto
);

module.exports =
  router;
