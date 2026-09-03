package auth;

import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;

public class AuthClient {

    private static final String HOST =
            System.getenv()
                    .getOrDefault(
                            "AUTH_RMI_HOST",
                            "localhost"
                    );

    private static final int PORT =
            Integer.parseInt(
                    System.getenv()
                            .getOrDefault(
                                    "AUTH_RMI_PORT",
                                    "1099"
                            )
            );

    private static final String SERVICE_NAME =
            "AuthService";

    public static void main(
            String[] args
    ) {

        String username =
                args.length > 0
                        ? args[0]
                        : System.getenv(
                                "AUTH_TEST_USER"
                        );

        String password =
                args.length > 1
                        ? args[1]
                        : System.getenv(
                                "AUTH_TEST_PASSWORD"
                        );

        if (
                username == null ||
                username.isBlank() ||
                password == null ||
                password.isBlank()
        ) {

            System.err.println(
                    "Uso:"
            );

            System.err.println(
                    "java auth.AuthClient <usuario> <password>"
            );

            System.err.println(
                    "o variables AUTH_TEST_USER / AUTH_TEST_PASSWORD"
            );

            System.exit(1);
        }

        try {

            Registry registry =
                    LocateRegistry.getRegistry(
                            HOST,
                            PORT
                    );

            IAuthService authService =
                    (IAuthService) registry.lookup(
                            SERVICE_NAME
                    );

            System.out.println(
                    "==================================="
            );
            System.out.println(
                    " 1. LOGIN RMI -> DIRECTORY SERVICE"
            );
            System.out.println(
                    "==================================="
            );

            AuthResult loginResult =
                    authService.login(
                            username,
                            password
                    );

            System.out.println(
                    "Success: "
                            + loginResult.isSuccess()
            );

            System.out.println(
                    "Mensaje: "
                            + loginResult.getMessage()
            );

            if (
                    !loginResult.isSuccess()
            ) {
                return;
            }

            System.out.println(
                    "User ID: "
                            + loginResult.getUserId()
            );

            System.out.println(
                    "Usuario: "
                            + loginResult.getUsername()
            );

            System.out.println(
                    "Rol: "
                            + loginResult.getRole()
            );

            String token =
                    loginResult.getToken();

            String userId =
                    loginResult.getUserId();

            System.out.println(
                    "Token generado: "
                            + (
                                    token != null &&
                                    !token.isBlank()
                            )
            );

            System.out.println(
                    "==================================="
            );
            System.out.println(
                    " 2. VALIDAR TOKEN"
            );
            System.out.println(
                    "==================================="
            );

            TokenResult tokenResult =
                    authService.validateToken(
                            token
                    );

            System.out.println(
                    "Válido: "
                            + tokenResult.isValid()
            );

            System.out.println(
                    "Mensaje: "
                            + tokenResult.getMessage()
            );

            System.out.println(
                    "User ID: "
                            + tokenResult.getUserId()
            );

            System.out.println(
                    "Rol: "
                            + tokenResult.getRole()
            );

            System.out.println(
                    "==================================="
            );
            System.out.println(
                    " 3. CONSULTAR USUARIO"
            );
            System.out.println(
                    "==================================="
            );

            AuthResult userResult =
                    authService.getUser(
                            userId
                    );

            System.out.println(
                    "Success: "
                            + userResult.isSuccess()
            );

            System.out.println(
                    "Mensaje: "
                            + userResult.getMessage()
            );

            System.out.println(
                    "Usuario: "
                            + userResult.getUsername()
            );

            System.out.println(
                    "Rol: "
                            + userResult.getRole()
            );

            System.out.println(
                    "==================================="
            );
            System.out.println(
                    " PRUEBA RMI COMPLETADA"
            );
            System.out.println(
                    "==================================="
            );

        } catch (
                Exception error
        ) {

            System.err.println(
                    "Error en Auth Client:"
            );

            error.printStackTrace();
        }
    }
}
