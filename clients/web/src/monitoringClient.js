const grpc =
  require("@grpc/grpc-js");

const protoLoader =
  require("@grpc/proto-loader");

const path =
  require("path");

const PROTO_PATH =
  path.resolve(
    __dirname,
    "../../../services/monitoring-service/proto/monitoring.proto"
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
  ).monitoring;

const client =
  new proto.MonitoringService(
    process.env.MONITORING_GRPC_ADDRESS ||
      "127.0.0.1:50051",
    grpc.credentials.createInsecure()
  );

function userMetadata(
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
  request = {},
  token
) {

  return new Promise(
    (resolve, reject) => {

      client[method](
        request,
        userMetadata(
          token
        ),
        {
          deadline:
            Date.now() +
            10000
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

async function getMetrics(
  token
) {

  return call(
    "getMetrics",
    {},
    token
  );
}

async function getServiceStatus(
  componentName,
  token
) {

  return call(
    "getServiceStatus",
    {
      componentName
    },
    token
  );
}

async function getAlerts(
  token
) {

  return call(
    "getAlerts",
    {},
    token
  );
}

async function getAlertRules(
  token
) {

  return call(
    "getAlertRules",
    {},
    token
  );
}


async function createAlertRule(
  rule,
  token
) {

  return call(
    "createAlertRule",
    rule,
    token
  );
}

async function updateAlertRule(
  rule,
  token
) {

  return call(
    "updateAlertRule",
    rule,
    token
  );
}

async function deleteAlertRule(
  id,
  token
) {

  return call(
    "deleteAlertRule",
    {
      id
    },
    token
  );
}

async function getHpcSummary(
  token
) {

  return call(
    "getHpcSummary",
    {},
    token
  );
}

async function getNodeStatus(
  nodeId,
  token
) {

  return call(
    "getNodeStatus",
    {
      nodeId
    },
    token
  );
}

module.exports = {
  getMetrics,
  getServiceStatus,
  getAlerts,
  getAlertRules,
  createAlertRule,
  updateAlertRule,
  deleteAlertRule,
  getHpcSummary,
  getNodeStatus
};
