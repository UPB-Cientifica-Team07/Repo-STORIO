package auth;

import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

public class AuthServiceImpl
        extends UnicastRemoteObject
        implements IAuthService {

    private static final long serialVersionUID = 1L;

    private final Map<String, User> usersByUsername;
    private final Map<String, User> usersById;
    private final Map<String, String> tokens;

    public AuthServiceImpl() throws RemoteException {
        super();

        usersByUsername = new HashMap<>();
        usersById = new HashMap<>();
        tokens = new HashMap<>();

        loadDefaultUsers();
    }

    // =====================================
    // USUARIOS POR DEFECTO
    // =====================================

// =====================================
// USUARIOS POR DEFECTO
// =====================================

private void loadDefaultUsers() {

    // =====================================
    // ADMINISTRADOR
    // =====================================

    User admin = new User(
            "user-001",
            "samuel",
            "123456",
            "ADMIN"
    );

    usersByUsername.put(
            admin.username,
            admin
    );

    usersById.put(
            admin.id,
            admin
    );

    // =====================================
    // USUARIO NORMAL 1
    // =====================================

    User normalUser = new User(
            "user-002",
            "prueba",
            "123456",
            "USUARIO"
    );

    usersByUsername.put(
            normalUser.username,
            normalUser
    );

    usersById.put(
            normalUser.id,
            normalUser
    );

    // =====================================
    // USUARIO NORMAL 2
    // =====================================

    User thirdUser = new User(
            "user-003",
            "tercero",
            "123456",
            "USUARIO"
    );

    usersByUsername.put(
            thirdUser.username,
            thirdUser
    );

    usersById.put(
            thirdUser.id,
            thirdUser
    );

    // =====================================
    // LOG DE USUARIOS
    // =====================================

    System.out.println(
            "Usuarios por defecto cargados:"
    );

    System.out.println(
            " - samuel | user-001 | ADMIN"
    );

    System.out.println(
            " - prueba | user-002 | USUARIO"
    );

    System.out.println(
            " - tercero | user-003 | USUARIO"
    );
}


    // =====================================
    // LOGIN
    // =====================================

    @Override
    public synchronized AuthResult login(
            String username,
            String password
    ) throws RemoteException {

        if (
                username == null ||
                username.isBlank()
        ) {

            return new AuthResult(
                    false,
                    "El usuario es obligatorio",
                    "",
                    "",
                    "",
                    ""
            );
        }

        if (
                password == null ||
                password.isBlank()
        ) {

            return new AuthResult(
                    false,
                    "La contraseña es obligatoria",
                    "",
                    "",
                    "",
                    ""
            );
        }

        User user =
                usersByUsername.get(
                        username
                );

        if (user == null) {

            return new AuthResult(
                    false,
                    "Usuario no encontrado",
                    "",
                    "",
                    "",
                    ""
            );
        }

        if (
                !user.password.equals(
                        password
                )
        ) {

            return new AuthResult(
                    false,
                    "Credenciales inválidas",
                    "",
                    "",
                    "",
                    ""
            );
        }

        String token =
                UUID.randomUUID()
                        .toString();

        tokens.put(
                token,
                user.id
        );

        System.out.println(
                "Login correcto"
                        + " | Usuario: "
                        + user.username
                        + " | ID: "
                        + user.id
                        + " | Rol: "
                        + user.role
        );

        return new AuthResult(
                true,
                "Login correcto",
                user.id,
                user.username,
                user.role,
                token
        );
    }

    // =====================================
    // VALIDAR TOKEN
    // =====================================

    @Override
    public synchronized TokenResult validateToken(
            String token
    ) throws RemoteException {

        if (
                token == null ||
                token.isBlank()
        ) {

            return new TokenResult(
                    false,
                    "El token es obligatorio",
                    "",
                    ""
            );
        }

        String userId =
                tokens.get(
                        token
                );

        if (userId == null) {

            return new TokenResult(
                    false,
                    "Token inválido",
                    "",
                    ""
            );
        }

        User user =
                usersById.get(
                        userId
                );

        if (user == null) {

            return new TokenResult(
                    false,
                    "Usuario asociado al token no encontrado",
                    "",
                    ""
            );
        }

        return new TokenResult(
                true,
                "Token válido",
                user.id,
                user.role
        );
    }

    // =====================================
    // CONSULTAR USUARIO
    // =====================================

    @Override
    public synchronized AuthResult getUser(
            String userId
    ) throws RemoteException {

        if (
                userId == null ||
                userId.isBlank()
        ) {

            return new AuthResult(
                    false,
                    "El user ID es obligatorio",
                    "",
                    "",
                    "",
                    ""
            );
        }

        User user =
                usersById.get(
                        userId
                );

        if (user == null) {

            return new AuthResult(
                    false,
                    "Usuario no encontrado",
                    "",
                    "",
                    "",
                    ""
            );
        }

        return new AuthResult(
                true,
                "Usuario encontrado",
                user.id,
                user.username,
                user.role,
                ""
        );
    }

    // =====================================
    // MODELO INTERNO DE USUARIO
    // =====================================

    private static class User {

        private final String id;
        private final String username;
        private final String password;
        private final String role;

        private User(
                String id,
                String username,
                String password,
                String role
        ) {

            this.id =
                    id;

            this.username =
                    username;

            this.password =
                    password;

            this.role =
                    role;
        }
    }
}