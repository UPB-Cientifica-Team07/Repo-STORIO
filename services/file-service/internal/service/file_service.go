package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/repository"
)

type FileService struct {
	repository         *repository.FileRepository
	homeRepository     *repository.HomeRepository
	metadataRepository *repository.FileMetadataRepository

	streamingNotifier func(
		fileID string,
	) error
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewFileService(
	fileRepository *repository.FileRepository,
	homeRepository *repository.HomeRepository,
	metadataRepository *repository.FileMetadataRepository,
) *FileService {

	return &FileService{
		repository:         fileRepository,
		homeRepository:     homeRepository,
		metadataRepository: metadataRepository,
	}
}

// =====================================
// STREAMING NOTIFIER
// =====================================

func (s *FileService) SetStreamingNotifier(
	notifier func(
		fileID string,
	) error,
) {
	s.streamingNotifier = notifier
}

// =====================================
// SUBIR ARCHIVO
// =====================================
//
// relativeDirectory:
//
// "":
//   conserva el comportamiento tradicional
//   de Shared File.
//
//   text/*  -> documentos/
//   image/* -> imagenes/
//   video/* -> videos/
//
// Con valor:
//
//   permite a File Sync conservar la
//   estructura lógica del árbol.
//
// Ejemplo:
//
//   relativeDirectory = "proyecto/codigo"
//
// Ruta final:
//
//   homes/user-003/proyecto/codigo/<uuid>-main.go
//
// IMPORTANTE:
//
// - File Service construye siempre el Home.
// - El cliente nunca puede escoger otro Home.
// - No se permiten rutas absolutas.
// - No se permite "..".
// - La capa Repository realiza además una
//   segunda validación de confinamiento.
//
// =====================================

func (s *FileService) UploadFile(
	userID string,
	name string,
	fileType string,
	content []byte,
	relativeDirectory string,
) (*repository.File, error) {

	if userID == "" {
		return nil, errors.New(
			"el user ID es obligatorio",
		)
	}

	if name == "" {
		return nil, errors.New(
			"el nombre del archivo es obligatorio",
		)
	}

	if s.repository == nil {
		return nil, errors.New(
			"File Repository no inicializado",
		)
	}

	if s.homeRepository == nil {
		return nil, errors.New(
			"Home Repository no inicializado",
		)
	}

	if s.metadataRepository == nil {
		return nil, errors.New(
			"File Metadata Repository no inicializado",
		)
	}

	size :=
		int64(
			len(content),
		)

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	// =====================================
	// RESERVA ATÓMICA DE CUOTA
	// =====================================

	home, err :=
		s.homeRepository.ReserveQuota(
			ctx,
			userID,
			size,
		)

	if err != nil {
		return nil, err
	}

	// =====================================
	// SANITIZAR NOMBRE
	// =====================================

	safeName :=
		sanitizeFileName(
			name,
		)

	if safeName == "" {

		rollbackErr :=
			s.homeRepository.ReleaseQuota(
				ctx,
				userID,
				size,
			)

		if rollbackErr != nil {

			return nil,
				fmt.Errorf(
					"nombre inválido y falló rollback de cuota: %v",
					rollbackErr,
				)
		}

		return nil,
			errors.New(
				"nombre de archivo inválido",
			)
	}

	// =====================================
	// DIRECTORIO DESTINO
	// =====================================

	destinationDirectory :=
		""

	// -------------------------------------
	// FILE SYNC
	// -------------------------------------
	//
	// Cuando Sync envía una ruta lógica,
	// File Service la valida.
	//
	// -------------------------------------

	if strings.TrimSpace(
		relativeDirectory,
	) != "" {

		cleanDirectory,
			directoryErr :=
			sanitizeRelativeDirectory(
				relativeDirectory,
			)

		if directoryErr != nil {

			rollbackErr :=
				s.homeRepository.ReleaseQuota(
					ctx,
					userID,
					size,
				)

			if rollbackErr != nil {

				return nil,
					fmt.Errorf(
						"directorio relativo inválido (%v) y falló rollback de cuota (%v)",
						directoryErr,
						rollbackErr,
					)
			}

			return nil,
				directoryErr
		}

		destinationDirectory =
			cleanDirectory

	} else {

		// -------------------------------------
		// SHARED FILE TRADICIONAL
		// -------------------------------------

		destinationDirectory =
			categoryForMIME(
				fileType,
			)
	}

	// =====================================
	// GENERAR ID
	// =====================================

	fileID :=
		uuid.New().
			String()

	// =====================================
	// CONSTRUIR RUTA
	// =====================================
	//
	// home.BasePath lo determina File Service.
	//
	// Ejemplo:
	//
	// homes/user-003
	// +
	// proyecto/codigo
	// +
	// <uuid>-main.go
	//
	// =====================================

	relativePath :=
		filepath.Join(
			home.BasePath,
			destinationDirectory,
			fileID+"-"+safeName,
		)

	file :=
		&repository.File{
			ID:           fileID,
			UserID:       userID,
			Name:         safeName,
			Type:         fileType,
			RelativePath: relativePath,
			UnixMode:     repository.DefaultUnixMode,
			Content:      content,
			Size:         size,
			CreatedAt:    time.Now(),
		}

	// =====================================
	// FILESYSTEM + metadata.json
	// =====================================

	if err :=
		s.repository.Save(
			file,
		); err != nil {

		rollbackErr :=
			s.homeRepository.ReleaseQuota(
				ctx,
				userID,
				size,
			)

		if rollbackErr != nil {

			return nil,
				fmt.Errorf(
					"falló almacenamiento (%v) y también rollback de cuota (%v)",
					err,
					rollbackErr,
				)
		}

		return nil, err
	}

	// =====================================
	// POSTGRESQL
	// =====================================

	if err :=
		s.metadataRepository.Create(
			ctx,
			file,
		); err != nil {

		deleteErr :=
			s.repository.Delete(
				file.ID,
			)

		quotaErr :=
			s.homeRepository.ReleaseQuota(
				ctx,
				userID,
				size,
			)

		if deleteErr != nil ||
			quotaErr != nil {

			return nil,
				fmt.Errorf(
					"falló metadata PostgreSQL (%v), rollback filesystem=%v, rollback cuota=%v",
					err,
					deleteErr,
					quotaErr,
				)
		}

		return nil,
			fmt.Errorf(
				"falló registro PostgreSQL; operación revertida: %w",
				err,
			)
	}

	// =====================================
	// STREAMING
	// =====================================
	//
	// El archivo ya quedó confirmado en:
	//
	// - Shared Storage
	// - PostgreSQL
	//
	// Para archivos video/* se solicita
	// de manera asíncrona el registro de
	// metadata en Streaming Service.
	//
	// Una caída de Streaming NO revierte
	// una carga válida del File Service.
	//
	// registerVideo es idempotente.
	//
	// =====================================

	if strings.HasPrefix(
		strings.ToLower(
			file.Type,
		),
		"video/",
	) &&
		s.streamingNotifier != nil {

		fileID := file.ID

		go func() {
			_ = s.streamingNotifier(
				fileID,
			)
		}()
	}

	return file, nil
}

// =====================================
// RESTAURAR ARCHIVO
// =====================================

func (s *FileService) RestoreFile(
	fileID string,
	userID string,
	name string,
	fileType string,
	content []byte,
) (*repository.File, error) {

	if fileID == "" {
		return nil,
			errors.New(
				"el file ID es obligatorio",
			)
	}

	if _, err :=
		uuid.Parse(
			fileID,
		); err != nil {

		return nil,
			errors.New(
				"el file ID no es un UUID válido",
			)
	}

	if userID == "" {
		return nil,
			errors.New(
				"el user ID es obligatorio",
			)
	}

	if name == "" {
		return nil,
			errors.New(
				"el nombre del archivo es obligatorio",
			)
	}

	if s.repository == nil {
		return nil,
			errors.New(
				"File Repository no inicializado",
			)
	}

	if s.homeRepository == nil {
		return nil,
			errors.New(
				"Home Repository no inicializado",
			)
	}

	if s.metadataRepository == nil {
		return nil,
			errors.New(
				"File Metadata Repository no inicializado",
			)
	}

	size :=
		int64(
			len(content),
		)

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	home, err :=
		s.homeRepository.ReserveQuota(
			ctx,
			userID,
			size,
		)

	if err != nil {
		return nil, err
	}

	safeName :=
		sanitizeFileName(
			name,
		)

	if safeName == "" {

		rollbackErr :=
			s.homeRepository.ReleaseQuota(
				ctx,
				userID,
				size,
			)

		if rollbackErr != nil {

			return nil,
				fmt.Errorf(
					"nombre inválido y falló rollback de cuota: %v",
					rollbackErr,
				)
		}

		return nil,
			errors.New(
				"nombre de archivo inválido",
			)
	}

	category :=
		categoryForMIME(
			fileType,
		)

	relativePath :=
		filepath.Join(
			home.BasePath,
			category,
			fileID+"-"+safeName,
		)

	file :=
		&repository.File{
			ID:           fileID,
			UserID:       userID,
			Name:         safeName,
			Type:         fileType,
			RelativePath: relativePath,
			UnixMode:     repository.DefaultUnixMode,
			Content:      content,
			Size:         size,
			CreatedAt:    time.Now(),
		}

	if err :=
		s.repository.Save(
			file,
		); err != nil {

		quotaErr :=
			s.homeRepository.ReleaseQuota(
				ctx,
				userID,
				size,
			)

		if quotaErr != nil {

			return nil,
				fmt.Errorf(
					"falló restauración (%v) y rollback cuota (%v)",
					err,
					quotaErr,
				)
		}

		return nil, err
	}

	if err :=
		s.metadataRepository.Create(
			ctx,
			file,
		); err != nil {

		deleteErr :=
			s.repository.Delete(
				file.ID,
			)

		quotaErr :=
			s.homeRepository.ReleaseQuota(
				ctx,
				userID,
				size,
			)

		if deleteErr != nil ||
			quotaErr != nil {

			return nil,
				fmt.Errorf(
					"falló metadata PostgreSQL (%v), rollback filesystem=%v, rollback cuota=%v",
					err,
					deleteErr,
					quotaErr,
				)
		}

		return nil,
			fmt.Errorf(
				"falló restauración PostgreSQL; operación revertida: %w",
				err,
			)
	}

	return file, nil
}

// =====================================
// OBTENER ARCHIVO
// =====================================

func (s *FileService) GetFile(
	fileID string,
) (*repository.File, error) {

	if fileID == "" {

		return nil,
			errors.New(
				"file ID es obligatorio",
			)
	}

	if s.repository == nil {

		return nil,
			errors.New(
				"File Repository no inicializado",
			)
	}

	return s.repository.FindByID(
		fileID,
	)
}

// =====================================
// LISTAR
// =====================================

func (s *FileService) ListFiles(
	userID string,
) []*repository.File {

	if userID == "" ||
		s.repository == nil {

		return []*repository.File{}
	}

	return s.repository.FindByUserID(
		userID,
	)
}

// =====================================
// OBTENER HOME
// =====================================

func (s *FileService) GetHome(
	userID string,
) (*repository.Home, error) {

	if userID == "" {
		return nil,
			errors.New(
				"el user ID es obligatorio",
			)
	}

	if s.homeRepository == nil {
		return nil,
			errors.New(
				"Home Repository no inicializado",
			)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	return s.homeRepository.FindByDirectoryID(
		ctx,
		userID,
	)
}

// =====================================
// ACTUALIZAR CONTENIDO
// =====================================
//
// Esta capa NO decide autorización.
//
// Autorización:
//
// OWNER
// ADMIN
// can_write
//
// se evalúa en la capa gRPC.
//
// =====================================

func (s *FileService) UpdateFile(
	fileID string,
	newContent []byte,
) (*repository.File, error) {

	if fileID == "" {

		return nil,
			errors.New(
				"file ID es obligatorio",
			)
	}

	if _, err :=
		uuid.Parse(
			fileID,
		); err != nil {

		return nil,
			errors.New(
				"el file ID no es un UUID válido",
			)
	}

	if s.repository == nil {

		return nil,
			errors.New(
				"File Repository no inicializado",
			)
	}

	// =====================================
	// ESTADO ORIGINAL
	// =====================================

	oldFile, err :=
		s.repository.FindByID(
			fileID,
		)

	if err != nil {
		return nil, err
	}

	oldSize :=
		oldFile.Size

	newSize :=
		int64(
			len(
				newContent,
			),
		)

	// =====================================
	// ARCHIVO LEGACY
	// =====================================

	if oldFile.RelativePath == "" {

		updatedFile, err :=
			s.repository.UpdateContent(
				fileID,
				newContent,
			)

		if err != nil {
			return nil, err
		}

		return updatedFile, nil
	}

	if s.homeRepository == nil {

		return nil,
			errors.New(
				"Home Repository no inicializado",
			)
	}

	if s.metadataRepository == nil {

		return nil,
			errors.New(
				"File Metadata Repository no inicializado",
			)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	delta :=
		newSize -
			oldSize

	// =====================================
	// CASO 1
	// ARCHIVO CRECE
	// =====================================

	if delta > 0 {

		_, err :=
			s.homeRepository.ReserveQuota(
				ctx,
				oldFile.UserID,
				delta,
			)

		if err != nil {
			return nil, err
		}

		updatedFile, err :=
			s.repository.UpdateContent(
				fileID,
				newContent,
			)

		if err != nil {

			quotaRollbackErr :=
				s.homeRepository.ReleaseQuota(
					ctx,
					oldFile.UserID,
					delta,
				)

			if quotaRollbackErr != nil {

				return nil,
					fmt.Errorf(
						"falló actualización local (%v) y rollback de cuota (%v)",
						err,
						quotaRollbackErr,
					)
			}

			return nil, err
		}

		if err :=
			s.metadataRepository.Update(
				ctx,
				updatedFile,
			); err != nil {

			localRollbackErr :=
				s.rollbackContent(
					fileID,
					oldFile.Content,
				)

			quotaRollbackErr :=
				s.homeRepository.ReleaseQuota(
					ctx,
					oldFile.UserID,
					delta,
				)

			if localRollbackErr != nil ||
				quotaRollbackErr != nil {

				return nil,
					fmt.Errorf(
						"falló actualización PostgreSQL (%v), rollback local=%v, rollback cuota=%v",
						err,
						localRollbackErr,
						quotaRollbackErr,
					)
			}

			return nil,
				fmt.Errorf(
					"falló actualización PostgreSQL; operación revertida: %w",
					err,
				)
		}

		return updatedFile, nil
	}

	// =====================================
	// CASO 2
	// MISMO TAMAÑO
	// =====================================

	if delta == 0 {

		updatedFile, err :=
			s.repository.UpdateContent(
				fileID,
				newContent,
			)

		if err != nil {
			return nil, err
		}

		if err :=
			s.metadataRepository.Update(
				ctx,
				updatedFile,
			); err != nil {

			localRollbackErr :=
				s.rollbackContent(
					fileID,
					oldFile.Content,
				)

			if localRollbackErr != nil {

				return nil,
					fmt.Errorf(
						"falló actualización PostgreSQL (%v) y rollback local (%v)",
						err,
						localRollbackErr,
					)
			}

			return nil,
				fmt.Errorf(
					"falló actualización PostgreSQL; operación revertida: %w",
					err,
				)
		}

		return updatedFile, nil
	}

	// =====================================
	// CASO 3
	// ARCHIVO DISMINUYE
	// =====================================

	releasedBytes :=
		-delta

	updatedFile, err :=
		s.repository.UpdateContent(
			fileID,
			newContent,
		)

	if err != nil {
		return nil, err
	}

	if err :=
		s.metadataRepository.Update(
			ctx,
			updatedFile,
		); err != nil {

		localRollbackErr :=
			s.rollbackContent(
				fileID,
				oldFile.Content,
			)

		if localRollbackErr != nil {

			return nil,
				fmt.Errorf(
					"falló actualización PostgreSQL (%v) y rollback local (%v)",
					err,
					localRollbackErr,
				)
		}

		return nil,
			fmt.Errorf(
				"falló actualización PostgreSQL; operación revertida: %w",
				err,
			)
	}

	if err :=
		s.homeRepository.ReleaseQuota(
			ctx,
			oldFile.UserID,
			releasedBytes,
		); err != nil {

		localRollbackErr :=
			s.rollbackContent(
				fileID,
				oldFile.Content,
			)

		metadataRollbackErr :=
			s.metadataRepository.Update(
				ctx,
				oldFile,
			)

		if localRollbackErr != nil ||
			metadataRollbackErr != nil {

			return nil,
				fmt.Errorf(
					"falló liberación de cuota (%v), rollback local=%v, rollback PostgreSQL=%v",
					err,
					localRollbackErr,
					metadataRollbackErr,
				)
		}

		return nil,
			fmt.Errorf(
				"falló liberación de cuota; actualización revertida: %w",
				err,
			)
	}

	return updatedFile, nil
}

// =====================================
// ROLLBACK DE CONTENIDO
// =====================================

func (s *FileService) rollbackContent(
	fileID string,
	oldContent []byte,
) error {

	if s.repository == nil {

		return errors.New(
			"File Repository no inicializado",
		)
	}

	_, err :=
		s.repository.UpdateContent(
			fileID,
			oldContent,
		)

	return err
}

// =====================================
// ELIMINAR
// =====================================

func (s *FileService) DeleteFile(
	fileID string,
) error {

	if fileID == "" {

		return errors.New(
			"file ID es obligatorio",
		)
	}

	if s.repository == nil {

		return errors.New(
			"File Repository no inicializado",
		)
	}

	file, err :=
		s.repository.FindByID(
			fileID,
		)

	if err != nil {
		return err
	}

	isHomeFile :=
		file.RelativePath != ""

	// =====================================
	// LEGACY
	// =====================================

	if !isHomeFile {

		return s.repository.Delete(
			fileID,
		)
	}

	if s.homeRepository == nil {

		return errors.New(
			"Home Repository no inicializado",
		)
	}

	if s.metadataRepository == nil {

		return errors.New(
			"File Metadata Repository no inicializado",
		)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	// =====================================
	// ELIMINAR FÍSICAMENTE
	// =====================================

	if err :=
		s.repository.Delete(
			fileID,
		); err != nil {

		return err
	}

	// =====================================
	// ELIMINAR POSTGRESQL
	// =====================================

	if err :=
		s.metadataRepository.Delete(
			ctx,
			fileID,
		); err != nil {

		restoreErr :=
			s.repository.Save(
				file,
			)

		if restoreErr != nil {

			return fmt.Errorf(
				"falló DELETE PostgreSQL (%v) y también rollback filesystem (%v)",
				err,
				restoreErr,
			)
		}

		return fmt.Errorf(
			"no se pudo eliminar metadata PostgreSQL; eliminación revertida: %w",
			err,
		)
	}

	// =====================================
	// LIBERAR CUOTA
	// =====================================

	if err :=
		s.homeRepository.ReleaseQuota(
			ctx,
			file.UserID,
			file.Size,
		); err != nil {

		restoreFileErr :=
			s.repository.Save(
				file,
			)

		restoreMetadataErr :=
			s.metadataRepository.Create(
				ctx,
				file,
			)

		if restoreFileErr != nil ||
			restoreMetadataErr != nil {

			return fmt.Errorf(
				"falló liberación de cuota (%v), rollback filesystem=%v, rollback metadata=%v",
				err,
				restoreFileErr,
				restoreMetadataErr,
			)
		}

		return fmt.Errorf(
			"no se pudo liberar cuota; eliminación revertida: %w",
			err,
		)
	}

	return nil
}

// =====================================
// MIME -> DIRECTORIO
// =====================================

func categoryForMIME(
	mime string,
) string {

	mime =
		strings.ToLower(
			strings.TrimSpace(
				mime,
			),
		)

	switch {

	case strings.HasPrefix(
		mime,
		"image/",
	):

		return "imagenes"

	case strings.HasPrefix(
		mime,
		"video/",
	):

		return "videos"

	default:

		return "documentos"
	}
}

// =====================================
// SANITIZAR NOMBRE
// =====================================

func sanitizeFileName(
	name string,
) string {

	name =
		strings.TrimSpace(
			name,
		)

	if name == "" {
		return ""
	}

	safeName :=
		filepath.Base(
			name,
		)

	if safeName == "." ||
		safeName == ".." {

		return ""
	}

	return safeName
}

// =====================================
// SANITIZAR DIRECTORIO RELATIVO
// =====================================
//
// Solo acepta rutas relativas dentro del
// Home lógico.
//
// Válidos:
//
// proyecto
// proyecto/codigo
// datos/2026/agosto
//
// Inválidos:
//
// /etc
// /tmp/test
// ../user-001
// proyecto/../../user-001
//
// =====================================

func sanitizeRelativeDirectory(
	directory string,
) (string, error) {

	directory =
		strings.TrimSpace(
			directory,
		)

	// =====================================
	// VACÍO
	// =====================================

	if directory == "" {
		return "", nil
	}

	// =====================================
	// NORMALIZAR SEPARADORES
	// =====================================
	//
	// Los clientes Windows pueden enviar:
	//
	// proyecto\codigo
	//
	// El contrato lógico usa "/".
	//
	// =====================================

	directory =
		strings.ReplaceAll(
			directory,
			"\\",
			"/",
		)

	clean :=
		filepath.Clean(
			directory,
		)

	// =====================================
	// ABSOLUTA
	// =====================================

	if filepath.IsAbs(
		clean,
	) {

		return "",
			errors.New(
				"directorio relativo absoluto no permitido",
			)
	}

	// =====================================
	// DIRECTORIO ACTUAL
	// =====================================

	if clean == "." {

		return "", nil
	}

	// =====================================
	// PATH TRAVERSAL
	// =====================================

	if clean == ".." ||
		strings.HasPrefix(
			clean,
			".."+string(
				os.PathSeparator,
			),
		) {

		return "",
			errors.New(
				"directorio relativo inválido",
			)
	}

	// =====================================
	// SEGUNDA VALIDACIÓN POR SEGMENTOS
	// =====================================
	//
	// Evita que alguna representación
	// inusual preserve segmentos "..".
	//
	// =====================================

	segments :=
		strings.FieldsFunc(
			filepath.ToSlash(
				clean,
			),
			func(r rune) bool {
				return r == '/'
			},
		)

	if len(segments) == 0 {
		return "", nil
	}

	for _, segment := range segments {

		segment =
			strings.TrimSpace(
				segment,
			)

		if segment == "" ||
			segment == "." {

			continue
		}

		if segment == ".." {

			return "",
				errors.New(
					"directorio relativo contiene traversal no permitido",
				)
		}

		// Evitar segmentos NUL.
		if strings.ContainsRune(
			segment,
			'\x00',
		) {

			return "",
				errors.New(
					"directorio relativo contiene caracteres inválidos",
				)
		}
	}

	// =====================================
	// REPRESENTACIÓN NORMALIZADA
	// =====================================

	return filepath.Join(
		segments...,
	), nil
}
