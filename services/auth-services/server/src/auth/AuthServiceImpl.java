package auth;

import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

import javax.naming.NamingException;

public class AuthServiceImpl
        extends UnicastRemoteObject
        implements IAuthService {

    private static final long serialVersionUID =
            1L;

    /*
     * Tokens permanecen gestionados por
     * Auth Service.
     *
     * token -> usuario autenticado
     */
    private final Map<String, User>
            sessions;

    private final LdapDirectoryClient
            directoryClient;

    public AuthServiceImpl()
            throws RemoteException {

        super();

        sessions =
                new HashMap<>();

        directoryClient =
                new LdapDirectoryClient();

        System.out.println(
                "Auth Service configurado con OpenLDAP"
        );

        System.out.println(
                " LDAP: "
                        + System.getenv()
                                .getOrDefault(
                                        "LDAP_URL",
                                        "ldap://127.0.0.1:389"
                                )
        );

        System.out.println(
                " Base DN: "
                        + System.getenv()
                                .getOrDefault(
                                        "LDAP_BASE_DN",
                                        "dc=upb-cientifica,dc=local"
                                )
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

        try {

            LdapDirectoryClient.DirectoryUser
                    directoryUser =
                    directoryClient
                            .authenticate(
                                    username.trim(),
                                    password
                            );

            if (
                    directoryUser == null
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

            User user =
                    new User(
                            directoryUser.id(),
                            directoryUser.username(),
                            directoryUser.role()
                    );

            String token =
                    UUID.randomUUID()
                            .toString();

            sessions.put(
                    token,
                    user
            );

            System.out.println(
                    "Login LDAP correcto"
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

        } catch (
                NamingException error
        ) {

            System.err.println(
                    "Error consultando OpenLDAP: "
                            + error.getMessage()
            );

            return new AuthResult(
                    false,
                    "Directory Service no disponible",
                    "",
                    "",
                    "",
                    ""
            );
        }
    }

    // =====================================
    // VALIDAR TOKEN
    // =====================================

    @Override
    public synchronized TokenResult
    validateToken(
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

        User user =
                sessions.get(
                        token
                );

        if (
                user == null
        ) {

            return new TokenResult(
                    false,
                    "Token inválido",
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

        /*
         * Primero buscamos entre sesiones
         * ya autenticadas.
         */
        for (
                User user
                        : sessions.values()
        ) {

            if (
                    user.id.equals(
                            userId
                    )
            ) {

                return userResult(
                        user
                );
            }
        }

        /*
         * Si el usuario todavía no ha iniciado
         * sesión, se consulta el directorio.
         */
        try {

            LdapDirectoryClient.DirectoryUser
                    directoryUser =
                    directoryClient
                            .findByDirectoryId(
                                    userId
                            );

            if (
                    directoryUser == null
            ) {

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
                    directoryUser.id(),
                    directoryUser.username(),
                    directoryUser.role(),
                    ""
            );

        } catch (
                NamingException error
        ) {

            return new AuthResult(
                    false,
                    "Directory Service no disponible",
                    "",
                    "",
                    "",
                    ""
            );
        }
    }

    private AuthResult userResult(
            User user
    ) {

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
    // SESIÓN AUTENTICADA
    // =====================================

    private static class User {

        private final String id;
        private final String username;
        private final String role;

        private User(
                String id,
                String username,
                String role
        ) {

            this.id =
                    id;

            this.username =
                    username;

            this.role =
                    role;
        }
    }
}
