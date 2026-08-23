package auth;

import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;

public class AuthClient {

    private static final String HOST = "localhost";
    private static final int PORT = 1099;
    private static final String SERVICE_NAME = "AuthService";

    public static void main(String[] args) {

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
                    " 1. LOGIN"
            );

            System.out.println(
                    "==================================="
            );

            AuthResult loginResult =
                    authService.login(
                            "samuel",
                            "123456"
                    );

            System.out.println(
                    "Success: "
                            + loginResult.isSuccess()
            );

            System.out.println(
                    "Mensaje: "
                            + loginResult.getMessage()
            );

            if (!loginResult.isSuccess()) {
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

            System.out.println(
                    "Token: "
                            + loginResult.getToken()
            );

            String token =
                    loginResult.getToken();

            String userId =
                    loginResult.getUserId();

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

        } catch (Exception e) {

            System.err.println(
                    "Error en Auth Client:"
            );

            e.printStackTrace();
        }
    }
}