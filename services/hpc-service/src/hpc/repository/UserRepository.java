package hpc.repository;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.UUID;

public class UserRepository {

    public UUID findIdByDirectoryId(
        String directoryId
    ) throws SQLException {

        String sql = """
            SELECT id_usuario
            FROM usuario
            WHERE
                directorio_id = ?
                AND estado = true
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
                directoryId
            );

            try (
                ResultSet result =
                    statement.executeQuery()
            ) {

                if (!result.next()) {

                    return null;
                }

                return result.getObject(
                    "id_usuario",
                    UUID.class
                );
            }
        }
    }
}
