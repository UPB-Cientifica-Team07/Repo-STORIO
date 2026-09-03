const express = require("express");
const path = require("path");
const axios = require("axios");
const multer = require("multer");

const fileClient =
  require("./src/fileClient");

const app = express();

const PORT =
  process.env.PORT ||
  8080;

const AUTH_SERVICE =
  process.env.AUTH_SERVICE ||
  "http://127.0.0.1:8081";

const PHOTO_SERVICE =
  process.env.PHOTO_SERVICE ||
  "http://127.0.0.1:50052";

const STREAMING_SERVICE =
  process.env.STREAMING_SERVICE ||
  "http://127.0.0.1:50054";

const upload =
  multer({
    storage:
      multer.memoryStorage(),

    limits: {
      fileSize:
        20 * 1024 * 1024
    }
  });

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

      const response =
        await axios.get(
          `${AUTH_SERVICE}/internal/auth/login`,
          {
            params: {
              username,
              password
            },

            timeout: 5000
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
// STATIC FILES
// =====================================

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
