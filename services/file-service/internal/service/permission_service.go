package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/repository"
)

type PermissionService struct {
	repository *repository.PermissionRepository
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewPermissionService(
	repository *repository.PermissionRepository,
) *PermissionService {

	return &PermissionService{
		repository: repository,
	}
}

// =====================================
// OBTENER PERMISO COMPLETO
// =====================================
//
// Devuelve la ACL explícita asignada a un
// usuario sobre un archivo.
//
// nil significa que no existe ACL.
//
// =====================================

func (s *PermissionService) GetPermission(
	fileID string,
	userID string,
) (*repository.ResourcePermission, error) {

	if fileID == "" {
		return nil,
			errors.New(
				"fileID es obligatorio",
			)
	}

	if userID == "" {
		return nil,
			errors.New(
				"userID es obligatorio",
			)
	}

	if s.repository == nil {
		return nil,
			errors.New(
				"Permission Repository no inicializado",
			)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	permission, err :=
		s.repository.FindPermission(
			ctx,
			fileID,
			userID,
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"no se pudo consultar permiso: %w",
				err,
			)
	}

	return permission, nil
}

// =====================================
// LISTAR PERMISOS ACCESIBLES
// =====================================
//
// Devuelve todas las ACL activas que
// permiten al usuario acceder de alguna
// forma a archivos compartidos.
//
// El PermissionRepository filtra ACL con:
//
// - puede_leer=true
// - puede_escribir=true
// - puede_compartir=true
//
// =====================================

func (s *PermissionService) ListAccessible(
	userID string,
) ([]*repository.ResourcePermission, error) {

	if userID == "" {
		return nil,
			errors.New(
				"userID es obligatorio",
			)
	}

	if s.repository == nil {
		return nil,
			errors.New(
				"Permission Repository no inicializado",
			)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	permissions, err :=
		s.repository.FindAccessibleByUser(
			ctx,
			userID,
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"no se pudieron listar permisos accesibles: %w",
				err,
			)
	}

	return permissions, nil
}

// =====================================
// PUEDE LEER
// =====================================

func (s *PermissionService) CanRead(
	fileID string,
	userID string,
) (bool, error) {

	permission, err :=
		s.GetPermission(
			fileID,
			userID,
		)

	if err != nil {
		return false, err
	}

	if permission == nil {
		return false, nil
	}

	return permission.CanRead, nil
}

// =====================================
// PUEDE ESCRIBIR
// =====================================

func (s *PermissionService) CanWrite(
	fileID string,
	userID string,
) (bool, error) {

	permission, err :=
		s.GetPermission(
			fileID,
			userID,
		)

	if err != nil {
		return false, err
	}

	if permission == nil {
		return false, nil
	}

	return permission.CanWrite, nil
}

// =====================================
// PUEDE COMPARTIR
// =====================================

func (s *PermissionService) CanShare(
	fileID string,
	userID string,
) (bool, error) {

	permission, err :=
		s.GetPermission(
			fileID,
			userID,
		)

	if err != nil {
		return false, err
	}

	if permission == nil {
		return false, nil
	}

	return permission.CanShare, nil
}

// =====================================
// CONCEDER / ACTUALIZAR PERMISO
// =====================================

func (s *PermissionService) Grant(
	fileID string,
	targetUserID string,
	canRead bool,
	canWrite bool,
	canShare bool,
) error {

	if fileID == "" {
		return errors.New(
			"fileID es obligatorio",
		)
	}

	if targetUserID == "" {
		return errors.New(
			"targetUserID es obligatorio",
		)
	}

	if s.repository == nil {
		return errors.New(
			"Permission Repository no inicializado",
		)
	}

	// =====================================
	// ACL VACÍA NO PERMITIDA
	// =====================================
	//
	// Para eliminar completamente una ACL
	// se debe utilizar Revoke().
	//
	// =====================================

	if !canRead &&
		!canWrite &&
		!canShare {

		return errors.New(
			"debe concederse al menos un permiso; use Revoke para eliminar el acceso",
		)
	}

	// =====================================
	// NORMALIZACIÓN DE PERMISOS
	// =====================================
	//
	// Escritura implica lectura.
	//
	// Compartir también implica lectura,
	// porque no tiene sentido permitir
	// delegar un recurso al que no se puede
	// acceder.
	//
	// =====================================

	if canWrite {
		canRead = true
	}

	if canShare {
		canRead = true
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	if err :=
		s.repository.UpsertPermission(
			ctx,
			fileID,
			targetUserID,
			canRead,
			canWrite,
			canShare,
		); err != nil {

		return fmt.Errorf(
			"no se pudo conceder permiso: %w",
			err,
		)
	}

	return nil
}

// =====================================
// REVOCAR PERMISO COMPLETO
// =====================================

func (s *PermissionService) Revoke(
	fileID string,
	targetUserID string,
) error {

	if fileID == "" {
		return errors.New(
			"fileID es obligatorio",
		)
	}

	if targetUserID == "" {
		return errors.New(
			"targetUserID es obligatorio",
		)
	}

	if s.repository == nil {
		return errors.New(
			"Permission Repository no inicializado",
		)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	if err :=
		s.repository.RevokePermission(
			ctx,
			fileID,
			targetUserID,
		); err != nil {

		return fmt.Errorf(
			"no se pudo revocar permiso: %w",
			err,
		)
	}

	return nil
}
