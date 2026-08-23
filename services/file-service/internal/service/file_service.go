package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/repository"
)

type FileService struct {
	repository *repository.FileRepository
}

func NewFileService(
	repository *repository.FileRepository,
) *FileService {

	return &FileService{
		repository: repository,
	}
}

func (s *FileService) UploadFile(
	userID string,
	name string,
	fileType string,
	content []byte,
) (*repository.File, error) {

	if userID == "" {
		return nil, errors.New("el user ID es obligatorio")
	}

	if name == "" {
		return nil, errors.New("el nombre del archivo es obligatorio")
	}

	file := &repository.File{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      name,
		Type:      fileType,
		Content:   content,
		Size:      int64(len(content)),
		CreatedAt: time.Now(),
	}

	err := s.repository.Save(file)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *FileService) GetFile(
	fileID string,
) (*repository.File, error) {

	return s.repository.FindByID(
		fileID,
	)
}

func (s *FileService) ListFiles(
	userID string,
) []*repository.File {

	return s.repository.FindByUserID(
		userID,
	)
}

func (s *FileService) DeleteFile(
	fileID string,
) error {

	return s.repository.Delete(
		fileID,
	)
}
