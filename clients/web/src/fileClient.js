const grpc =
  require("@grpc/grpc-js");

const protoLoader =
  require("@grpc/proto-loader");

const path =
  require("path");

const PROTO_PATH =
  path.resolve(
    __dirname,
    "../../../services/file-service/proto/file.proto"
  );

const packageDefinition =
  protoLoader.loadSync(
    PROTO_PATH,
    {
      keepCase:
        false,

      longs:
        String,

      enums:
        String,

      defaults:
        true,

      oneofs:
        true
    }
  );

const proto =
  grpc.loadPackageDefinition(
    packageDefinition
  ).file;

const client =
  new proto.FileService(
    "127.0.0.1:50053",
    grpc.credentials.createInsecure()
  );

function metadataFromToken(
  token
) {

  const metadata =
    new grpc.Metadata();

  metadata.set(
    "authorization",
    `Bearer ${token}`
  );

  return metadata;
}

function call(
  method,
  request,
  token
) {

  return new Promise(
    (resolve, reject) => {

      client[method](
        request,
        metadataFromToken(
          token
        ),
        {
          deadline:
            Date.now() +
            15000
        },
        (error, response) => {

          if (error) {

            reject(
              error
            );

            return;
          }

          resolve(
            response
          );
        }
      );
    }
  );
}

async function getHome(
  token,
  userId
) {

  return call(
    "getHome",
    {
      userId
    },
    token
  );
}

async function listFiles(
  token,
  userId
) {

  return call(
    "listFiles",
    {
      userId
    },
    token
  );
}

async function getFile(
  token,
  fileId
) {

  return call(
    "getFile",
    {
      fileId
    },
    token
  );
}

async function uploadFile(
  token,
  {
    userId,
    fileName,
    fileType,
    content,
    relativeDirectory = ""
  }
) {

  return call(
    "uploadFile",
    {
      userId,
      fileName,
      fileType,
      content,
      relativeDirectory
    },
    token
  );
}

async function updateFile(
  token,
  fileId,
  content
) {

  return call(
    "updateFile",
    {
      fileId,
      content
    },
    token
  );
}

async function deleteFile(
  token,
  fileId
) {

  return call(
    "deleteFile",
    {
      fileId
    },
    token
  );
}

async function shareFile(
  token,
  {
    fileId,
    targetUserId,
    canRead,
    canWrite,
    canShare
  }
) {

  return call(
    "shareFile",
    {
      fileId,
      targetUserId,
      canRead,
      canWrite,
      canShare
    },
    token
  );
}

async function revokeFileAccess(
  token,
  {
    fileId,
    targetUserId
  }
) {

  return call(
    "revokeFileAccess",
    {
      fileId,
      targetUserId
    },
    token
  );
}

module.exports = {
  getHome,
  listFiles,
  getFile,
  uploadFile,
  updateFile,
  deleteFile,
  shareFile,
  revokeFileAccess
};
