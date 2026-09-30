const express = require("express");
const path = require("path");
const fs = require("fs");
const https = require("https");
const axios = require("axios");
const multer = require("multer");

const fileClient =
  require("./src/fileClient");

const monitoringClient =
  require(
    "./src/monitoringClient"
  );

const gridClient =
  require("./src/gridClient");

const app = express();

// =====================================
// HTTP SECURITY
// =====================================

app.disable(
  "x-powered-by"
);

app.use(
  (req, res, next) => {

    res.setHeader(
      "X-Content-Type-Options",
      "nosniff"
    );

    res.setHeader(
      "X-Frame-Options",
      "DENY"
    );

    res.setHeader(
      "Referrer-Policy",
      "no-referrer"
    );

    res.setHeader(
      "Permissions-Policy",
      "camera=(), microphone=(), geolocation=()"
    );

    res.setHeader(
      "Content-Security-Policy",
      "default-src 'self'; " +
        "script-src 'self' 'unsafe-inline'; " +
        "style-src 'self' 'unsafe-inline'; " +
        "img-src 'self' data: blob:; " +
        "connect-src 'self'; " +
        "font-src 'self'; " +
        "object-src 'none'; " +
        "base-uri 'self'; " +
        "frame-ancestors 'none'"
    );

    next();
  }
);

const PORT =
  process.env.PORT ||
  8080;

const AUTH_SERVICE =
  process.env.AUTH_SERVICE ||
  "https://127.0.0.1:8081";

const AUTH_TLS_CA_FILE =
  process.env.AUTH_TLS_CA_FILE ||
  path.resolve(
    __dirname,
    "../../security/pki/upb_dev_ca.crt"
  );

const AUTH_HTTPS_AGENT =
  new https.Agent({
    ca:
      fs.readFileSync(
        AUTH_TLS_CA_FILE
      ),

    minVersion:
      "TLSv1.2",

    rejectUnauthorized:
      true
  });

const PHOTO_SERVICE =
  process.env.PHOTO_SERVICE ||
  "http://127.0.0.1:50052";

const STREAMING_SERVICE =
  process.env.STREAMING_SERVICE ||
  "http://127.0.0.1:8082";

const upload =
  multer({
    storage:
      multer.memoryStorage(),

    limits: {
      fileSize:
        20 * 1024 * 1024
    }
  });

const scientificUpload =
  multer({
    storage:
      multer.memoryStorage(),

    limits: {
      fileSize:
        2 * 1024 * 1024,

      files:
        2,

      fields:
        4
    }
  });

const scientificUploadFields =
  scientificUpload.fields(
    [
      {
        name:
          "source",

        maxCount:
          1
      },

      {
        name:
          "dataset",

        maxCount:
          1
      }
    ]
  );

app.use(
  express.json()
);

// =====================================
// API - AUTH
// =====================================

app.post(
  "/api/auth/login",
  async (req, res) => {

    try {

      const {
        username,
        password
      } = req.body;

      if (
        !username ||
        !password
      ) {

        return res
          .status(400)
          .json({
            success: false,
            message:
              "Usuario y contraseña son obligatorios"
          });
      }

      const loginBody =
        new URLSearchParams({
          username,
          password
        }).toString();

      const response =
        await axios.post(
          `${AUTH_SERVICE}/internal/auth/login`,
          loginBody,
          {
            headers: {
              "Content-Type":
                "application/x-www-form-urlencoded"
            },

            timeout: 5000,

            httpsAgent:
              AUTH_HTTPS_AGENT
          }
        );

      const raw =
        String(
          response.data
        );

      const parts =
        raw.split("|");

      if (
        parts.length < 5 ||
        parts[0] !== "true"
      ) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              parts[4] ||
              "Credenciales inválidas"
          });
      }

      res.cookie(
        "upb_token",
        parts[3],
        {
          httpOnly: true,
          sameSite: "strict",
          secure:
            process.env.NODE_ENV ===
            "production",

          path: "/"
        }
      );

      return res.json({
        success: true,

        user: {
          userId:
            parts[1],

          role:
            parts[2]
        },

        token:
          parts[3],

        message:
          parts[4]
      });

    } catch (error) {

      console.error(
        "AUTH PROXY ERROR:",
        error.message
      );

      return res
        .status(502)
        .json({
          success: false,
          message:
            "No fue posible comunicarse con Auth Service"
        });
    }
  }
);

// =====================================
// LOGOUT
// =====================================

app.post(
  "/api/auth/logout",
  (req, res) => {

    res.clearCookie(
      "upb_token",
      {
        httpOnly: true,
        sameSite: "strict",
        secure:
          process.env.NODE_ENV ===
          "production",

        path: "/"
      }
    );

    return res.json({
      success: true,
      message:
        "Sesión cerrada"
    });
  }
);

// =====================================
// HEALTH
// =====================================

app.get(
  "/health",
  (req, res) => {

    res.json({
      service:
        "UPB-CIENTIFICA Web Client",

      status:
        "ACTIVE",

      technology:
        "JavaScript + Node.js + Express",

      port:
        Number(PORT),

      authService:
        AUTH_SERVICE,

      photoService:
        PHOTO_SERVICE,

      streamingService:
        STREAMING_SERVICE,

      timestamp:
        new Date().toISOString()
    });
  }
);

// =====================================
// PHOTO API PROXY
// =====================================

function getBearerToken(req) {

  const authorization =
    req.headers.authorization;

  if (
    !authorization ||
    !authorization.startsWith("Bearer ")
  ) {

    return null;
  }

  return authorization;
}

// -------------------------------------
// LIST PHOTOS
// -------------------------------------

app.get(
  "/api/photos",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.get(
          `${PHOTO_SERVICE}/photos`,
          {
            headers: {
              Authorization:
                authorization
            },

            params:
              req.query,

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error comunicándose con Photo Service"
          }
        );
    }
  }
);

// -------------------------------------
// GET PHOTO METADATA
// -------------------------------------

app.get(
  "/api/photos/:id",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.get(
          `${PHOTO_SERVICE}/photos/${req.params.id}`,
          {
            headers: {
              Authorization:
                authorization
            },

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error obteniendo fotografía"
          }
        );
    }
  }
);

// -------------------------------------
// GET PHOTO CONTENT
// -------------------------------------

app.get(
  "/api/photos/:id/content",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.get(
          `${PHOTO_SERVICE}/photos/${req.params.id}/content`,
          {
            headers: {
              Authorization:
                authorization
            },

            responseType:
              "arraybuffer",

            timeout:
              10000
          }
        );

      if (
        response.headers[
          "content-type"
        ]
      ) {

        res.setHeader(
          "Content-Type",
          response.headers[
            "content-type"
          ]
        );
      }

      if (
        response.headers[
          "content-disposition"
        ]
      ) {

        res.setHeader(
          "Content-Disposition",
          response.headers[
            "content-disposition"
          ]
        );
      }

      return res.send(
        Buffer.from(
          response.data
        )
      );

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json({
          success: false,
          message:
            "Error obteniendo contenido de fotografía"
        });
    }
  }
);

// -------------------------------------
// DELETE PHOTO
// -------------------------------------

app.delete(
  "/api/photos/:id",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.delete(
          `${PHOTO_SERVICE}/photos/${req.params.id}`,
          {
            headers: {
              Authorization:
                authorization
            },

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error eliminando fotografía"
          }
        );
    }
  }
);

// -------------------------------------
// UPLOAD PHOTO
// -------------------------------------

app.post(
  "/api/photos",
  upload.single("file"),
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      if (!req.file) {

        return res
          .status(400)
          .json({
            success: false,
            message:
              "Debe seleccionar una imagen"
          });
      }

      const FormData =
        (await import("form-data"))
          .default;

      const form =
        new FormData();

      form.append(
        "file",
        req.file.buffer,
        {
          filename:
            req.file.originalname,

          contentType:
            req.file.mimetype
        }
      );

      if (
        req.body.tags
      ) {

        form.append(
          "tags",
          req.body.tags
        );
      }

      const response =
        await axios.post(
          `${PHOTO_SERVICE}/photos`,
          form,
          {
            headers: {
              ...form.getHeaders(),

              Authorization:
                authorization
            },

            maxBodyLength:
              Infinity,

            timeout:
              20000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error cargando fotografía"
          }
        );
    }
  }
);

// =====================================
// PHOTO API PROXY
// =====================================

function getBearerToken(req) {

  const authorization =
    req.headers.authorization;

  if (
    !authorization ||
    !authorization.startsWith("Bearer ")
  ) {

    return null;
  }

  return authorization;
}

// -------------------------------------
// LIST PHOTOS
// -------------------------------------

app.get(
  "/api/photos",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.get(
          `${PHOTO_SERVICE}/photos`,
          {
            headers: {
              Authorization:
                authorization
            },

            params:
              req.query,

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error comunicándose con Photo Service"
          }
        );
    }
  }
);

// -------------------------------------
// GET PHOTO METADATA
// -------------------------------------

app.get(
  "/api/photos/:id",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.get(
          `${PHOTO_SERVICE}/photos/${req.params.id}`,
          {
            headers: {
              Authorization:
                authorization
            },

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error obteniendo fotografía"
          }
        );
    }
  }
);

// -------------------------------------
// GET PHOTO CONTENT
// -------------------------------------

app.get(
  "/api/photos/:id/content",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.get(
          `${PHOTO_SERVICE}/photos/${req.params.id}/content`,
          {
            headers: {
              Authorization:
                authorization
            },

            responseType:
              "arraybuffer",

            timeout:
              10000
          }
        );

      if (
        response.headers[
          "content-type"
        ]
      ) {

        res.setHeader(
          "Content-Type",
          response.headers[
            "content-type"
          ]
        );
      }

      if (
        response.headers[
          "content-disposition"
        ]
      ) {

        res.setHeader(
          "Content-Disposition",
          response.headers[
            "content-disposition"
          ]
        );
      }

      return res.send(
        Buffer.from(
          response.data
        )
      );

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json({
          success: false,
          message:
            "Error obteniendo contenido de fotografía"
        });
    }
  }
);

// -------------------------------------
// DELETE PHOTO
// -------------------------------------

app.delete(
  "/api/photos/:id",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.delete(
          `${PHOTO_SERVICE}/photos/${req.params.id}`,
          {
            headers: {
              Authorization:
                authorization
            },

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error eliminando fotografía"
          }
        );
    }
  }
);

// -------------------------------------
// UPLOAD PHOTO
// -------------------------------------

app.post(
  "/api/photos",
  upload.single("file"),
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      if (!req.file) {

        return res
          .status(400)
          .json({
            success: false,
            message:
              "Debe seleccionar una imagen"
          });
      }

      const FormData =
        (await import("form-data"))
          .default;

      const form =
        new FormData();

      form.append(
        "file",
        req.file.buffer,
        {
          filename:
            req.file.originalname,

          contentType:
            req.file.mimetype
        }
      );

      if (
        req.body.tags
      ) {

        form.append(
          "tags",
          req.body.tags
        );
      }

      const response =
        await axios.post(
          `${PHOTO_SERVICE}/photos`,
          form,
          {
            headers: {
              ...form.getHeaders(),

              Authorization:
                authorization
            },

            maxBodyLength:
              Infinity,

            timeout:
              20000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error cargando fotografía"
          }
        );
    }
  }
);

// =====================================
// ALBUM API PROXY
// =====================================

// -------------------------------------
// LIST ALBUMS
// -------------------------------------

app.get(
  "/api/albums",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.get(
          `${PHOTO_SERVICE}/albums`,
          {
            headers: {
              Authorization:
                authorization
            },

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error consultando álbumes"
          }
        );
    }
  }
);

// -------------------------------------
// CREATE ALBUM
// -------------------------------------

app.post(
  "/api/albums",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.post(
          `${PHOTO_SERVICE}/albums`,
          req.body,
          {
            headers: {
              Authorization:
                authorization,

              "Content-Type":
                "application/json"
            },

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error creando álbum"
          }
        );
    }
  }
);

// -------------------------------------
// GET ALBUM
// -------------------------------------

app.get(
  "/api/albums/:id",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.get(
          `${PHOTO_SERVICE}/albums/${req.params.id}`,
          {
            headers: {
              Authorization:
                authorization
            },

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error consultando álbum"
          }
        );
    }
  }
);

// -------------------------------------
// ADD PHOTO TO ALBUM
// -------------------------------------

app.post(
  "/api/albums/:id/photos/:photoId",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.post(
          `${PHOTO_SERVICE}/albums/${req.params.id}/photos/${req.params.photoId}`,
          {},
          {
            headers: {
              Authorization:
                authorization
            },

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error agregando fotografía al álbum"
          }
        );
    }
  }
);

// -------------------------------------
// REMOVE PHOTO FROM ALBUM
// -------------------------------------

app.delete(
  "/api/albums/:id/photos/:photoId",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.delete(
          `${PHOTO_SERVICE}/albums/${req.params.id}/photos/${req.params.photoId}`,
          {
            headers: {
              Authorization:
                authorization
            },

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error quitando fotografía del álbum"
          }
        );
    }
  }
);

// -------------------------------------
// ADD PHOTO TO ALBUM
// -------------------------------------

app.post(
  "/api/albums/:id/photos/:photoId",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.post(
          `${PHOTO_SERVICE}/albums/${req.params.id}/photos/${req.params.photoId}`,
          {},
          {
            headers: {
              Authorization:
                authorization
            },

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error agregando fotografía al álbum"
          }
        );
    }
  }
);

// -------------------------------------
// REMOVE PHOTO FROM ALBUM
// -------------------------------------

app.delete(
  "/api/albums/:id/photos/:photoId",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.delete(
          `${PHOTO_SERVICE}/albums/${req.params.id}/photos/${req.params.photoId}`,
          {
            headers: {
              Authorization:
                authorization
            },

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error quitando fotografía del álbum"
          }
        );
    }
  }
);

// -------------------------------------
// DELETE ALBUM
// -------------------------------------

app.delete(
  "/api/albums/:id",
  async (req, res) => {

    try {

      const authorization =
        getBearerToken(req);

      if (!authorization) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await axios.delete(
          `${PHOTO_SERVICE}/albums/${req.params.id}`,
          {
            headers: {
              Authorization:
                authorization
            },

            timeout:
              10000
          }
        );

      return res
        .status(response.status)
        .json(response.data);

    } catch (error) {

      const status =
        error.response?.status ||
        502;

      return res
        .status(status)
        .json(
          error.response?.data || {
            success: false,
            message:
              "Error eliminando álbum"
          }
        );
    }
  }
);

// =====================================
// STREAMING CATALOG API
// =====================================

function xmlDecode(
  value
) {

  return String(
    value || ""
  )
    .replaceAll(
      "&lt;",
      "<"
    )
    .replaceAll(
      "&gt;",
      ">"
    )
    .replaceAll(
      "&quot;",
      '"'
    )
    .replaceAll(
      "&apos;",
      "'"
    )
    .replaceAll(
      "&amp;",
      "&"
    );
}

function soapTagValue(
  xml,
  tag
) {

  const expression =
    new RegExp(
      `<(?:[A-Za-z0-9_]+:)?${tag}>([\\s\\S]*?)<\\/(?:[A-Za-z0-9_]+:)?${tag}>`
    );

  const match =
    String(xml)
      .match(
        expression
      );

  return match
    ? xmlDecode(
        match[1]
      )
    : "";
}

function parseVideoListSoap(
  xml
) {

  const videos = [];

  const expression =
    /<(?:[A-Za-z0-9_]+:)?video>([\s\S]*?)<\/(?:[A-Za-z0-9_]+:)?video>/g;

  let match;

  while (
    (
      match =
        expression.exec(
          String(xml)
        )
    ) !== null
  ) {

    const block =
      match[1];

    videos.push({
      idVideo:
        soapTagValue(
          block,
          "idVideo"
        ),

      idArchivo:
        soapTagValue(
          block,
          "idArchivo"
        ),

      nombre:
        soapTagValue(
          block,
          "nombre"
        ),

      mimeType:
        soapTagValue(
          block,
          "mimeType"
        ),

      tamano:
        Number(
          soapTagValue(
            block,
            "tamano"
          ) ||
          0
        ),

      duracionSegundos:
        Number(
          soapTagValue(
            block,
            "duracionSegundos"
          ) ||
          0
        ),

      calidad:
        soapTagValue(
          block,
          "calidad"
        ),

      formato:
        soapTagValue(
          block,
          "formato"
        )
    });
  }

  return videos;
}

app.get(
  "/api/videos",
  async (req, res) => {

    const token =
      tokenOnly(
        req
      ) ||
      cookieValue(
        req,
        "upb_token"
      );

    if (!token) {

      return res
        .status(401)
        .json({
          success: false,
          message:
            "Token requerido"
        });
    }

    const escapedToken =
      String(token)
        .replaceAll(
          "&",
          "&amp;"
        )
        .replaceAll(
          "<",
          "&lt;"
        )
        .replaceAll(
          ">",
          "&gt;"
        )
        .replaceAll(
          '"',
          "&quot;"
        )
        .replaceAll(
          "'",
          "&apos;"
        );

    const payload =
      `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope
  xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:tns="urn:UPBCientificaStreaming">
  <soap:Body>
    <tns:listVideosRequest>
      <tns:token>${escapedToken}</tns:token>
    </tns:listVideosRequest>
  </soap:Body>
</soap:Envelope>`;

    try {

      const response =
        await axios.post(
          `${STREAMING_SERVICE}/`,
          payload,
          {
            headers: {
              "Content-Type":
                "text/xml; charset=utf-8",

              SOAPAction:
                '"urn:UPBCientificaStreaming#listVideos"'
            },

            timeout:
              10000,

            validateStatus:
              () => true
          }
        );

      const xml =
        String(
          response.data ||
          ""
        );

      if (
        response.status !== 200
      ) {

        return res
          .status(
            response.status
          )
          .json({
            success: false,
            message:
              soapTagValue(
                xml,
                "faultstring"
              ) ||
              "Streaming Service rechazó la solicitud"
          });
      }

      return res.json({
        success: true,

        message:
          soapTagValue(
            xml,
            "message"
          ) ||
          "Videos encontrados",

        videos:
          parseVideoListSoap(
            xml
          )
      });

    } catch (error) {

      console.error(
        "STREAMING CATALOG ERROR:",
        error.message
      );

      return res
        .status(502)
        .json({
          success: false,
          message:
            "No fue posible consultar Streaming Service"
        });
    }
  }
);

// =====================================
// STREAMING API PROXY
// =====================================

app.get(
  "/api/stream/:idVideo",
  async (req, res) => {

    const token =
      cookieValue(
        req,
        "upb_token"
      );

    if (!token) {

      return res
        .status(401)
        .json({
          success: false,
          message:
            "Sesión requerida para reproducir video"
        });
    }

    try {

      const headers = {
        Authorization:
          `Bearer ${token}`
      };

      if (req.headers.range) {

        headers.Range =
          req.headers.range;
      }

      const response =
        await axios.get(
          `${STREAMING_SERVICE}/stream/${encodeURIComponent(
            req.params.idVideo
          )}`,
          {
            headers,

            responseType:
              "stream",

            timeout:
              15000,

            validateStatus:
              () => true
          }
        );

      const forwardedHeaders = [
        "content-type",
        "content-length",
        "content-range",
        "accept-ranges",
        "cache-control"
      ];

      for (
        const headerName
        of forwardedHeaders
      ) {

        const value =
          response.headers[
            headerName
          ];

        if (value !== undefined) {

          res.setHeader(
            headerName,
            value
          );
        }
      }

      res.status(
        response.status
      );

      response.data.on(
        "error",
        error => {

          console.error(
            "STREAM PROXY BODY ERROR:",
            error.message
          );

          if (!res.headersSent) {

            res
              .status(502)
              .end();
          } else {

            res.destroy(
              error
            );
          }
        }
      );

      return response.data.pipe(
        res
      );

    } catch (error) {

      console.error(
        "STREAM PROXY ERROR:",
        error.message
      );

      if (res.headersSent) {

        return res.end();
      }

      return res
        .status(502)
        .json({
          success: false,
          message:
            "No fue posible comunicarse con Streaming Service"
        });
    }
  }
);

// =====================================
// FILE API - HTTP -> gRPC
// =====================================

function cookieValue(
  req,
  name
) {

  const raw =
    req.headers.cookie ||
    "";

  const cookies =
    raw.split(";");

  for (
    const cookie of cookies
  ) {

    const [
      key,
      ...valueParts
    ] =
      cookie
        .trim()
        .split("=");

    if (key === name) {

      return decodeURIComponent(
        valueParts.join("=")
      );
    }
  }

  return null;
}

function tokenOnly(
  req
) {

  const authorization =
    req.headers.authorization;

  if (
    !authorization ||
    !authorization.startsWith(
      "Bearer "
    )
  ) {

    return null;
  }

  return authorization
    .slice(
      "Bearer ".length
    )
    .trim();
}

// -------------------------------------
// GET HOME / QUOTA
// -------------------------------------

app.get(
  "/api/home",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await fileClient.getHome(
          token,
          req.query.userId ||
          ""
        );

      if (
        !response.success ||
        !response.home
      ) {

        return res
          .status(404)
          .json({
            success: false,
            message:
              response.message ||
              "Home no encontrado"
          });
      }

      const quotaBytes =
        Number(
          response.home.quotaBytes ||
          0
        );

      const usedBytes =
        Number(
          response.home.usedBytes ||
          0
        );

      return res.json({
        success: true,
        message:
          response.message,

        home: {
          userId:
            response.home.userId,

          basePath:
            response.home.basePath,

          quotaBytes,

          usedBytes,

          availableBytes:
            Math.max(
              quotaBytes -
              usedBytes,
              0
            ),

          percentage:
            quotaBytes > 0
              ? (
                  usedBytes /
                  quotaBytes
                ) * 100
              : 0
        }
      });

    } catch (error) {

      console.error(
        "GET HOME ERROR:",
        error.message
      );

      return res
        .status(
          error.code === 7
            ? 403
            : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error consultando Home"
        });
    }
  }
);

// -------------------------------------
// LIST FILES
// -------------------------------------

app.get(
  "/api/files",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await fileClient.listFiles(
          token,
          req.query.userId ||
          ""
        );

      return res.json({
        success:
          response.success,

        message:
          response.message,

        files:
          (response.files || [])
            .map(
              file => ({
                id:
                  file.fileId,

                ownerId:
                  file.userId,

                name:
                  file.fileName,

                type:
                  file.fileType,

                size:
                  Number(
                    file.size ||
                    0
                  ),

                createdAt:
                  file.createdAt
              })
            )
      });

    } catch (error) {

      console.error(
        "FILE LIST ERROR:",
        error.message
      );

      return res
        .status(502)
        .json({
          success: false,
          message:
            error.details ||
            "Error consultando File Service"
        });
    }
  }
);

// -------------------------------------
// GET FILE CONTENT
// -------------------------------------

app.get(
  "/api/files/:id/content",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await fileClient.getFile(
          token,
          req.params.id
        );

      if (
        !response.success ||
        !response.file
      ) {

        return res
          .status(404)
          .json({
            success: false,
            message:
              response.message ||
              "Archivo no encontrado"
          });
      }

      const file =
        response.file;

      res.setHeader(
        "Content-Type",
        file.fileType ||
        "application/octet-stream"
      );

      res.setHeader(
        "Content-Disposition",
        `attachment; filename="${encodeURIComponent(
          file.fileName ||
          "archivo"
        )}"`
      );

      return res.send(
        Buffer.from(
          file.content ||
          []
        )
      );

    } catch (error) {

      const code =
        error.code;

      return res
        .status(
          code === 7
            ? 403
            : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error obteniendo archivo"
        });
    }
  }
);

// -------------------------------------
// UPLOAD FILE
// -------------------------------------

app.post(
  "/api/files",
  upload.single("file"),
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      if (!req.file) {

        return res
          .status(400)
          .json({
            success: false,
            message:
              "Debe seleccionar un archivo"
          });
      }

      const response =
        await fileClient.uploadFile(
          token,
          {
            userId:
              req.body.userId ||
              "",

            fileName:
              req.file.originalname,

            fileType:
              req.file.mimetype ||
              "application/octet-stream",

            content:
              req.file.buffer,

            relativeDirectory:
              req.body.relativeDirectory ||
              ""
          }
        );

      return res.json({
        success:
          response.success,

        message:
          response.message,

        fileId:
          response.fileId
      });

    } catch (error) {

      return res
        .status(502)
        .json({
          success: false,
          message:
            error.details ||
            "Error subiendo archivo"
        });
    }
  }
);

// -------------------------------------
// UPDATE FILE
// -------------------------------------

app.put(
  "/api/files/:id",
  upload.single("file"),
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      if (!req.file) {

        return res
          .status(400)
          .json({
            success: false,
            message:
              "Debe seleccionar el nuevo contenido"
          });
      }

      const response =
        await fileClient.updateFile(
          token,
          req.params.id,
          req.file.buffer
        );

      return res.json({
        success:
          response.success,

        message:
          response.message,

        file:
          response.file
            ? {
                id:
                  response.file.fileId,

                ownerId:
                  response.file.userId,

                name:
                  response.file.fileName,

                type:
                  response.file.fileType,

                size:
                  Number(
                    response.file.size ||
                    0
                  )
              }
            : null
      });

    } catch (error) {

      return res
        .status(
          error.code === 7
            ? 403
            : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error actualizando archivo"
        });
    }
  }
);

// -------------------------------------
// DELETE FILE
// -------------------------------------

app.delete(
  "/api/files/:id",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await fileClient.deleteFile(
          token,
          req.params.id
        );

      return res.json({
        success:
          response.success,

        message:
          response.message
      });

    } catch (error) {

      return res
        .status(
          error.code === 7
            ? 403
            : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error eliminando archivo"
        });
    }
  }
);

// -------------------------------------
// SHARE FILE
// -------------------------------------

app.post(
  "/api/files/:id/share",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const {
        targetUserId,
        canRead = false,
        canWrite = false,
        canShare = false
      } =
        req.body;

      if (!targetUserId) {

        return res
          .status(400)
          .json({
            success: false,
            message:
              "targetUserId es obligatorio"
          });
      }

      const response =
        await fileClient.shareFile(
          token,
          {
            fileId:
              req.params.id,

            targetUserId,

            canRead:
              Boolean(
                canRead
              ),

            canWrite:
              Boolean(
                canWrite
              ),

            canShare:
              Boolean(
                canShare
              )
          }
        );

      return res.json({
        success:
          response.success,

        message:
          response.message
      });

    } catch (error) {

      return res
        .status(
          error.code === 7
            ? 403
            : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error compartiendo archivo"
        });
    }
  }
);

// -------------------------------------
// REVOKE FILE
// -------------------------------------

app.delete(
  "/api/files/:id/share/:userId",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await fileClient
          .revokeFileAccess(
            token,
            {
              fileId:
                req.params.id,

              targetUserId:
                req.params.userId
            }
          );

      return res.json({
        success:
          response.success,

        message:
          response.message
      });

    } catch (error) {

      return res
        .status(
          error.code === 7
            ? 403
            : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error revocando acceso"
        });
    }
  }
);


// =====================================
// MONITORING API
// =====================================

// -------------------------------------
// METRICS
// -------------------------------------

app.get(
  "/api/monitoring/metrics",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await monitoringClient
          .getMetrics(
            token
          );

      return res.json({
        success: true,
        metrics:
          response.metrics || []
      });

    } catch (error) {

      console.error(
        "MONITORING METRICS ERROR:",
        error
      );

      return res
        .status(
          error.code === 7
            ? 403
            : error.code === 16
              ? 401
              : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error consultando métricas"
        });
    }
  }
);


// -------------------------------------
// HISTORICAL METRICS HELPERS
// -------------------------------------

function filterHistoricalMetrics(
  metrics,
  query
) {

  const component =
    String(
      query.component || ""
    )
      .trim()
      .toLowerCase();

  const from =
    Number(
      query.from || 0
    );

  const to =
    Number(
      query.to || 0
    );

  const requestedLimit =
    Number(
      query.limit || 500
    );

  const limit =
    Number.isFinite(
      requestedLimit
    )
      ? Math.min(
          Math.max(
            Math.trunc(
              requestedLimit
            ),
            1
          ),
          5000
        )
      : 500;

  return (
    Array.isArray(metrics)
      ? metrics
      : []
  )
    .filter(
      metric => {

        const timestamp =
          Number(
            metric.timestamp || 0
          );

        if (
          component &&
          String(
            metric.componentName ||
            ""
          )
            .trim()
            .toLowerCase() !==
            component
        ) {

          return false;
        }

        if (
          from > 0 &&
          timestamp < from
        ) {

          return false;
        }

        if (
          to > 0 &&
          timestamp > to
        ) {

          return false;
        }

        return true;
      }
    )
    .sort(
      (a, b) =>
        Number(
          b.timestamp || 0
        ) -
        Number(
          a.timestamp || 0
        )
    )
    .slice(
      0,
      limit
    );
}

function csvEscape(
  value
) {

  const stringValue =
    String(
      value ?? ""
    );

  if (
    /[",\n\r]/.test(
      stringValue
    )
  ) {

    return (
      '"' +
      stringValue
        .replace(
          /"/g,
          '""'
        ) +
      '"'
    );
  }

  return stringValue;
}

// -------------------------------------
// HISTORICAL METRICS
// -------------------------------------

app.get(
  "/api/monitoring/history",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await monitoringClient
          .getMetrics(
            token
          );

      const metrics =
        filterHistoricalMetrics(
          response.metrics || [],
          req.query
        );

      return res.json({
        success: true,

        count:
          metrics.length,

        filters: {
          component:
            req.query.component ||
            "",

          from:
            Number(
              req.query.from ||
              0
            ),

          to:
            Number(
              req.query.to ||
              0
            ),

          limit:
            Math.min(
              Math.max(
                Number(
                  req.query.limit ||
                  500
                ) || 500,
                1
              ),
              5000
            )
        },

        metrics
      });

    } catch (error) {

      console.error(
        "MONITORING HISTORY ERROR:",
        error
      );

      return res
        .status(
          error.code === 7
            ? 403
            : error.code === 16
              ? 401
              : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error consultando histórico"
        });
    }
  }
);

// -------------------------------------
// CSV REPORT
// -------------------------------------

app.get(
  "/api/monitoring/report.csv",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await monitoringClient
          .getMetrics(
            token
          );

      const metrics =
        filterHistoricalMetrics(
          response.metrics || [],
          {
            ...req.query,

            limit:
              req.query.limit ||
              5000
          }
        );

      const header = [
        "timestamp",
        "component_id",
        "component_name",
        "cpu_usage",
        "memory_usage",
        "storage_usage",
        "active_connections",
        "total_requests"
      ];

      const rows =
        metrics.map(
          metric => [
            metric.timestamp,
            metric.componentId,
            metric.componentName,
            metric.cpuUsage,
            metric.memoryUsage,
            metric.storageUsage,
            metric.activeConnections,
            metric.totalRequests
          ]
            .map(
              csvEscape
            )
            .join(",")
        );

      const csv =
        [
          header.join(","),
          ...rows
        ]
          .join("\n") +
        "\n";

      const filename =
        `monitoring-report-${
          new Date()
            .toISOString()
            .replace(
              /[:.]/g,
              "-"
            )
        }.csv`;

      res.setHeader(
        "Content-Type",
        "text/csv; charset=utf-8"
      );

      res.setHeader(
        "Content-Disposition",
        `attachment; filename="${filename}"`
      );

      return res.send(
        csv
      );

    } catch (error) {

      console.error(
        "MONITORING CSV ERROR:",
        error
      );

      return res
        .status(
          error.code === 7
            ? 403
            : error.code === 16
              ? 401
              : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error generando reporte CSV"
        });
    }
  }
);

// -------------------------------------
// SERVICE STATUS
// -------------------------------------

app.get(
  "/api/monitoring/services/:name",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await monitoringClient
          .getServiceStatus(
            req.params.name,
            token
          );

      return res.json({
        success: true,
        service:
          response
      });

    } catch (error) {

      const status =
        error.code === 5
          ? 404
          : 502;

      return res
        .status(status)
        .json({
          success: false,
          message:
            error.details ||
            "Error consultando estado del servicio"
        });
    }
  }
);

// -------------------------------------
// ALERTS
// -------------------------------------

app.get(
  "/api/monitoring/alerts",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await monitoringClient
          .getAlerts(
            token
          );

      return res.json({
        success: true,
        alerts:
          response.alerts || []
      });

    } catch (error) {

      return res
        .status(
          error.code === 7
            ? 403
            : error.code === 16
              ? 401
              : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error consultando alertas"
        });
    }
  }
);

// -------------------------------------
// ALERT RULES
// -------------------------------------

app.get(
  "/api/monitoring/rules",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await monitoringClient
          .getAlertRules(
            token
          );

      return res.json({
        success: true,
        rules:
          response.rules || []
      });

    } catch (error) {

      return res
        .status(
          error.code === 7
            ? 403
            : error.code === 16
              ? 401
              : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error consultando reglas de alerta"
        });
    }
  }
);


// -------------------------------------
// CREATE ALERT RULE
// -------------------------------------

app.post(
  "/api/monitoring/rules",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const {
        id = "",
        name,
        metric,
        operator,
        threshold,
        componentName,
        enabled = true
      } = req.body || {};

      if (
        !name ||
        !metric ||
        !operator ||
        !componentName
      ) {

        return res
          .status(400)
          .json({
            success: false,
            message:
              "Faltan datos obligatorios"
          });
      }

      const response =
        await monitoringClient
          .createAlertRule(
            {
              id,
              name,
              metric,
              operator,
              threshold:
                Number(
                  threshold
                ),
              componentName,
              enabled:
                Boolean(
                  enabled
                )
            },
            token
          );

      return res.json({
        success:
          response.success,
        message:
          response.message,
        ruleId:
          response.ruleId
      });

    } catch (error) {

      return res
        .status(
          error.code === 7
            ? 403
            : error.code === 16
              ? 401
              : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error creando regla"
        });
    }
  }
);

// -------------------------------------
// UPDATE ALERT RULE
// -------------------------------------

app.put(
  "/api/monitoring/rules/:id",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const {
        name,
        metric,
        operator,
        threshold,
        componentName,
        enabled
      } = req.body || {};

      const response =
        await monitoringClient
          .updateAlertRule(
            {
              id:
                req.params.id,
              name,
              metric,
              operator,
              threshold:
                Number(
                  threshold
                ),
              componentName,
              enabled:
                Boolean(
                  enabled
                )
            },
            token
          );

      return res.json({
        success:
          response.success,
        message:
          response.message,
        rule:
          response.rule
      });

    } catch (error) {

      return res
        .status(
          error.code === 7
            ? 403
            : error.code === 16
              ? 401
              : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error actualizando regla"
        });
    }
  }
);

// -------------------------------------
// DELETE ALERT RULE
// -------------------------------------

app.delete(
  "/api/monitoring/rules/:id",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await monitoringClient
          .deleteAlertRule(
            req.params.id,
            token
          );

      return res.json({
        success:
          response.success,
        message:
          response.message,
        ruleId:
          response.ruleId
      });

    } catch (error) {

      return res
        .status(
          error.code === 7
            ? 403
            : error.code === 16
              ? 401
              : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error eliminando regla"
        });
    }
  }
);

// -------------------------------------
// HPC SUMMARY
// -------------------------------------

app.get(
  "/api/monitoring/hpc",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await monitoringClient
          .getHpcSummary(
            token
          );

      return res.json({
        success: true,
        hpc:
          response
      });

    } catch (error) {

      console.error(
        "MONITORING HPC ERROR:",
        error
      );

      return res
        .status(
          error.code === 7
            ? 403
            : error.code === 16
              ? 401
              : 502
        )
        .json({
          success: false,
          message:
            error.details ||
            "Error consultando resumen HPC"
        });
    }
  }
);

// -------------------------------------
// HPC NODE
// -------------------------------------

app.get(
  "/api/monitoring/nodes/:id",
  async (req, res) => {

    try {

      const token =
        tokenOnly(req);

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,
            message:
              "Token requerido"
          });
      }

      const response =
        await monitoringClient
          .getNodeStatus(
            req.params.id,
            token
          );

      return res.json({
        success: true,
        node:
          response
      });

    } catch (error) {

      const status =
        error.code === 5
          ? 404
          : 502;

      return res
        .status(status)
        .json({
          success: false,
          message:
            error.details ||
            "Error consultando nodo HPC"
        });
    }
  }
);

// =====================================
// STATIC FILES
// =====================================


// =====================================
// SCIENTIFIC HPC API - HTTP -> GRID
// =====================================

app.post(
  "/api/hpc/scientific",

  (
    req,
    res,
    next
  ) => {

    scientificUploadFields(
      req,
      res,
      error => {

        if (error) {

          return res
            .status(400)
            .json({
              success: false,

              message:
                error.code ===
                "LIMIT_FILE_SIZE"
                  ? "Los archivos científicos exceden el tamaño permitido"
                  : (
                    error.message ||
                    "Carga científica inválida"
                  )
            });
        }

        next();
      }
    );
  },

  async (req, res) => {

    try {

      const token =
        tokenOnly(
          req
        );

      if (!token) {

        return res
          .status(401)
          .json({
            success: false,

            message:
              "Token requerido"
          });
      }

      const source =
        req.files?.source?.[0];

      const dataset =
        req.files?.dataset?.[0];

      if (!source) {

        return res
          .status(400)
          .json({
            success: false,

            message:
              "source.c es obligatorio"
          });
      }

      if (!dataset) {

        return res
          .status(400)
          .json({
            success: false,

            message:
              "dataset es obligatorio"
          });
      }

      if (
        !String(
          source.originalname ||
          ""
        )
          .toLowerCase()
          .endsWith(
            ".c"
          )
      ) {

        return res
          .status(400)
          .json({
            success: false,

            message:
              "El código fuente debe ser un archivo .c"
          });
      }

      if (
        !source.buffer ||
        source.buffer.length === 0
      ) {

        return res
          .status(400)
          .json({
            success: false,

            message:
              "source.c está vacío"
          });
      }

      if (
        source.buffer.length >
        512 * 1024
      ) {

        return res
          .status(400)
          .json({
            success: false,

            message:
              "source.c excede 512 KiB"
          });
      }

      if (
        !dataset.buffer ||
        dataset.buffer.length === 0
      ) {

        return res
          .status(400)
          .json({
            success: false,

            message:
              "dataset está vacío"
          });
      }

      if (
        dataset.buffer.length >
        2 * 1024 * 1024
      ) {

        return res
          .status(400)
          .json({
            success: false,

            message:
              "dataset excede 2 MiB"
          });
      }

      const processes =
        Number.parseInt(
          String(
            req.body?.processes ||
            ""
          ),
          10
        );

      if (
        !Number.isInteger(
          processes
        ) ||
        processes < 1 ||
        processes > 4
      ) {

        return res
          .status(400)
          .json({
            success: false,

            message:
              "processes debe estar entre 1 y 4"
          });
      }

      const response =
        await gridClient
          .submitScientificJob(
            token,
            {
              sourceCode:
                source.buffer,

              dataset:
                dataset.buffer,

              processes
            }
          );

      return res.json({
        success: true,

        job: {
          jobId:
            response.jobId,

          success:
            response.success,

          exitCode:
            response.exitCode,

          launcherNode:
            response.launcherNode,

          durationMs:
            response.durationMs,

          output:
            response.output,

          sourceSha256:
            response.sourceSha256,

          datasetSha256:
            response.datasetSha256
        }
      });

    } catch (error) {

      console.error(
        "GRID SCIENTIFIC ERROR:",
        error
      );

      const statusCode =
        error.code === 3
          ? 400
          : error.code === 7
            ? 403
            : error.code === 9
              ? 409
              : error.code === 14
                ? 503
                : error.code === 16
                  ? 401
                  : 502;

      return res
        .status(
          statusCode
        )
        .json({
          success: false,

          message:
            error.details ||
            error.message ||
            "Error ejecutando trabajo científico"
        });
    }
  }
);

app.use(
  express.static(
    path.join(
      __dirname,
      "public"
    )
  )
);

// =====================================
// FALLBACK
// =====================================

app.use(
  (req, res) => {

    res.sendFile(
      path.join(
        __dirname,
        "public",
        "index.html"
      )
    );
  }
);

// =====================================
// SERVER
// =====================================

app.listen(
  PORT,
  () => {

    console.log(
      "==================================="
    );

    console.log(
      " UPB-CIENTIFICA WEB"
    );

    console.log(
      ` Puerto: ${PORT}`
    );

    console.log(
      ` Auth: ${AUTH_SERVICE}`
    );

    console.log(
      ` Photo: ${PHOTO_SERVICE}`
    );

    console.log(
      " Estado: ACTIVO"
    );

    console.log(
      "==================================="
    );
  }
);
