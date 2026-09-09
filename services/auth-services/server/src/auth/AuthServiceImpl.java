package auth;

import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;
import java.security.SecureRandom;
import java.util.Base64;
import java.util.HashMap;
import java.util.Map;

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
    private static final int TOKEN_BYTES = 32;

    private static final long DEFAULT_TOKEN_TTL_SECONDS =
            3600L;

    private final Map<String, Session>
            sessions;

    private final SecureRandom
            secureRandom;

    private final long
            tokenTtlMillis;

    private final LdapDirectoryClient
            directoryClient;

    public AuthServiceImpl()
            throws RemoteException {

        super();

        sessions =
                new HashMap<>();

        secureRandom =
                new SecureRandom();

        tokenTtlMillis =
                loadTokenTtlMillis();

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

        System.out.println(
                " Token TTL: "
                        + (tokenTtlMillis / 1000L)
                        + " segundos"
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
                    generateSecureToken();

            long expiresAt =
                    System.currentTimeMillis()
                            + tokenTtlMillis;

            sessions.put(
                    token,
                    new Session(
                            user,
                            expiresAt
                    )
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

        Session session =
                sessions.get(
                        token
                );

        if (
                session == null
        ) {

            return new TokenResult(
                    false,
                    "Token inválido",
                    "",
                    ""
            );
        }

        if (
                session.isExpired()
        ) {

            sessions.remove(
                    token
            );

            return new TokenResult(
                    false,
                    "Token expirado",
                    "",
                    ""
            );
        }

        return new TokenResult(
                true,
                "Token válido",
                session.user.id,
                session.user.role
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
        cleanupExpiredSessions();

        for (
                Session session
                        : sessions.values()
        ) {

            User user =
                    session.user;

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
    // TOKEN / SESIONES
    // =====================================

    private String generateSecureToken() {

        byte[] bytes =
                new byte[TOKEN_BYTES];

        secureRandom.nextBytes(
                bytes
        );

        return Base64
                .getUrlEncoder()
                .withoutPadding()
                .encodeToString(
                        bytes
                );
    }

    private long loadTokenTtlMillis() {

        String value =
                System.getenv()
                        .getOrDefault(
                                "AUTH_TOKEN_TTL_SECONDS",
                                String.valueOf(
                                        DEFAULT_TOKEN_TTL_SECONDS
                                )
                        );

        try {

            long seconds =
                    Long.parseLong(
                            value
                    );

            if (seconds <= 0) {
                throw new NumberFormatException();
            }

            return Math.multiplyExact(
                    seconds,
                    1000L
            );

        } catch (
                ArithmeticException |
                NumberFormatException error
        ) {

            throw new IllegalArgumentException(
                    "AUTH_TOKEN_TTL_SECONDS debe ser un entero positivo válido"
            );
        }
    }

    private void cleanupExpiredSessions() {

        sessions.entrySet()
                .removeIf(
                        entry ->
                                entry
                                        .getValue()
                                        .isExpired()
                );
    }

    public synchronized boolean logout(
            String token
    ) {

        if (
                token == null ||
                token.isBlank()
        ) {
            return false;
        }

        return sessions.remove(
                token
        ) != null;
    }

    private static class Session {

        private final User user;
        private final long expiresAtMillis;

        private Session(
                User user,
                long expiresAtMillis
        ) {

            this.user =
                    user;

            this.expiresAtMillis =
                    expiresAtMillis;
        }

        private boolean isExpired() {

            return System.currentTimeMillis()
                    >= expiresAtMillis;
        }
    }

    // =====================================
    // USUARIO AUTENTICADO
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
