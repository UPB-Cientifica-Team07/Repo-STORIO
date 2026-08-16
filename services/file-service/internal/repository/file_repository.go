package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID        string
	Name      string
	Content   string
	CreatedAt time.Time
}

type FileRepository struct {
	mu    sync.RWMutex
	files map[string]File
}

func NewFileRepository() *FileRepository {
	return &FileRepository{
		files: make(map[string]File),
	}
}

func (r *FileRepository) Save(
	name string,
	content string,
) File {

	r.mu.Lock()
	defer r.mu.Unlock()

	file := File{
		ID:        uuid.NewString(),
		Name:      name,
		Content:   content,
		CreatedAt: time.Now(),
	}

	r.files[file.ID] = file

	return file
}

func (r *FileRepository) Get(
	id string,
) (File, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	file, exists := r.files[id]

	if !exists {
		return File{}, errors.New("archivo no encontrado")
	}

	return file, nil
}

func (r *FileRepository) Delete(
	id string,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.files[id]

	if !exists {
		return errors.New("archivo no encontrado")
	}

	delete(
		r.files,
		id,
	)

	return nil
}

func (r *FileRepository) List() []File {

	r.mu.RLock()
	defer r.mu.RUnlock()

	files := make(
		[]File,
		0,
		len(r.files),
	)

	for _, file := range r.files {
		files = append(
			files,
			file,
		)
	}

	return files
}
