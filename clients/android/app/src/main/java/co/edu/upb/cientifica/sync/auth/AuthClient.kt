package co.edu.upb.cientifica.sync.auth

import android.content.Context
import co.edu.upb.cientifica.sync.R
import java.security.KeyStore
import java.security.cert.CertificateFactory
import javax.net.ssl.HttpsURLConnection
import javax.net.ssl.SSLContext
import javax.net.ssl.TrustManagerFactory
import javax.net.ssl.X509TrustManager
import java.net.URL
import java.net.URLEncoder

data class LoginResult(
    val success: Boolean,
    val userId: String,
    val role: String,
    val token: String,
    val message: String
)

class AuthClient(
    context: Context,
    private val baseUrl: String
) {

    private val sslContext:
        SSLContext

    init {

        val certificateFactory =
            CertificateFactory
                .getInstance(
                    "X.509"
                )

        val caCertificate =
            context.resources
                .openRawResource(
                    R.raw.upb_dev_ca
                )
                .use {
                    input ->

                    certificateFactory
                        .generateCertificate(
                            input
                        )
                }

        val keyStore =
            KeyStore
                .getInstance(
                    KeyStore
                        .getDefaultType()
                )

        keyStore.load(
            null,
            null
        )

        keyStore.setCertificateEntry(
            "upb-dev-ca",
            caCertificate
        )

        val trustManagerFactory =
            TrustManagerFactory
                .getInstance(
                    TrustManagerFactory
                        .getDefaultAlgorithm()
                )

        trustManagerFactory.init(
            keyStore
        )

        val trustManager =
            trustManagerFactory
                .trustManagers
                .filterIsInstance<
                    X509TrustManager
                >()
                .firstOrNull()
                ?: throw IllegalStateException(
                    "X509TrustManager no disponible"
                )

        sslContext =
            SSLContext
                .getInstance(
                    "TLS"
                )

        sslContext.init(
            null,
            arrayOf(
                trustManager
            ),
            null
        )
    }

    fun login(
        username: String,
        password: String
    ): LoginResult {

        val encodedUsername =
            URLEncoder.encode(
                username,
                "UTF-8"
            )

        val encodedPassword =
            URLEncoder.encode(
                password,
                "UTF-8"
            )

        val endpoint =
            URL(
                "$baseUrl/internal/auth/login"
            )

        val requestBody =
            "username=$encodedUsername" +
                "&password=$encodedPassword"

        val connection =
            endpoint.openConnection()
                as HttpsURLConnection

        connection.sslSocketFactory =
            sslContext.socketFactory

        try {

            connection.requestMethod =
                "POST"

            connection.doOutput =
                true

            connection.setRequestProperty(
                "Content-Type",
                "application/x-www-form-urlencoded"
            )

            connection.connectTimeout =
                5000

            connection.readTimeout =
                5000

            connection.outputStream.use {
                output ->

                output.write(
                    requestBody
                        .toByteArray(
                            Charsets.UTF_8
                        )
                )
            }

            val status =
                connection.responseCode

            val stream =
                if (status in 200..299) {
                    connection.inputStream
                } else {
                    connection.errorStream
                }

            val response =
                stream
                    ?.bufferedReader()
                    ?.use {
                        it.readText()
                    }
                    ?.trim()
                    ?: ""

            if (response.isBlank()) {

                return LoginResult(
                    success = false,
                    userId = "",
                    role = "",
                    token = "",
                    message =
                        "Auth Service respondió vacío"
                )
            }

            val parts =
                response.split(
                    "|"
                )

            if (parts.size < 5) {

                return LoginResult(
                    success = false,
                    userId = "",
                    role = "",
                    token = "",
                    message =
                        "Formato Auth inválido: $response"
                )
            }

            return LoginResult(
                success =
                    parts[0]
                        .trim()
                        .equals(
                            "true",
                            ignoreCase = true
                        ),

                userId =
                    parts[1].trim(),

                role =
                    parts[2].trim(),

                token =
                    parts[3].trim(),

                message =
                    parts
                        .drop(4)
                        .joinToString("|")
                        .trim()
            )

        } finally {

            connection.disconnect()
        }
    }
}
