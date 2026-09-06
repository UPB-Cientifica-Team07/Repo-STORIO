package service

import (
	"errors"
	"strings"
	"time"

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
// REGISTER FILE
// =====================================

func (s *SyncService) RegisterFile(
	fileID string,
	userID string,
	deviceID string,
	fileName string,
	fileType string,
	size int64,
	relativePath string,
) (*repository.FileMetadata, error) {

	if s == nil ||
		s.repository == nil {

		return nil,
			errors.New(
				"Sync Repository no inicializado",
			)
	}

	fileID =
		strings.TrimSpace(fileID)

	userID =
		strings.TrimSpace(userID)

	deviceID =
		strings.TrimSpace(deviceID)

	fileName =
		strings.TrimSpace(fileName)

	fileType =
		strings.TrimSpace(fileType)

	relativePath =
		strings.TrimSpace(relativePath)

	if fileID == "" {
		return nil, errors.New("el file ID es obligatorio")
	}

	if userID == "" {
		return nil, errors.New("el user ID es obligatorio")
	}

	if deviceID == "" {
		return nil, errors.New("el device ID es obligatorio")
	}

	if fileName == "" {
		return nil, errors.New("el nombre del archivo es obligatorio")
	}

	if size < 0 {
		return nil, errors.New("el tamaño del archivo no puede ser negativo")
	}

	now :=
		time.Now()

	file :=
		&repository.FileMetadata{
			FileID:       fileID,
			UserID:       userID,
			FileName:     fileName,
			FileType:     fileType,
			RelativePath: relativePath,
			Size:         size,
			Version:      1,
			Deleted:      false,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

	if err :=
		s.repository.SaveFile(
			file,
		); err != nil {

		return nil, err
	}

	change :=
		&repository.FileChange{
			Type:           ChangeTypeCreated,
			FileID:         file.FileID,
			FileName:       file.FileName,
			RelativePath:   file.RelativePath,
			Version:        file.Version,
			OriginDeviceID: deviceID,
			Timestamp:      now,
		}

	if err :=
		s.repository.AddChange(
			userID,
			change,
		); err != nil {

		rollbackErr :=
			s.repository.RemoveFile(
				file.FileID,
			)

		if rollbackErr != nil {

			return nil,
				errors.New(
					"falló registro del cambio y también rollback de metadata Sync",
				)
		}

		return nil, err
	}

	return file, nil
}

// =====================================
// GET FILE
// =====================================

func (s *SyncService) GetFile(
	fileID string,
) (*repository.FileMetadata, error) {

	if s == nil ||
		s.repository == nil {

		return nil,
			errors.New(
				"Sync Repository no inicializado",
			)
	}

	fileID =
		strings.TrimSpace(
			fileID,
		)

	if fileID == "" {
		return nil, errors.New("el file ID es obligatorio")
	}

	return s.repository.FindFileByID(
		fileID,
	)
}

// =====================================
// LIST FILES
// =====================================

func (s *SyncService) ListFiles(
	userID string,
) ([]*repository.FileMetadata, error) {

	if s == nil ||
		s.repository == nil {

		return nil,
			errors.New(
				"Sync Repository no inicializado",
			)
	}

	userID =
		strings.TrimSpace(
			userID,
		)

	if userID == "" {
		return nil, errors.New("el user ID es obligatorio")
	}

	return s.repository.FindFilesByUserID(
		userID,
	)
}

// =====================================
// SYNC
// =====================================

func (s *SyncService) Sync(
	userID string,
	deviceID string,
) ([]*repository.FileChange, error) {

	if s == nil ||
		s.repository == nil {

		return nil,
			errors.New(
				"Sync Repository no inicializado",
			)
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
		return nil, errors.New("el user ID es obligatorio")
	}

	if deviceID == "" {
		return nil, errors.New("el device ID es obligatorio")
	}

	return s.repository.GetChanges(
		userID,
		deviceID,
	)
}

// =====================================
// ACKNOWLEDGE CHANGES
// =====================================

func (s *SyncService) AcknowledgeChanges(
	userID string,
	deviceID string,
	changeID int64,
) error {

	if s == nil ||
		s.repository == nil {

		return errors.New(
			"Sync Repository no inicializado",
		)
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
		return errors.New(
			"el user ID es obligatorio",
		)
	}

	if deviceID == "" {
		return errors.New(
			"el device ID es obligatorio",
		)
	}

	if changeID <= 0 {
		return errors.New(
			"el change ID debe ser mayor que cero",
		)
	}

	return s.repository.AcknowledgeChanges(
		userID,
		deviceID,
		changeID,
	)
}

// =====================================
// UPDATE FILE
// =====================================

func (s *SyncService) UpdateFile(
	userID string,
	deviceID string,
	fileID string,
	fileName string,
	fileType string,
	size int64,
) (*repository.FileMetadata, error) {

	if s == nil ||
		s.repository == nil {

		return nil,
			errors.New(
				"Sync Repository no inicializado",
			)
	}

	userID =
		strings.TrimSpace(userID)

	deviceID =
		strings.TrimSpace(deviceID)

	fileID =
		strings.TrimSpace(fileID)

	fileName =
		strings.TrimSpace(fileName)

	fileType =
		strings.TrimSpace(fileType)

	if userID == "" {
		return nil, errors.New("el user ID es obligatorio")
	}

	if deviceID == "" {
		return nil, errors.New("el device ID es obligatorio")
	}

	if fileID == "" {
		return nil, errors.New("el file ID es obligatorio")
	}

	if size < 0 {
		return nil, errors.New("el tamaño del archivo no puede ser negativo")
	}

	file, err :=
		s.repository.FindFileByID(
			fileID,
		)

	if err != nil {
		return nil, err
	}

	if file.UserID !=
		userID {

		return nil,
			errors.New(
				"el archivo no pertenece al usuario",
			)
	}

	if fileName != "" {
		file.FileName = fileName
	}

	if fileType != "" {
		file.FileType = fileType
	}

	file.Size =
		size

	file.Version++

	file.UpdatedAt =
		time.Now()

	if err :=
		s.repository.SaveFile(
			file,
		); err != nil {

		return nil, err
	}

	change :=
		&repository.FileChange{
			Type:           ChangeTypeChanged,
			FileID:         file.FileID,
			FileName:       file.FileName,
			RelativePath:   file.RelativePath,
			Version:        file.Version,
			OriginDeviceID: deviceID,
			Timestamp:      file.UpdatedAt,
		}

	if err :=
		s.repository.AddChange(
			userID,
			change,
		); err != nil {

		return nil, err
	}

	return file, nil
}

// =====================================
// DELETE FILE
// =====================================

func (s *SyncService) DeleteFile(
	userID string,
	deviceID string,
	fileID string,
) error {

	if s == nil ||
		s.repository == nil {

		return errors.New(
			"Sync Repository no inicializado",
		)
	}

	userID =
		strings.TrimSpace(userID)

	deviceID =
		strings.TrimSpace(deviceID)

	fileID =
		strings.TrimSpace(fileID)

	if userID == "" {
		return errors.New("el user ID es obligatorio")
	}

	if deviceID == "" {
		return errors.New("el device ID es obligatorio")
	}

	if fileID == "" {
		return errors.New("el file ID es obligatorio")
	}

	file, err :=
		s.repository.FindFileByID(
			fileID,
		)

	if err != nil {
		return err
	}

	if file.UserID !=
		userID {

		return errors.New(
			"el archivo no pertenece al usuario",
		)
	}

	deletedFile, err :=
		s.repository.DeleteFile(
			fileID,
		)

	if err != nil {
		return err
	}

	change :=
		&repository.FileChange{
			Type:           ChangeTypeDeleted,
			FileID:         deletedFile.FileID,
			FileName:       deletedFile.FileName,
			RelativePath:   deletedFile.RelativePath,
			Version:        deletedFile.Version,
			OriginDeviceID: deviceID,
			Timestamp:      deletedFile.UpdatedAt,
		}

	if err :=
		s.repository.AddChange(
			userID,
			change,
		); err != nil {

		return err
	}

	return nil
}
