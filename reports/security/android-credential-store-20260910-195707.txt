package co.edu.upb.cientifica.sync.config

import android.content.Context

data class StoredCredentials(
    val username: String,
    val password: String
)

class CredentialStore(
    context: Context
) {

    companion object {

        private const val PREFERENCES =
            "upb_sync_credentials"

        private const val KEY_USERNAME =
            "username"

        private const val KEY_PASSWORD =
            "password"
    }

    private val preferences =
        context.applicationContext
            .getSharedPreferences(
                PREFERENCES,
                Context.MODE_PRIVATE
            )

    fun save(
        username: String,
        password: String
    ) {

        require(
            username.isNotBlank()
        ) {
            "El usuario es obligatorio"
        }

        require(
            password.isNotBlank()
        ) {
            "La contraseña es obligatoria"
        }

        preferences
            .edit()
            .putString(
                KEY_USERNAME,
                username.trim()
            )
            .putString(
                KEY_PASSWORD,
                password
            )
            .apply()
    }

    fun load():
        StoredCredentials? {

        val username =
            preferences
                .getString(
                    KEY_USERNAME,
                    null
                )
                ?.trim()

        val password =
            preferences
                .getString(
                    KEY_PASSWORD,
                    null
                )

        if (
            username.isNullOrBlank() ||
            password.isNullOrBlank()
        ) {

            return null
        }

        return StoredCredentials(
            username =
                username,

            password =
                password
        )
    }

    fun clear() {

        preferences
            .edit()
            .remove(
                KEY_USERNAME
            )
            .remove(
                KEY_PASSWORD
            )
            .apply()
    }
}
