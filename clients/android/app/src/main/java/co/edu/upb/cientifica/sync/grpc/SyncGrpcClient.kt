package co.edu.upb.cientifica.sync.grpc

import co.edu.upb.cientifica.sync.proto.AcknowledgeChangesRequest
import co.edu.upb.cientifica.sync.proto.AcknowledgeChangesResponse
import co.edu.upb.cientifica.sync.proto.AuthenticateRequest
import co.edu.upb.cientifica.sync.proto.AuthenticateResponse
import co.edu.upb.cientifica.sync.proto.DownloadRequest
import co.edu.upb.cientifica.sync.proto.DownloadResponse
import co.edu.upb.cientifica.sync.proto.FileChange
import co.edu.upb.cientifica.sync.proto.FileMetadata
import co.edu.upb.cientifica.sync.proto.SyncRequest
import co.edu.upb.cientifica.sync.proto.SyncServiceGrpc
import co.edu.upb.cientifica.sync.proto.UploadMetadata
import co.edu.upb.cientifica.sync.proto.UploadRequest
import co.edu.upb.cientifica.sync.proto.UploadResponse
import com.google.protobuf.ByteString
import io.grpc.ManagedChannel
import io.grpc.Metadata
import io.grpc.okhttp.OkHttpChannelBuilder
import io.grpc.stub.MetadataUtils
import io.grpc.stub.StreamObserver
import java.io.ByteArrayOutputStream
import java.util.concurrent.CountDownLatch
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicReference

data class UploadResult(
    val success: Boolean,
    val message: String,
    val fileId: String,
    val bytesReceived: Long
)

data class DownloadedFile(
    val fileId: String,
    val relativePath: String,
    val fileName: String,
    val version: Long,
    val content: ByteArray
)

class SyncGrpcClient(
    host: String,
    port: Int
) {

    companion object {

        private val AUTHORIZATION_KEY =
            Metadata.Key.of(
                "authorization",
                Metadata.ASCII_STRING_MARSHALLER
            )
    }

    private val channel: ManagedChannel =
        OkHttpChannelBuilder
            .forAddress(
                host,
                port
            )
            .usePlaintext()
            .build()

    private val blockingStub =
        SyncServiceGrpc
            .newBlockingStub(
                channel
            )

    fun authenticate(
        token: String,
        deviceId: String
    ): AuthenticateResponse {

        val request =
            AuthenticateRequest
                .newBuilder()
                .setToken(token)
                .setDeviceId(deviceId)
                .build()

        return blockingStub
            .withDeadlineAfter(
                5,
                TimeUnit.SECONDS
            )
            .authenticate(
                request
            )
    }

    private fun authenticatedBlockingStub(
        token: String
    ): SyncServiceGrpc.SyncServiceBlockingStub {

        val headers =
            Metadata().apply {

                put(
                    AUTHORIZATION_KEY,
                    "Bearer $token"
                )
            }

        return SyncServiceGrpc
            .newBlockingStub(
                channel
            )
            .withInterceptors(
                MetadataUtils
                    .newAttachHeadersInterceptor(
                        headers
                    )
            )
    }

    fun sync(
        token: String,
        userId: String,
        deviceId: String
    ): List<FileChange> {

        val request =
            SyncRequest
                .newBuilder()
                .setUserId(
                    userId
                )
                .setDeviceId(
                    deviceId
                )
                .build()

        val response =
            authenticatedBlockingStub(
                token
            )
                .withDeadlineAfter(
                    10,
                    TimeUnit.SECONDS
                )
                .sync(
                    request
                )

        if (!response.success) {

            throw RuntimeException(
                "Sync respondió false: ${response.message}"
            )
        }

        return response.changesList
    }

    fun download(
        token: String,
        userId: String,
        fileId: String
    ): DownloadedFile {

        val request =
            DownloadRequest
                .newBuilder()
                .setUserId(
                    userId
                )
                .setFileId(
                    fileId
                )
                .build()

        val iterator =
            authenticatedBlockingStub(
                token
            )
                .withDeadlineAfter(
                    20,
                    TimeUnit.SECONDS
                )
                .download(
                    request
                )

        var metadata:
            FileMetadata? =
            null

        val output =
            ByteArrayOutputStream()

        while (
            iterator.hasNext()
        ) {

            val response =
                iterator.next()

            when (
                response.dataCase
            ) {

                DownloadResponse.DataCase.METADATA -> {

                    metadata =
                        response.metadata
                }

                DownloadResponse.DataCase.CHUNK -> {

                    output.write(
                        response.chunk
                            .toByteArray()
                    )
                }

                else -> Unit
            }
        }

        val fileMetadata =
            metadata
                ?: throw RuntimeException(
                    "Download no devolvió metadata"
                )

        return DownloadedFile(
            fileId =
                fileMetadata.fileId,

            relativePath =
                fileMetadata.relativePath,

            fileName =
                fileMetadata.fileName,

            version =
                fileMetadata.version,

            content =
                output.toByteArray()
        )
    }

    fun upload(
        token: String,
        userId: String,
        deviceId: String,
        fileName: String,
        fileType: String,
        relativePath: String,
        content: ByteArray
    ): UploadResult {

        val headers =
            Metadata().apply {

                put(
                    AUTHORIZATION_KEY,
                    "Bearer $token"
                )
            }

        val asyncStub =
            SyncServiceGrpc
                .newStub(
                    channel
                )
                .withInterceptors(
                    MetadataUtils
                        .newAttachHeadersInterceptor(
                            headers
                        )
                )

        val latch =
            CountDownLatch(1)

        val responseRef =
            AtomicReference<
                UploadResponse?
            >(null)

        val errorRef =
            AtomicReference<
                Throwable?
            >(null)

        val responseObserver =
            object :
                StreamObserver<
                    UploadResponse
                > {

                override fun onNext(
                    value: UploadResponse
                ) {

                    responseRef.set(
                        value
                    )
                }

                override fun onError(
                    error: Throwable
                ) {

                    errorRef.set(
                        error
                    )

                    latch.countDown()
                }

                override fun onCompleted() {

                    latch.countDown()
                }
            }

        val requestObserver =
            asyncStub
                .withDeadlineAfter(
                    20,
                    TimeUnit.SECONDS
                )
                .upload(
                    responseObserver
                )

        try {

            val metadata =
                UploadMetadata
                    .newBuilder()
                    .setUserId(
                        userId
                    )
                    .setDeviceId(
                        deviceId
                    )
                    .setFileName(
                        fileName
                    )
                    .setFileType(
                        fileType
                    )
                    .setSize(
                        content.size.toLong()
                    )
                    .setRelativePath(
                        relativePath
                    )
                    .build()

            requestObserver
                .onNext(
                    UploadRequest
                        .newBuilder()
                        .setMetadata(
                            metadata
                        )
                        .build()
                )

            val chunkSize =
                64 * 1024

            var offset =
                0

            while (
                offset <
                content.size
            ) {

                val end =
                    minOf(
                        offset +
                            chunkSize,
                        content.size
                    )

                val chunk =
                    ByteString.copyFrom(
                        content,
                        offset,
                        end - offset
                    )

                requestObserver
                    .onNext(
                        UploadRequest
                            .newBuilder()
                            .setChunk(
                                chunk
                            )
                            .build()
                    )

                offset =
                    end
            }

            requestObserver
                .onCompleted()

        } catch (
            error: Throwable
        ) {

            requestObserver
                .onError(
                    error
                )

            throw error
        }

        val completed =
            latch.await(
                25,
                TimeUnit.SECONDS
            )

        if (!completed) {

            throw RuntimeException(
                "Timeout esperando respuesta de Upload"
            )
        }

        errorRef
            .get()
            ?.let {

                throw RuntimeException(
                    "Upload gRPC falló: ${it.message}",
                    it
                )
            }

        val response =
            responseRef.get()
                ?: throw RuntimeException(
                    "Upload no devolvió respuesta"
                )

        return UploadResult(
            success =
                response.success,

            message =
                response.message,

            fileId =
                response.fileId,

            bytesReceived =
                response.bytesReceived
        )
    }

    fun acknowledgeChanges(
        token: String,
        deviceId: String,
        changeId: Long
    ): AcknowledgeChangesResponse {

        val request =
            AcknowledgeChangesRequest
                .newBuilder()
                .setDeviceId(
                    deviceId
                )
                .setChangeId(
                    changeId
                )
                .build()

        val response =
            authenticatedBlockingStub(
                token
            )
                .withDeadlineAfter(
                    10,
                    TimeUnit.SECONDS
                )
                .acknowledgeChanges(
                    request
                )

        if (!response.success) {

            throw RuntimeException(
                "ACK respondió false: ${response.message}"
            )
        }

        return response
    }

    fun close() {

        channel.shutdown()

        try {

            channel.awaitTermination(
                3,
                TimeUnit.SECONDS
            )

        } catch (
            interrupted:
            InterruptedException
        ) {

            Thread
                .currentThread()
                .interrupt()
        }
    }
}
