const axios = require("axios");

const config = require("../config/config");

async function authMiddleware(req, res, next) {
  try {
    const authorizationHeader =
      req.headers.authorization;

    if (!authorizationHeader) {
      return res.status(401).json({
        success: false,
        message: "Token de autenticación requerido"
      });
    }

    const parts =
      authorizationHeader.split(" ");

    if (
      parts.length !== 2 ||
      parts[0] !== "Bearer"
    ) {
      return res.status(401).json({
        success: false,
        message:
          "Formato de autorización inválido. Use: Bearer <token>"
      });
    }

    const token =
      parts[1];

    if (!token) {
      return res.status(401).json({
        success: false,
        message: "Token vacío"
      });
    }

    const response =
      await axios.get(
        `${config.authServiceUrl}/internal/auth/validate`,
        {
          headers: {
            Authorization:
              `Bearer ${token}`
          },
          timeout: 3000
        }
      );

    const rawResponse =
      String(response.data);

    const partsResponse =
      rawResponse.split("|");

    if (partsResponse.length < 4) {
      return res.status(502).json({
        success: false,
        message:
          "Respuesta inválida del Auth Service"
      });
    }

    const valid =
      partsResponse[0] === "true";

    const userId =
      partsResponse[1];

    const role =
      partsResponse[2];

    const message =
      partsResponse
        .slice(3)
        .join("|");

    if (!valid) {
      return res.status(401).json({
        success: false,
        message:
          message || "Token inválido"
      });
    }

    req.user = {
      userId,
      role,
      token
    };

    next();
  } catch (error) {
    if (error.code === "ECONNREFUSED") {
      return res.status(503).json({
        success: false,
        message:
          "Auth Service no disponible"
      });
    }

    if (error.code === "ECONNABORTED") {
      return res.status(504).json({
        success: false,
        message:
          "Tiempo de espera agotado consultando Auth Service"
      });
    }

    return res.status(500).json({
      success: false,
      message:
        "Error validando autenticación",
      detail:
        error.message
    });
  }
}

module.exports =
  authMiddleware;
