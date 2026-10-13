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

const {
  sharePhoto,
  revokePhotoShare,
  listPhotoShares,
  listSharedWithMe
} =
  require("../controllers/photoShareController");

const {
  createExternalShare,
  listExternalShares,
  revokeExternalShare,
  getExternalPhoto,
  getExternalPhotoContent
} =
  require("../controllers/photoExternalShareController");

const router =
  express.Router();

// Public external-share routes deliberately
// execute before authentication middleware.

router.get(
  "/public/:token/content",
  getExternalPhotoContent
);

router.get(
  "/public/:token",
  getExternalPhoto
);

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
  "/shared-with-me",
  listSharedWithMe
);

router.get(
  "/:id/shares",
  listPhotoShares
);

router.post(
  "/:id/share",
  sharePhoto
);

router.delete(
  "/:id/share/:targetUserId",
  revokePhotoShare
);

router.post(
  "/:id/external-shares",
  createExternalShare
);

router.get(
  "/:id/external-shares",
  listExternalShares
);

router.delete(
  "/:id/external-shares/:shareId",
  revokeExternalShare
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
