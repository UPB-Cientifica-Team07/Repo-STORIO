package co.edu.upb.cientifica.sync.config

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64

import java.security.KeyStore

import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

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

        /*
         * Nombre utilizado por versiones anteriores.
         *
         * Solo se conserva para migrar instalaciones
         * que todavía tengan la contraseña en claro.
         */
        private const val LEGACY_KEY_PASSWORD =
            "password"

        private const val KEY_PASSWORD_CIPHERTEXT =
            "password_ciphertext"

        private const val KEY_PASSWORD_IV =
            "password_iv"

        private const val KEYSTORE_PROVIDER =
            "AndroidKeyStore"

        private const val KEY_ALIAS =
            "upb_sync_credentials_aes"

        private const val TRANSFORMATION =
            "AES/GCM/NoPadding"

        private const val GCM_TAG_LENGTH_BITS =
            128
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

        val encrypted =
            encryptPassword(
                password
            )

        preferences
            .edit()
            .putString(
                KEY_USERNAME,
                username.trim()
            )
            .putString(
                KEY_PASSWORD_CIPHERTEXT,
                encrypted.ciphertext
            )
            .putString(
                KEY_PASSWORD_IV,
                encrypted.iv
            )
            /*
             * Si existe una instalación anterior,
             * eliminamos cualquier contraseña plaintext.
             */
            .remove(
                LEGACY_KEY_PASSWORD
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

        if (
            username.isNullOrBlank()
        ) {
            return null
        }

        val ciphertext =
            preferences
                .getString(
                    KEY_PASSWORD_CIPHERTEXT,
                    null
                )

        val iv =
            preferences
                .getString(
                    KEY_PASSWORD_IV,
                    null
                )

        if (
            !ciphertext.isNullOrBlank() &&
            !iv.isNullOrBlank()
        ) {

            return try {

                StoredCredentials(
                    username =
                        username,

                    password =
                        decryptPassword(
                            ciphertext,
                            iv
                        )
                )

            } catch (
                error: Exception
            ) {

                /*
                 * Puede ocurrir si Android restauró
                 * SharedPreferences en un dispositivo
                 * donde la clave Keystore no existe.
                 *
                 * Nunca degradamos a almacenamiento
                 * plaintext.
                 */
                clearEncryptedPassword()

                null
            }
        }

        /*
         * Migración única desde instalaciones antiguas.
         *
         * Se lee la contraseña plaintext, se cifra con
         * Android Keystore y posteriormente se elimina
         * la entrada antigua.
         */
        val legacyPassword =
            preferences
                .getString(
                    LEGACY_KEY_PASSWORD,
                    null
                )

        if (
            legacyPassword.isNullOrBlank()
        ) {
            return null
        }

        save(
            username,
            legacyPassword
        )

        return StoredCredentials(
            username =
                username,

            password =
                legacyPassword
        )
    }

    fun clear() {

        preferences
            .edit()
            .remove(
                KEY_USERNAME
            )
            .remove(
                LEGACY_KEY_PASSWORD
            )
            .remove(
                KEY_PASSWORD_CIPHERTEXT
            )
            .remove(
                KEY_PASSWORD_IV
            )
            .apply()
    }

    private fun encryptPassword(
        password: String
    ): EncryptedPassword {

        val cipher =
            Cipher.getInstance(
                TRANSFORMATION
            )

        cipher.init(
            Cipher.ENCRYPT_MODE,
            getOrCreateSecretKey()
        )

        val ciphertext =
            cipher.doFinal(
                password.toByteArray(
                    Charsets.UTF_8
                )
            )

        return EncryptedPassword(
            ciphertext =
                Base64.encodeToString(
                    ciphertext,
                    Base64.NO_WRAP
                ),

            iv =
                Base64.encodeToString(
                    cipher.iv,
                    Base64.NO_WRAP
                )
        )
    }

    private fun decryptPassword(
        ciphertextBase64: String,
        ivBase64: String
    ): String {

        val ciphertext =
            Base64.decode(
                ciphertextBase64,
                Base64.NO_WRAP
            )

        val iv =
            Base64.decode(
                ivBase64,
                Base64.NO_WRAP
            )

        val cipher =
            Cipher.getInstance(
                TRANSFORMATION
            )

        cipher.init(
            Cipher.DECRYPT_MODE,
            getOrCreateSecretKey(),
            GCMParameterSpec(
                GCM_TAG_LENGTH_BITS,
                iv
            )
        )

        return String(
            cipher.doFinal(
                ciphertext
            ),
            Charsets.UTF_8
        )
    }

    private fun getOrCreateSecretKey():
        SecretKey {

        val keyStore =
            KeyStore.getInstance(
                KEYSTORE_PROVIDER
            )

        keyStore.load(
            null
        )

        val existingKey =
            keyStore.getKey(
                KEY_ALIAS,
                null
            )

        if (
            existingKey is SecretKey
        ) {
            return existingKey
        }

        val keyGenerator =
            KeyGenerator.getInstance(
                KeyProperties.KEY_ALGORITHM_AES,
                KEYSTORE_PROVIDER
            )

        val specification =
            KeyGenParameterSpec.Builder(
                KEY_ALIAS,
                KeyProperties.PURPOSE_ENCRYPT or
                    KeyProperties.PURPOSE_DECRYPT
            )
                .setBlockModes(
                    KeyProperties.BLOCK_MODE_GCM
                )
                .setEncryptionPaddings(
                    KeyProperties.ENCRYPTION_PADDING_NONE
                )
                .setKeySize(
                    256
                )
                .setRandomizedEncryptionRequired(
                    true
                )
                .build()

        keyGenerator.init(
            specification
        )

        return keyGenerator.generateKey()
    }

    private fun clearEncryptedPassword() {

        preferences
            .edit()
            .remove(
                KEY_PASSWORD_CIPHERTEXT
            )
            .remove(
                KEY_PASSWORD_IV
            )
            .remove(
                LEGACY_KEY_PASSWORD
            )
            .apply()
    }

    private data class EncryptedPassword(
        val ciphertext: String,
        val iv: String
    )
}
