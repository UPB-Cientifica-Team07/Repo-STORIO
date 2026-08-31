package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Home struct {
	ID          string
	UserUUID    string
	DirectoryID string
	BasePath    string
	QuotaBytes  int64
	UsedBytes   int64
}

type HomeRepository struct {
	db *sql.DB
}

func NewHomeRepository(db *sql.DB) *HomeRepository {
	return &HomeRepository{
		db: db,
	}
}

// =====================================
// BUSCAR HOME POR DIRECTORIO_ID
// =====================================

func (r *HomeRepository) FindByDirectoryID(
	ctx context.Context,
	directoryID string,
) (*Home, error) {

	if directoryID == "" {
		return nil, errors.New(
			"directorio_id es obligatorio",
		)
	}

	const query = `
SELECT
    h.id_home::text,
    u.id_usuario::text,
    u.directorio_id,
    h.ruta_base,
    h.cuota_maxima,
    h.usado_bytes
FROM home h
JOIN usuario u
    ON u.id_usuario = h.id_usuario
WHERE
    u.directorio_id = $1
    AND u.estado = true
LIMIT 1;
`

	home := &Home{}

	err := r.db.QueryRowContext(
		ctx,
		query,
		directoryID,
	).Scan(
		&home.ID,
		&home.UserUUID,
		&home.DirectoryID,
		&home.BasePath,
		&home.QuotaBytes,
		&home.UsedBytes,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf(
			"el usuario %s no tiene Home configurado",
			directoryID,
		)
	}

	if err != nil {
		return nil, fmt.Errorf(
			"error consultando Home: %w",
			err,
		)
	}

	return home, nil
}

// =====================================
// LISTAR HOMES ACTIVOS
// =====================================

func (r *HomeRepository) ListActive(
	ctx context.Context,
) ([]*Home, error) {

	const query = `
SELECT
    h.id_home::text,
    u.id_usuario::text,
    u.directorio_id,
    h.ruta_base,
    h.cuota_maxima,
    h.usado_bytes
FROM home h
JOIN usuario u
    ON u.id_usuario = h.id_usuario
WHERE u.estado = true
ORDER BY u.directorio_id;
`

	rows, err := r.db.QueryContext(
		ctx,
		query,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"error listando Homes: %w",
			err,
		)
	}

	defer rows.Close()

	homes := make(
		[]*Home,
		0,
	)

	for rows.Next() {
		home := &Home{}

		if err := rows.Scan(
			&home.ID,
			&home.UserUUID,
			&home.DirectoryID,
			&home.BasePath,
			&home.QuotaBytes,
			&home.UsedBytes,
		); err != nil {
			return nil, fmt.Errorf(
				"error leyendo Home: %w",
				err,
			)
		}

		homes = append(
			homes,
			home,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"error recorriendo Homes: %w",
			err,
		)
	}

	return homes, nil
}

// =====================================
// VALIDAR CUOTA SIN MODIFICAR
// =====================================

func (r *HomeRepository) ValidateQuota(
	ctx context.Context,
	directoryID string,
	additionalBytes int64,
) (*Home, error) {

	if additionalBytes < 0 {
		return nil, errors.New(
			"el tamaño adicional no puede ser negativo",
		)
	}

	home, err := r.FindByDirectoryID(
		ctx,
		directoryID,
	)

	if err != nil {
		return nil, err
	}

	if home.UsedBytes+additionalBytes >
		home.QuotaBytes {

		available :=
			home.QuotaBytes -
				home.UsedBytes

		return nil, fmt.Errorf(
			"cuota excedida: disponibles=%d bytes, requeridos=%d bytes",
			available,
			additionalBytes,
		)
	}

	return home, nil
}

// =====================================
// RESERVAR CUOTA ATÓMICAMENTE
// =====================================
//
// PostgreSQL ejecuta el UPDATE de forma
// atómica.
//
// Dos UploadFile concurrentes no pueden
// utilizar simultáneamente el mismo espacio
// disponible.
//
// =====================================

func (r *HomeRepository) ReserveQuota(
	ctx context.Context,
	directoryID string,
	bytes int64,
) (*Home, error) {

	if directoryID == "" {
		return nil, errors.New(
			"directorio_id es obligatorio",
		)
	}

	if bytes < 0 {
		return nil, errors.New(
			"los bytes a reservar no pueden ser negativos",
		)
	}

	if bytes == 0 {
		return r.FindByDirectoryID(
			ctx,
			directoryID,
		)
	}

	const query = `
UPDATE home AS h
SET usado_bytes = h.usado_bytes + $2
FROM usuario AS u
WHERE
    h.id_usuario = u.id_usuario
    AND u.directorio_id = $1
    AND u.estado = true
    AND h.usado_bytes + $2 <= h.cuota_maxima
RETURNING
    h.id_home::text,
    u.id_usuario::text,
    u.directorio_id,
    h.ruta_base,
    h.cuota_maxima,
    h.usado_bytes;
`

	home := &Home{}

	err := r.db.QueryRowContext(
		ctx,
		query,
		directoryID,
		bytes,
	).Scan(
		&home.ID,
		&home.UserUUID,
		&home.DirectoryID,
		&home.BasePath,
		&home.QuotaBytes,
		&home.UsedBytes,
	)

	if errors.Is(err, sql.ErrNoRows) {
		currentHome, findErr :=
			r.FindByDirectoryID(
				ctx,
				directoryID,
			)

		if findErr != nil {
			return nil, findErr
		}

		available :=
			currentHome.QuotaBytes -
				currentHome.UsedBytes

		return nil, fmt.Errorf(
			"cuota excedida: disponibles=%d bytes, requeridos=%d bytes",
			available,
			bytes,
		)
	}

	if err != nil {
		return nil, fmt.Errorf(
			"no se pudo reservar cuota: %w",
			err,
		)
	}

	return home, nil
}

// =====================================
// LIBERAR CUOTA
// =====================================

func (r *HomeRepository) ReleaseQuota(
	ctx context.Context,
	directoryID string,
	bytes int64,
) error {

	if directoryID == "" {
		return errors.New(
			"directorio_id es obligatorio",
		)
	}

	if bytes < 0 {
		return errors.New(
			"los bytes a liberar no pueden ser negativos",
		)
	}

	if bytes == 0 {
		return nil
	}

	const query = `
UPDATE home AS h
SET usado_bytes = GREATEST(
    h.usado_bytes - $2,
    0
)
FROM usuario AS u
WHERE
    h.id_usuario = u.id_usuario
    AND u.directorio_id = $1
    AND u.estado = true;
`

	result, err := r.db.ExecContext(
		ctx,
		query,
		directoryID,
		bytes,
	)

	if err != nil {
		return fmt.Errorf(
			"no se pudo liberar cuota: %w",
			err,
		)
	}

	rowsAffected, err :=
		result.RowsAffected()

	if err != nil {
		return fmt.Errorf(
			"no se pudo verificar liberación de cuota: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(
			"no se encontró Home activo para %s",
			directoryID,
		)
	}

	return nil
}

// =====================================
// FIJAR USO REAL
// =====================================
//
// Se utiliza durante reconciliación.
//
// =====================================

func (r *HomeRepository) SetUsage(
	ctx context.Context,
	directoryID string,
	usedBytes int64,
) error {

	if directoryID == "" {
		return errors.New(
			"directorio_id es obligatorio",
		)
	}

	if usedBytes < 0 {
		return errors.New(
			"usedBytes no puede ser negativo",
		)
	}

	const query = `
UPDATE home AS h
SET usado_bytes = $2
FROM usuario AS u
WHERE
    h.id_usuario = u.id_usuario
    AND u.directorio_id = $1
    AND u.estado = true
    AND $2 <= h.cuota_maxima;
`

	result, err := r.db.ExecContext(
		ctx,
		query,
		directoryID,
		usedBytes,
	)

	if err != nil {
		return fmt.Errorf(
			"no se pudo reconciliar cuota: %w",
			err,
		)
	}

	rowsAffected, err :=
		result.RowsAffected()

	if err != nil {
		return fmt.Errorf(
			"no se pudo verificar reconciliación: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(
			"no se pudo reconciliar %s: Home inexistente o uso físico supera la cuota",
			directoryID,
		)
	}

	return nil
}
