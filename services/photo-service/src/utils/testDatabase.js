const pool = require("../config/database");

async function testDatabase() {
  try {
    const result =
      await pool.query(
        "SELECT current_database() AS database, current_user AS user, NOW() AS timestamp"
      );

    console.log("===================================");
    console.log(" DATABASE CONNECTION");
    console.log(" Estado: OK");
    console.log(
      ` Base de datos: ${result.rows[0].database}`
    );
    console.log(
      ` Usuario: ${result.rows[0].user}`
    );
    console.log(
      ` Timestamp: ${result.rows[0].timestamp}`
    );
    console.log("===================================");
  } catch (error) {
    console.error("===================================");
    console.error(" DATABASE CONNECTION");
    console.error(" Estado: ERROR");
    console.error(error.message);
    console.error("===================================");

    process.exitCode = 1;
  } finally {
    await pool.end();
  }
}

testDatabase();