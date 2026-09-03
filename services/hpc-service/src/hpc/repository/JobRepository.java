package hpc.repository;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.SQLException;
import java.util.UUID;

public class JobRepository {

    public void createJob(
        UUID jobId,
        UUID userId,
        String language,
        String resources,
        String description
    ) throws SQLException {

        String sql = """
            INSERT INTO trabajo_hpc (
                id_job,
                id_usuario,
                lenguaje,
                estado,
                recursos,
                descripcion
            )
            VALUES (?, ?, ?, 'PENDIENTE', ?, ?)
            """;

        try (
            Connection connection =
                Database.getConnection();

            PreparedStatement statement =
                connection.prepareStatement(
                    sql
                )
        ) {

            statement.setObject(
                1,
                jobId
            );

            statement.setObject(
                2,
                userId
            );

            statement.setString(
                3,
                language
            );

            statement.setString(
                4,
                resources
            );

            statement.setString(
                5,
                description
            );

            statement.executeUpdate();
        }
    }

    public void markRunning(
        UUID jobId,
        UUID nodeId
    ) throws SQLException {

        String sql = """
            UPDATE trabajo_hpc
            SET
                id_nodo = ?,
                estado = 'EJECUTANDO',
                inicio = CURRENT_TIMESTAMP
            WHERE id_job = ?
            """;

        try (
            Connection connection =
                Database.getConnection();

            PreparedStatement statement =
                connection.prepareStatement(
                    sql
                )
        ) {

            statement.setObject(
                1,
                nodeId
            );

            statement.setObject(
                2,
                jobId
            );

            statement.executeUpdate();
        }
    }

    public void finishJob(
        UUID jobId,
        boolean success
    ) throws SQLException {

        String sql = """
            UPDATE trabajo_hpc
            SET
                estado = ?,
                fin = CURRENT_TIMESTAMP
            WHERE id_job = ?
            """;

        try (
            Connection connection =
                Database.getConnection();

            PreparedStatement statement =
                connection.prepareStatement(
                    sql
                )
        ) {

            statement.setString(
                1,
                success
                    ? "FINALIZADO"
                    : "ERROR"
            );

            statement.setObject(
                2,
                jobId
            );

            statement.executeUpdate();
        }
    }
}
