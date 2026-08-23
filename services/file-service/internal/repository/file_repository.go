package repository

import (
	"errors"
	"sync"
	"time"
)

type File struct {
	ID        string
	UserID    string
	Name      string
	Type      string
	Content   []byte
	Size      int64
	CreatedAt time.Time
}

type FileRepository struct {
	mu    sync.RWMutex
	files map[string]*File
}

func NewFileRepository() *FileRepository {
	return &FileRepository{
		files: make(map[string]*File),
	}
}

func (r *FileRepository) Save(file *File) error {
	if file == nil {
		return errors.New("el archivo no puede ser nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.files[file.ID] = file

	return nil
}

func (r *FileRepository) FindByID(
	fileID string,
) (*File, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	file, exists := r.files[fileID]

	if !exists {
		return nil, errors.New("archivo no encontrado")
	}

	return file, nil
}

func (r *FileRepository) FindByUserID(
	userID string,
) []*File {

	r.mu.RLock()
	defer r.mu.RUnlock()

	var files []*File

	for _, file := range r.files {

		if file.UserID == userID {
			files = append(
				files,
				file,
			)
		}
	}

	return files
}

func (r *FileRepository) Delete(
	fileID string,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.files[fileID]

	if !exists {
		return errors.New("archivo no encontrado")
	}

	delete(
		r.files,
		fileID,
	)

	return nil
}
