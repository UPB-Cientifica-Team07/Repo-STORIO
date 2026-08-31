package co.edu.upb.cientifica.sync.sync

import android.content.Context
import co.edu.upb.cientifica.sync.auth.AuthClient
import co.edu.upb.cientifica.sync.grpc.SyncGrpcClient
import co.edu.upb.cientifica.sync.network.LocalNetworkPolicy
import java.io.File

data class SyncRunResult(
    val success: Boolean,
    val skipped: Boolean,
    val message: String,
    val userId: String = "",
    val deviceId: String = "",
    val changes: Int = 0,
    val resources: Int = 0,
    val downloaded: Int = 0,
    val deleted: Int = 0,
    val errors: Int = 0
)

class SyncEngine(
    context: Context
) {

    companion object {

        const val SERVER_HOST =
            "192.168.10.13"

        const val AUTH_URL =
            "http://192.168.10.13:8081"

        const val SYNC_PORT =
            50055

        const val USERNAME =
            "tercero"

        const val PASSWORD =
            "123456"
    }

    private val context =
        context.applicationContext

    private val networkPolicy =
        LocalNetworkPolicy(
            context =
                this.context,

            serverHost =
                SERVER_HOST,

            serverPort =
                SYNC_PORT
        )

    fun synchronize():
        SyncRunResult {

        val network =
            networkPolicy.evaluate()

        if (!network.allowed) {

            return SyncRunResult(
                success =
                    true,

                skipped =
                    true,

                message =
                    "Sincronización omitida: red local no autorizada"
            )
        }

        var grpcClient:
            SyncGrpcClient? =
            null

        try {

            val deviceId =
                DeviceIdentity
                    .getOrCreate(
                        context
                    )

            val authClient =
                AuthClient(
                    AUTH_URL
                )

            val login =
                authClient.login(
                    USERNAME,
                    PASSWORD
                )

            if (!login.success) {

                throw RuntimeException(
                    "Auth HTTP falló: ${login.message}"
                )
            }

            grpcClient =
                SyncGrpcClient(
                    SERVER_HOST,
                    SYNC_PORT
                )

            val authentication =
                grpcClient.authenticate(
                    login.token,
                    deviceId
                )

            if (
                !authentication.success
            ) {

                throw RuntimeException(
                    "gRPC Authenticate falló: " +
                        authentication.message
                )
            }

            val userId =
                authentication.userId

            val changes =
                grpcClient.sync(
                    token =
                        login.token,

                    userId =
                        userId,

                    deviceId =
                        deviceId
                )

            /*
             * Un mismo file_id puede tener varios
             * eventos pendientes.
             *
             * Aplicamos únicamente su estado final.
             */
            val latestChanges =
                LinkedHashMap<
                    String,
                    co.edu.upb.cientifica.sync.proto.FileChange
                >()

            for (
                change in changes
            ) {

                latestChanges[
                    change.fileId
                ] = change
            }

            var downloaded =
                0

            var deleted =
                0

            var errors =
                0

            for (
                change in
                latestChanges.values
            ) {

                try {

                    when (
                        change.type
                    ) {

                        "FILE_CREATED",
                        "FILE_CHANGED" -> {

                            val file =
                                grpcClient.download(
                                    token =
                                        login.token,

                                    userId =
                                        userId,

                                    fileId =
                                        change.fileId
                                )

                            writeSyncedFile(
                                relativePath =
                                    file.relativePath,

                                content =
                                    file.content
                            )

                            downloaded++
                        }

                        "FILE_DELETED" -> {

                            deleteSyncedFile(
                                change.relativePath
                            )

                            deleted++
                        }
                    }

                } catch (
                    error: Exception
                ) {

                    errors++
                }
            }

            return SyncRunResult(
                success =
                    errors == 0,

                skipped =
                    false,

                message =
                    if (
                        errors == 0
                    ) {

                        "Sincronización completada correctamente"

                    } else {

                        "Sincronización completada con $errors error(es)"
                    },

                userId =
                    userId,

                deviceId =
                    deviceId,

                changes =
                    changes.size,

                resources =
                    latestChanges.size,

                downloaded =
                    downloaded,

                deleted =
                    deleted,

                errors =
                    errors
            )

        } finally {

            grpcClient?.close()
        }
    }

    private fun writeSyncedFile(
        relativePath: String,
        content: ByteArray
    ) {

        val root =
            File(
                context.filesDir,
                "sync"
            )

        if (
            !root.exists() &&
            !root.mkdirs()
        ) {

            throw RuntimeException(
                "No se pudo crear directorio Sync"
            )
        }

        val canonicalRoot =
            root.canonicalFile

        val destination =
            File(
                canonicalRoot,
                relativePath
            ).canonicalFile

        validateDestination(
            canonicalRoot,
            destination,
            relativePath
        )

        val parent =
            destination.parentFile

        if (
            parent != null &&
            !parent.exists() &&
            !parent.mkdirs()
        ) {

            throw RuntimeException(
                "No se pudo crear directorio para $relativePath"
            )
        }

        destination.writeBytes(
            content
        )
    }

    private fun deleteSyncedFile(
        relativePath: String
    ) {

        val root =
            File(
                context.filesDir,
                "sync"
            ).canonicalFile

        val destination =
            File(
                root,
                relativePath
            ).canonicalFile

        validateDestination(
            root,
            destination,
            relativePath
        )

        if (
            destination.exists() &&
            !destination.delete()
        ) {

            throw RuntimeException(
                "No se pudo eliminar $relativePath"
            )
        }

        removeEmptyParents(
            destination.parentFile,
            root
        )
    }

    private fun validateDestination(
        root: File,
        destination: File,
        relativePath: String
    ) {

        if (
            destination.path !=
                root.path &&
            !destination.path.startsWith(
                root.path +
                    File.separator
            )
        ) {

            throw SecurityException(
                "Ruta fuera del directorio Sync: $relativePath"
            )
        }
    }

    private fun removeEmptyParents(
        start: File?,
        root: File
    ) {

        var current =
            start

        while (
            current != null &&
            current.path != root.path
        ) {

            val children =
                current.listFiles()

            if (
                children == null ||
                children.isNotEmpty()
            ) {

                break
            }

            if (
                !current.delete()
            ) {

                break
            }

            current =
                current.parentFile
        }
    }
}
