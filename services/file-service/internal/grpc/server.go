package grpc

import (
	"context"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/auth"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/service"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedFileServiceServer

	fileService       *service.FileService
	permissionService *service.PermissionService
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewServer(
	fileService *service.FileService,
	permissionService *service.PermissionService,
) *Server {

	return &Server{
		fileService:       fileService,
		permissionService: permissionService,
	}
}

// =====================================
// IDENTIDAD
// =====================================

func authenticatedIdentity(
	ctx context.Context,
) (auth.Identity, error) {

	identity,
		ok :=
		auth.IdentityFromContext(
			ctx,
		)

	if !ok ||
		identity.UserID == "" {

		return auth.Identity{},
			status.Error(
				codes.Unauthenticated,
				"identidad autenticada no encontrada",
			)
	}

	return identity, nil
}

// =====================================
// OWNER / ADMIN
// =====================================

func canAccess(
	identity auth.Identity,
	ownerID string,
) bool {

	return identity.UserID == ownerID ||
		identity.Role == "ADMIN"
}

// =====================================
// UPLOAD
// =====================================

func (s *Server) UploadFile(
	ctx context.Context,
	request *pb.UploadFileRequest,
) (*pb.UploadFileResponse, error) {

	identity,
		err :=
		authenticatedIdentity(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	if s.fileService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"File Service no inicializado",
			)
	}

	// =====================================
	// IDENTIDAD REAL
	// =====================================
	//
	// El propietario nunca se toma de
	// request.UserId.
	//
	// La identidad real proviene del token
	// autenticado.
	//
	// =====================================

	file, err :=
		s.fileService.UploadFile(
			identity.UserID,
			request.FileName,
			request.FileType,
			request.Content,
			request.RelativeDirectory,
		)

	if err != nil {

		return &pb.UploadFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.UploadFileResponse{
		Success: true,
		Message: "Archivo almacenado correctamente",
		FileId:  file.ID,
	}, nil
}

// =====================================
// RESTORE
// =====================================

func (s *Server) RestoreFile(
	ctx context.Context,
	request *pb.RestoreFileRequest,
) (*pb.RestoreFileResponse, error) {

	identity,
		err :=
		authenticatedIdentity(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	if s.fileService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"File Service no inicializado",
			)
	}

	ownerID :=
		identity.UserID

	// =====================================
	// ADMIN
	// =====================================
	//
	// Un ADMIN puede restaurar explícitamente
	// para otro propietario.
	//
	// =====================================

	if identity.Role == "ADMIN" &&
		request.UserId != "" {

		ownerID =
			request.UserId
	}

	// =====================================
	// USUARIO NORMAL
	// =====================================
	//
	// Nunca puede restaurar bajo la identidad
	// de otro propietario.
	//
	// =====================================

	if identity.Role != "ADMIN" &&
		request.UserId != "" &&
		request.UserId != identity.UserID {

		return nil,
			status.Error(
				codes.PermissionDenied,
				"no puede restaurar archivos de otro usuario",
			)
	}

	file,
		err :=
		s.fileService.RestoreFile(
			request.FileId,
			ownerID,
			request.FileName,
			request.FileType,
			request.Content,
		)

	if err != nil {

		return &pb.RestoreFileResponse{
			Success: false,
			Message: err.Error(),
			FileId:  request.FileId,
		}, nil
	}

	return &pb.RestoreFileResponse{
		Success: true,
		Message: "Archivo restaurado correctamente",
		FileId:  file.ID,
	}, nil
}

// =====================================
// GET
// =====================================
//
// Puede obtener:
//
// - OWNER
// - ADMIN
// - usuario con can_read=true
//
// =====================================

func (s *Server) GetFile(
	ctx context.Context,
	request *pb.GetFileRequest,
) (*pb.GetFileResponse, error) {

	identity,
		err :=
		authenticatedIdentity(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	if s.fileService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"File Service no inicializado",
			)
	}

	if request.FileId == "" {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"file_id es obligatorio",
			)
	}

	file,
		err :=
		s.fileService.GetFile(
			request.FileId,
		)

	if err != nil {

		return &pb.GetFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// =====================================
	// OWNER / ADMIN
	// =====================================

	if canAccess(
		identity,
		file.UserID,
	) {

		return &pb.GetFileResponse{
			Success: true,
			Message: "Archivo encontrado",
			File:    toProtoFile(file),
		}, nil
	}

	// =====================================
	// ACL DE LECTURA
	// =====================================

	if s.permissionService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"servicio de permisos no inicializado",
			)
	}

	canRead,
		err :=
		s.permissionService.CanRead(
			file.ID,
			identity.UserID,
		)

	if err != nil {

		return nil,
			status.Errorf(
				codes.Internal,
				"no se pudo verificar permiso de lectura: %v",
				err,
			)
	}

	if !canRead {

		return nil,
			status.Error(
				codes.PermissionDenied,
				"no tiene permisos para acceder a este archivo",
			)
	}

	return &pb.GetFileResponse{
		Success: true,
		Message: "Archivo compartido encontrado",
		File:    toProtoFile(file),
	}, nil
}

// =====================================
// UPDATE
// =====================================
//
// Puede modificar:
//
// - OWNER
// - ADMIN
// - usuario con can_write=true
//
// El permiso can_write no transfiere
// propiedad.
//
// =====================================

func (s *Server) UpdateFile(
	ctx context.Context,
	request *pb.UpdateFileRequest,
) (*pb.UpdateFileResponse, error) {

	identity,
		err :=
		authenticatedIdentity(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	if s.fileService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"File Service no inicializado",
			)
	}

	if request.FileId == "" {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"file_id es obligatorio",
			)
	}

	// =====================================
	// OBTENER ARCHIVO
	// =====================================

	file,
		err :=
		s.fileService.GetFile(
			request.FileId,
		)

	if err != nil {

		return &pb.UpdateFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// =====================================
	// OWNER / ADMIN
	// =====================================

	authorized :=
		canAccess(
			identity,
			file.UserID,
		)

	// =====================================
	// ACL DE ESCRITURA
	// =====================================

	if !authorized {

		if s.permissionService == nil {

			return nil,
				status.Error(
					codes.Internal,
					"servicio de permisos no inicializado",
				)
		}

		canWrite,
			permissionErr :=
			s.permissionService.CanWrite(
				file.ID,
				identity.UserID,
			)

		if permissionErr != nil {

			return nil,
				status.Errorf(
					codes.Internal,
					"no se pudo verificar permiso de escritura: %v",
					permissionErr,
				)
		}

		authorized =
			canWrite
	}

	if !authorized {

		return nil,
			status.Error(
				codes.PermissionDenied,
				"no tiene permisos para modificar este archivo",
			)
	}

	// =====================================
	// ACTUALIZAR
	// =====================================

	updatedFile,
		err :=
		s.fileService.UpdateFile(
			request.FileId,
			request.Content,
		)

	if err != nil {

		return &pb.UpdateFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.UpdateFileResponse{
		Success: true,
		Message: "Archivo actualizado correctamente",
		File:    toProtoFile(updatedFile),
	}, nil
}

// =====================================
// DELETE
// =====================================
//
// El permiso can_write NO autoriza borrado.
//
// Delete está reservado a:
//
// - OWNER
// - ADMIN
//
// =====================================

func (s *Server) DeleteFile(
	ctx context.Context,
	request *pb.DeleteFileRequest,
) (*pb.DeleteFileResponse, error) {

	identity,
		err :=
		authenticatedIdentity(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	if s.fileService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"File Service no inicializado",
			)
	}

	if request.FileId == "" {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"file_id es obligatorio",
			)
	}

	file,
		err :=
		s.fileService.GetFile(
			request.FileId,
		)

	if err != nil {

		return &pb.DeleteFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	if !canAccess(
		identity,
		file.UserID,
	) {

		return nil,
			status.Error(
				codes.PermissionDenied,
				"no tiene permisos para eliminar este archivo",
			)
	}

	err =
		s.fileService.DeleteFile(
			request.FileId,
		)

	if err != nil {

		return &pb.DeleteFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.DeleteFileResponse{
		Success: true,
		Message: "Archivo eliminado correctamente",
	}, nil
}

// =====================================
// SHARE FILE
// =====================================
//
// OWNER / ADMIN:
//
// pueden conceder cualquier ACL válida.
//
// DELEGADO:
//
// debe tener can_share=true.
//
// Además solamente puede conceder un
// subconjunto de los permisos que él mismo
// posee.
//
// Esto evita escalamiento de privilegios.
//
// =====================================

func (s *Server) ShareFile(
	ctx context.Context,
	request *pb.ShareFileRequest,
) (*pb.ShareFileResponse, error) {

	identity,
		err :=
		authenticatedIdentity(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	if s.fileService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"File Service no inicializado",
			)
	}

	if request.FileId == "" {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"file_id es obligatorio",
			)
	}

	if request.TargetUserId == "" {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"target_user_id es obligatorio",
			)
	}

	if s.permissionService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"servicio de permisos no inicializado",
			)
	}

	// =====================================
	// OBTENER ARCHIVO
	// =====================================

	file,
		err :=
		s.fileService.GetFile(
			request.FileId,
		)

	if err != nil {

		return &pb.ShareFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// =====================================
	// NO CREAR ACL PARA OWNER
	// =====================================

	if request.TargetUserId ==
		file.UserID {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"el propietario ya posee acceso completo al archivo",
			)
	}

	// =====================================
	// OWNER / ADMIN
	// =====================================

	privileged :=
		canAccess(
			identity,
			file.UserID,
		)

	// =====================================
	// USUARIO DELEGADO
	// =====================================

	if !privileged {

		permission,
			permissionErr :=
			s.permissionService.GetPermission(
				file.ID,
				identity.UserID,
			)

		if permissionErr != nil {

			return nil,
				status.Errorf(
					codes.Internal,
					"no se pudo verificar ACL del usuario: %v",
					permissionErr,
				)
		}

		// =====================================
		// REQUIERE CAN_SHARE
		// =====================================

		if permission == nil ||
			!permission.CanShare {

			return nil,
				status.Error(
					codes.PermissionDenied,
					"no tiene permisos para compartir este archivo",
				)
		}

		// =====================================
		// NORMALIZAR SOLICITUD
		// =====================================
		//
		// Grant normaliza:
		//
		// write -> read
		// share -> read
		//
		// Realizamos la misma normalización
		// antes de verificar anti-escalamiento.
		//
		// =====================================

		requestedRead :=
			request.CanRead ||
				request.CanWrite ||
				request.CanShare

		requestedWrite :=
			request.CanWrite

		requestedShare :=
			request.CanShare

		// =====================================
		// ANTI-ESCALAMIENTO: READ
		// =====================================

		if requestedRead &&
			!permission.CanRead {

			return nil,
				status.Error(
					codes.PermissionDenied,
					"no puede conceder permiso de lectura que no posee",
				)
		}

		// =====================================
		// ANTI-ESCALAMIENTO: WRITE
		// =====================================

		if requestedWrite &&
			!permission.CanWrite {

			return nil,
				status.Error(
					codes.PermissionDenied,
					"no puede conceder permiso de escritura que no posee",
				)
		}

		// =====================================
		// ANTI-ESCALAMIENTO: SHARE
		// =====================================

		if requestedShare &&
			!permission.CanShare {

			return nil,
				status.Error(
					codes.PermissionDenied,
					"no puede conceder permiso de compartición que no posee",
				)
		}
	}

	// =====================================
	// CREAR / ACTUALIZAR ACL
	// =====================================

	err =
		s.permissionService.Grant(
			file.ID,
			request.TargetUserId,
			request.CanRead,
			request.CanWrite,
			request.CanShare,
		)

	if err != nil {

		return &pb.ShareFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.ShareFileResponse{
		Success: true,
		Message: "Permisos del archivo actualizados correctamente",
	}, nil
}

// =====================================
// REVOKE FILE ACCESS
// =====================================
//
// Puede revocar:
//
// - OWNER
// - ADMIN
// - delegado con can_share=true
//
// El propietario nunca puede ser revocado.
//
// =====================================

func (s *Server) RevokeFileAccess(
	ctx context.Context,
	request *pb.RevokeFileAccessRequest,
) (*pb.RevokeFileAccessResponse, error) {

	identity,
		err :=
		authenticatedIdentity(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	if s.fileService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"File Service no inicializado",
			)
	}

	if request.FileId == "" {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"file_id es obligatorio",
			)
	}

	if request.TargetUserId == "" {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"target_user_id es obligatorio",
			)
	}

	if s.permissionService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"servicio de permisos no inicializado",
			)
	}

	// =====================================
	// OBTENER ARCHIVO
	// =====================================

	file,
		err :=
		s.fileService.GetFile(
			request.FileId,
		)

	if err != nil {

		return &pb.RevokeFileAccessResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// =====================================
	// OWNER NO TIENE ACL REVOCABLE
	// =====================================

	if request.TargetUserId ==
		file.UserID {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"no se puede revocar al propietario del archivo",
			)
	}

	// =====================================
	// OWNER / ADMIN
	// =====================================

	authorized :=
		canAccess(
			identity,
			file.UserID,
		)

	// =====================================
	// DELEGADO CON SHARE
	// =====================================

	if !authorized {

		canShare,
			permissionErr :=
			s.permissionService.CanShare(
				file.ID,
				identity.UserID,
			)

		if permissionErr != nil {

			return nil,
				status.Errorf(
					codes.Internal,
					"no se pudo verificar permiso para revocar: %v",
					permissionErr,
				)
		}

		authorized =
			canShare
	}

	if !authorized {

		return nil,
			status.Error(
				codes.PermissionDenied,
				"no tiene permisos para revocar acceso a este archivo",
			)
	}

	// =====================================
	// REVOCAR
	// =====================================

	err =
		s.permissionService.Revoke(
			file.ID,
			request.TargetUserId,
		)

	if err != nil {

		return &pb.RevokeFileAccessResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.RevokeFileAccessResponse{
		Success: true,
		Message: "Acceso al archivo revocado correctamente",
	}, nil
}

// =====================================
// LIST
// =====================================
//
// USUARIO NORMAL:
//
// devuelve:
//
// - archivos propios
// - archivos compartidos mediante ACL
//
// ADMIN:
//
// - sin request.UserId:
//   lista sus propios archivos +
//   compartidos.
//
// - con request.UserId:
//   conserva la funcionalidad administrativa
//   previa y lista únicamente los archivos
//   propios del usuario solicitado.
//
// Los archivos compartidos se obtienen a
// partir de permiso_recurso y después se
// cargan mediante FileService.GetFile(),
// conservando como fuente real el
// repositorio de archivos.
//
// =====================================

func (s *Server) ListFiles(
	ctx context.Context,
	request *pb.ListFilesRequest,
) (*pb.ListFilesResponse, error) {

	identity,
		err :=
		authenticatedIdentity(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	if s.fileService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"File Service no inicializado",
			)
	}

	// =====================================
	// ADMIN LISTANDO OTRO USUARIO
	// =====================================
	//
	// Conservamos exactamente la semántica
	// administrativa anterior.
	//
	// No mezclamos aquí los compartidos del
	// usuario solicitado.
	//
	// =====================================

	if identity.Role == "ADMIN" &&
		request.UserId != "" {

		files :=
			s.fileService.ListFiles(
				request.UserId,
			)

		responseFiles :=
			make(
				[]*pb.FileData,
				0,
				len(files),
			)

		for _, file := range files {

			if file == nil {
				continue
			}

			responseFiles =
				append(
					responseFiles,
					toProtoFile(file),
				)
		}

		return &pb.ListFilesResponse{
			Success: true,
			Message: "Archivos del usuario encontrados correctamente",
			Files:   responseFiles,
		}, nil
	}

	// =====================================
	// USUARIO AUTENTICADO
	// =====================================

	userID :=
		identity.UserID

	// =====================================
	// ARCHIVOS PROPIOS
	// =====================================

	ownedFiles :=
		s.fileService.ListFiles(
			userID,
		)

	// =====================================
	// MAPA DE ARCHIVOS
	// =====================================
	//
	// Se utiliza para:
	//
	// - eliminar duplicados
	// - evitar insertar dos veces el mismo ID
	//
	// =====================================

	filesByID :=
		make(
			map[string]*repository.File,
		)

	for _, file := range ownedFiles {

		if file == nil ||
			file.ID == "" {

			continue
		}

		filesByID[file.ID] =
			file
	}

	// =====================================
	// ACL
	// =====================================

	if s.permissionService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"servicio de permisos no inicializado",
			)
	}

	permissions,
		err :=
		s.permissionService.ListAccessible(
			userID,
		)

	if err != nil {

		return nil,
			status.Errorf(
				codes.Internal,
				"no se pudieron listar archivos compartidos: %v",
				err,
			)
	}

	// =====================================
	// CARGAR ARCHIVOS COMPARTIDOS
	// =====================================

	for _, permission := range permissions {

		if permission == nil ||
			permission.FileID == "" {

			continue
		}

		// =====================================
		// EVITAR DUPLICADO
		// =====================================

		if _, exists :=
			filesByID[permission.FileID]; exists {

			continue
		}

		// =====================================
		// CARGAR ARCHIVO REAL
		// =====================================
		//
		// PostgreSQL determina la ACL.
		//
		// FileService determina el contenido
		// real del archivo.
		//
		// =====================================

		file,
			getErr :=
			s.fileService.GetFile(
				permission.FileID,
			)

		if getErr != nil {

			// =====================================
			// ACL HUÉRFANA / INCONSISTENCIA
			// =====================================
			//
			// Una ACL antigua puede apuntar a un
			// archivo físico ya inexistente.
			//
			// No bloqueamos todo el listado por
			// una única entrada inconsistente.
			//
			// =====================================

			continue
		}

		if file == nil ||
			file.ID == "" {

			continue
		}

		filesByID[file.ID] =
			file
	}

	// =====================================
	// CONSTRUIR RESPUESTA
	// =====================================

	responseFiles :=
		make(
			[]*pb.FileData,
			0,
			len(filesByID),
		)

	// =====================================
	// PRIMERO: ARCHIVOS PROPIOS
	// =====================================
	//
	// Conservamos el orden que ya devuelve
	// FileService.ListFiles().
	//
	// =====================================

	for _, file := range ownedFiles {

		if file == nil ||
			file.ID == "" {

			continue
		}

		storedFile,
			exists :=
			filesByID[file.ID]

		if !exists {
			continue
		}

		responseFiles =
			append(
				responseFiles,
				toProtoFile(storedFile),
			)

		// Una vez agregado lo eliminamos del
		// mapa para garantizar que nunca vuelva
		// a aparecer.
		delete(
			filesByID,
			file.ID,
		)
	}

	// =====================================
	// DESPUÉS: ARCHIVOS COMPARTIDOS
	// =====================================
	//
	// Utilizamos el orden de las ACL obtenido
	// desde PermissionRepository.
	//
	// =====================================

	for _, permission := range permissions {

		if permission == nil ||
			permission.FileID == "" {

			continue
		}

		file,
			exists :=
			filesByID[permission.FileID]

		if !exists {
			continue
		}

		responseFiles =
			append(
				responseFiles,
				toProtoFile(file),
			)

		delete(
			filesByID,
			permission.FileID,
		)
	}

	return &pb.ListFilesResponse{
		Success: true,
		Message: "Archivos propios y compartidos encontrados correctamente",
		Files:   responseFiles,
	}, nil
}

// =====================================
// CONVERSIÓN
// =====================================

func toProtoFile(
	file *repository.File,
) *pb.FileData {

	if file == nil {
		return nil
	}

	return &pb.FileData{
		FileId:   file.ID,
		UserId:   file.UserID,
		FileName: file.Name,
		FileType: file.Type,
		Content:  file.Content,
		Size:     file.Size,

		CreatedAt: file.CreatedAt.Format(
			"2006-01-02 15:04:05",
		),
	}
}
