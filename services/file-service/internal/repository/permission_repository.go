package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type ResourcePermission struct {
	ID          string
	FileID      string
	UserUUID    string
	DirectoryID string

	CanRead  bool
	CanWrite bool
	CanShare bool
}

type PermissionRepository struct {
	db *sql.DB
}

func NewPermissionRepository(
	db *sql.DB,
) *PermissionRepository {

	return &PermissionRepository{
		db: db,
	}
}

// =====================================
// BUSCAR PERMISO
// =====================================

func (r *PermissionRepository) FindPermission(
	ctx context.Context,
	fileID string,
	directoryID string,
) (*ResourcePermission, error) {

	if fileID == "" {
		return nil,
			errors.New(
				"fileID es obligatorio",
			)
	}

	if directoryID == "" {
		return nil,
			errors.New(
				"directoryID es obligatorio",
			)
	}

	const query = `
SELECT
    p.id_permiso::text,
    p.id_archivo::text,
    u.id_usuario::text,
    u.directorio_id,
    p.puede_leer,
    p.puede_escribir,
    p.puede_compartir
FROM permiso_recurso p
JOIN usuario u
    ON u.id_usuario = p.id_usuario
WHERE
    p.id_archivo = $1::uuid
    AND u.directorio_id = $2
    AND u.estado = true
LIMIT 1;
`

	permission :=
		&ResourcePermission{}

	err :=
		r.db.QueryRowContext(
			ctx,
			query,
			fileID,
			directoryID,
		).Scan(
			&permission.ID,
			&permission.FileID,
			&permission.UserUUID,
			&permission.DirectoryID,
			&permission.CanRead,
			&permission.CanWrite,
			&permission.CanShare,
		)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return nil, nil
	}

	if err != nil {
		return nil,
			fmt.Errorf(
				"error consultando permiso: %w",
				err,
			)
	}

	return permission, nil
}

// =====================================
// LISTAR PERMISOS DE UN USUARIO
// =====================================
//
// Devuelve todas las ACL activas asociadas
// al usuario indicado.
//
// Se consideran accesibles las ACL que
// tengan al menos uno de:
//
// - puede_leer
// - puede_escribir
// - puede_compartir
//
// Aunque Grant normaliza write/share -> read,
// usamos OR para tolerar filas antiguas o
// inconsistentes creadas antes de esa regla.
//
// =====================================

func (r *PermissionRepository) FindAccessibleByUser(
	ctx context.Context,
	directoryID string,
) ([]*ResourcePermission, error) {

	if directoryID == "" {
		return nil,
			errors.New(
				"directoryID es obligatorio",
			)
	}

	const query = `
SELECT
    p.id_permiso::text,
    p.id_archivo::text,
    u.id_usuario::text,
    u.directorio_id,
    p.puede_leer,
    p.puede_escribir,
    p.puede_compartir
FROM permiso_recurso p
JOIN usuario u
    ON u.id_usuario = p.id_usuario
JOIN archivo a
    ON a.id_archivo = p.id_archivo
WHERE
    u.directorio_id = $1
    AND u.estado = true
    AND (
        p.puede_leer = true
        OR p.puede_escribir = true
        OR p.puede_compartir = true
    )
ORDER BY
    p.id_archivo;
`

	rows,
		err :=
		r.db.QueryContext(
			ctx,
			query,
			directoryID,
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"error listando permisos del usuario: %w",
				err,
			)
	}

	defer rows.Close()

	permissions :=
		make(
			[]*ResourcePermission,
			0,
		)

	for rows.Next() {

		permission :=
			&ResourcePermission{}

		if err :=
			rows.Scan(
				&permission.ID,
				&permission.FileID,
				&permission.UserUUID,
				&permission.DirectoryID,
				&permission.CanRead,
				&permission.CanWrite,
				&permission.CanShare,
			); err != nil {

			return nil,
				fmt.Errorf(
					"error leyendo permiso: %w",
					err,
				)
		}

		permissions =
			append(
				permissions,
				permission,
			)
	}

	if err :=
		rows.Err(); err != nil {

		return nil,
			fmt.Errorf(
				"error recorriendo permisos: %w",
				err,
			)
	}

	return permissions, nil
}

// =====================================
// CREAR O ACTUALIZAR PERMISO
// =====================================

func (r *PermissionRepository) UpsertPermission(
	ctx context.Context,
	fileID string,
	directoryID string,
	canRead bool,
	canWrite bool,
	canShare bool,
) error {

	if fileID == "" {
		return errors.New(
			"fileID es obligatorio",
		)
	}

	if directoryID == "" {
		return errors.New(
			"directoryID es obligatorio",
		)
	}

	const query = `
INSERT INTO permiso_recurso (
    id_archivo,
    id_usuario,
    puede_leer,
    puede_escribir,
    puede_compartir
)
SELECT
    $1::uuid,
    u.id_usuario,
    $3,
    $4,
    $5
FROM usuario u
WHERE
    u.directorio_id = $2
    AND u.estado = true
ON CONFLICT (
    id_archivo,
    id_usuario
)
DO UPDATE SET
    puede_leer = EXCLUDED.puede_leer,
    puede_escribir = EXCLUDED.puede_escribir,
    puede_compartir = EXCLUDED.puede_compartir;
`

	result,
		err :=
		r.db.ExecContext(
			ctx,
			query,
			fileID,
			directoryID,
			canRead,
			canWrite,
			canShare,
		)

	if err != nil {
		return fmt.Errorf(
			"no se pudo crear o actualizar permiso: %w",
			err,
		)
	}

	rowsAffected,
		err :=
		result.RowsAffected()

	if err != nil {
		return fmt.Errorf(
			"no se pudo verificar permiso: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(
			"usuario %s no encontrado o inactivo",
			directoryID,
		)
	}

	return nil
}

// =====================================
// REVOCAR PERMISO
// =====================================

func (r *PermissionRepository) RevokePermission(
	ctx context.Context,
	fileID string,
	directoryID string,
) error {

	if fileID == "" {
		return errors.New(
			"fileID es obligatorio",
		)
	}

	if directoryID == "" {
		return errors.New(
			"directoryID es obligatorio",
		)
	}

	const query = `
DELETE FROM permiso_recurso
WHERE
    id_archivo = $1::uuid
    AND id_usuario = (
        SELECT id_usuario
        FROM usuario
        WHERE directorio_id = $2
    );
`

	_,
		err :=
		r.db.ExecContext(
			ctx,
			query,
			fileID,
			directoryID,
		)

	if err != nil {
		return fmt.Errorf(
			"no se pudo revocar permiso: %w",
			err,
		)
	}

	return nil
}
