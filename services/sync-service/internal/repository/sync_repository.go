package repository

import (
	"errors"
	"sync"
	"time"
)

type FileMetadata struct {
	FileID    string
	UserID    string
	FileName  string
	FileType  string
	Size      int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileChange struct {
	Type           string
	FileID         string
	FileName       string
	OriginDeviceID string
	Timestamp      time.Time
}

type SyncRepository struct {
	mu sync.RWMutex

	files map[string]*FileMetadata

	userFiles map[string][]string

	changes map[string][]*FileChange
}

func NewSyncRepository() *SyncRepository {

	return &SyncRepository{
		files:     make(map[string]*FileMetadata),
		userFiles: make(map[string][]string),
		changes:   make(map[string][]*FileChange),
	}
}

// =====================================
// ARCHIVOS
// =====================================

func (r *SyncRepository) SaveFile(
	file *FileMetadata,
) error {

	if file == nil {
		return errors.New("el archivo no puede ser nil")
	}

	if file.FileID == "" {
		return errors.New("el file ID es obligatorio")
	}

	if file.UserID == "" {
		return errors.New("el user ID es obligatorio")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.files[file.FileID]

	r.files[file.FileID] = file

	if !exists {

		r.userFiles[file.UserID] = append(
			r.userFiles[file.UserID],
			file.FileID,
		)
	}

	return nil
}

func (r *SyncRepository) FindFileByID(
	fileID string,
) (*FileMetadata, error) {

	if fileID == "" {
		return nil, errors.New(
			"el file ID es obligatorio",
		)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	file, exists := r.files[fileID]

	if !exists {
		return nil, errors.New(
			"archivo no encontrado",
		)
	}

	return file, nil
}

func (r *SyncRepository) FindFilesByUserID(
	userID string,
) ([]*FileMetadata, error) {

	if userID == "" {
		return nil, errors.New(
			"el user ID es obligatorio",
		)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	fileIDs := r.userFiles[userID]

	files := make(
		[]*FileMetadata,
		0,
		len(fileIDs),
	)

	for _, fileID := range fileIDs {

		file, exists := r.files[fileID]

		if exists {
			files = append(
				files,
				file,
			)
		}
	}

	return files, nil
}

func (r *SyncRepository) DeleteFile(
	fileID string,
) error {

	if fileID == "" {
		return errors.New(
			"el file ID es obligatorio",
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	file, exists := r.files[fileID]

	if !exists {
		return errors.New(
			"archivo no encontrado",
		)
	}

	delete(
		r.files,
		fileID,
	)

	fileIDs := r.userFiles[file.UserID]

	updatedIDs := make(
		[]string,
		0,
		len(fileIDs),
	)

	for _, id := range fileIDs {

		if id != fileID {
			updatedIDs = append(
				updatedIDs,
				id,
			)
		}
	}

	r.userFiles[file.UserID] = updatedIDs

	return nil
}

// =====================================
// CAMBIOS DE SINCRONIZACIÓN
// =====================================

func (r *SyncRepository) AddChange(
	userID string,
	change *FileChange,
) error {

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

	r.mu.Lock()
	defer r.mu.Unlock()

	r.changes[userID] = append(
		r.changes[userID],
		change,
	)

	return nil
}

func (r *SyncRepository) GetChanges(
	userID string,
	deviceID string,
) ([]*FileChange, error) {

	if userID == "" {
		return nil, errors.New(
			"el user ID es obligatorio",
		)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	userChanges := r.changes[userID]

	result := make(
		[]*FileChange,
		0,
		len(userChanges),
	)

	for _, change := range userChanges {

		// El dispositivo que originó el cambio
		// no necesita recibir su propia notificación.
		if deviceID != "" &&
			change.OriginDeviceID == deviceID {

			continue
		}

		result = append(
			result,
			change,
		)
	}

	return result, nil
}
