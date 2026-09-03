<?php

declare(strict_types=1);

require_once __DIR__ . '/../src/VideoRepository.php';
require_once __DIR__ . '/../src/StreamingService.php';
require_once __DIR__ . '/../src/AuthClient.php';
require_once __DIR__ . '/../src/VideoStreamHandler.php';

$config =
    require __DIR__ . '/../config/config.php';

$repository =
    new VideoRepository(
        $config['database']
    );

$authClient =
    new AuthClient(
        (string)$config['auth']['base_url']
    );

$service =
    new StreamingService(
        $repository,
        $authClient,
        (string)$config['storage']['root']
    );

$streamHandler =
    new VideoStreamHandler(
        $repository,
        $authClient,
        (string)$config['storage']['root']
    );

// =====================================
// HEALTH
// =====================================

if (
    $_SERVER['REQUEST_METHOD'] === 'GET' &&
    ($_SERVER['REQUEST_URI'] ?? '/') === '/health'
) {
    header('Content-Type: application/json; charset=utf-8');

    echo json_encode(
        [
            'service' =>
                'UPB-CIENTIFICA Streaming Service',

            'status' =>
                'ACTIVE',

            'technology' =>
                'PHP + SOAP + WSDL',

            'port' =>
                (int)$config['service']['port'],

            'timestamp' =>
                gmdate('c'),
        ],
        JSON_UNESCAPED_UNICODE |
        JSON_PRETTY_PRINT
    );

    exit;
}

// =====================================
// HTTP VIDEO STREAM
// =====================================

$path =
    parse_url(
        $_SERVER['REQUEST_URI'] ?? '/',
        PHP_URL_PATH
    );

if (
    $_SERVER['REQUEST_METHOD'] === 'GET' &&
    preg_match(
        '#^/stream/([0-9a-fA-F-]{36})$#',
        $path,
        $matches
    )
) {
    $streamHandler->handle(
        $matches[1]
    );

    exit;
}

// =====================================
// WSDL
// =====================================

if (
    $_SERVER['REQUEST_METHOD'] === 'GET' &&
    isset($_GET['wsdl'])
) {
    header('Content-Type: text/xml; charset=utf-8');

    readfile(
        __DIR__ .
        '/../wsdl/streaming.wsdl'
    );

    exit;
}

// =====================================
// SOLO SOAP POST
// =====================================

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    http_response_code(405);

    header('Content-Type: application/json; charset=utf-8');

    echo json_encode(
        [
            'success' => false,
            'message' =>
                'Use POST para SOAP, /?wsdl para WSDL o /health para estado',
        ],
        JSON_UNESCAPED_UNICODE
    );

    exit;
}

// =====================================
// SOAP SERVER
// =====================================

try {
    $soapServer =
        new SoapServer(
            __DIR__ .
            '/../wsdl/streaming.wsdl',
            [
                'cache_wsdl' =>
                    WSDL_CACHE_NONE,

                'exceptions' =>
                    true,
            ]
        );

    $soapServer->setObject(
        $service
    );

    $soapServer->handle();

} catch (Throwable $error) {

    error_log(
        'Streaming SOAP ERROR: ' .
        $error->getMessage()
    );

    if (!headers_sent()) {
        http_response_code(500);
        header(
            'Content-Type: text/xml; charset=utf-8'
        );
    }

    $fault =
        new SoapFault(
            'SERVER_ERROR',
            'Error interno del Streaming Service'
        );

    $faultServer =
        new SoapServer(
            null,
            [
                'uri' =>
                    'urn:UPBCientificaStreaming',
            ]
        );

    $faultServer->fault(
        $fault->faultcode,
        $fault->faultstring
    );
}
