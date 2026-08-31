package co.edu.upb.cientifica.sync.auth

import java.net.HttpURLConnection
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
    private val baseUrl: String
) {

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
                "$baseUrl/internal/auth/login" +
                    "?username=$encodedUsername" +
                    "&password=$encodedPassword"
            )

        val connection =
            endpoint.openConnection()
                as HttpURLConnection

        try {

            connection.requestMethod =
                "GET"

            connection.connectTimeout =
                5000

            connection.readTimeout =
                5000

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
