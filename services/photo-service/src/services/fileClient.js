const grpc =
  require("@grpc/grpc-js");

const protoLoader =
  require("@grpc/proto-loader");

const path =
  require("path");

const config =
  require("../config/config");

const PROTO_PATH =
  path.resolve(
    __dirname,
    "../../../file-service/proto/file.proto"
  );

const packageDefinition =
  protoLoader.loadSync(
    PROTO_PATH,
    {
      keepCase: false,
      longs: String,
      enums: String,
      defaults: true,
      oneofs: true
    }
  );

const fileProto =
  grpc.loadPackageDefinition(
    packageDefinition
  ).file;

const client =
  new fileProto.FileService(
    config.fileServiceAddress,
    grpc.credentials.createInsecure()
  );

// =====================================
// AUTH METADATA
// =====================================

function createAuthMetadata(
  token
) {

  if (!token) {
    throw new Error(
      "Token requerido para File Service"
    );
  }

  const metadata =
    new grpc.Metadata();

  metadata.set(
    "authorization",
    `Bearer ${token}`
  );

  return metadata;
}

// =====================================
// UPLOAD
// =====================================

function uploadFile({
  token,
  userId,
  fileName,
  fileType,
  content
}) {

  return new Promise(
    (resolve, reject) => {

      let metadata;

      try {

        metadata =
          createAuthMetadata(
            token
          );

      } catch (error) {

        return reject(
          error
        );
      }

      client.UploadFile(
        {
          // Se conserva por compatibilidad
          // con el proto, pero File Service
          // ya NO confía en este valor.
          userId,

          fileName,
          fileType,
          content
        },

        metadata,

        (error, response) => {

          if (error) {
            return reject(
              error
            );
          }

          if (!response.success) {

            return reject(
              new Error(
                response.message ||
                "File Service rechazó el archivo"
              )
            );
          }

          if (!response.fileId) {

            return reject(
              new Error(
                "File Service no devolvió fileId"
              )
            );
          }

          resolve({
            fileId:
              response.fileId,

            message:
              response.message
          });
        }
      );
    }
  );
}

// =====================================
// RESTORE
// =====================================

function restoreFile({
  token,
  fileId,
  userId,
  fileName,
  fileType,
  content
}) {

  return new Promise(
    (resolve, reject) => {

      let metadata;

      try {

        metadata =
          createAuthMetadata(
            token
          );

      } catch (error) {

        return reject(
          error
        );
      }

      client.RestoreFile(
        {
          fileId,
          userId,
          fileName,
          fileType,
          content
        },

        metadata,

        (error, response) => {

          if (error) {

            return reject(
              error
            );
          }

          if (!response.success) {

            return reject(
              new Error(
                response.message ||
                "File Service no pudo restaurar el archivo"
              )
            );
          }

          if (
            response.fileId !==
            fileId
          ) {

            return reject(
              new Error(
                "File Service restauró un ID diferente al solicitado"
              )
            );
          }

          resolve({
            fileId:
              response.fileId,

            message:
              response.message
          });
        }
      );
    }
  );
}

// =====================================
// GET
// =====================================

function getFile(
  fileId,
  token
) {

  return new Promise(
    (resolve, reject) => {

      let metadata;

      try {

        metadata =
          createAuthMetadata(
            token
          );

      } catch (error) {

        return reject(
          error
        );
      }

      client.GetFile(
        {
          fileId
        },

        metadata,

        (error, response) => {

          if (error) {

            return reject(
              error
            );
          }

          if (
            !response.success ||
            !response.file
          ) {

            return reject(
              new Error(
                response.message ||
                "Archivo no encontrado"
              )
            );
          }

          resolve(
            response.file
          );
        }
      );
    }
  );
}

// =====================================
// DELETE
// =====================================

function deleteFile(
  fileId,
  token
) {

  return new Promise(
    (resolve, reject) => {

      let metadata;

      try {

        metadata =
          createAuthMetadata(
            token
          );

      } catch (error) {

        return reject(
          error
        );
      }

      client.DeleteFile(
        {
          fileId
        },

        metadata,

        (error, response) => {

          if (error) {

            return reject(
              error
            );
          }

          if (!response.success) {

            return reject(
              new Error(
                response.message ||
                "No se pudo eliminar archivo"
              )
            );
          }

          resolve(
            response
          );
        }
      );
    }
  );
}

module.exports = {
  uploadFile,
  restoreFile,
  getFile,
  deleteFile
};