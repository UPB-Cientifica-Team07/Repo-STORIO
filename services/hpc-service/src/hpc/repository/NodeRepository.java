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
                ubicacion
            )
            VALUES (?, ?, ?, ?, ?, ?)
            ON CONFLICT (hostname)
            DO UPDATE SET
                cpu = EXCLUDED.cpu,
                memoria_mb = EXCLUDED.memoria_mb,
                estado = EXCLUDED.estado,
                ip = EXCLUDED.ip,
                ubicacion = EXCLUDED.ubicacion
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
