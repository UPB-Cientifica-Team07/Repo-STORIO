package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// =====================================
// POLÍTICA DE PERMISOS FÍSICOS
// =====================================
//
// Archivos gestionados por File Service:
//
// 0640
//
// owner:
//   read + write
//
// group:
//   read
//
// others:
//   sin acceso
//
// Los usuarios de UPB-CIENTÍFICA son
// identidades lógicas de aplicación.
// Las ACL efectivas se controlan mediante
// permiso_recurso.
//
// =====================================

const (
	DefaultUnixMode = "0640"

	DefaultFileMode os.FileMode = 0640
	DefaultDirMode  os.FileMode = 0750

	// Se mantiene únicamente para la carpeta
	// histórica files/ y evitar cambiar la
	// compatibilidad existente.
	LegacyDirMode os.FileMode = 0755
)

// =====================================
// MODELO
// =====================================

type File struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	RelativePath string    `json:"relative_path,omitempty"`
	UnixMode     string    `json:"unix_mode,omitempty"`
	Content      []byte    `json:"-"`
	Size         int64     `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
}

type FileRepository struct {
	mu sync.RWMutex

	storageDir string
	filesDir   string
	indexPath  string

	files map[string]*File
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewFileRepository(
	storageDir string,
) (*FileRepository, error) {

	if storageDir == "" {
		return nil,
			errors.New(
				"storageDir es obligatorio",
			)
	}

	filesDir :=
		filepath.Join(
			storageDir,
			"files",
		)

	indexPath :=
		filepath.Join(
			storageDir,
			"metadata.json",
		)

	homesDir :=
		filepath.Join(
			storageDir,
			"homes",
		)

	// =====================================
	// STORAGE LEGACY
	// =====================================

	if err :=
		os.MkdirAll(
			filesDir,
			LegacyDirMode,
		); err != nil {

		return nil,
			fmt.Errorf(
				"no se pudo crear almacenamiento legacy: %w",
				err,
			)
	}

	// =====================================
	// STORAGE HOMES
	// =====================================

	if err :=
		os.MkdirAll(
			homesDir,
			DefaultDirMode,
		); err != nil {

		return nil,
			fmt.Errorf(
				"no se pudo crear directorio homes: %w",
				err,
			)
	}

	repository :=
		&FileRepository{
			storageDir: storageDir,
			filesDir:   filesDir,
			indexPath:  indexPath,
			files:      make(map[string]*File),
		}

	if err :=
		repository.loadMetadata(); err != nil {

		return nil, err
	}

	return repository, nil
}

// =====================================
// GUARDAR ARCHIVO
// =====================================

func (r *FileRepository) Save(
	file *File,
) error {

	if file == nil {
		return errors.New(
			"el archivo no puede ser nil",
		)
	}

	if file.ID == "" {
		return errors.New(
			"el ID del archivo es obligatorio",
		)
	}

	if file.UserID == "" {
		return errors.New(
			"el user ID del archivo es obligatorio",
		)
	}

	if file.Name == "" {
		return errors.New(
			"el nombre del archivo es obligatorio",
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists :=
		r.files[file.ID]; exists {

		return errors.New(
			"ya existe un archivo con ese ID",
		)
	}

	contentPath,
		err :=
		r.contentPathForFile(
			file,
		)

	if err != nil {
		return err
	}

	// =====================================
	// CREAR DIRECTORIO FÍSICO
	// =====================================

	if err :=
		os.MkdirAll(
			filepath.Dir(
				contentPath,
			),
			DefaultDirMode,
		); err != nil {

		return fmt.Errorf(
			"no se pudo crear directorio del Home: %w",
			err,
		)
	}

	// =====================================
	// ESCRIBIR ARCHIVO
	// =====================================

	if err :=
		os.WriteFile(
			contentPath,
			file.Content,
			DefaultFileMode,
		); err != nil {

		return fmt.Errorf(
			"no se pudo escribir el archivo físico: %w",
			err,
		)
	}

	// =====================================
	// NORMALIZAR UNIX MODE
	// =====================================

	unixMode :=
		file.UnixMode

	if unixMode == "" {
		unixMode =
			DefaultUnixMode
	}

	metadata :=
		&File{
			ID:           file.ID,
			UserID:       file.UserID,
			Name:         file.Name,
			Type:         file.Type,
			RelativePath: file.RelativePath,
			UnixMode:     unixMode,
			Size:         file.Size,
			CreatedAt:    file.CreatedAt,
		}

	r.files[file.ID] =
		metadata

	// =====================================
	// PERSISTIR metadata.json
	// =====================================

	if err :=
		r.persistMetadata(); err != nil {

		delete(
			r.files,
			file.ID,
		)

		_ =
			os.Remove(
				contentPath,
			)

		return err
	}

	return nil
}

// =====================================
// BUSCAR POR ID
// =====================================

func (r *FileRepository) FindByID(
	fileID string,
) (*File, error) {

	if fileID == "" {
		return nil,
			errors.New(
				"file ID es obligatorio",
			)
	}

	r.mu.RLock()

	metadata,
		exists :=
		r.files[fileID]

	if !exists {
		r.mu.RUnlock()

		return nil,
			errors.New(
				"archivo no encontrado",
			)
	}

	fileCopy :=
		*metadata

	r.mu.RUnlock()

	contentPath,
		err :=
		r.contentPathForFile(
			&fileCopy,
		)

	if err != nil {
		return nil, err
	}

	content,
		err :=
		os.ReadFile(
			contentPath,
		)

	if err != nil {

		if os.IsNotExist(
			err,
		) {

			return nil,
				errors.New(
					"metadata encontrada pero archivo físico inexistente",
				)
		}

		return nil,
			fmt.Errorf(
				"no se pudo leer el archivo físico: %w",
				err,
			)
	}

	fileCopy.Content =
		content

	fileCopy.Size =
		int64(
			len(
				content,
			),
		)

	return &fileCopy, nil
}

// =====================================
// ACTUALIZAR CONTENIDO
// =====================================
//
// Actualiza:
//
// - contenido físico
// - size en metadata.json
//
// Conserva:
//
// - ID
// - propietario
// - nombre
// - tipo MIME
// - ruta relativa
// - permisos Unix
// - CreatedAt
//
// Estrategia:
//
// archivo actual
//   -> .update.bak
//
// contenido nuevo
//   -> .update.tmp
//   -> archivo actual
//
// Si falla metadata.json:
// se restaura el backup.
//
// =====================================

func (r *FileRepository) UpdateContent(
	fileID string,
	newContent []byte,
) (*File, error) {

	if fileID == "" {
		return nil,
			errors.New(
				"file ID es obligatorio",
			)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// =====================================
	// METADATA ACTUAL
	// =====================================

	current,
		exists :=
		r.files[fileID]

	if !exists {
		return nil,
			errors.New(
				"archivo no encontrado",
			)
	}

	oldMetadata :=
		*current

	contentPath,
		err :=
		r.contentPathForFile(
			current,
		)

	if err != nil {
		return nil, err
	}

	// =====================================
	// CONTENIDO ACTUAL
	// =====================================

	oldContent,
		err :=
		os.ReadFile(
			contentPath,
		)

	if err != nil {

		if os.IsNotExist(
			err,
		) {
			return nil,
				errors.New(
					"metadata encontrada pero archivo físico inexistente",
				)
		}

		return nil,
			fmt.Errorf(
				"no se pudo leer contenido actual: %w",
				err,
			)
	}

	// =====================================
	// RUTAS TEMPORALES
	// =====================================

	tempPath :=
		contentPath +
			".update.tmp"

	backupPath :=
		contentPath +
			".update.bak"

	_ =
		os.Remove(
			tempPath,
		)

	_ =
		os.Remove(
			backupPath,
		)

	// =====================================
	// ESCRIBIR NUEVO CONTENIDO TEMPORAL
	// =====================================

	if err :=
		os.WriteFile(
			tempPath,
			newContent,
			DefaultFileMode,
		); err != nil {

		return nil,
			fmt.Errorf(
				"no se pudo escribir contenido temporal: %w",
				err,
			)
	}

	// =====================================
	// CREAR BACKUP
	// =====================================

	if err :=
		os.Rename(
			contentPath,
			backupPath,
		); err != nil {

		_ =
			os.Remove(
				tempPath,
			)

		return nil,
			fmt.Errorf(
				"no se pudo crear backup del archivo: %w",
				err,
			)
	}

	// =====================================
	// CONFIRMAR NUEVO CONTENIDO
	// =====================================

	if err :=
		os.Rename(
			tempPath,
			contentPath,
		); err != nil {

		restoreErr :=
			os.Rename(
				backupPath,
				contentPath,
			)

		if restoreErr != nil {
			return nil,
				fmt.Errorf(
					"falló actualización física (%v) y rollback (%v)",
					err,
					restoreErr,
				)
		}

		return nil,
			fmt.Errorf(
				"no se pudo confirmar nuevo contenido: %w",
				err,
			)
	}

	// =====================================
	// ACTUALIZAR METADATA EN MEMORIA
	// =====================================

	current.Size =
		int64(
			len(
				newContent,
			),
		)

	// =====================================
	// NORMALIZAR UNIX MODE
	// =====================================

	if current.UnixMode == "" {
		current.UnixMode =
			DefaultUnixMode
	}

	// =====================================
	// PERSISTIR metadata.json
	// =====================================

	if err :=
		r.persistMetadata(); err != nil {

		// Restaurar metadata en memoria.
		*current =
			oldMetadata

		// Retirar archivo nuevo.
		_ =
			os.Remove(
				contentPath,
			)

		// Restaurar archivo original.
		restoreErr :=
			os.Rename(
				backupPath,
				contentPath,
			)

		if restoreErr != nil {

			// Último intento utilizando
			// el contenido ya leído.
			writeErr :=
				os.WriteFile(
					contentPath,
					oldContent,
					DefaultFileMode,
				)

			if writeErr != nil {
				return nil,
					fmt.Errorf(
						"falló metadata (%v), rollback rename (%v) y rollback write (%v)",
						err,
						restoreErr,
						writeErr,
					)
			}
		}

		return nil,
			fmt.Errorf(
				"no se pudo persistir metadata actualizada: %w",
				err,
			)
	}

	// =====================================
	// LIMPIAR BACKUP
	// =====================================
	//
	// El estado nuevo ya está confirmado
	// tanto físicamente como en metadata.json.
	//
	// Si la limpieza falla no revertimos
	// una actualización válida.
	//
	// =====================================

	if err :=
		os.Remove(
			backupPath,
		); err != nil {

		fmt.Printf(
			"Advertencia: no se pudo limpiar backup %s: %v\n",
			backupPath,
			err,
		)
	}

	// =====================================
	// RESPUESTA
	// =====================================

	updated :=
		*current

	updated.Content =
		append(
			[]byte(nil),
			newContent...,
		)

	return &updated, nil
}

// =====================================
// LISTAR POR USUARIO
// =====================================

func (r *FileRepository) FindByUserID(
	userID string,
) []*File {

	r.mu.RLock()

	ids :=
		make(
			[]string,
			0,
		)

	for id, file := range r.files {

		if file.UserID ==
			userID {

			ids =
				append(
					ids,
					id,
				)
		}
	}

	r.mu.RUnlock()

	files :=
		make(
			[]*File,
			0,
			len(
				ids,
			),
		)

	for _, id := range ids {

		file,
			err :=
			r.FindByID(
				id,
			)

		if err != nil {
			continue
		}

		files =
			append(
				files,
				file,
			)
	}

	return files
}

// =====================================
// ELIMINAR ARCHIVO
// =====================================

func (r *FileRepository) Delete(
	fileID string,
) error {

	if fileID == "" {
		return errors.New(
			"file ID es obligatorio",
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	file,
		exists :=
		r.files[fileID]

	if !exists {
		return errors.New(
			"archivo no encontrado",
		)
	}

	contentPath,
		err :=
		r.contentPathForFile(
			file,
		)

	if err != nil {
		return err
	}

	tempDeletePath :=
		contentPath +
			".delete.tmp"

	physicalExists :=
		true

	_,
		err =
		os.Stat(
			contentPath,
		)

	if err != nil {

		if os.IsNotExist(
			err,
		) {

			physicalExists =
				false

		} else {

			return fmt.Errorf(
				"no se pudo consultar archivo físico: %w",
				err,
			)
		}
	}

	// =====================================
	// PREPARAR ELIMINACIÓN
	// =====================================

	if physicalExists {

		_ =
			os.Remove(
				tempDeletePath,
			)

		if err :=
			os.Rename(
				contentPath,
				tempDeletePath,
			); err != nil {

			return fmt.Errorf(
				"no se pudo preparar eliminación física: %w",
				err,
			)
		}
	}

	// =====================================
	// ELIMINAR METADATA EN MEMORIA
	// =====================================

	delete(
		r.files,
		fileID,
	)

	// =====================================
	// PERSISTIR metadata.json
	// =====================================

	if err :=
		r.persistMetadata(); err != nil {

		r.files[fileID] =
			file

		if physicalExists {

			if restoreErr :=
				os.Rename(
					tempDeletePath,
					contentPath,
				); restoreErr != nil {

				return fmt.Errorf(
					"falló metadata y también rollback físico: metadata=%v rollback=%v",
					err,
					restoreErr,
				)
			}
		}

		return err
	}

	// =====================================
	// LIMPIAR TOMBSTONE
	// =====================================

	if physicalExists {

		if err :=
			os.Remove(
				tempDeletePath,
			); err != nil {

			return fmt.Errorf(
				"metadata eliminada pero no se pudo limpiar archivo temporal: %w",
				err,
			)
		}
	}

	return nil
}

// =====================================
// CARGAR METADATA
// =====================================

func (r *FileRepository) loadMetadata() error {

	data,
		err :=
		os.ReadFile(
			r.indexPath,
		)

	if err != nil {

		if os.IsNotExist(
			err,
		) {

			return r.persistMetadata()
		}

		return fmt.Errorf(
			"no se pudo leer metadata: %w",
			err,
		)
	}

	if len(
		data,
	) == 0 {

		return nil
	}

	var storedFiles []*File

	if err :=
		json.Unmarshal(
			data,
			&storedFiles,
		); err != nil {

		return fmt.Errorf(
			"metadata inválida: %w",
			err,
		)
	}

	for _, file := range storedFiles {

		if file == nil ||
			file.ID == "" {

			continue
		}

		file.Content =
			nil

		// =====================================
		// NORMALIZACIÓN POSIX
		// =====================================
		//
		// Metadata antigua sin UnixMode se
		// interpreta utilizando la política
		// actual de File Service.
		//
		// Ya no utilizamos 0644 como fallback.
		//
		// =====================================

		if file.UnixMode == "" {
			file.UnixMode =
				DefaultUnixMode
		}

		r.files[file.ID] =
			file
	}

	return nil
}

// =====================================
// PERSISTIR METADATA
// =====================================

func (r *FileRepository) persistMetadata() error {

	files :=
		make(
			[]*File,
			0,
			len(
				r.files,
			),
		)

	for _, file := range r.files {

		files =
			append(
				files,
				file,
			)
	}

	data,
		err :=
		json.MarshalIndent(
			files,
			"",
			"  ",
		)

	if err != nil {
		return fmt.Errorf(
			"no se pudo serializar metadata: %w",
			err,
		)
	}

	tempPath :=
		r.indexPath +
			".tmp"

	if err :=
		os.WriteFile(
			tempPath,
			data,
			DefaultFileMode,
		); err != nil {

		return fmt.Errorf(
			"no se pudo escribir metadata temporal: %w",
			err,
		)
	}

	if err :=
		os.Rename(
			tempPath,
			r.indexPath,
		); err != nil {

		return fmt.Errorf(
			"no se pudo actualizar metadata: %w",
			err,
		)
	}

	return nil
}

// =====================================
// RESOLVER RUTA FÍSICA
// =====================================

func (r *FileRepository) contentPathForFile(
	file *File,
) (string, error) {

	if file == nil {
		return "",
			errors.New(
				"archivo nil",
			)
	}

	// =====================================
	// NUEVO MODELO BASADO EN HOME
	// =====================================

	if file.RelativePath != "" {

		clean :=
			filepath.Clean(
				file.RelativePath,
			)

		if filepath.IsAbs(
			clean,
		) {

			return "",
				errors.New(
					"ruta relativa absoluta no permitida",
				)
		}

		if clean == "." ||
			clean == ".." ||
			strings.HasPrefix(
				clean,
				".."+string(
					os.PathSeparator,
				),
			) {

			return "",
				errors.New(
					"ruta relativa inválida",
				)
		}

		fullPath :=
			filepath.Join(
				r.storageDir,
				clean,
			)

		relativeToStorage,
			err :=
			filepath.Rel(
				r.storageDir,
				fullPath,
			)

		if err != nil {
			return "",
				fmt.Errorf(
					"no se pudo validar ruta física: %w",
					err,
				)
		}

		if relativeToStorage == ".." ||
			strings.HasPrefix(
				relativeToStorage,
				".."+string(
					os.PathSeparator,
				),
			) {

			return "",
				errors.New(
					"ruta fuera de shared-storage no permitida",
				)
		}

		return fullPath, nil
	}

	// =====================================
	// COMPATIBILIDAD LEGACY
	// =====================================

	return filepath.Join(
		r.filesDir,
		file.ID+".bin",
	), nil
}
