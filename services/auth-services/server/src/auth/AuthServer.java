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

public class AuthServer {

    private static final int RMI_PORT = 1099;
    private static final int HTTP_PORT = 8081;

    private static final String SERVICE_NAME = "AuthService";

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
                    .equalsIgnoreCase("GET")) {

                sendResponse(
                        exchange,
                        405,
                        "false||||METHOD_NOT_ALLOWED"
                );

                return;
            }

            String query =
                    exchange
                            .getRequestURI()
                            .getRawQuery();

            String username =
                    getQueryParameter(
                            query,
                            "username"
                    );

            String password =
                    getQueryParameter(
                            query,
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

            AuthResult result =
                    authService.login(
                            username,
                            password
                    );

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

            String query =
                    exchange
                            .getRequestURI()
                            .getRawQuery();

            String token =
                    getQueryParameter(
                            query,
                            "token"
                    );

            if (token == null ||
                    token.isBlank()) {

                sendResponse(
                        exchange,
                        400,
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
    // QUERY PARAMETER
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