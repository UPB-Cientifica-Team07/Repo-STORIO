const grpc =
  require("@grpc/grpc-js");

const protoLoader =
  require("@grpc/proto-loader");

const path =
  require("path");

const PROTO_PATH =
  path.resolve(
    __dirname,
    "../../../services/grid-service/proto/grid.proto"
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
  ).grid;

const client =
  new proto.GridService(
    process.env.GRID_GRPC_ADDRESS ||
      "127.0.0.1:50056",

    grpc.credentials.createInsecure()
  );

function metadataFromToken(
  token
) {

  if (
    !token ||
    typeof token !== "string" ||
    !token.trim()
  ) {

    throw new Error(
      "Token de usuario requerido"
    );
  }

  const metadata =
    new grpc.Metadata();

  metadata.set(
    "authorization",
    `Bearer ${token.trim()}`
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
            180000
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

async function submitScientificJob(
  token,
  {
    sourceCode,
    dataset,
    processes
  }
) {

  if (
    !Buffer.isBuffer(
      sourceCode
    ) ||
    sourceCode.length === 0
  ) {

    throw new Error(
      "sourceCode es obligatorio"
    );
  }

  if (
    !Buffer.isBuffer(
      dataset
    ) ||
    dataset.length === 0
  ) {

    throw new Error(
      "dataset es obligatorio"
    );
  }

  return call(
    "submitScientificJob",
    {
      sourceCode,
      dataset,
      processes
    },
    token
  );
}

module.exports = {
  submitScientificJob
};
