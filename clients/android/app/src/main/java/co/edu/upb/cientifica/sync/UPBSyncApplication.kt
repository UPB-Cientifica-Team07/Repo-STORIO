package co.edu.upb.cientifica.sync

import android.app.Application
import androidx.work.Constraints
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.ExistingWorkPolicy
import androidx.work.NetworkType
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import co.edu.upb.cientifica.sync.sync.SyncWorker
import java.util.concurrent.TimeUnit

class UPBSyncApplication :
    Application() {

    companion object {

        private const val PERIODIC_WORK_NAME =
            "upb-file-sync-periodic"

        private const val STARTUP_WORK_NAME =
            "upb-file-sync-startup"
    }

    override fun onCreate() {

        super.onCreate()

        schedulePeriodicSync()

        scheduleStartupSync()
    }

    private fun networkConstraints():
        Constraints {

        return Constraints
            .Builder()
            .setRequiredNetworkType(
                NetworkType.CONNECTED
            )
            .build()
    }

    private fun schedulePeriodicSync() {

        val request =
            PeriodicWorkRequestBuilder<
                SyncWorker
            >(
                15,
                TimeUnit.MINUTES
            )
                .setConstraints(
                    networkConstraints()
                )
                .build()

        WorkManager
            .getInstance(
                applicationContext
            )
            .enqueueUniquePeriodicWork(
                PERIODIC_WORK_NAME,
                ExistingPeriodicWorkPolicy.KEEP,
                request
            )
    }

    private fun scheduleStartupSync() {

        val request =
            OneTimeWorkRequestBuilder<
                SyncWorker
            >()
                .setConstraints(
                    networkConstraints()
                )
                .build()

        WorkManager
            .getInstance(
                applicationContext
            )
            .enqueueUniqueWork(
                STARTUP_WORK_NAME,
                ExistingWorkPolicy.REPLACE,
                request
            )
    }
}
