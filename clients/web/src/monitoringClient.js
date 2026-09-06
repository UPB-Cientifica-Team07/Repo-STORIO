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

function call(
  method,
  request = {}
) {

  return new Promise(
    (resolve, reject) => {

      client[method](
        request,
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

async function getMetrics() {

  return call(
    "getMetrics"
  );
}

async function getServiceStatus(
  componentName
) {

  return call(
    "getServiceStatus",
    {
      componentName
    }
  );
}

async function getAlerts() {

  return call(
    "getAlerts"
  );
}

async function getAlertRules() {

  return call(
    "getAlertRules"
  );
}


async function createAlertRule(
  rule
) {

  return call(
    "createAlertRule",
    rule
  );
}

async function updateAlertRule(
  rule
) {

  return call(
    "updateAlertRule",
    rule
  );
}

async function deleteAlertRule(
  id
) {

  return call(
    "deleteAlertRule",
    {
      id
    }
  );
}

async function getHpcSummary() {

  return call(
    "getHpcSummary"
  );
}

async function getNodeStatus(
  nodeId
) {

  return call(
    "getNodeStatus",
    {
      nodeId
    }
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
