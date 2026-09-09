<?php

declare(strict_types=1);

final class AuthClient
{
    private string $baseUrl;

    public function __construct(
        string $baseUrl
    ) {
        $this->baseUrl =
            rtrim(
                $baseUrl,
                '/'
            );
    }

    // =====================================
    // VALIDATE TOKEN
    // =====================================

    public function validateToken(
        string $token
    ): array {
        $token =
            trim(
                $token
            );

        if ($token === '') {
            throw new RuntimeException(
                'Token obligatorio'
            );
        }

        $url =
            $this->baseUrl .
            '/internal/auth/validate';

        $context =
            stream_context_create([
                'http' => [
                    'method' =>
                        'GET',

                    'timeout' =>
                        3,

                    'ignore_errors' =>
                        true,

                    'header' =>
                        "Authorization: Bearer " .
                        $token . "\r\n",
                ],
            ]);

        $body =
            @file_get_contents(
                $url,
                false,
                $context
            );

        if ($body === false) {
            throw new RuntimeException(
                'No fue posible conectar con Auth Service'
            );
        }

        $parts =
            explode(
                '|',
                trim($body),
                4
            );

        if (count($parts) !== 4) {
            throw new RuntimeException(
                'Respuesta inválida de Auth Service'
            );
        }

        return [
            'valid' =>
                strtolower(
                    trim($parts[0])
                ) === 'true',

            'userId' =>
                trim(
                    $parts[1]
                ),

            'role' =>
                strtoupper(
                    trim(
                        $parts[2]
                    )
                ),

            'message' =>
                trim(
                    $parts[3]
                ),
        ];
    }
}
