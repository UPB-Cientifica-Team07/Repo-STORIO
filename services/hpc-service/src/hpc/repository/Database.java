package hpc.repository;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;

public final class Database {

    private static final String URL =
        System.getenv()
            .getOrDefault(
                "HPC_DB_URL",
                "jdbc:postgresql://127.0.0.1:5434/upb_cientifica"
            );

    private static final String USER =
        System.getenv()
            .getOrDefault(
                "HPC_DB_USER",
                "upb_app"
            );

    private static final String PASSWORD =
        System.getenv()
            .getOrDefault(
                "HPC_DB_PASSWORD",
                "upb_dev_2026"
            );

    private Database() {
    }

    public static Connection getConnection()
        throws SQLException {

        return DriverManager.getConnection(
            URL,
            USER,
            PASSWORD
        );
    }
}
