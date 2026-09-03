<?php

declare(strict_types=1);

$defaultStorageRoot =
    realpath(
        __DIR__ .
        '/../../file-service/data/shared-storage'
    );

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
            getenv('DB_PASSWORD')
                ?: 'upb_dev_2026',
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
