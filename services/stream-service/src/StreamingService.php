<?php

declare(strict_types=1);

require_once __DIR__ . '/VideoRepository.php';
require_once __DIR__ . '/AuthClient.php';

final class StreamingService
{
    private VideoRepository $repository;

    private AuthClient $authClient;

    private string $storageRoot;

    public function __construct(
        VideoRepository $repository,
        AuthClient $authClient,
        string $storageRoot
    ) {
        if ($storageRoot === '') {
            throw new RuntimeException(
                'Shared Storage no configurado'
            );
        }

        $this->repository =
            $repository;

        $this->authClient =
            $authClient;

        $this->storageRoot =
            rtrim(
                $storageRoot,
                DIRECTORY_SEPARATOR
            );
    }

    // =====================================
    // REGISTER VIDEO - INTERNO
    // =====================================

    public function registerVideo(
        object $request
    ): array {
        $fileId =
            trim(
                (string)(
                    $request->idArchivo
                    ?? ''
                )
            );

        if ($fileId === '') {
            throw new SoapFault(
                'INVALID_REQUEST',
                'idArchivo es obligatorio'
            );
        }

        $file =
            $this->repository
                ->findFileById(
                    $fileId
                );

        if ($file === null) {
            throw new SoapFault(
                'FILE_NOT_FOUND',
                'El archivo solicitado no existe'
            );
        }

        if (
            strtoupper(
                (string)$file['tipo_archivo']
            ) !== 'VIDEO'
        ) {
            throw new SoapFault(
                'INVALID_FILE_TYPE',
                'El archivo no es de tipo VIDEO'
            );
        }

        $physicalPath =
            $this->physicalPath(
                (string)$file[
                    'ruta_relativa'
                ]
            );

        $metadata =
            $this->probeVideo(
                $physicalPath
            );

        $durationSeconds =
            max(
                0,
                (int)round(
                    (float)(
                        $metadata['duration']
                        ?? 0
                    )
                )
            );

        $height =
            (int)(
                $metadata['height']
                ?? 0
            );

        $quality =
            $this->qualityFromHeight(
                $height
            );

        $format =
            strtoupper(
                (string)(
                    $metadata['format']
                    ?? 'UNKNOWN'
                )
            );

        $video =
            $this->repository
                ->upsertVideo(
                    $fileId,
                    $durationSeconds,
                    $quality,
                    $format
                );

        return [
            'success' =>
                true,

            'message' =>
                'Video registrado correctamente',

            'idVideo' =>
                $video['id_video'],

            'idArchivo' =>
                $video['id_archivo'],

            'duracionSegundos' =>
                (int)$video[
                    'duracion_segundos'
                ],

            'calidad' =>
                $video['calidad'],

            'formato' =>
                $video['formato'],
        ];
    }

    // =====================================
    // LIST VIDEOS
    // =====================================

    public function listVideos(
        object $request
    ): array {
        $identity =
            $this->authenticatedIdentity(
                $request
            );

        $videos =
            $this->repository
                ->listAccessibleByDirectoryId(
                    $identity['userId'],
                    $identity['role'] ===
                        'ADMIN'
                );

        return [
            'success' =>
                true,

            'message' =>
                'Videos autorizados encontrados correctamente',

            'videos' =>
                array_map(
                    fn(array $video) =>
                        $this->mapVideo(
                            $video
                        ),
                    $videos
                ),
        ];
    }

    // =====================================
    // GET VIDEO
    // =====================================

    public function getVideo(
        object $request
    ): array {
        $videoId =
            trim(
                (string)(
                    $request->idVideo
                    ?? ''
                )
            );

        if ($videoId === '') {
            throw new SoapFault(
                'INVALID_REQUEST',
                'idVideo es obligatorio'
            );
        }

        $identity =
            $this->authenticatedIdentity(
                $request
            );

        $video =
            $this->repository
                ->findById(
                    $videoId
                );

        if ($video === null) {
            throw new SoapFault(
                'VIDEO_NOT_FOUND',
                'El video solicitado no existe'
            );
        }

        if (
            !$this->canAccessVideo(
                $video,
                $identity
            )
        ) {
            throw new SoapFault(
                'ACCESS_DENIED',
                'El usuario no tiene permiso READ sobre este video'
            );
        }

        return [
            'success' =>
                true,

            'message' =>
                'Video encontrado correctamente',

            'video' =>
                $this->mapVideo(
                    $video
                ),
        ];
    }

    // =====================================
    // GET STREAMING INFO
    // =====================================

    public function getStreamingInfo(
        object $request
    ): array {
        $videoResponse =
            $this->getVideo(
                $request
            );

        $video =
            $videoResponse[
                'video'
            ];

        return [
            'success' =>
                true,

            'message' =>
                'Información de streaming disponible',

            'streaming' => [
                'idVideo' =>
                    $video['idVideo'],

                'idArchivo' =>
                    $video['idArchivo'],

                'nombre' =>
                    $video['nombre'],

                'mimeType' =>
                    $video['mimeType'],

                'formato' =>
                    $video['formato'],

                'tamano' =>
                    $video['tamano'],

                'modo' =>
                    'HTTP_RANGE',

                'reproduccionDisponible' =>
                    true,
            ],
        ];
    }

    // =====================================
    // IDENTIDAD AUTENTICADA
    // =====================================

    private function authenticatedIdentity(
        object $request
    ): array {
        $token =
            trim(
                (string)(
                    $request->token
                    ?? ''
                )
            );

        if ($token === '') {
            throw new SoapFault(
                'UNAUTHORIZED',
                'token es obligatorio'
            );
        }

        try {
            $identity =
                $this->authClient
                    ->validateToken(
                        $token
                    );
        } catch (Throwable $error) {
            throw new SoapFault(
                'AUTH_UNAVAILABLE',
                'No fue posible validar el token'
            );
        }

        if (
            !$identity['valid'] ||
            $identity['userId'] === ''
        ) {
            throw new SoapFault(
                'UNAUTHORIZED',
                'Token inválido'
            );
        }

        return $identity;
    }

    // =====================================
    // AUTORIZACIÓN
    // =====================================

    private function canAccessVideo(
        array $video,
        array $identity
    ): bool {
        if (
            $identity['role'] ===
            'ADMIN'
        ) {
            return true;
        }

        if (
            $video['owner_id'] ===
            $identity['userId']
        ) {
            return true;
        }

        return $this->repository
            ->canRead(
                (string)$video[
                    'id_archivo'
                ],
                (string)$identity[
                    'userId'
                ]
            );
    }

    // =====================================
    // FÍSICO
    // =====================================

    private function physicalPath(
        string $relativePath
    ): string {
        $normalized =
            str_replace(
                '\\',
                '/',
                $relativePath
            );

        if (
            $normalized === '' ||
            str_starts_with(
                $normalized,
                '/'
            ) ||
            str_contains(
                $normalized,
                '../'
            )
        ) {
            throw new SoapFault(
                'INVALID_PATH',
                'Ruta de archivo inválida'
            );
        }

        $candidate =
            $this->storageRoot .
            DIRECTORY_SEPARATOR .
            str_replace(
                '/',
                DIRECTORY_SEPARATOR,
                $normalized
            );

        $real =
            realpath(
                $candidate
            );

        if ($real === false) {
            throw new SoapFault(
                'FILE_NOT_FOUND',
                'No fue posible resolver el archivo físico'
            );
        }

        $root =
            realpath(
                $this->storageRoot
            );

        if (
            $root === false ||
            !str_starts_with(
                $real,
                $root .
                DIRECTORY_SEPARATOR
            )
        ) {
            throw new SoapFault(
                'INVALID_PATH',
                'Archivo fuera de Shared Storage'
            );
        }

        return $real;
    }

    // =====================================
    // FFPROBE
    // =====================================

    private function probeVideo(
        string $path
    ): array {
        $command =
            sprintf(
                'ffprobe -v error ' .
                '-select_streams v:0 ' .
                '-show_entries stream=height ' .
                '-show_entries format=duration,format_name ' .
                '-of json %s 2>&1',
                escapeshellarg(
                    $path
                )
            );

        $output = [];
        $exitCode = 0;

        exec(
            $command,
            $output,
            $exitCode
        );

        if ($exitCode !== 0) {
            throw new SoapFault(
                'MEDIA_ANALYSIS_ERROR',
                'No fue posible analizar el video'
            );
        }

        $json =
            json_decode(
                implode(
                    PHP_EOL,
                    $output
                ),
                true
            );

        if (!is_array($json)) {
            throw new SoapFault(
                'MEDIA_ANALYSIS_ERROR',
                'Respuesta inválida de ffprobe'
            );
        }

        $stream =
            $json['streams'][0]
            ?? [];

        $format =
            $json['format']
            ?? [];

        $formatName =
            (string)(
                $format['format_name']
                ?? ''
            );

        $canonicalFormat =
            str_contains(
                strtolower(
                    $formatName
                ),
                'mp4'
            )
                ? 'MP4'
                : strtoupper(
                    explode(
                        ',',
                        $formatName
                    )[0]
                    ?? 'UNKNOWN'
                );

        return [
            'duration' =>
                (float)(
                    $format['duration']
                    ?? 0
                ),

            'height' =>
                (int)(
                    $stream['height']
                    ?? 0
                ),

            'format' =>
                $canonicalFormat,
        ];
    }

    // =====================================
    // CALIDAD
    // =====================================

    private function qualityFromHeight(
        int $height
    ): string {
        if ($height >= 2160) {
            return '2160p';
        }

        if ($height >= 1440) {
            return '1440p';
        }

        if ($height >= 1080) {
            return '1080p';
        }

        if ($height >= 720) {
            return '720p';
        }

        if ($height >= 480) {
            return '480p';
        }

        if ($height >= 360) {
            return '360p';
        }

        if ($height > 0) {
            return $height . 'p';
        }

        return 'UNKNOWN';
    }

    // =====================================
    // MAP VIDEO
    // =====================================

    private function mapVideo(
        array $video
    ): array {
        return [
            'idVideo' =>
                $video['id_video'],

            'idArchivo' =>
                $video['id_archivo'],

            'nombre' =>
                $video['nombre'],

            'mimeType' =>
                $video['mime_type'],

            'tamano' =>
                (int)$video[
                    'tamano'
                ],

            'duracionSegundos' =>
                $video[
                    'duracion_segundos'
                ] !== null
                    ? (int)$video[
                        'duracion_segundos'
                    ]
                    : null,

            'calidad' =>
                $video['calidad'],

            'formato' =>
                $video['formato'],
        ];
    }
}
