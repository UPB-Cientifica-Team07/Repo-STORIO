package storage

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Storage struct {
	basePath string
	mu       sync.RWMutex
}

func NewStorage(
	basePath string,
) (*Storage, error) {

	if basePath == "" {
		return nil, errors.New(
			"la ruta base de almacenamiento es obligatoria",
		)
	}

	err := os.MkdirAll(
		basePath,
		0755,
	)

	if err != nil {
		return nil, err
	}

	return &Storage{
		basePath: basePath,
	}, nil
}

func (s *Storage) Save(
	fileID string,
	content []byte,
) error {

	if fileID == "" {
		return errors.New(
			"el file ID es obligatorio",
		)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(
		s.basePath,
		fileID,
	)

	err := os.WriteFile(
		path,
		content,
		0644,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) Read(
	fileID string,
) ([]byte, error) {

	if fileID == "" {
		return nil, errors.New(
			"el file ID es obligatorio",
		)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(
		s.basePath,
		fileID,
	)

	content, err := os.ReadFile(
		path,
	)

	if err != nil {

		if os.IsNotExist(err) {
			return nil, errors.New(
				"archivo físico no encontrado",
			)
		}

		return nil, err
	}

	return content, nil
}

func (s *Storage) Delete(
	fileID string,
) error {

	if fileID == "" {
		return errors.New(
			"el file ID es obligatorio",
		)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(
		s.basePath,
		fileID,
	)

	err := os.Remove(
		path,
	)

	if err != nil {

		if os.IsNotExist(err) {
			return errors.New(
				"archivo físico no encontrado",
			)
		}

		return err
	}

	return nil
}

func (s *Storage) Exists(
	fileID string,
) bool {

	if fileID == "" {
		return false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(
		s.basePath,
		fileID,
	)

	_, err := os.Stat(
		path,
	)

	return err == nil
}

func (s *Storage) Path(
	fileID string,
) string {

	return filepath.Join(
		s.basePath,
		fileID,
	)
}
