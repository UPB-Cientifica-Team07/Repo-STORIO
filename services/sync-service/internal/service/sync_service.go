package service

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/repository"
)

const (
	ChangeTypeCreated = "FILE_CREATED"
	ChangeTypeChanged = "FILE_CHANGED"
	ChangeTypeDeleted = "FILE_DELETED"
)

type SyncService struct {
	repository *repository.SyncRepository
}

func NewSyncService(
	repository *repository.SyncRepository,
) *SyncService {

	return &SyncService{
		repository: repository,
	}
}

// =====================================
// REGISTRAR ARCHIVO
// =====================================

func (s *SyncService) RegisterFile(
	userID string,
	deviceID string,
	fileName string,
	fileType string,
	size int64,
) (*repository.FileMetadata, error) {

	if userID == "" {
		return nil, errors.New(
			"el user ID es obligatorio",
		)
	}

	if deviceID == "" {
		return nil, errors.New(
			"el device ID es obligatorio",
		)
	}

	if fileName == "" {
		return nil, errors.New(
			"el nombre del archivo es obligatorio",
		)
	}

	if size < 0 {
		return nil, errors.New(
			"el tamaño del archivo no puede ser negativo",
		)
	}

	now := time.Now()

	file := &repository.FileMetadata{
		FileID:    uuid.New().String(),
		UserID:    userID,
		FileName:  fileName,
		FileType:  fileType,
		Size:      size,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.repository.SaveFile(
		file,
	)

	if err != nil {
		return nil, err
	}

	change := &repository.FileChange{
		Type:           ChangeTypeCreated,
		FileID:         file.FileID,
		FileName:       file.FileName,
		OriginDeviceID: deviceID,
		Timestamp:      now,
	}

	err = s.repository.AddChange(
		userID,
		change,
	)

	if err != nil {
		return nil, err
	}

	return file, nil
}

// =====================================
// OBTENER ARCHIVO
// =====================================

func (s *SyncService) GetFile(
	fileID string,
) (*repository.FileMetadata, error) {

	if fileID == "" {
		return nil, errors.New(
			"el file ID es obligatorio",
		)
	}

	return s.repository.FindFileByID(
		fileID,
	)
}

// =====================================
// LISTAR ARCHIVOS
// =====================================

func (s *SyncService) ListFiles(
	userID string,
) ([]*repository.FileMetadata, error) {

	if userID == "" {
		return nil, errors.New(
			"el user ID es obligatorio",
		)
	}

	return s.repository.FindFilesByUserID(
		userID,
	)
}

// =====================================
// SINCRONIZAR
// =====================================

func (s *SyncService) Sync(
	userID string,
	deviceID string,
) ([]*repository.FileChange, error) {

	if userID == "" {
		return nil, errors.New(
			"el user ID es obligatorio",
		)
	}

	if deviceID == "" {
		return nil, errors.New(
			"el device ID es obligatorio",
		)
	}

	changes, err := s.repository.GetChanges(
		userID,
		deviceID,
	)

	if err != nil {
		return nil, err
	}

	return changes, nil
}

// =====================================
// ACTUALIZAR ARCHIVO
// =====================================

func (s *SyncService) UpdateFile(
	userID string,
	deviceID string,
	fileID string,
	fileName string,
	fileType string,
	size int64,
) (*repository.FileMetadata, error) {

	if userID == "" {
		return nil, errors.New(
			"el user ID es obligatorio",
		)
	}

	if deviceID == "" {
		return nil, errors.New(
			"el device ID es obligatorio",
		)
	}

	if fileID == "" {
		return nil, errors.New(
			"el file ID es obligatorio",
		)
	}

	file, err := s.repository.FindFileByID(
		fileID,
	)

	if err != nil {
		return nil, err
	}

	if file.UserID != userID {
		return nil, errors.New(
			"el archivo no pertenece al usuario",
		)
	}

	if fileName != "" {
		file.FileName = fileName
	}

	if fileType != "" {
		file.FileType = fileType
	}

	if size >= 0 {
		file.Size = size
	}

	file.UpdatedAt = time.Now()

	err = s.repository.SaveFile(
		file,
	)

	if err != nil {
		return nil, err
	}

	change := &repository.FileChange{
		Type:           ChangeTypeChanged,
		FileID:         file.FileID,
		FileName:       file.FileName,
		OriginDeviceID: deviceID,
		Timestamp:      file.UpdatedAt,
	}

	err = s.repository.AddChange(
		userID,
		change,
	)

	if err != nil {
		return nil, err
	}

	return file, nil
}

// =====================================
// ELIMINAR ARCHIVO
// =====================================

func (s *SyncService) DeleteFile(
	userID string,
	deviceID string,
	fileID string,
) error {

	if userID == "" {
		return errors.New(
			"el user ID es obligatorio",
		)
	}

	if deviceID == "" {
		return errors.New(
			"el device ID es obligatorio",
		)
	}

	if fileID == "" {
		return errors.New(
			"el file ID es obligatorio",
		)
	}

	file, err := s.repository.FindFileByID(
		fileID,
	)

	if err != nil {
		return err
	}

	if file.UserID != userID {
		return errors.New(
			"el archivo no pertenece al usuario",
		)
	}

	err = s.repository.DeleteFile(
		fileID,
	)

	if err != nil {
		return err
	}

	change := &repository.FileChange{
		Type:           ChangeTypeDeleted,
		FileID:         file.FileID,
		FileName:       file.FileName,
		OriginDeviceID: deviceID,
		Timestamp:      time.Now(),
	}

	err = s.repository.AddChange(
		userID,
		change,
	)

	if err != nil {
		return err
	}

	return nil
}
