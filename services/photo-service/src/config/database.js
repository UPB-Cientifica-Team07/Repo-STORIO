const { Pool } = require("pg");

const pool = new Pool({
  host: process.env.DB_HOST || "localhost",
  port: Number(process.env.DB_PORT || 5434),
  user: process.env.DB_USER || "upb_app",
  password: process.env.DB_PASSWORD || "upb_dev_2026",
  database: process.env.DB_NAME || "upb_cientifica",
  max: 10,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 3000
});

module.exports = pool;