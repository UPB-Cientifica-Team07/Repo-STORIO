const { Pool } = require("pg");

const dbPassword =
  process.env.PHOTO_DB_PASSWORD ||
  process.env.DB_PASSWORD;

if (!dbPassword) {
  throw new Error(
    "PHOTO_DB_PASSWORD o DB_PASSWORD es obligatorio"
  );
}

const pool = new Pool({
  host: process.env.DB_HOST || "localhost",
  port: Number(process.env.DB_PORT || 5434),
  user: process.env.DB_USER || "upb_app",
  password: dbPassword,
  database: process.env.DB_NAME || "upb_cientifica",
  max: 10,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 3000
});

module.exports = pool;