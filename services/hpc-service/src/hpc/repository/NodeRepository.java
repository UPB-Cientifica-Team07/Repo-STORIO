package hpc.repository;

import hpc.common.NodeInfo;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.UUID;

public class NodeRepository {

    public UUID upsertNode(
        NodeInfo node
    ) throws SQLException {

        String sql = """
            INSERT INTO nodo_hpc (
                hostname,
                cpu,
                memoria_mb,
                estado,
                ip,
                ubicacion,
                last_heartbeat
            )
            VALUES (
                ?, ?, ?, ?, ?, ?,
                CURRENT_TIMESTAMP
            )
            ON CONFLICT (hostname)
            DO UPDATE SET
                cpu = EXCLUDED.cpu,
                memoria_mb = EXCLUDED.memoria_mb,
                estado = EXCLUDED.estado,
                ip = EXCLUDED.ip,
                ubicacion = EXCLUDED.ubicacion,
                cpu_usage = 0,
                memory_usage = 0,
                storage_usage = 0,
                last_heartbeat = CURRENT_TIMESTAMP
            RETURNING id_nodo
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
                node.getHostname()
            );

            statement.setInt(
                2,
                node.getCpuCores()
            );

            statement.setLong(
                3,
                node.getMemoryMb()
            );

            statement.setString(
                4,
                "DISPONIBLE"
            );

            statement.setString(
                5,
                node.getIp()
            );

            statement.setString(
                6,
                "HPC Cluster"
            );

            try (
                ResultSet result =
                    statement.executeQuery()
            ) {

                if (!result.next()) {

                    throw new SQLException(
                        "No se obtuvo id_nodo"
                    );
                }

                return result.getObject(
                    "id_nodo",
                    UUID.class
                );
            }
        }
    }

    public void markAllInactive()
        throws SQLException {

        String sql = """
            UPDATE nodo_hpc
            SET estado = 'INACTIVO'
            WHERE estado <> 'INACTIVO'
            """;

        try (
            Connection connection =
                Database.getConnection();

            PreparedStatement statement =
                connection.prepareStatement(
                    sql
                )
        ) {

            statement.executeUpdate();
        }
    }

    public void updateTelemetry(
        UUID nodeId,
        double cpuUsage,
        double memoryUsage,
        long storageUsage
    ) throws SQLException {

        String sql = """
            UPDATE nodo_hpc
            SET
                cpu_usage = ?,
                memory_usage = ?,
                storage_usage = ?,
                last_heartbeat = CURRENT_TIMESTAMP
            WHERE id_nodo = ?
            """;

        try (
            Connection connection =
                Database.getConnection();

            PreparedStatement statement =
                connection.prepareStatement(
                    sql
                )
        ) {

            statement.setDouble(
                1,
                cpuUsage
            );

            statement.setDouble(
                2,
                memoryUsage
            );

            statement.setLong(
                3,
                storageUsage
            );

            statement.setObject(
                4,
                nodeId
            );

            statement.executeUpdate();
        }
    }

    public void updateStatus(
        UUID nodeId,
        String status
    ) throws SQLException {

        String sql = """
            UPDATE nodo_hpc
            SET estado = ?
            WHERE id_nodo = ?
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
                status
            );

            statement.setObject(
                2,
                nodeId
            );

            statement.executeUpdate();
        }
    }
}
