const metricsState =
  require("../services/metricsState");

function metricsMiddleware(req, res, next) {
  metricsState.requestReceived();
  next();
}

module.exports = metricsMiddleware;
