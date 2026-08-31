const grpc =
  require("@grpc/grpc-js");

const protoLoader =
  require("@grpc/proto-loader");

const path =
  require("path");

const pool =
  require("../config/database");

const config =
  require("../config/config");

const metricsState =
  require("./metricsState");

// =====================================
// CONFIGURACIÓN
// =====================================

const COMPONENT_ID =
  "photo-service-01";

const COMPONENT_NAME =
  "Photo Service";

const PROTO_PATH =
  path.resolve(
    __dirname,
    "../../../monitoring-service/proto/monitoring.proto"
  );

// =====================================
// CARGAR PROTO
// =====================================

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

const monitoringProto =
  grpc.loadPackageDefinition(
    packageDefinition
  ).monitoring;

// =====================================
// CLIENTE
// =====================================

const client =
  new monitoringProto.MonitoringService(
    config.monitoringServiceAddress,
    grpc.credentials.createInsecure()
  );

// =====================================
// CPU
// =====================================

let previousCpuUsage =
  process.cpuUsage();

let previousCpuTime =
  process.hrtime.bigint();

let cpuInitialized = false;

function getCpuUsage() {
  const currentCpu =
    process.cpuUsage();

  const currentTime =
    process.hrtime.bigint();

  if (!cpuInitialized){

    previousCpuUsage = currentCpu;
    previousCpuTime = currentTime;
    cpuInitialized = true;

    return 0;

  }




  const userDifference =
    currentCpu.user -
    previousCpuUsage.user;

  const systemDifference =
    currentCpu.system -
    previousCpuUsage.system;

  const elapsedMicroseconds =
    Number(
      currentTime -
      previousCpuTime
    ) / 1000;

  previousCpuUsage =
    currentCpu;

  previousCpuTime =
    currentTime;

  if (elapsedMicroseconds <= 0) {
    return 0;
  }

  const cpuPercentage =
    (
      (
        userDifference +
        systemDifference
      ) /
      elapsedMicroseconds
    ) * 100;

  return Number(
    cpuPercentage.toFixed(2)
  );
}

// =====================================
// MEMORIA
// =====================================

function getMemoryUsage() {
  const memory =
    process.memoryUsage();

  const memoryMb =
    memory.rss /
    1024 /
    1024;

  return Number(
    memoryMb.toFixed(2)
  );
}

// =====================================
// STORAGE LÓGICO DE IMÁGENES
// =====================================

async function getStorageUsage() {
  const result =
    await pool.query(
      `
      SELECT
        COALESCE(
          SUM(a.tamano),
          0
        ) AS total
      FROM archivo a
      JOIN imagen i
        ON i.id_archivo = a.id_archivo
      `
    );

  return Number(
    result.rows[0].total || 0
  );
}

// =====================================
// REPORTAR ESTADO
// =====================================

function reportStatus(
  status,
  message
) {
  return new Promise(
    (resolve, reject) => {
      client.ReportStatus(
        {
          componentId:
            COMPONENT_ID,

          componentName:
            COMPONENT_NAME,

          status,

          message,

          timestamp:
            Math.floor(
              Date.now() / 1000
            )
        },

        (error, response) => {
          if (error) {
            return reject(error);
          }

          if (!response.success) {
            return reject(
              new Error(
                response.message ||
                "Monitoring rechazó el estado"
              )
            );
          }

          resolve(response);
        }
      );
    }
  );
}

// =====================================
// REPORTAR MÉTRICAS
// =====================================

async function reportMetrics() {
  const storageUsage =
    await getStorageUsage();

  const payload = {
    componentId:
      COMPONENT_ID,

    componentName:
      COMPONENT_NAME,

    cpuUsage:
      getCpuUsage(),

    memoryUsage:
      getMemoryUsage(),

    activeConnections:
      metricsState
        .getActiveConnections(),

    totalRequests:
      metricsState
        .getTotalRequests(),

    timestamp:
      Math.floor(
        Date.now() / 1000
      ),

    storageUsage
  };

  return new Promise(
    (resolve, reject) => {
      client.ReportMetrics(
        payload,

        (error, response) => {
          if (error) {
            return reject(error);
          }

          if (!response.success) {
            return reject(
              new Error(
                response.message ||
                "Monitoring rechazó las métricas"
              )
            );
          }

          resolve({
            response,
            metrics: payload
          });
        }
      );
    }
  );
}

// =====================================
// HEARTBEAT
// =====================================

let heartbeatTimer = null;

function startMonitoring() {
  if (heartbeatTimer) {
    return;
  }

  const executeHeartbeat =
    async () => {
      try {
        await reportStatus(
          "ACTIVE",
          "Photo Service operativo"
        );

        const {
          metrics
        } =
          await reportMetrics();

        console.log(
          `[Monitoring] ACTIVE | ` +
          `CPU: ${metrics.cpuUsage}% | ` +
          `RAM: ${metrics.memoryUsage} MB | ` +
          `Storage: ${metrics.storageUsage} bytes | ` +
          `Connections: ${metrics.activeConnections} | ` +
          `Requests: ${metrics.totalRequests}`
        );
      } catch (error) {
        console.warn(
          `[Monitoring] No disponible: ${error.message}`
        );
      }
    };

  executeHeartbeat();

  heartbeatTimer =
    setInterval(
      executeHeartbeat,
      5000
    );
}

async function stopMonitoring() {
  if (heartbeatTimer) {
    clearInterval(
      heartbeatTimer
    );

    heartbeatTimer = null;
  }

  try {
    await reportStatus(
      "INACTIVE",
      "Photo Service detenido"
    );
  } catch (error) {
    console.warn(
      `[Monitoring] No se pudo reportar cierre: ${error.message}`
    );
  }

  client.close();
}

module.exports = {
  startMonitoring,
  stopMonitoring,
  reportStatus,
  reportMetrics
};
