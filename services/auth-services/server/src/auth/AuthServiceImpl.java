package auth;

import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;
import java.security.SecureRandom;
import java.util.Base64;
import java.util.HashMap;
import java.util.Locale;
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

    private final Map<String, LoginAttempt>
            loginAttempts;

    private final int
            loginMaxAttempts;

    private final long
            loginWindowMillis;

    private final long
            loginBlockMillis;

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

        loginAttempts =
                new HashMap<>();

        loginMaxAttempts =
                loadPositiveInt(
                        "AUTH_LOGIN_MAX_ATTEMPTS",
                        5
                );

        loginWindowMillis =
                Math.multiplyExact(
                        loadPositiveLong(
                                "AUTH_LOGIN_WINDOW_SECONDS",
                                300L
                        ),
                        1000L
                );

        loginBlockMillis =
                Math.multiplyExact(
                        loadPositiveLong(
                                "AUTH_LOGIN_BLOCK_SECONDS",
                                300L
                        ),
                        1000L
                );

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

        System.out.println(
                " RMI login rate limit: "
                        + loginMaxAttempts
                        + " intentos"
        );

        System.out.println(
                " RMI ventana login: "
                        + (loginWindowMillis / 1000L)
                        + " segundos"
        );

        System.out.println(
                " RMI bloqueo login: "
                        + (loginBlockMillis / 1000L)
                        + " segundos"
        );
    }

    // =====================================
    // LOGIN RMI
    // =====================================

    @Override
    public synchronized AuthResult login(
            String username,
            String password
    ) throws RemoteException {

        return authenticate(
                username,
                password,
                true
        );
    }

    // =====================================
    // LOGIN HTTP
    // =====================================

    public synchronized AuthResult loginHttp(
            String username,
            String password
    ) throws RemoteException {

        /*
         * AuthServer aplica su propio limitador
         * por usuario + IP.
         *
         * No aplicamos aquí nuevamente el
         * limitador RMI para evitar contabilizar
         * dos veces un mismo intento HTTP.
         */
        return authenticate(
                username,
                password,
                false
        );
    }

    // =====================================
    // AUTENTICACIÓN COMÚN
    // =====================================

    private AuthResult authenticate(
            String username,
            String password,
            boolean applyLoginRateLimit
    ) {

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

        String normalizedUsername =
                username
                        .trim()
                        .toLowerCase(
                                Locale.ROOT
                        );

        long now =
                System.currentTimeMillis();

        if (
                applyLoginRateLimit &&
                isLoginBlocked(
                        normalizedUsername,
                        now
                )
        ) {

            System.err.println(
                    "Login RMI bloqueado por rate limit"
                            + " | Usuario: "
                            + normalizedUsername
            );

            return new AuthResult(
                    false,
                    "Demasiados intentos. Intente más tarde.",
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

                if (applyLoginRateLimit) {

                    boolean blocked =
                            registerLoginFailure(
                                    normalizedUsername,
                                    now
                            );

                    System.err.println(
                            "Login RMI LDAP rechazado"
                                    + " | Usuario: "
                                    + normalizedUsername
                                    + " | Bloqueado: "
                                    + blocked
                    );
                }

                return new AuthResult(
                        false,
                        "Credenciales inválidas",
                        "",
                        "",
                        "",
                        ""
                );
            }

            if (applyLoginRateLimit) {

                clearLoginFailures(
                        normalizedUsername
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
    // RATE LIMIT RMI
    // =====================================

    private boolean isLoginBlocked(
            String username,
            long now
    ) {

        LoginAttempt attempt =
                loginAttempts.get(
                        username
                );

        if (attempt == null) {
            return false;
        }

        if (
                attempt.blockedUntilMillis
                        > now
        ) {
            return true;
        }

        if (
                attempt.blockedUntilMillis
                        > 0L
        ) {

            loginAttempts.remove(
                    username
            );

            return false;
        }

        if (
                now - attempt.windowStartedMillis
                        >= loginWindowMillis
        ) {

            loginAttempts.remove(
                    username
            );

            return false;
        }

        return false;
    }

    private boolean registerLoginFailure(
            String username,
            long now
    ) {

        LoginAttempt attempt =
                loginAttempts.get(
                        username
                );

        if (
                attempt == null ||
                now - attempt.windowStartedMillis
                        >= loginWindowMillis
        ) {

            attempt =
                    new LoginAttempt(
                            now
                    );

            loginAttempts.put(
                    username,
                    attempt
            );
        }

        attempt.failedAttempts++;

        if (
                attempt.failedAttempts
                        >= loginMaxAttempts
        ) {

            attempt.blockedUntilMillis =
                    now
                            + loginBlockMillis;

            return true;
        }

        return false;
    }

    private void clearLoginFailures(
            String username
    ) {

        loginAttempts.remove(
                username
        );
    }

    private static int loadPositiveInt(
            String key,
            int defaultValue
    ) {

        String value =
                System.getenv(
                        key
                );

        if (
                value == null ||
                value.isBlank()
        ) {
            return defaultValue;
        }

        try {

            int parsed =
                    Integer.parseInt(
                            value.trim()
                    );

            if (parsed <= 0) {
                throw new NumberFormatException();
            }

            return parsed;

        } catch (
                NumberFormatException error
        ) {

            throw new IllegalArgumentException(
                    key
                            + " debe ser un entero positivo"
            );
        }
    }

    private static long loadPositiveLong(
            String key,
            long defaultValue
    ) {

        String value =
                System.getenv(
                        key
                );

        if (
                value == null ||
                value.isBlank()
        ) {
            return defaultValue;
        }

        try {

            long parsed =
                    Long.parseLong(
                            value.trim()
                    );

            if (parsed <= 0L) {
                throw new NumberFormatException();
            }

            return parsed;

        } catch (
                NumberFormatException error
        ) {

            throw new IllegalArgumentException(
                    key
                            + " debe ser un entero positivo"
            );
        }
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

    private static class LoginAttempt {

        private final long
                windowStartedMillis;

        private int
                failedAttempts;

        private long
                blockedUntilMillis;

        private LoginAttempt(
                long windowStartedMillis
        ) {

            this.windowStartedMillis =
                    windowStartedMillis;

            this.failedAttempts =
                    0;

            this.blockedUntilMillis =
                    0L;
        }
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
