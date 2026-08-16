-- ============================================================
-- UPB-CIENTÍFICA
-- Datos iniciales de prueba
-- Motor: PostgreSQL
-- ============================================================


-- ============================================================
-- USUARIOS
-- ============================================================

INSERT INTO usuario (
    id_usuario,
    usuario,
    correo,
    password_hash,
    rol,
    estado
)
VALUES
(
    '11111111-1111-1111-1111-111111111111',
    'admin',
    'admin@upbcientifica.edu.co',
    '$2a$10$example_admin_hash',
    'ADMINISTRADOR',
    TRUE
),
(
    '22222222-2222-2222-2222-222222222222',
    'investigador1',
    'investigador@upbcientifica.edu.co',
    '$2a$10$example_investigador_hash',
    'INVESTIGADOR',
    TRUE
),
(
    '33333333-3333-3333-3333-333333333333',
    'usuario1',
    'usuario@upbcientifica.edu.co',
    '$2a$10$example_usuario_hash',
    'USUARIO',
    TRUE
);


-- ============================================================
-- NODOS HPC
-- ============================================================

INSERT INTO nodo_hpc (
    id_nodo,
    hostname,
    cpu,
    memoria,
    estado,
    ip,
    ubicacion
)
VALUES
(
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'hpc-node-01',
    16,
    32768,
    'DISPONIBLE',
    '192.168.1.101',
    'Cluster Principal'
),
(
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'hpc-node-02',
    32,
    65536,
    'OCUPADO',
    '192.168.1.102',
    'Cluster Principal'
);


-- ============================================================
-- ARCHIVOS
-- ============================================================

INSERT INTO archivo (
    id_archivo,
    nombre,
    ruta,
    tamano,
    tipo_archivo,
    id_usuario
)
VALUES
(
    '10000000-0000-0000-0000-000000000001',
    'investigacion.pdf',
    '/storage/investigador1/investigacion.pdf',
    2457600,
    'DOCUMENTO',
    '22222222-2222-2222-2222-222222222222'
),
(
    '10000000-0000-0000-0000-000000000002',
    'microscopia.png',
    '/storage/investigador1/microscopia.png',
    5242880,
    'IMAGEN',
    '22222222-2222-2222-2222-222222222222'
),
(
    '10000000-0000-0000-0000-000000000003',
    'experimento.mp4',
    '/storage/usuario1/experimento.mp4',
    104857600,
    'VIDEO',
    '33333333-3333-3333-3333-333333333333'
);


-- ============================================================
-- IMAGEN
-- ============================================================

INSERT INTO imagen (
    id_imagen,
    id_archivo,
    resolucion,
    formato
)
VALUES
(
    '20000000-0000-0000-0000-000000000001',
    '10000000-0000-0000-0000-000000000002',
    '1920x1080',
    'PNG'
);


-- ============================================================
-- VIDEO
-- ============================================================

INSERT INTO video (
    id_video,
    id_archivo,
    duracion,
    calidad,
    formato
)
VALUES
(
    '30000000-0000-0000-0000-000000000001',
    '10000000-0000-0000-0000-000000000003',
    360,
    '1080p',
    'MP4'
);


-- ============================================================
-- TOKENS
-- ============================================================

INSERT INTO token (
    id_token,
    id_usuario,
    token,
    expiracion,
    estado
)
VALUES
(
    '40000000-0000-0000-0000-000000000001',
    '11111111-1111-1111-1111-111111111111',
    'token_admin_prueba_001',
    '2026-12-31 23:59:59',
    TRUE
),
(
    '40000000-0000-0000-0000-000000000002',
    '22222222-2222-2222-2222-222222222222',
    'token_investigador_prueba_001',
    '2026-12-31 23:59:59',
    TRUE
);


-- ============================================================
-- TRABAJOS HPC
-- ============================================================

INSERT INTO trabajo_hpc (
    id_job,
    id_usuario,
    id_nodo,
    lenguaje,
    estado,
    inicio,
    fin,
    recursos,
    descripcion
)
VALUES
(
    '50000000-0000-0000-0000-000000000001',
    '22222222-2222-2222-2222-222222222222',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'MPI',
    'FINALIZADO',
    '2026-08-15 10:00:00',
    '2026-08-15 10:30:00',
    '4 procesos MPI',
    'Simulacion inicial de prueba'
),
(
    '50000000-0000-0000-0000-000000000002',
    '22222222-2222-2222-2222-222222222222',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'MPI',
    'EJECUTANDO',
    '2026-08-15 14:00:00',
    NULL,
    '8 procesos MPI',
    'Procesamiento distribuido en ejecucion'
);