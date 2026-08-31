package co.edu.upb.cientifica.sync.sync

import android.content.Context
import java.util.UUID

object DeviceIdentity {

    private const val PREFERENCES =
        "upb_sync_preferences"

    private const val DEVICE_ID_KEY =
        "device_id"

    fun getOrCreate(
        context: Context
    ): String {

        val preferences =
            context.getSharedPreferences(
                PREFERENCES,
                Context.MODE_PRIVATE
            )

        val existing =
            preferences.getString(
                DEVICE_ID_KEY,
                null
            )

        if (
            !existing.isNullOrBlank()
        ) {

            return existing
        }

        val generated =
            "android-" +
                UUID
                    .randomUUID()
                    .toString()
                    .take(8)

        preferences
            .edit()
            .putString(
                DEVICE_ID_KEY,
                generated
            )
            .apply()

        return generated
    }
}
