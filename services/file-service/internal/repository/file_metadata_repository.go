package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type FileMetadata struct {
	ID           string
	HomeID       string
	UserUUID     string
	DirectoryID  string
	Name         string
	RelativePath string
	Size         int64
	MIMEType     string
	FileType     string
	UnixMode     string
}

type FileMetadataRepository struct {
	db *sql.DB
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewFileMetadataRepository(
	db *sql.DB,
) *FileMetadataRepository {

	return &FileMetadataRepository{
		db: db,
	}
}

// =====================================
// CREAR METADATA
// =====================================

func (r *FileMetadataRepository) Create(
	ctx context.Context,
	file *File,
) error {

	if file == nil {
		return errors.New(
			"archivo nil",
		)
	}

	if file.ID == "" {
		return errors.New(
			"file ID es obligatorio",
		)
	}

	if file.UserID == "" {
		return errors.New(
			"user ID es obligatorio",
		)
	}

	if file.RelativePath == "" {
		return errors.New(
			"relative_path es obligatorio para archivos Home",
		)
	}

	fileType :=
		databaseFileType(
			file.Type,
		)

	unixMode :=
		file.UnixMode

	if unixMode == "" {
		unixMode =
			"0640"
	}

	const query = `
INSERT INTO archivo (
    id_archivo,
    id_home,
    id_usuario,
    nombre,
    ruta_relativa,
    tamano,
    mime_type,
    tipo_archivo,
    permisos_unix
)
SELECT
    $1::uuid,
    h.id_home,
    u.id_usuario,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
FROM usuario u
JOIN home h
    ON h.id_usuario = u.id_usuario
WHERE
    u.directorio_id = $2
    AND u.estado = true
ON CONFLICT (id_archivo)
DO NOTHING;
`

	result,
		err :=
		r.db.ExecContext(
			ctx,
			query,
			file.ID,
			file.UserID,
			file.Name,
			file.RelativePath,
			file.Size,
			file.Type,
			fileType,
			unixMode,
		)

	if err != nil {
		return fmt.Errorf(
			"no se pudo registrar metadata PostgreSQL: %w",
			err,
		)
	}

	rowsAffected,
		err :=
		result.RowsAffected()

	if err != nil {
		return fmt.Errorf(
			"no se pudo verificar metadata PostgreSQL: %w",
			err,
		)
	}

	// Puede ser 0 si ya estaba registrado.
	if rowsAffected == 0 {
		return nil
	}

	return nil
}

// =====================================
// ACTUALIZAR METADATA
// =====================================
//
// Actualiza la representación persistente
// del archivo en PostgreSQL.
//
// El propietario y el Home NO cambian.
//
// =====================================

func (r *FileMetadataRepository) Update(
	ctx context.Context,
	file *File,
) error {

	if file == nil {
		return errors.New(
			"archivo nil",
		)
	}

	if file.ID == "" {
		return errors.New(
			"fileID es obligatorio",
		)
	}

	if file.UserID == "" {
		return errors.New(
			"userID es obligatorio",
		)
	}

	if file.RelativePath == "" {
		return errors.New(
			"relative_path es obligatorio para archivos Home",
		)
	}

	fileType :=
		databaseFileType(
			file.Type,
		)

	unixMode :=
		file.UnixMode

	if unixMode == "" {
		unixMode =
			"0640"
	}

	const query = `
UPDATE archivo
SET
    nombre = $2,
    ruta_relativa = $3,
    tamano = $4,
    mime_type = $5,
    tipo_archivo = $6,
    permisos_unix = $7
WHERE id_archivo = $1::uuid;
`

	result,
		err :=
		r.db.ExecContext(
			ctx,
			query,
			file.ID,
			file.Name,
			file.RelativePath,
			file.Size,
			file.Type,
			fileType,
			unixMode,
		)

	if err != nil {
		return fmt.Errorf(
			"no se pudo actualizar metadata PostgreSQL: %w",
			err,
		)
	}

	rowsAffected,
		err :=
		result.RowsAffected()

	if err != nil {
		return fmt.Errorf(
			"no se pudo verificar actualización PostgreSQL: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return errors.New(
			"archivo no registrado en PostgreSQL",
		)
	}

	return nil
}

// =====================================
// ELIMINAR METADATA
// =====================================

func (r *FileMetadataRepository) Delete(
	ctx context.Context,
	fileID string,
) error {

	if fileID == "" {
		return errors.New(
			"fileID es obligatorio",
		)
	}

	const query = `
DELETE FROM archivo
WHERE id_archivo = $1::uuid;
`

	_,
		err :=
		r.db.ExecContext(
			ctx,
			query,
			fileID,
		)

	if err != nil {
		return fmt.Errorf(
			"no se pudo eliminar metadata PostgreSQL: %w",
			err,
		)
	}

	return nil
}

// =====================================
// COMPROBAR EXISTENCIA
// =====================================

func (r *FileMetadataRepository) Exists(
	ctx context.Context,
	fileID string,
) (bool, error) {

	if fileID == "" {
		return false,
			errors.New(
				"fileID es obligatorio",
			)
	}

	const query = `
SELECT EXISTS (
    SELECT 1
    FROM archivo
    WHERE id_archivo = $1::uuid
);
`

	var exists bool

	if err :=
		r.db.QueryRowContext(
			ctx,
			query,
			fileID,
		).Scan(
			&exists,
		); err != nil {

		return false,
			fmt.Errorf(
				"no se pudo consultar metadata PostgreSQL: %w",
				err,
			)
	}

	return exists, nil
}

// =====================================
// MAPEAR MIME -> tipo_archivo
// =====================================

func databaseFileType(
	mime string,
) string {

	mime =
		strings.ToLower(
			strings.TrimSpace(
				mime,
			),
		)

	switch {

	case strings.HasPrefix(
		mime,
		"image/",
	):
		return "IMAGEN"

	case strings.HasPrefix(
		mime,
		"video/",
	):
		return "VIDEO"

	case strings.HasPrefix(
		mime,
		"text/",
	):
		return "DOCUMENTO"

	case strings.Contains(
		mime,
		"pdf",
	):
		return "DOCUMENTO"

	default:
		return "OTRO"
	}
}
