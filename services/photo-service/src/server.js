const express =
  require("express");

const config =
  require("./config/config");

const photoRoutes =
  require("./routes/photoRoutes");

const albumRoutes =
  require("./routes/albumRoutes");

const metricsMiddleware =
  require("./middleware/metricsMiddleware");

const metricsState =
  require("./services/metricsState");

const {
  startMonitoring,
  stopMonitoring
} =
  require("./services/monitoringClient");

const app =
  express();

// =====================================
// MIDDLEWARE
// =====================================

app.use(
  express.json()
);

app.use(
  metricsMiddleware
);

// =====================================
// HEALTH
// =====================================

app.get(
  "/health",
  (req, res) => {
    res.status(200).json({
      service:
        "Photo Service",

      status:
        "ACTIVE",

      technology:
        "JavaScript + Node.js + Express",

      port:
        Number(config.port),

      authService:
        config.authServiceUrl,

      monitoringService:
        config.monitoringServiceAddress,

      timestamp:
        new Date().toISOString()
    });
  }
);

// =====================================
// ROUTES
// =====================================

app.use(
  "/photos",
  photoRoutes
);

app.use(
  "/albums",
  albumRoutes
);

// =====================================
// 404
// =====================================

app.use(
  (req, res) => {
    res.status(404).json({
      success: false,
      message:
        "Ruta no encontrada"
    });
  }
);

// =====================================
// START SERVER
// =====================================

const server =
  app.listen(
    config.port,
    () => {
      console.log(
        "==================================="
      );

      console.log(
        " PHOTO SERVICE"
      );

      console.log(
        " Tecnología: JavaScript + Node.js"
      );

      console.log(
        ` Puerto: ${config.port}`
      );

      console.log(
        ` Auth Service: ${config.authServiceUrl}`
      );

      console.log(
        ` Monitoring: ${config.monitoringServiceAddress}`
      );

      console.log(
        " Estado: ACTIVO"
      );

      console.log(
        "==================================="
      );

      startMonitoring();
    }
  );

// =====================================
// CONEXIONES TCP
// =====================================

server.on(
  "connection",
  (socket) => {
    metricsState.connectionOpened();

    socket.on(
      "close",
      () => {
        metricsState.connectionClosed();
      }
    );
  }
);

// =====================================
// SHUTDOWN CONTROLADO
// =====================================

let shuttingDown =
  false;

async function shutdown(signal) {
  if (shuttingDown) {
    return;
  }

  shuttingDown = true;

  console.log(
    `\nCerrando Photo Service (${signal})...`
  );

  await stopMonitoring();

  server.close(
    () => {
      console.log(
        "Photo Service detenido."
      );

      process.exit(0);
    }
  );

  setTimeout(
    () => {
      process.exit(1);
    },
    5000
  ).unref();
}

process.on(
  "SIGINT",
  () => shutdown("SIGINT")
);

process.on(
  "SIGTERM",
  () => shutdown("SIGTERM")
);
