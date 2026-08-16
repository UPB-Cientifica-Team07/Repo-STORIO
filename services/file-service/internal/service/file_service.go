package service

import (
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
	name string,
	content string,
) repository.File {

	return s.repository.Save(
		name,
		content,
	)
}

func (s *FileService) GetFile(
	id string,
) (repository.File, error) {

	return s.repository.Get(id)
}

func (s *FileService) DeleteFile(
	id string,
) error {

	return s.repository.Delete(id)
}

func (s *FileService) ListFiles() []repository.File {

	return s.repository.List()
}
