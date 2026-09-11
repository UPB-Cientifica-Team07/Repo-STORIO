package hpc.auth;

import java.io.IOException;
import java.io.FileInputStream;
import java.security.KeyStore;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;

import javax.net.ssl.SSLContext;
import javax.net.ssl.TrustManagerFactory;
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

        try {

            String caFile =
                System.getenv()
                    .getOrDefault(
                        "AUTH_TLS_CA_FILE",
                        "security/pki/upb_dev_ca.crt"
                    );

            CertificateFactory
                certificateFactory =
                    CertificateFactory.getInstance(
                        "X.509"
                    );

            X509Certificate caCertificate;

            try (
                FileInputStream input =
                    new FileInputStream(
                        caFile
                    )
            ) {

                caCertificate =
                    (X509Certificate)
                        certificateFactory
                            .generateCertificate(
                                input
                            );
            }

            KeyStore trustStore =
                KeyStore.getInstance(
                    KeyStore
                        .getDefaultType()
                );

            trustStore.load(
                null,
                null
            );

            trustStore.setCertificateEntry(
                "upb-dev-ca",
                caCertificate
            );

            TrustManagerFactory
                trustManagerFactory =
                    TrustManagerFactory
                        .getInstance(
                            TrustManagerFactory
                                .getDefaultAlgorithm()
                        );

            trustManagerFactory.init(
                trustStore
            );

            SSLContext sslContext =
                SSLContext.getInstance(
                    "TLS"
                );

            sslContext.init(
                null,
                trustManagerFactory
                    .getTrustManagers(),
                null
            );

            this.client =
                HttpClient
                    .newBuilder()
                    .sslContext(
                        sslContext
                    )
                    .connectTimeout(
                        Duration.ofSeconds(
                            3
                        )
                    )
                    .build();

        } catch (Exception error) {

            throw new IllegalStateException(
                "No fue posible configurar TLS para Auth Service",
                error
            );
        }
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
