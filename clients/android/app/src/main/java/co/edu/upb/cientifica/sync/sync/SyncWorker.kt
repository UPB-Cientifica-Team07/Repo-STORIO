package co.edu.upb.cientifica.sync.sync

import android.content.Context
import android.util.Log
import androidx.work.Worker
import androidx.work.WorkerParameters

class SyncWorker(
    appContext: Context,
    workerParams: WorkerParameters
) :
    Worker(
        appContext,
        workerParams
    ) {

    companion object {

        const val TAG =
            "UPB-SYNC-WORKER"

        private const val PREFERENCES =
            "upb_background_sync"

        private const val KEY_LAST_TIME =
            "last_time"

        private const val KEY_LAST_RESULT =
            "last_result"

        private const val KEY_CHANGES =
            "changes"

        private const val KEY_DOWNLOADED =
            "downloaded"

        private const val KEY_DELETED =
            "deleted"

        private const val KEY_ERRORS =
            "errors"
    }

    override fun doWork():
        Result {

        Log.i(
            TAG,
            "WorkManager inició File Sync"
        )

        return try {

            val engine =
                SyncEngine(
                    applicationContext
                )

            val sync =
                engine.synchronize()

            saveResult(
                sync
            )

            if (
                sync.skipped
            ) {

                Log.i(
                    TAG,
                    "Sync omitida: ${sync.message}"
                )

                /*
                 * No hacemos retry si está conectado
                 * por datos móviles o fuera de la LAN.
                 *
                 * El siguiente trabajo periódico
                 * volverá a comprobar la política.
                 */
                Result.success()

            } else if (
                sync.errors == 0
            ) {

                Log.i(
                    TAG,
                    "Sync OK | cambios=${sync.changes} " +
                        "descargados=${sync.downloaded} " +
                        "eliminados=${sync.deleted}"
                )

                Result.success()

            } else {

                Log.w(
                    TAG,
                    "Sync con errores=${sync.errors}"
                )

                Result.retry()
            }

        } catch (
            error: Exception
        ) {

            Log.e(
                TAG,
                "Error en WorkManager Sync",
                error
            )

            saveFailure(
                error
            )

            Result.retry()
        }
    }

    private fun saveResult(
        sync: SyncRunResult
    ) {

        applicationContext
            .getSharedPreferences(
                PREFERENCES,
                Context.MODE_PRIVATE
            )
            .edit()
            .putLong(
                KEY_LAST_TIME,
                System.currentTimeMillis()
            )
            .putString(
                KEY_LAST_RESULT,
                if (
                    sync.skipped
                ) {
                    "SKIPPED: ${sync.message}"
                } else if (
                    sync.errors == 0
                ) {
                    "OK"
                } else {
                    "ERRORS"
                }
            )
            .putInt(
                KEY_CHANGES,
                sync.changes
            )
            .putInt(
                KEY_DOWNLOADED,
                sync.downloaded
            )
            .putInt(
                KEY_DELETED,
                sync.deleted
            )
            .putInt(
                KEY_ERRORS,
                sync.errors
            )
            .apply()
    }

    private fun saveFailure(
        error: Exception
    ) {

        applicationContext
            .getSharedPreferences(
                PREFERENCES,
                Context.MODE_PRIVATE
            )
            .edit()
            .putLong(
                KEY_LAST_TIME,
                System.currentTimeMillis()
            )
            .putString(
                KEY_LAST_RESULT,
                "FAILED: ${error.message}"
            )
            .apply()
    }
}
