module.exports = {
  host:
    process.env.PHOTO_BIND_ADDRESS ||
    "127.0.0.1",

  port: process.env.PORT || 50052,

  authServiceUrl:
    process.env.AUTH_SERVICE_URL ||
    "http://localhost:8081",

  monitoringServiceAddress:
    process.env.MONITORING_SERVICE_ADDRESS ||
    "localhost:50051",

  fileServiceAddress:
    process.env.FILE_SERVICE_ADDRESS ||
    "localhost:50053"
};
