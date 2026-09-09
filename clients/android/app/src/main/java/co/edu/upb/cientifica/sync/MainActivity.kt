package co.edu.upb.cientifica.sync

import android.os.Bundle
import android.widget.Button
import android.widget.EditText
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import co.edu.upb.cientifica.sync.auth.AuthClient
import co.edu.upb.cientifica.sync.config.CredentialStore
import co.edu.upb.cientifica.sync.grpc.SyncGrpcClient
import co.edu.upb.cientifica.sync.network.LocalNetworkPolicy
import co.edu.upb.cientifica.sync.network.NetworkMonitor
import co.edu.upb.cientifica.sync.sync.DeviceIdentity
import co.edu.upb.cientifica.sync.sync.SyncEngine
import co.edu.upb.cientifica.sync.sync.SyncRunResult
import java.io.File
import java.util.concurrent.Executors
import java.util.concurrent.atomic.AtomicBoolean

class MainActivity :
    AppCompatActivity() {

    companion object {

        private const val TEST_RELATIVE_PATH =
            "Android/Prueba/android-sync.txt"
    }

    private val executor =
        Executors.newSingleThreadExecutor()

    private val automaticSyncRunning =
        AtomicBoolean(false)

    @Volatile
    private var lastNetworkAllowed =
        false

    private lateinit var networkText:
        TextView

    private lateinit var statusText:
        TextView

    private lateinit var usernameInput:
        EditText

    private lateinit var passwordInput:
        EditText

    private lateinit var saveCredentialsButton:
        Button

    private lateinit var authButton:
        Button

    private lateinit var uploadButton:
        Button

    private lateinit var syncButton:
        Button

    private lateinit var networkMonitor:
        NetworkMonitor

    private lateinit var networkPolicy:
        LocalNetworkPolicy

    override fun onCreate(
        savedInstanceState: Bundle?
    ) {

        super.onCreate(
            savedInstanceState
        )

        setContentView(
            R.layout.activity_main
        )

        networkText =
            findViewById(
                R.id.networkText
            )

        statusText =
            findViewById(
                R.id.statusText
            )

        usernameInput =
            findViewById(
                R.id.usernameInput
            )

        passwordInput =
            findViewById(
                R.id.passwordInput
            )

        saveCredentialsButton =
            findViewById(
                R.id.saveCredentialsButton
            )

        authButton =
            findViewById(
                R.id.authButton
            )

        uploadButton =
            findViewById(
                R.id.uploadButton
            )

        syncButton =
            findViewById(
                R.id.syncButton
            )

        networkPolicy =
            LocalNetworkPolicy(
                context =
                    applicationContext,

                serverHost =
                    SyncEngine.SERVER_HOST,

                serverPort =
                    SyncEngine.SYNC_PORT
            )

        networkMonitor =
            NetworkMonitor(
                applicationContext
            ) {

                evaluateNetwork()
            }

        CredentialStore(
            applicationContext
        ).load()
            ?.let { credentials ->

                usernameInput.setText(
                    credentials.username
                )
            }

        saveCredentialsButton
            .setOnClickListener {

                try {

                    CredentialStore(
                        applicationContext
                    ).save(
                        usernameInput
                            .text
                            .toString(),

                        passwordInput
                            .text
                            .toString()
                    )

                    passwordInput.text.clear()

                    updateStatus(
                        "Credenciales guardadas localmente"
                    )

                } catch (
                    error: Exception
                ) {

                    showError(
                        error
                    )
                }
            }

        authButton
            .setOnClickListener {

                executeIfNetworkAllowed {

                    runAuthenticationTest()
                }
            }

        uploadButton
            .setOnClickListener {

                executeIfNetworkAllowed {

                    runUploadTest()
                }
            }

        syncButton
            .setOnClickListener {

                executor.execute {

                    runSync(
                        "MANUAL"
                    )
                }
            }

        evaluateNetwork()
    }

    override fun onStart() {

        super.onStart()

        lastNetworkAllowed =
            false

        networkMonitor.start()

        evaluateNetwork()
    }

    override fun onStop() {

        networkMonitor.stop()

        lastNetworkAllowed =
            false

        super.onStop()
    }

    private fun executeIfNetworkAllowed(
        action: () -> Unit
    ) {

        statusText.text =
            "Estado: verificando red..."

        executor.execute {

            val result =
                networkPolicy.evaluate()

            if (!result.allowed) {

                updateStatus(
                    """
                    OPERACIÓN BLOQUEADA

                    ${result.message}

                    IP:
                    ${result.localIp ?: "sin IP"}
                    """.trimIndent()
                )

                return@execute
            }

            action()
        }
    }

    private fun evaluateNetwork() {

        executor.execute {

            val result =
                networkPolicy.evaluate()

            val becameAllowed =
                result.allowed &&
                    !lastNetworkAllowed

            lastNetworkAllowed =
                result.allowed

            val text =
                """
                Red:
                ${if (result.allowed) "AUTORIZADA" else "NO AUTORIZADA"}

                Wi-Fi:
                ${if (result.wifi) "OK" else "NO"}

                IP local:
                ${result.localIp ?: "No disponible"}

                Subred 192.168.10.0/24:
                ${if (result.subnetOk) "OK" else "NO"}

                Sync :50055:
                ${if (result.serverReachable) "ALCANZABLE" else "NO ALCANZABLE"}

                ${result.message}
                """.trimIndent()

            runOnUiThread {

                networkText.text =
                    text

                authButton.isEnabled =
                    result.allowed

                uploadButton.isEnabled =
                    result.allowed

                syncButton.isEnabled =
                    result.allowed
            }

            if (
                becameAllowed
            ) {

                runAutomaticSync()
            }
        }
    }

    private fun runAutomaticSync() {

        if (
            !automaticSyncRunning
                .compareAndSet(
                    false,
                    true
                )
        ) {

            return
        }

        try {

            updateStatus(
                """
                SINCRONIZACIÓN AUTOMÁTICA

                Red local detectada.

                Consultando cambios pendientes...
                """.trimIndent()
            )

            runSync(
                "AUTOMÁTICA"
            )

        } finally {

            automaticSyncRunning
                .set(
                    false
                )
        }
    }

    private fun runSync(
        mode: String
    ) {

        try {

            val engine =
                SyncEngine(
                    applicationContext
                )

            val result =
                engine.synchronize()

            showSyncResult(
                mode,
                result
            )

        } catch (
            error: Exception
        ) {

            showError(
                error
            )
        }
    }

    private fun showSyncResult(
        mode: String,
        result: SyncRunResult
    ) {

        if (
            result.skipped
        ) {

            updateStatus(
                """
                SYNC $mode: OMITIDA

                ${result.message}
                """.trimIndent()
            )

            return
        }

        updateStatus(
            """
            SYNC $mode: ${if (result.errors == 0) "OK" else "CON ERRORES"}

            Usuario:
            ${result.userId}

            Device:
            ${result.deviceId}

            Cambios recibidos:
            ${result.changes}

            Recursos finales:
            ${result.resources}

            Descargados/actualizados:
            ${result.downloaded}

            Eliminados:
            ${result.deleted}

            Errores:
            ${result.errors}

            Directorio local:
            ${File(filesDir, "sync").absolutePath}
            """.trimIndent()
        )
    }

    private fun loginAndAuthenticate(
        grpcClient: SyncGrpcClient,
        deviceId: String
    ): Triple<
        String,
        String,
        String
    > {

        val authClient =
            AuthClient(
                SyncEngine.AUTH_URL
            )

        val credentials =
            CredentialStore(
                applicationContext
            ).load()
                ?: throw RuntimeException(
                    "Credenciales no configuradas"
                )

        val login =
            authClient.login(
                credentials.username,
                credentials.password
            )

        if (
            !login.success
        ) {

            throw RuntimeException(
                "Auth HTTP falló: ${login.message}"
            )
        }

        val authResponse =
            grpcClient.authenticate(
                login.token,
                deviceId
            )

        if (
            !authResponse.success
        ) {

            throw RuntimeException(
                "gRPC Authenticate falló: ${authResponse.message}"
            )
        }

        return Triple(
            login.token,
            authResponse.userId,
            login.role
        )
    }

    private fun runAuthenticationTest() {

        var grpcClient:
            SyncGrpcClient? =
            null

        try {

            grpcClient =
                SyncGrpcClient(
                    SyncEngine.SERVER_HOST,
                    SyncEngine.SYNC_PORT
                )

            val deviceId =
                DeviceIdentity
                    .getOrCreate(
                        applicationContext
                    )

            val session =
                loginAndAuthenticate(
                    grpcClient,
                    deviceId
                )

            updateStatus(
                """
                AUTH HTTP: OK

                Rol:
                ${session.third}

                Token:
                recibido

                Device ID:
                $deviceId

                gRPC Authenticate:
                OK

                User ID Sync:
                ${session.second}

                Token validado.
                """.trimIndent()
            )

        } catch (
            error: Exception
        ) {

            showError(
                error
            )

        } finally {

            grpcClient?.close()
        }
    }

    private fun runUploadTest() {

        var grpcClient:
            SyncGrpcClient? =
            null

        try {

            grpcClient =
                SyncGrpcClient(
                    SyncEngine.SERVER_HOST,
                    SyncEngine.SYNC_PORT
                )

            val deviceId =
                DeviceIdentity
                    .getOrCreate(
                        applicationContext
                    )

            val session =
                loginAndAuthenticate(
                    grpcClient,
                    deviceId
                )

            val localDirectory =
                File(
                    filesDir,
                    "sync/Android/Prueba"
                )

            if (
                !localDirectory.exists()
            ) {

                localDirectory.mkdirs()
            }

            val localFile =
                File(
                    localDirectory,
                    "android-sync.txt"
                )

            val content =
                """
                Archivo enviado desde Android - SYNC
                Usuario: ${session.second}
                Device: $deviceId
                """.trimIndent()

            localFile.writeText(
                content,
                Charsets.UTF_8
            )

            val upload =
                grpcClient.upload(
                    token =
                        session.first,

                    userId =
                        session.second,

                    deviceId =
                        deviceId,

                    fileName =
                        localFile.name,

                    fileType =
                        "text/plain",

                    relativePath =
                        TEST_RELATIVE_PATH,

                    content =
                        localFile.readBytes()
                )

            if (
                !upload.success
            ) {

                throw RuntimeException(
                    "Upload respondió false: ${upload.message}"
                )
            }

            updateStatus(
                """
                UPLOAD ANDROID: OK

                Ruta lógica:
                $TEST_RELATIVE_PATH

                Usuario:
                ${session.second}

                Device:
                $deviceId

                File ID:
                ${upload.fileId}

                Bytes recibidos:
                ${upload.bytesReceived}

                Mensaje:
                ${upload.message}
                """.trimIndent()
            )

        } catch (
            error: Exception
        ) {

            showError(
                error
            )

        } finally {

            grpcClient?.close()
        }
    }

    private fun showError(
        error: Exception
    ) {

        updateStatus(
            """
            ERROR

            ${error.javaClass.simpleName}

            ${error.message}
            """.trimIndent()
        )
    }

    private fun updateStatus(
        message: String
    ) {

        runOnUiThread {

            statusText.text =
                message
        }
    }

    override fun onDestroy() {

        executor.shutdownNow()

        super.onDestroy()
    }
}
