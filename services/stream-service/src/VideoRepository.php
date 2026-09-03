<?php

declare(strict_types=1);

final class VideoRepository
{
    private PDO $db;

    public function __construct(
        array $config
    ) {
        $dsn =
            sprintf(
                'pgsql:host=%s;port=%s;dbname=%s',
                $config['host'],
                $config['port'],
                $config['name']
            );

        $this->db =
            new PDO(
                $dsn,
                $config['user'],
                $config['password'],
                [
                    PDO::ATTR_ERRMODE =>
                        PDO::ERRMODE_EXCEPTION,

                    PDO::ATTR_DEFAULT_FETCH_MODE =>
                        PDO::FETCH_ASSOC,
                ]
            );
    }

    // =====================================
    // LISTAR VIDEOS DE UN USUARIO
    // =====================================

    public function listByDirectoryId(
        string $directoryId
    ): array {
        $sql = <<<'SQL'
SELECT
    v.id_video::text AS id_video,
    a.id_archivo::text AS id_archivo,
    a.nombre,
    a.mime_type,
    a.tamano,
    a.ruta_relativa,
    v.duracion_segundos,
    v.calidad,
    v.formato,
    v.fecha_registro
FROM video v
JOIN archivo a
    ON a.id_archivo = v.id_archivo
JOIN usuario u
    ON u.id_usuario = a.id_usuario
WHERE
    u.directorio_id = :directory_id
    AND a.tipo_archivo = 'VIDEO'
ORDER BY
    v.fecha_registro DESC;
SQL;

        $stmt =
            $this->db->prepare(
                $sql
            );

        $stmt->execute([
            ':directory_id' =>
                $directoryId,
        ]);

        return $stmt->fetchAll();
    }

    // =====================================
    // BUSCAR VIDEO POR ID VIDEO
    // =====================================

    public function findById(
        string $videoId
    ): ?array {
        $sql = <<<'SQL'
SELECT
    v.id_video::text AS id_video,
    a.id_archivo::text AS id_archivo,
    u.directorio_id AS owner_id,
    a.nombre,
    a.mime_type,
    a.tamano,
    a.ruta_relativa,
    v.duracion_segundos,
    v.calidad,
    v.formato,
    v.fecha_registro
FROM video v
JOIN archivo a
    ON a.id_archivo = v.id_archivo
JOIN usuario u
    ON u.id_usuario = a.id_usuario
WHERE
    v.id_video = CAST(:video_id AS uuid)
    AND a.tipo_archivo = 'VIDEO'
LIMIT 1;
SQL;

        $stmt =
            $this->db->prepare(
                $sql
            );

        $stmt->execute([
            ':video_id' =>
                $videoId,
        ]);

        $video =
            $stmt->fetch();

        return $video ?: null;
    }

    // =====================================
    // BUSCAR ARCHIVO PARA REGISTRO
    // =====================================

    public function findFileById(
        string $fileId
    ): ?array {
        $sql = <<<'SQL'
SELECT
    a.id_archivo::text AS id_archivo,
    u.directorio_id AS owner_id,
    a.nombre,
    a.ruta_relativa,
    a.tamano,
    a.mime_type,
    a.tipo_archivo
FROM archivo a
JOIN usuario u
    ON u.id_usuario = a.id_usuario
WHERE
    a.id_archivo = CAST(:file_id AS uuid)
LIMIT 1;
SQL;

        $stmt =
            $this->db->prepare(
                $sql
            );

        $stmt->execute([
            ':file_id' =>
                $fileId,
        ]);

        $file =
            $stmt->fetch();

        return $file ?: null;
    }

    // =====================================
    // UPSERT METADATA VIDEO
    // =====================================

    public function upsertVideo(
        string $fileId,
        int $durationSeconds,
        string $quality,
        string $format
    ): array {
        $sql = <<<'SQL'
INSERT INTO video (
    id_archivo,
    duracion_segundos,
    calidad,
    formato
)
VALUES (
    CAST(:file_id AS uuid),
    :duration_seconds,
    :quality,
    :format
)
ON CONFLICT (id_archivo)
DO UPDATE SET
    duracion_segundos =
        EXCLUDED.duracion_segundos,
    calidad =
        EXCLUDED.calidad,
    formato =
        EXCLUDED.formato
RETURNING
    id_video::text AS id_video,
    id_archivo::text AS id_archivo,
    duracion_segundos,
    calidad,
    formato,
    fecha_registro;
SQL;

        $stmt =
            $this->db->prepare(
                $sql
            );

        $stmt->execute([
            ':file_id' =>
                $fileId,

            ':duration_seconds' =>
                $durationSeconds,

            ':quality' =>
                $quality,

            ':format' =>
                $format,
        ]);

        return $stmt->fetch();
    }


    // =====================================
    // ACL READ
    // =====================================

    public function canRead(
        string $fileId,
        string $directoryId
    ): bool {
        $sql = <<<'SQL'
SELECT EXISTS (
    SELECT 1
    FROM permiso_recurso p
    JOIN usuario u
        ON u.id_usuario = p.id_usuario
    WHERE
        p.id_archivo = CAST(:file_id AS uuid)
        AND u.directorio_id = :directory_id
        AND u.estado = true
        AND p.puede_leer = true
) AS allowed;
SQL;

        $stmt =
            $this->db->prepare(
                $sql
            );

        $stmt->execute([
            ':file_id' =>
                $fileId,

            ':directory_id' =>
                $directoryId,
        ]);

        return filter_var(
            $stmt->fetchColumn(),
            FILTER_VALIDATE_BOOLEAN
        );
    }



    // =====================================
    // LISTAR VIDEOS AUTORIZADOS
    // =====================================

    public function listAccessibleByDirectoryId(
        string $directoryId,
        bool $isAdmin
    ): array {
        if ($isAdmin) {
            $sql = <<<'SQL'
SELECT
    v.id_video::text AS id_video,
    a.id_archivo::text AS id_archivo,
    owner.directorio_id AS owner_id,
    a.nombre,
    a.mime_type,
    a.tamano,
    a.ruta_relativa,
    v.duracion_segundos,
    v.calidad,
    v.formato,
    v.fecha_registro
FROM video v
JOIN archivo a
    ON a.id_archivo = v.id_archivo
JOIN usuario owner
    ON owner.id_usuario = a.id_usuario
WHERE
    a.tipo_archivo = 'VIDEO'
ORDER BY
    v.fecha_registro DESC;
SQL;

            $stmt =
                $this->db->prepare(
                    $sql
                );

            $stmt->execute();

            return $stmt->fetchAll();
        }

        $sql = <<<'SQL'
SELECT
    v.id_video::text AS id_video,
    a.id_archivo::text AS id_archivo,
    owner.directorio_id AS owner_id,
    a.nombre,
    a.mime_type,
    a.tamano,
    a.ruta_relativa,
    v.duracion_segundos,
    v.calidad,
    v.formato,
    v.fecha_registro
FROM video v
JOIN archivo a
    ON a.id_archivo = v.id_archivo
JOIN usuario owner
    ON owner.id_usuario = a.id_usuario
WHERE
    a.tipo_archivo = 'VIDEO'
    AND (
        owner.directorio_id = :directory_id
        OR EXISTS (
            SELECT 1
            FROM permiso_recurso p
            JOIN usuario granted
                ON granted.id_usuario = p.id_usuario
            WHERE
                p.id_archivo = a.id_archivo
                AND granted.directorio_id = :directory_id
                AND granted.estado = true
                AND p.puede_leer = true
        )
    )
ORDER BY
    v.fecha_registro DESC;
SQL;

        $stmt =
            $this->db->prepare(
                $sql
            );

        $stmt->execute([
            ':directory_id' =>
                $directoryId,
        ]);

        return $stmt->fetchAll();
    }

}
