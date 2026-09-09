package hpc.auth;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;

public class AuthClient {

    private final String baseUrl;

    private final HttpClient client;

    public AuthClient(
        String baseUrl
    ) {

        this.baseUrl =
            baseUrl.replaceAll(
                "/+$",
                ""
            );

        this.client =
            HttpClient
                .newBuilder()
                .connectTimeout(
                    Duration.ofSeconds(
                        3
                    )
                )
                .build();
    }

    public AuthIdentity validate(
        String token
    ) throws IOException,
        InterruptedException {

        if (
            token == null ||
            token.isBlank()
        ) {

            return new AuthIdentity(
                false,
                "",
                "",
                "Token requerido"
            );
        }

        URI uri =
            URI.create(
                baseUrl +
                "/internal/auth/validate"
            );

        HttpRequest request =
            HttpRequest
                .newBuilder(
                    uri
                )
                .header(
                    "Authorization",
                    "Bearer " + token
                )
                .GET()
                .timeout(
                    Duration.ofSeconds(
                        5
                    )
                )
                .build();

        HttpResponse<String> response =
            client.send(
                request,
                HttpResponse
                    .BodyHandlers
                    .ofString()
            );

        if (
            response.statusCode()
                != 200
        ) {

            throw new IOException(
                "Auth Service respondió HTTP " +
                response.statusCode()
            );
        }

        String[] parts =
            response
                .body()
                .trim()
                .split(
                    "\\|",
                    4
                );

        if (
            parts.length < 3
        ) {

            throw new IOException(
                "Respuesta inválida de Auth Service"
            );
        }

        boolean valid =
            Boolean.parseBoolean(
                parts[0]
            );

        String userId =
            parts.length > 1
                ? parts[1]
                : "";

        String role =
            parts.length > 2
                ? parts[2]
                : "";

        String message =
            parts.length > 3
                ? parts[3]
                : "";

        return new AuthIdentity(
            valid,
            userId,
            role,
            message
        );
    }
}
