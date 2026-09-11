<?php

/*
 * Headers defensivos globales.
 *
 * No se incluye HSTS porque este servicio HTTP
 * se despliega detrás de la capa de red/proxy y
 * este listener local no ofrece TLS directamente.
 */
header_remove('X-Powered-By');

header(
    'X-Content-Type-Options: nosniff'
);

header(
    'X-Frame-Options: DENY'
);

header(
    'Referrer-Policy: no-referrer'
);

header(
    'Permissions-Policy: camera=(), microphone=(), geolocation=()'
);

header(
    "Content-Security-Policy: default-src 'none'; frame-ancestors 'none'; base-uri 'none'"
);


$path =
    parse_url(
        $_SERVER['REQUEST_URI'],
        PHP_URL_PATH
    );

$file =
    __DIR__ . $path;

if (
    $path !== '/' &&
    is_file($file)
) {
    return false;
}

require __DIR__ . '/index.php';
