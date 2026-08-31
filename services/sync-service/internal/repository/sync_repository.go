package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// =====================================
// FILE METADATA
// =====================================

type FileMetadata struct {
	FileID       string
	UserID       string
	FileName     string
	FileType     string
	RelativePath string
	Size         int64
	Version      int64
	Deleted      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// =====================================
// FILE CHANGE
// =====================================

type FileChange struct {
	ChangeID       int64
	Type           string
	FileID         string
	FileName       string
	RelativePath   string
	Version        int64
	OriginDeviceID string
	Timestamp      time.Time
}

// =====================================
// SYNC REPOSITORY
// =====================================
//
// El estado de Sync es persistente.
//
// Ya NO se utilizan:
//
//	map[string]*FileMetadata
//	map[string][]*FileChange
//
// PostgreSQL es ahora la fuente de verdad
// para metadata lógica, cambios y cursores.
//
// =====================================

type SyncRepository struct {
	db *sql.DB
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewSyncRepository(
	db *sql.DB,
) *SyncRepository {

	return &SyncRepository{
		db: db,
	}
}

// =====================================
// VALIDACIÓN DB
// =====================================

func (r *SyncRepository) validateDB() error {

	if r == nil ||
		r.db == nil {

		return errors.New(
			"Sync Repository PostgreSQL no inicializado",
		)
	}

	return nil
}

// =====================================
// GUARDAR ARCHIVO
// =====================================

func (r *SyncRepository) SaveFile(
	file *FileMetadata,
) error {

	if err :=
		r.validateDB(); err != nil {

		return err
	}

	if file == nil {

		return errors.New(
			"el archivo no puede ser nil",
		)
	}

	file.FileID =
		strings.TrimSpace(
			file.FileID,
		)

	file.UserID =
		strings.TrimSpace(
			file.UserID,
		)

	file.FileName =
		strings.TrimSpace(
			file.FileName,
		)

	file.FileType =
		strings.TrimSpace(
			file.FileType,
		)

	file.RelativePath =
		strings.TrimSpace(
			file.RelativePath,
		)

	if file.FileID == "" {

		return errors.New(
			"el file ID es obligatorio",
		)
	}

	if file.UserID == "" {

		return errors.New(
			"el user ID es obligatorio",
		)
	}

	if file.FileName == "" {

		return errors.New(
			"el nombre del archivo es obligatorio",
		)
	}

	if file.Size < 0 {

		return errors.New(
			"el tamaño del archivo no puede ser negativo",
		)
	}

	if file.Version <= 0 {

		file.Version =
			1
	}

	if file.CreatedAt.IsZero() {

		file.CreatedAt =
			time.Now()
	}

	if file.UpdatedAt.IsZero() {

		file.UpdatedAt =
			file.CreatedAt
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	result, err :=
		r.db.ExecContext(
			ctx,
			`
			INSERT INTO sync_file (
				id_archivo,
				id_usuario,
				nombre,
				mime_type,
				relative_path,
				tamano,
				version,
				eliminado,
				fecha_creacion,
				fecha_modificacion
			)
			SELECT
				$1::uuid,
				u.id_usuario,
				$2,
				NULLIF($3, ''),
				NULLIF($4, ''),
				$5,
				$6,
				$7,
				$8,
				$9
			FROM usuario u
			WHERE u.directorio_id = $10

			ON CONFLICT (id_archivo)
			DO UPDATE SET
				nombre = EXCLUDED.nombre,
				mime_type = EXCLUDED.mime_type,
				relative_path = EXCLUDED.relative_path,
				tamano = EXCLUDED.tamano,
				version = EXCLUDED.version,
				eliminado = EXCLUDED.eliminado,
				fecha_modificacion = EXCLUDED.fecha_modificacion

			WHERE sync_file.id_usuario = EXCLUDED.id_usuario
			`,
			file.FileID,
			file.FileName,
			file.FileType,
			file.RelativePath,
			file.Size,
			file.Version,
			file.Deleted,
			file.CreatedAt,
			file.UpdatedAt,
			file.UserID,
		)

	if err != nil {

		return fmt.Errorf(
			"no se pudo persistir metadata Sync: %w",
			err,
		)
	}

	rowsAffected, err :=
		result.RowsAffected()

	if err != nil {

		return fmt.Errorf(
			"no se pudo verificar persistencia Sync: %w",
			err,
		)
	}

	if rowsAffected == 0 {

		return errors.New(
			"usuario no encontrado o intento de cambiar propietario del archivo",
		)
	}

	return nil
}

// =====================================
// BUSCAR ARCHIVO ACTIVO
// =====================================

func (r *SyncRepository) FindFileByID(
	fileID string,
) (*FileMetadata, error) {

	if err :=
		r.validateDB(); err != nil {

		return nil, err
	}

	fileID =
		strings.TrimSpace(
			fileID,
		)

	if fileID == "" {

		return nil,
			errors.New(
				"el file ID es obligatorio",
			)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	file :=
		&FileMetadata{}

	var (
		fileType     sql.NullString
		relativePath sql.NullString
	)

	err :=
		r.db.QueryRowContext(
			ctx,
			`
			SELECT
				sf.id_archivo::text,
				u.directorio_id,
				sf.nombre,
				sf.mime_type,
				sf.relative_path,
				sf.tamano,
				sf.version,
				sf.eliminado,
				sf.fecha_creacion,
				sf.fecha_modificacion
			FROM sync_file sf
			INNER JOIN usuario u
				ON u.id_usuario = sf.id_usuario
			WHERE
				sf.id_archivo = $1::uuid
				AND sf.eliminado = FALSE
			`,
			fileID,
		).Scan(
			&file.FileID,
			&file.UserID,
			&file.FileName,
			&fileType,
			&relativePath,
			&file.Size,
			&file.Version,
			&file.Deleted,
			&file.CreatedAt,
			&file.UpdatedAt,
		)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {

		return nil,
			errors.New(
				"archivo no encontrado",
			)
	}

	if err != nil {

		return nil,
			fmt.Errorf(
				"error consultando archivo Sync: %w",
				err,
			)
	}

	if fileType.Valid {

		file.FileType =
			fileType.String
	}

	if relativePath.Valid {

		file.RelativePath =
			relativePath.String
	}

	return file, nil
}

// =====================================
// LISTAR ARCHIVOS ACTIVOS
// =====================================

func (r *SyncRepository) FindFilesByUserID(
	userID string,
) ([]*FileMetadata, error) {

	if err :=
		r.validateDB(); err != nil {

		return nil, err
	}

	userID =
		strings.TrimSpace(
			userID,
		)

	if userID == "" {

		return nil,
			errors.New(
				"el user ID es obligatorio",
			)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	rows, err :=
		r.db.QueryContext(
			ctx,
			`
			SELECT
				sf.id_archivo::text,
				u.directorio_id,
				sf.nombre,
				sf.mime_type,
				sf.relative_path,
				sf.tamano,
				sf.version,
				sf.eliminado,
				sf.fecha_creacion,
				sf.fecha_modificacion
			FROM sync_file sf
			INNER JOIN usuario u
				ON u.id_usuario = sf.id_usuario
			WHERE
				u.directorio_id = $1
				AND sf.eliminado = FALSE
			ORDER BY
				sf.relative_path NULLS LAST,
				sf.nombre,
				sf.id_archivo
			`,
			userID,
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"error listando archivos Sync: %w",
				err,
			)
	}

	defer rows.Close()

	files :=
		make(
			[]*FileMetadata,
			0,
		)

	for rows.Next() {

		file :=
			&FileMetadata{}

		var (
			fileType     sql.NullString
			relativePath sql.NullString
		)

		err :=
			rows.Scan(
				&file.FileID,
				&file.UserID,
				&file.FileName,
				&fileType,
				&relativePath,
				&file.Size,
				&file.Version,
				&file.Deleted,
				&file.CreatedAt,
				&file.UpdatedAt,
			)

		if err != nil {

			return nil,
				fmt.Errorf(
					"error leyendo archivo Sync: %w",
					err,
				)
		}

		if fileType.Valid {

			file.FileType =
				fileType.String
		}

		if relativePath.Valid {

			file.RelativePath =
				relativePath.String
		}

		files =
			append(
				files,
				file,
			)
	}

	if err :=
		rows.Err(); err != nil {

		return nil,
			fmt.Errorf(
				"error recorriendo archivos Sync: %w",
				err,
			)
	}

	return files, nil
}

// =====================================
// ELIMINACIÓN LÓGICA
// =====================================
//
// El archivo no desaparece de sync_file.
//
// Esto permite mantener:
// - versión
// - historial
// - estado después de reinicios
//
// File Service sí elimina el contenido físico.
//
// =====================================

func (r *SyncRepository) DeleteFile(
	fileID string,
) (*FileMetadata, error) {

	if err :=
		r.validateDB(); err != nil {

		return nil, err
	}

	fileID =
		strings.TrimSpace(
			fileID,
		)

	if fileID == "" {

		return nil,
			errors.New(
				"el file ID es obligatorio",
			)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	deleted :=
		&FileMetadata{}

	var (
		fileType     sql.NullString
		relativePath sql.NullString
	)

	err :=
		r.db.QueryRowContext(
			ctx,
			`
			UPDATE sync_file sf
			SET
				eliminado = TRUE,
				version = version + 1,
				fecha_modificacion = CURRENT_TIMESTAMP
			FROM usuario u
			WHERE
				sf.id_archivo = $1::uuid
				AND sf.id_usuario = u.id_usuario
				AND sf.eliminado = FALSE
			RETURNING
				sf.id_archivo::text,
				u.directorio_id,
				sf.nombre,
				sf.mime_type,
				sf.relative_path,
				sf.tamano,
				sf.version,
				sf.eliminado,
				sf.fecha_creacion,
				sf.fecha_modificacion
			`,
			fileID,
		).Scan(
			&deleted.FileID,
			&deleted.UserID,
			&deleted.FileName,
			&fileType,
			&relativePath,
			&deleted.Size,
			&deleted.Version,
			&deleted.Deleted,
			&deleted.CreatedAt,
			&deleted.UpdatedAt,
		)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {

		return nil,
			errors.New(
				"archivo no encontrado",
			)
	}

	if err != nil {

		return nil,
			fmt.Errorf(
				"error eliminando estado Sync: %w",
				err,
			)
	}

	if fileType.Valid {

		deleted.FileType =
			fileType.String
	}

	if relativePath.Valid {

		deleted.RelativePath =
			relativePath.String
	}

	return deleted, nil
}

// =====================================
// ELIMINACIÓN FÍSICA DE METADATA SYNC
// =====================================
//
// Solo se usa como compensación cuando se
// está registrando un archivo nuevo y falla
// antes de completar el proceso.
//
// NO elimina el archivo físico de File Service.
//
// =====================================

func (r *SyncRepository) RemoveFile(
	fileID string,
) error {

	if err :=
		r.validateDB(); err != nil {

		return err
	}

	fileID =
		strings.TrimSpace(
			fileID,
		)

	if fileID == "" {

		return errors.New(
			"el file ID es obligatorio",
		)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	_, err :=
		r.db.ExecContext(
			ctx,
			`
			DELETE FROM sync_file
			WHERE id_archivo = $1::uuid
			`,
			fileID,
		)

	if err != nil {

		return fmt.Errorf(
			"error removiendo metadata Sync: %w",
			err,
		)
	}

	return nil
}

// =====================================
// AGREGAR CAMBIO
// =====================================

func (r *SyncRepository) AddChange(
	userID string,
	change *FileChange,
) error {

	if err :=
		r.validateDB(); err != nil {

		return err
	}

	userID =
		strings.TrimSpace(
			userID,
		)

	if userID == "" {

		return errors.New(
			"el user ID es obligatorio",
		)
	}

	if change == nil {

		return errors.New(
			"el cambio no puede ser nil",
		)
	}

	change.Type =
		strings.TrimSpace(
			change.Type,
		)

	change.FileID =
		strings.TrimSpace(
			change.FileID,
		)

	change.FileName =
		strings.TrimSpace(
			change.FileName,
		)

	change.RelativePath =
		strings.TrimSpace(
			change.RelativePath,
		)

	change.OriginDeviceID =
		strings.TrimSpace(
			change.OriginDeviceID,
		)

	if change.Type == "" {

		return errors.New(
			"el tipo de cambio es obligatorio",
		)
	}

	if change.FileID == "" {

		return errors.New(
			"el file ID del cambio es obligatorio",
		)
	}

	if change.FileName == "" {

		return errors.New(
			"el nombre del cambio es obligatorio",
		)
	}

	if change.OriginDeviceID == "" {

		return errors.New(
			"el dispositivo origen es obligatorio",
		)
	}

	if change.Version <= 0 {

		return errors.New(
			"la versión del cambio debe ser mayor que cero",
		)
	}

	if change.Timestamp.IsZero() {

		change.Timestamp =
			time.Now()
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	result, err :=
		r.db.ExecContext(
			ctx,
			`
			INSERT INTO sync_change (
				id_usuario,
				id_archivo,
				tipo,
				nombre,
				relative_path,
				version,
				origin_device_id,
				fecha_cambio
			)
			SELECT
				u.id_usuario,
				$1::uuid,
				$2,
				$3,
				NULLIF($4, ''),
				$5,
				$6,
				$7
			FROM usuario u
			WHERE u.directorio_id = $8
			`,
			change.FileID,
			change.Type,
			change.FileName,
			change.RelativePath,
			change.Version,
			change.OriginDeviceID,
			change.Timestamp,
			userID,
		)

	if err != nil {

		return fmt.Errorf(
			"error persistiendo cambio Sync: %w",
			err,
		)
	}

	rowsAffected, err :=
		result.RowsAffected()

	if err != nil {

		return fmt.Errorf(
			"error verificando cambio Sync: %w",
			err,
		)
	}

	if rowsAffected == 0 {

		return errors.New(
			"usuario no encontrado al registrar cambio Sync",
		)
	}

	return nil
}

// =====================================
// OBTENER CAMBIOS PENDIENTES
// =====================================
//
// Cada usuario/dispositivo posee un cursor.
//
// Solo se consultan cambios:
//
//	id_cambio > last_change_id
//
// Los cambios creados por el mismo dispositivo
// no se devuelven, pero SÍ hacen avanzar su
// cursor.
//
// De esta forma no son revisados infinitamente.
//
// =====================================

func (r *SyncRepository) GetChanges(
	userID string,
	deviceID string,
) ([]*FileChange, error) {

	if err :=
		r.validateDB(); err != nil {

		return nil, err
	}

	userID =
		strings.TrimSpace(
			userID,
		)

	deviceID =
		strings.TrimSpace(
			deviceID,
		)

	if userID == "" {

		return nil,
			errors.New(
				"el user ID es obligatorio",
			)
	}

	if deviceID == "" {

		return nil,
			errors.New(
				"el device ID es obligatorio",
			)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

	defer cancel()

	tx, err :=
		r.db.BeginTx(
			ctx,
			&sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
			},
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"no se pudo iniciar transacción Sync: %w",
				err,
			)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	var (
		internalUserID string
		lastChangeID   int64
	)

	err =
		tx.QueryRowContext(
			ctx,
			`
			SELECT
				id_usuario::text
			FROM usuario
			WHERE directorio_id = $1
			`,
			userID,
		).Scan(
			&internalUserID,
		)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {

		return nil,
			errors.New(
				"usuario no encontrado",
			)
	}

	if err != nil {

		return nil,
			fmt.Errorf(
				"error resolviendo usuario Sync: %w",
				err,
			)
	}

	// Crear cursor en cero si el dispositivo
	// todavía nunca ha sincronizado.
	_, err =
		tx.ExecContext(
			ctx,
			`
			INSERT INTO sync_device_cursor (
				id_usuario,
				device_id,
				last_change_id
			)
			VALUES (
				$1::uuid,
				$2,
				0
			)
			ON CONFLICT (
				id_usuario,
				device_id
			)
			DO NOTHING
			`,
			internalUserID,
			deviceID,
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"error inicializando cursor Sync: %w",
				err,
			)
	}

	err =
		tx.QueryRowContext(
			ctx,
			`
			SELECT
				last_change_id
			FROM sync_device_cursor
			WHERE
				id_usuario = $1::uuid
				AND device_id = $2
			FOR UPDATE
			`,
			internalUserID,
			deviceID,
		).Scan(
			&lastChangeID,
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"error consultando cursor Sync: %w",
				err,
			)
	}

	rows, err :=
		tx.QueryContext(
			ctx,
			`
			SELECT
				id_cambio,
				tipo,
				id_archivo::text,
				nombre,
				relative_path,
				version,
				origin_device_id,
				fecha_cambio
			FROM sync_change
			WHERE
				id_usuario = $1::uuid
				AND id_cambio > $2
			ORDER BY id_cambio ASC
			`,
			internalUserID,
			lastChangeID,
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"error consultando cambios Sync: %w",
				err,
			)
	}

	defer rows.Close()

	changes :=
		make(
			[]*FileChange,
			0,
		)

	maxChangeID :=
		lastChangeID

	for rows.Next() {

		change :=
			&FileChange{}

		var relativePath sql.NullString

		err :=
			rows.Scan(
				&change.ChangeID,
				&change.Type,
				&change.FileID,
				&change.FileName,
				&relativePath,
				&change.Version,
				&change.OriginDeviceID,
				&change.Timestamp,
			)

		if err != nil {

			return nil,
				fmt.Errorf(
					"error leyendo cambio Sync: %w",
					err,
				)
		}

		if relativePath.Valid {

			change.RelativePath =
				relativePath.String
		}

		if change.ChangeID >
			maxChangeID {

			maxChangeID =
				change.ChangeID
		}

		// El dispositivo que originó el cambio
		// no necesita recibir su propio evento.
		if change.OriginDeviceID ==
			deviceID {

			continue
		}

		changes =
			append(
				changes,
				change,
			)
	}

	if err :=
		rows.Err(); err != nil {

		return nil,
			fmt.Errorf(
				"error recorriendo cambios Sync: %w",
				err,
			)
	}

	if maxChangeID >
		lastChangeID {

		_, err =
			tx.ExecContext(
				ctx,
				`
				UPDATE sync_device_cursor
				SET
					last_change_id = $1,
					fecha_actualizacion = CURRENT_TIMESTAMP
				WHERE
					id_usuario = $2::uuid
					AND device_id = $3
				`,
				maxChangeID,
				internalUserID,
				deviceID,
			)

		if err != nil {

			return nil,
				fmt.Errorf(
					"error actualizando cursor Sync: %w",
					err,
				)
		}
	}

	if err :=
		tx.Commit(); err != nil {

		return nil,
			fmt.Errorf(
				"error confirmando cursor Sync: %w",
				err,
			)
	}

	return changes, nil
}
