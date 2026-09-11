<?php

declare(strict_types=1);

$defaultStorageRoot =
    realpath(
        __DIR__ .
        '/../../file-service/data/shared-storage'
    );

$dbPassword =
    getenv('STREAM_DB_PASSWORD');

if (
    $dbPassword === false ||
    trim($dbPassword) === ''
) {
    $dbPassword =
        getenv('DB_PASSWORD');
}

if (
    $dbPassword === false ||
    trim($dbPassword) === ''
) {
    throw new RuntimeException(
        'STREAM_DB_PASSWORD o DB_PASSWORD es obligatorio'
    );
}

return [
    'database' => [
        'host' =>
            getenv('DB_HOST')
                ?: '127.0.0.1',

        'port' =>
            getenv('DB_PORT')
                ?: '5434',

        'name' =>
            getenv('DB_NAME')
                ?: 'upb_cientifica',

        'user' =>
            getenv('DB_USER')
                ?: 'upb_app',

        'password' =>
            $dbPassword,
    ],

    'service' => [
        'host' =>
            getenv('STREAM_HOST')
                ?: '127.0.0.1',

        'port' =>
            getenv('STREAM_PORT')
                ?: '50054',
    ],

    'auth' => [
        'base_url' =>
            getenv('AUTH_SERVICE_URL')
                ?: 'http://127.0.0.1:8081',
    ],

    'storage' => [
        'root' =>
            getenv('SHARED_STORAGE_ROOT')
                ?: $defaultStorageRoot,
    ],
];
