package auth;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;

import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;

import java.util.HashMap;
import java.util.Locale;
import java.util.Map;

public class AuthServer {

    private static final int RMI_PORT = 1099;
    private static final int HTTP_PORT = 8081;

    private static final String SERVICE_NAME = "AuthService";

    private static final int LOGIN_MAX_ATTEMPTS =
            loadPositiveInt(
                    "AUTH_LOGIN_MAX_ATTEMPTS",
                    5
            );

    private static final long LOGIN_WINDOW_MILLIS =
            loadPositiveLong(
                    "AUTH_LOGIN_WINDOW_SECONDS",
                    300L
            ) * 1000L;

    private static final long LOGIN_BLOCK_MILLIS =
            loadPositiveLong(
                    "AUTH_LOGIN_BLOCK_SECONDS",
                    300L
            ) * 1000L;

    private static final Map<String, LoginAttempt>
            LOGIN_ATTEMPTS =
            new HashMap<>();

    public static void main(String[] args) {

        try {

            System.out.println(
                    "==================================="
            );

            System.out.println(
                    " AUTH SERVICE"
            );

            System.out.println(
                    " Tecnología principal: Java RMI"
            );

            System.out.println(
                    " RMI Registry: " + RMI_PORT
            );

            System.out.println(
                    " Bridge HTTP interno: " + HTTP_PORT
            );

            System.out.println(
                    " Login rate limit: "
                            + LOGIN_MAX_ATTEMPTS
                            + " intentos"
            );

            System.out.println(
                    " Ventana login: "
                            + (LOGIN_WINDOW_MILLIS / 1000L)
                            + " segundos"
            );

            System.out.println(
                    " Bloqueo login: "
                            + (LOGIN_BLOCK_MILLIS / 1000L)
                            + " segundos"
            );

            System.out.println(
                    " Estado: INICIANDO"
            );

            System.out.println(
                    "==================================="
            );

            /*
             * Una sola instancia del servicio.
             *
             * RMI y HTTP comparten exactamente el mismo
             * AuthServiceImpl y, por lo tanto, los mismos tokens.
             */
            AuthServiceImpl authService =
                    new AuthServiceImpl();

            // =====================================
            // RMI
            // =====================================

            Registry registry;

            try {

                registry =
                        LocateRegistry.createRegistry(
                                RMI_PORT
                        );

                System.out.println(
                        "RMI Registry creado en puerto "
                                + RMI_PORT
                );

            } catch (Exception e) {

                registry =
                        LocateRegistry.getRegistry(
                                RMI_PORT
                        );

                System.out.println(
                        "RMI Registry existente detectado"
                );
            }

            registry.rebind(
                    SERVICE_NAME,
                    authService
            );

            // =====================================
            // HTTP BRIDGE
            // =====================================

            HttpServer httpServer =
                    HttpServer.create(
                            new InetSocketAddress(
                                    HTTP_PORT
                            ),
                            0
                    );

            // Validar token
            httpServer.createContext(
                    "/internal/auth/validate",
                    exchange ->
                            handleValidateToken(
                                    exchange,
                                    authService
                            )
            );

            // Login
            httpServer.createContext(
                    "/internal/auth/login",
                    exchange ->
                            handleLogin(
                                    exchange,
                                    authService
                            )
            );

            // Logout / revocación
            httpServer.createContext(
                    "/internal/auth/logout",
                    exchange ->
                            handleLogout(
                                    exchange,
                                    authService
                            )
            );

            httpServer.setExecutor(
                    null
            );

            httpServer.start();

            System.out.println(
                    "==================================="
            );

            System.out.println(
                    " AUTH SERVICE ACTIVO"
            );

            System.out.println(
                    " Servicio RMI: "
                            + SERVICE_NAME
            );

            System.out.println(
                    " RMI: "
                            + RMI_PORT
            );

            System.out.println(
                    " HTTP Bridge: "
                            + HTTP_PORT
            );

            System.out.println(
                    " Endpoint Login:"
            );

            System.out.println(
                    " /internal/auth/login"
            );

            System.out.println(
                    " Endpoint Validate:"
            );

            System.out.println(
                    " /internal/auth/validate"
            );

            System.out.println(
                    "==================================="
            );

        } catch (Exception e) {

            System.err.println(
                    "Error iniciando Auth Service:"
            );

            e.printStackTrace();
        }
    }

    // =====================================
    // LOGIN HTTP
    // =====================================

    private static void handleLogin(
            HttpExchange exchange,
            AuthServiceImpl authService
    ) throws IOException {

        try {

            if (!exchange
                    .getRequestMethod()
                    .equalsIgnoreCase("POST")) {

                sendResponse(
                        exchange,
                        405,
                        "false||||METHOD_NOT_ALLOWED"
                );

                return;
            }

            String requestBody =
                    new String(
                            exchange
                                    .getRequestBody()
                                    .readAllBytes(),
                            StandardCharsets.UTF_8
                    );

            String username =
                    getQueryParameter(
                            requestBody,
                            "username"
                    );

            String password =
                    getQueryParameter(
                            requestBody,
                            "password"
                    );

            if (username == null ||
                    username.isBlank()) {

                sendResponse(
                        exchange,
                        400,
                        "false||||USERNAME_REQUIRED"
                );

                return;
            }

            if (password == null ||
                    password.isBlank()) {

                sendResponse(
                        exchange,
                        400,
                        "false||||PASSWORD_REQUIRED"
                );

                return;
            }

            String normalizedUsername =
                    username
                            .trim()
                            .toLowerCase(
                                    Locale.ROOT
                            );

            String remoteAddress =
                    getRemoteAddress(
                            exchange
                    );

            String rateLimitKey =
                    normalizedUsername
                            + "|"
                            + remoteAddress;

            long now =
                    System.currentTimeMillis();

            long retryAfterSeconds =
                    getRetryAfterSeconds(
                            rateLimitKey,
                            now
                    );

            if (retryAfterSeconds > 0L) {

                exchange
                        .getResponseHeaders()
                        .set(
                                "Retry-After",
                                Long.toString(
                                        retryAfterSeconds
                                )
                        );

                System.err.println(
                        "Login bloqueado por rate limit"
                                + " | Usuario: "
                                + normalizedUsername
                                + " | IP: "
                                + remoteAddress
                );

                sendResponse(
                        exchange,
                        429,
                        "false||||TOO_MANY_ATTEMPTS"
                );

                return;
            }

            AuthResult result =
                    authService.loginHttp(
                            username,
                            password
                    );

            if (result.isSuccess()) {

                clearLoginFailures(
                        rateLimitKey
                );

            } else if (
                    "Credenciales inválidas"
                            .equals(
                                    result.getMessage()
                            )
            ) {

                boolean blocked =
                        registerLoginFailure(
                                rateLimitKey,
                                now
                        );

                System.err.println(
                        "Login LDAP rechazado"
                                + " | Usuario: "
                                + normalizedUsername
                                + " | IP: "
                                + remoteAddress
                                + " | Bloqueado: "
                                + blocked
                );
            }

            /*
             * Formato:
             *
             * success|userId|role|token|message
             */

            String response =
                    result.isSuccess()
                            + "|"
                            + safe(result.getUserId())
                            + "|"
                            + safe(result.getRole())
                            + "|"
                            + safe(result.getToken())
                            + "|"
                            + safe(result.getMessage());

            sendResponse(
                    exchange,
                    200,
                    response
            );

        } catch (Exception e) {

            sendResponse(
                    exchange,
                    500,
                    "false||||INTERNAL_ERROR"
            );
        }
    }

    // =====================================
    // VALIDATE TOKEN HTTP
    // =====================================

    private static void handleValidateToken(
            HttpExchange exchange,
            AuthServiceImpl authService
    ) throws IOException {

        try {

            if (!exchange
                    .getRequestMethod()
                    .equalsIgnoreCase("GET")) {

                sendResponse(
                        exchange,
                        405,
                        "false|||METHOD_NOT_ALLOWED"
                );

                return;
            }

            String token =
                    getBearerToken(
                            exchange
                    );

            if (token == null ||
                    token.isBlank()) {

                sendResponse(
                        exchange,
                        401,
                        "false|||TOKEN_REQUIRED"
                );

                return;
            }

            TokenResult result =
                    authService.validateToken(
                            token
                    );

            /*
             * Formato:
             *
             * valid|userId|role|message
             */

            String response =
                    result.isValid()
                            + "|"
                            + safe(result.getUserId())
                            + "|"
                            + safe(result.getRole())
                            + "|"
                            + safe(result.getMessage());

            sendResponse(
                    exchange,
                    200,
                    response
            );

        } catch (Exception e) {

            sendResponse(
                    exchange,
                    500,
                    "false|||INTERNAL_ERROR"
            );
        }
    }

    // =====================================
    // LOGOUT HTTP
    // =====================================

    private static void handleLogout(
            HttpExchange exchange,
            AuthServiceImpl authService
    ) throws IOException {

        try {

            if (!exchange
                    .getRequestMethod()
                    .equalsIgnoreCase("POST")) {

                sendResponse(
                        exchange,
                        405,
                        "false|METHOD_NOT_ALLOWED"
                );

                return;
            }

            String token =
                    getBearerToken(
                            exchange
                    );

            if (token == null ||
                    token.isBlank()) {

                sendResponse(
                        exchange,
                        401,
                        "false|TOKEN_REQUIRED"
                );

                return;
            }

            boolean revoked =
                    authService.logout(
                            token
                    );

            sendResponse(
                    exchange,
                    200,
                    revoked
                            ? "true|TOKEN_REVOKED"
                            : "false|TOKEN_INVALID"
            );

        } catch (Exception e) {

            sendResponse(
                    exchange,
                    500,
                    "false|INTERNAL_ERROR"
            );
        }
    }

    // =====================================
    // LOGIN RATE LIMIT
    // =====================================

    private static synchronized long
    getRetryAfterSeconds(
            String key,
            long now
    ) {

        LoginAttempt attempt =
                LOGIN_ATTEMPTS.get(
                        key
                );

        if (attempt == null) {
            return 0L;
        }

        if (
                attempt.blockedUntilMillis
                        > now
        ) {

            long remainingMillis =
                    attempt.blockedUntilMillis
                            - now;

            return Math.max(
                    1L,
                    (remainingMillis + 999L)
                            / 1000L
            );
        }

        if (
                attempt.blockedUntilMillis
                        > 0L
        ) {

            LOGIN_ATTEMPTS.remove(
                    key
            );

            return 0L;
        }

        if (
                now - attempt.windowStartedMillis
                        >= LOGIN_WINDOW_MILLIS
        ) {

            LOGIN_ATTEMPTS.remove(
                    key
            );
        }

        return 0L;
    }

    private static synchronized boolean
    registerLoginFailure(
            String key,
            long now
    ) {

        LoginAttempt attempt =
                LOGIN_ATTEMPTS.get(
                        key
                );

        if (
                attempt == null ||
                now - attempt.windowStartedMillis
                        >= LOGIN_WINDOW_MILLIS
        ) {

            attempt =
                    new LoginAttempt(
                            now
                    );

            LOGIN_ATTEMPTS.put(
                    key,
                    attempt
            );
        }

        attempt.failedAttempts++;

        if (
                attempt.failedAttempts
                        >= LOGIN_MAX_ATTEMPTS
        ) {

            attempt.blockedUntilMillis =
                    now
                            + LOGIN_BLOCK_MILLIS;

            return true;
        }

        return false;
    }

    private static synchronized void
    clearLoginFailures(
            String key
    ) {

        LOGIN_ATTEMPTS.remove(
                key
        );
    }

    private static String getRemoteAddress(
            HttpExchange exchange
    ) {

        if (
                exchange.getRemoteAddress() == null
        ) {
            return "unknown";
        }

        if (
                exchange
                        .getRemoteAddress()
                        .getAddress()
                        != null
        ) {

            return exchange
                    .getRemoteAddress()
                    .getAddress()
                    .getHostAddress();
        }

        return exchange
                .getRemoteAddress()
                .getHostString();
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

    private static final class LoginAttempt {

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

    // =====================================
    // BEARER TOKEN
    // =====================================

    private static String getBearerToken(
            HttpExchange exchange
    ) {

        String authorization =
                exchange
                        .getRequestHeaders()
                        .getFirst(
                                "Authorization"
                        );

        if (authorization == null) {
            return null;
        }

        String prefix =
                "Bearer ";

        if (!authorization
                .regionMatches(
                        true,
                        0,
                        prefix,
                        0,
                        prefix.length()
                )) {

            return null;
        }

        String token =
                authorization
                        .substring(
                                prefix.length()
                        )
                        .trim();

        return token.isEmpty()
                ? null
                : token;
    }

    // =====================================
    // FORM PARAMETER
    // =====================================

    private static String getQueryParameter(
            String query,
            String name
    ) {

        if (query == null ||
                query.isBlank()) {

            return null;
        }

        String[] parameters =
                query.split("&");

        for (String parameter :
                parameters) {

            String[] parts =
                    parameter.split(
                            "=",
                            2
                    );

            if (parts.length != 2) {
                continue;
            }

            String key =
                    URLDecoder.decode(
                            parts[0],
                            StandardCharsets.UTF_8
                    );

            if (!key.equals(name)) {
                continue;
            }

            return URLDecoder.decode(
                    parts[1],
                    StandardCharsets.UTF_8
            );
        }

        return null;
    }

    // =====================================
    // SAFE VALUE
    // =====================================

    private static String safe(
            String value
    ) {

        if (value == null) {
            return "";
        }

        return value.replace(
                "|",
                "/"
        );
    }

    // =====================================
    // HTTP RESPONSE
    // =====================================

    private static void sendResponse(
            HttpExchange exchange,
            int status,
            String response
    ) throws IOException {

        byte[] bytes =
                response.getBytes(
                        StandardCharsets.UTF_8
                );

        exchange
                .getResponseHeaders()
                .set(
                        "Content-Type",
                        "text/plain; charset=UTF-8"
                );

        exchange.sendResponseHeaders(
                status,
                bytes.length
        );

        try (
                OutputStream output =
                        exchange.getResponseBody()
        ) {

            output.write(
                    bytes
            );
        }
    }
}