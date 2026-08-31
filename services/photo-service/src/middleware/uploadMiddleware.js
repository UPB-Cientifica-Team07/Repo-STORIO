const multer =
  require("multer");

const storage =
  multer.memoryStorage();

const upload =
  multer({
    storage,

    limits: {
      fileSize:
        20 * 1024 * 1024
    },

    fileFilter(
      req,
      file,
      callback
    ) {
      if (
        !file.mimetype ||
        !file.mimetype.startsWith(
          "image/"
        )
      ) {
        return callback(
          new Error(
            "Solo se permiten archivos de imagen"
          )
        );
      }

      callback(
        null,
        true
      );
    }
  });

module.exports =
  upload;
