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
