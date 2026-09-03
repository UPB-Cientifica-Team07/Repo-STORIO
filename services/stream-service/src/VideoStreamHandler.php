<?php

declare(strict_types=1);

final class VideoStreamHandler
{
    private VideoRepository $repository;

    private AuthClient $authClient;

    private string $storageRoot;

    public function __construct(
        VideoRepository $repository,
        AuthClient $authClient,
        string $storageRoot
    ) {
        $this->repository =
            $repository;

        $this->authClient =
            $authClient;

        $root =
            realpath(
                $storageRoot
            );

        if ($root === false) {
            throw new RuntimeException(
                'Shared Storage inválido'
            );
        }

        $this->storageRoot =
            rtrim(
                $root,
                DIRECTORY_SEPARATOR
            );
    }

    // =====================================
    // STREAM
    // =====================================

    public function handle(
        string $videoId
    ): void {
        $identity =
            $this->authenticate();

        $video =
            $this->repository
                ->findById(
                    $videoId
                );

        if ($video === null) {
            $this->error(
                404,
                'VIDEO_NOT_FOUND',
                'El video solicitado no existe'
            );

            return;
        }

        $authorized =
            $identity['role'] === 'ADMIN' ||
            $video['owner_id'] ===
                $identity['userId'];

        if (!$authorized) {
            $authorized =
                $this->repository
                    ->canRead(
                        (string)$video[
                            'id_archivo'
                        ],
                        (string)$identity[
                            'userId'
                        ]
                    );
        }

        if (!$authorized) {
            $this->error(
                403,
                'ACCESS_DENIED',
                'El usuario no tiene permiso READ sobre este video'
            );

            return;
        }

        try {
            $path =
                $this->resolvePhysicalPath(
                    (string)$video[
                        'ruta_relativa'
                    ]
                );
        } catch (Throwable $error) {
            $this->error(
                404,
                'FILE_NOT_FOUND',
                'El archivo físico del video no existe'
            );

            return;
        }

        $this->sendFile(
            $path,
            (string)(
                $video['mime_type']
                ?: 'application/octet-stream'
            )
        );
    }

    // =====================================
    // AUTH
    // =====================================

    private function authenticate(): array
    {
        $header =
            $_SERVER[
                'HTTP_AUTHORIZATION'
            ]
            ?? '';

        if (
            !preg_match(
                '/^Bearer\s+(.+)$/i',
                trim($header),
                $matches
            )
        ) {
            $this->error(
                401,
                'UNAUTHORIZED',
                'Bearer token obligatorio'
            );

            exit;
        }

        try {
            $identity =
                $this->authClient
                    ->validateToken(
                        trim(
                            $matches[1]
                        )
                    );
        } catch (Throwable $error) {
            $this->error(
                503,
                'AUTH_UNAVAILABLE',
                'No fue posible validar la autenticación'
            );

            exit;
        }

        if (
            !$identity['valid'] ||
            $identity['userId'] === ''
        ) {
            $this->error(
                401,
                'UNAUTHORIZED',
                'Token inválido'
            );

            exit;
        }

        return $identity;
    }

    // =====================================
    // PATH CONFINEMENT
    // =====================================

    private function resolvePhysicalPath(
        string $relativePath
    ): string {
        $relativePath =
            str_replace(
                '\\',
                '/',
                trim($relativePath)
            );

        if (
            $relativePath === '' ||
            str_starts_with(
                $relativePath,
                '/'
            ) ||
            str_contains(
                $relativePath,
                '../'
            )
        ) {
            throw new RuntimeException(
                'Ruta inválida'
            );
        }

        $candidate =
            $this->storageRoot .
            DIRECTORY_SEPARATOR .
            str_replace(
                '/',
                DIRECTORY_SEPARATOR,
                $relativePath
            );

        $real =
            realpath(
                $candidate
            );

        if (
            $real === false ||
            !is_file($real)
        ) {
            throw new RuntimeException(
                'Archivo inexistente'
            );
        }

        if (
            !str_starts_with(
                $real,
                $this->storageRoot .
                DIRECTORY_SEPARATOR
            )
        ) {
            throw new RuntimeException(
                'Archivo fuera de Shared Storage'
            );
        }

        return $real;
    }

    // =====================================
    // HTTP RANGE
    // =====================================

    private function sendFile(
        string $path,
        string $mimeType
    ): void {
        $size =
            filesize(
                $path
            );

        if ($size === false) {
            $this->error(
                500,
                'STREAM_ERROR',
                'No fue posible determinar el tamaño del video'
            );

            return;
        }

        $start = 0;
        $end = $size - 1;

        $range =
            $_SERVER[
                'HTTP_RANGE'
            ]
            ?? '';

        if ($range !== '') {
            if (
                !preg_match(
                    '/^bytes=(\d*)-(\d*)$/',
                    trim($range),
                    $matches
                )
            ) {
                $this->rangeNotSatisfiable(
                    $size
                );

                return;
            }

            $rawStart =
                $matches[1];

            $rawEnd =
                $matches[2];

            // bytes=-500
            if (
                $rawStart === '' &&
                $rawEnd !== ''
            ) {
                $suffix =
                    (int)$rawEnd;

                if ($suffix <= 0) {
                    $this->rangeNotSatisfiable(
                        $size
                    );

                    return;
                }

                $suffix =
                    min(
                        $suffix,
                        $size
                    );

                $start =
                    $size -
                    $suffix;

                $end =
                    $size - 1;

            } else {

                if ($rawStart === '') {
                    $this->rangeNotSatisfiable(
                        $size
                    );

                    return;
                }

                $start =
                    (int)$rawStart;

                if ($rawEnd !== '') {
                    $end =
                        (int)$rawEnd;
                }
            }

            if (
                $start < 0 ||
                $start >= $size ||
                $end < $start
            ) {
                $this->rangeNotSatisfiable(
                    $size
                );

                return;
            }

            if ($end >= $size) {
                $end =
                    $size - 1;
            }

            http_response_code(
                206
            );

            header(
                sprintf(
                    'Content-Range: bytes %d-%d/%d',
                    $start,
                    $end,
                    $size
                )
            );
        } else {
            http_response_code(
                200
            );
        }

        $length =
            $end -
            $start +
            1;

        header(
            'Content-Type: ' .
            $mimeType
        );

        header(
            'Accept-Ranges: bytes'
        );

        header(
            'Content-Length: ' .
            $length
        );

        header(
            'Cache-Control: private, no-store'
        );

        $handle =
            fopen(
                $path,
                'rb'
            );

        if ($handle === false) {
            $this->error(
                500,
                'STREAM_ERROR',
                'No fue posible abrir el video'
            );

            return;
        }

        if ($start > 0) {
            fseek(
                $handle,
                $start
            );
        }

        $remaining =
            $length;

        $chunkSize =
            1024 * 1024;

        while (
            $remaining > 0 &&
            !feof($handle)
        ) {
            $readSize =
                min(
                    $chunkSize,
                    $remaining
                );

            $data =
                fread(
                    $handle,
                    $readSize
                );

            if (
                $data === false ||
                $data === ''
            ) {
                break;
            }

            echo $data;

            $remaining -=
                strlen(
                    $data
                );

            if (
                function_exists(
                    'fastcgi_finish_request'
                )
            ) {
                // No se invoca aquí porque
                // terminaría la respuesta.
            }

            flush();
        }

        fclose(
            $handle
        );
    }

    private function rangeNotSatisfiable(
        int $size
    ): void {
        http_response_code(
            416
        );

        header(
            'Content-Range: bytes */' .
            $size
        );

        header(
            'Content-Type: application/json; charset=utf-8'
        );

        echo json_encode(
            [
                'success' =>
                    false,

                'code' =>
                    'INVALID_RANGE',

                'message' =>
                    'Rango solicitado no válido',
            ],
            JSON_UNESCAPED_UNICODE
        );
    }

    // =====================================
    // HTTP ERROR
    // =====================================

    private function error(
        int $status,
        string $code,
        string $message
    ): void {
        http_response_code(
            $status
        );

        header(
            'Content-Type: application/json; charset=utf-8'
        );

        echo json_encode(
            [
                'success' =>
                    false,

                'code' =>
                    $code,

                'message' =>
                    $message,
            ],
            JSON_UNESCAPED_UNICODE
        );
    }
}
