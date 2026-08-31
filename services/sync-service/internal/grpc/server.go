package grpc

import (
	"context"
	"errors"
	"io"
	"path"
	"strings"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/fileclient"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/service"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	downloadChunkSize = 64 * 1024
)

type Server struct {
	pb.UnimplementedSyncServiceServer

	syncService *service.SyncService
	fileClient  *fileclient.Client
	authClient  *auth.Client
}

func NewServer(
	syncService *service.SyncService,
	fileClient *fileclient.Client,
	authClient *auth.Client,
) *Server {

	return &Server{
		syncService: syncService,
		fileClient:  fileClient,
		authClient:  authClient,
	}
}

// =====================================
// IDENTIDAD AUTENTICADA
// =====================================

func authenticatedIdentity(
	ctx context.Context,
) (auth.Identity, error) {

	identity, ok :=
		auth.IdentityFromContext(
			ctx,
		)

	if !ok ||
		strings.TrimSpace(identity.UserID) == "" {

		return auth.Identity{},
			status.Error(
				codes.Unauthenticated,
				"identidad autenticada no encontrada",
			)
	}

	return identity, nil
}

// =====================================
// NORMALIZAR RELATIVE PATH
// =====================================
//
// Contrato lógico de Sync:
//
// proyecto/codigo/main.go
//
// Se convierte en:
//
// relativePath:
// proyecto/codigo/main.go
//
// relativeDirectory:
// proyecto/codigo
//
// fileName:
// main.go
//
// File Service realizará posteriormente
// la validación definitiva y construirá:
//
// homes/<usuario>/proyecto/codigo/<uuid>-main.go
//
// =====================================

func normalizeUploadRelativePath(
	relativePath string,
	fileName string,
) (
	string,
	string,
	error,
) {

	fileName =
		strings.TrimSpace(
			fileName,
		)

	if fileName == "" {

		return "",
			"",
			errors.New(
				"el nombre del archivo es obligatorio",
			)
	}

	// file_name representa solamente el nombre
	// del archivo, nunca una ruta completa.

	if strings.ContainsAny(
		fileName,
		"/\\",
	) {

		return "",
			"",
			errors.New(
				"file_name no puede contener directorios",
			)
	}

	relativePath =
		strings.TrimSpace(
			relativePath,
		)

	// =====================================
	// COMPATIBILIDAD
	// =====================================
	//
	// Clientes antiguos pueden no enviar
	// relative_path.
	//
	// =====================================

	if relativePath == "" {

		return "",
			"",
			nil
	}

	// =====================================
	// CARACTERES INVÁLIDOS
	// =====================================

	if strings.ContainsRune(
		relativePath,
		'\x00',
	) {

		return "",
			"",
			errors.New(
				"relative_path contiene caracteres inválidos",
			)
	}

	// =====================================
	// NORMALIZAR WINDOWS A POSIX LÓGICO
	// =====================================

	relativePath =
		strings.ReplaceAll(
			relativePath,
			"\\",
			"/",
		)

	// =====================================
	// RUTA ABSOLUTA POSIX
	// =====================================

	if strings.HasPrefix(
		relativePath,
		"/",
	) {

		return "",
			"",
			errors.New(
				"relative_path debe ser relativo",
			)
	}

	// =====================================
	// RUTA ABSOLUTA WINDOWS
	// =====================================
	//
	// Ejemplo:
	//
	// C:/Users/test/file.txt
	//
	// =====================================

	if len(relativePath) >= 3 &&
		relativePath[1] == ':' &&
		relativePath[2] == '/' {

		return "",
			"",
			errors.New(
				"relative_path absoluto de Windows no permitido",
			)
	}

	cleanPath :=
		path.Clean(
			relativePath,
		)

	if cleanPath == "." ||
		cleanPath == "" {

		return "",
			"",
			errors.New(
				"relative_path inválido",
			)
	}

	// =====================================
	// PATH TRAVERSAL
	// =====================================

	if cleanPath == ".." ||
		strings.HasPrefix(
			cleanPath,
			"../",
		) {

		return "",
			"",
			errors.New(
				"relative_path contiene traversal no permitido",
			)
	}

	// =====================================
	// FILE NAME DEBE COINCIDIR
	// =====================================

	pathFileName :=
		path.Base(
			cleanPath,
		)

	if pathFileName !=
		fileName {

		return "",
			"",
			errors.New(
				"relative_path no coincide con file_name",
			)
	}

	relativeDirectory :=
		path.Dir(
			cleanPath,
		)

	if relativeDirectory == "." {

		relativeDirectory =
			""
	}

	return cleanPath,
		relativeDirectory,
		nil
}

// =====================================
// AUTHENTICATE
// =====================================

func (s *Server) Authenticate(
	ctx context.Context,
	request *pb.AuthenticateRequest,
) (*pb.AuthenticateResponse, error) {

	if request == nil {

		return &pb.AuthenticateResponse{
			Success: false,
			Message: "La solicitud es obligatoria",
		}, nil
	}

	if strings.TrimSpace(
		request.Token,
	) == "" {

		return &pb.AuthenticateResponse{
			Success: false,
			Message: "El token es obligatorio",
		}, nil
	}

	if strings.TrimSpace(
		request.DeviceId,
	) == "" {

		return &pb.AuthenticateResponse{
			Success: false,
			Message: "El device ID es obligatorio",
		}, nil
	}

	if s == nil ||
		s.authClient == nil {

		return &pb.AuthenticateResponse{
			Success: false,
			Message: "Auth Client no inicializado",
		}, nil
	}

	result, err :=
		s.authClient.ValidateToken(
			request.Token,
		)

	if err != nil {

		return &pb.AuthenticateResponse{
			Success: false,
			Message: "No fue posible validar el token: " +
				err.Error(),
		}, nil
	}

	if !result.Valid ||
		strings.TrimSpace(
			result.UserID,
		) == "" {

		return &pb.AuthenticateResponse{
			Success: false,
			Message: result.Message,
		}, nil
	}

	return &pb.AuthenticateResponse{
		Success: true,
		Message: result.Message,
		UserId:  result.UserID,
	}, nil
}

// =====================================
// SYNC
// =====================================

func (s *Server) Sync(
	ctx context.Context,
	request *pb.SyncRequest,
) (*pb.SyncResponse, error) {

	identity, err :=
		authenticatedIdentity(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	if s == nil ||
		s.syncService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"Sync Service no inicializado",
			)
	}

	if request == nil {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"la solicitud es obligatoria",
			)
	}

	if strings.TrimSpace(
		request.DeviceId,
	) == "" {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"el device ID es obligatorio",
			)
	}

	// request.UserId se ignora deliberadamente.
	// La identidad válida es la proveniente del token.

	changes, err :=
		s.syncService.Sync(
			identity.UserID,
			request.DeviceId,
		)

	if err != nil {

		return &pb.SyncResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	responseChanges :=
		make(
			[]*pb.FileChange,
			0,
			len(changes),
		)

	for _, change := range changes {

		if change == nil {
			continue
		}

		responseChanges =
			append(
				responseChanges,
				toProtoChange(
					change,
				),
			)
	}

	return &pb.SyncResponse{
		Success: true,
		Message: "Sincronización consultada correctamente",
		Changes: responseChanges,
	}, nil
}

// =====================================
// LIST FILES
// =====================================

func (s *Server) ListFiles(
	ctx context.Context,
	request *pb.ListFilesRequest,
) (*pb.ListFilesResponse, error) {

	identity, err :=
		authenticatedIdentity(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	if s == nil ||
		s.syncService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"Sync Service no inicializado",
			)
	}

	if request == nil {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"la solicitud es obligatoria",
			)
	}

	// request.UserId se ignora deliberadamente.

	files, err :=
		s.syncService.ListFiles(
			identity.UserID,
		)

	if err != nil {

		return &pb.ListFilesResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	responseFiles :=
		make(
			[]*pb.FileMetadata,
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
				toProtoFile(
					file,
				),
			)
	}

	return &pb.ListFilesResponse{
		Success: true,
		Message: "Archivos encontrados correctamente",
		Files:   responseFiles,
	}, nil
}

// =====================================
// UPLOAD
// =====================================
//
// Cliente
//   ↓
// Sync
//   ↓ relative_path
// File Service
//   ↓
// Home central
//
// =====================================

func (s *Server) Upload(
	stream grpc.ClientStreamingServer[
		pb.UploadRequest,
		pb.UploadResponse,
	],
) error {

	identity, err :=
		authenticatedIdentity(
			stream.Context(),
		)

	if err != nil {
		return err
	}

	if s == nil ||
		s.syncService == nil {

		return status.Error(
			codes.Internal,
			"Sync Service no inicializado",
		)
	}

	if s.fileClient == nil {

		return status.Error(
			codes.Internal,
			"File Service Client no inicializado",
		)
	}

	var uploadMetadata *pb.UploadMetadata

	var content []byte

	var bytesReceived int64

	firstMessage :=
		true

	for {

		request, recvErr :=
			stream.Recv()

		if recvErr == io.EOF {
			break
		}

		if recvErr != nil {

			return status.Errorf(
				codes.Internal,
				"error recibiendo upload: %v",
				recvErr,
			)
		}

		if request == nil {

			return status.Error(
				codes.InvalidArgument,
				"se recibió una solicitud vacía",
			)
		}

		// =====================================
		// METADATA
		// =====================================

		if firstMessage {

			uploadMetadata =
				request.GetMetadata()

			if uploadMetadata == nil {

				return status.Error(
					codes.InvalidArgument,
					"el primer mensaje debe contener metadata",
				)
			}

			if strings.TrimSpace(
				uploadMetadata.DeviceId,
			) == "" {

				return status.Error(
					codes.InvalidArgument,
					"el device ID es obligatorio",
				)
			}

			if strings.TrimSpace(
				uploadMetadata.FileName,
			) == "" {

				return status.Error(
					codes.InvalidArgument,
					"el nombre del archivo es obligatorio",
				)
			}

			if uploadMetadata.Size < 0 {

				return status.Error(
					codes.InvalidArgument,
					"el tamaño del archivo no puede ser negativo",
				)
			}

			firstMessage =
				false

			continue
		}

		// =====================================
		// CHUNKS
		// =====================================

		chunk :=
			request.GetChunk()

		if chunk == nil {

			return status.Error(
				codes.InvalidArgument,
				"se esperaba un chunk de contenido",
			)
		}

		content =
			append(
				content,
				chunk...,
			)

		bytesReceived +=
			int64(
				len(
					chunk,
				),
			)

		if bytesReceived >
			uploadMetadata.Size {

			return status.Error(
				codes.InvalidArgument,
				"se recibieron más bytes de los indicados en metadata",
			)
		}
	}

	if uploadMetadata == nil {

		return status.Error(
			codes.InvalidArgument,
			"no se recibieron metadatos",
		)
	}

	if bytesReceived !=
		uploadMetadata.Size {

		return status.Errorf(
			codes.InvalidArgument,
			"tamaño incorrecto: esperados=%d recibidos=%d",
			uploadMetadata.Size,
			bytesReceived,
		)
	}

	// =====================================
	// NORMALIZAR RUTA
	// =====================================

	normalizedRelativePath,
		relativeDirectory,
		pathErr :=
		normalizeUploadRelativePath(
			uploadMetadata.RelativePath,
			uploadMetadata.FileName,
		)

	if pathErr != nil {

		return status.Error(
			codes.InvalidArgument,
			pathErr.Error(),
		)
	}

	// =====================================
	// FILE SERVICE
	// =====================================

	fileResponse, err :=
		s.fileClient.Upload(
			stream.Context(),
			uploadMetadata.FileName,
			uploadMetadata.FileType,
			content,
			relativeDirectory,
		)

	if err != nil {

		return status.Errorf(
			codes.Internal,
			"error cargando archivo en File Service: %v",
			err,
		)
	}

	if fileResponse == nil {

		return status.Error(
			codes.Internal,
			"File Service devolvió una respuesta vacía",
		)
	}

	if !fileResponse.Success {

		return status.Error(
			codes.FailedPrecondition,
			"File Service rechazó el archivo: "+
				fileResponse.Message,
		)
	}

	if strings.TrimSpace(
		fileResponse.FileId,
	) == "" {

		return status.Error(
			codes.Internal,
			"File Service no devolvió un file ID",
		)
	}

	// =====================================
	// REGISTRAR EN SYNC
	// =====================================

	file, err :=
		s.syncService.RegisterFile(
			fileResponse.FileId,
			identity.UserID,
			uploadMetadata.DeviceId,
			uploadMetadata.FileName,
			uploadMetadata.FileType,
			bytesReceived,
			normalizedRelativePath,
		)

	if err != nil {

		_, rollbackErr :=
			s.fileClient.Delete(
				stream.Context(),
				fileResponse.FileId,
			)

		if rollbackErr != nil {

			return status.Errorf(
				codes.Internal,
				"error registrando archivo en Sync: %v; además falló rollback en File Service: %v",
				err,
				rollbackErr,
			)
		}

		return status.Errorf(
			codes.Internal,
			"error registrando archivo en Sync: %v",
			err,
		)
	}

	return stream.SendAndClose(
		&pb.UploadResponse{
			Success: true,
			Message: "Archivo cargado correctamente en Home central",

			FileId: file.FileID,

			BytesReceived: bytesReceived,
		},
	)
}

// =====================================
// DOWNLOAD
// =====================================

func (s *Server) Download(
	request *pb.DownloadRequest,
	stream grpc.ServerStreamingServer[pb.DownloadResponse],
) error {

	identity, err :=
		authenticatedIdentity(
			stream.Context(),
		)

	if err != nil {
		return err
	}

	if s == nil ||
		s.syncService == nil {

		return status.Error(
			codes.Internal,
			"Sync Service no inicializado",
		)
	}

	if s.fileClient == nil {

		return status.Error(
			codes.Internal,
			"File Service Client no inicializado",
		)
	}

	if request == nil {

		return status.Error(
			codes.InvalidArgument,
			"la solicitud es obligatoria",
		)
	}

	if strings.TrimSpace(
		request.FileId,
	) == "" {

		return status.Error(
			codes.InvalidArgument,
			"el file ID es obligatorio",
		)
	}

	// request.UserId se ignora deliberadamente.

	file, err :=
		s.syncService.GetFile(
			request.FileId,
		)

	if err != nil {

		return status.Errorf(
			codes.NotFound,
			"archivo no registrado en Sync: %v",
			err,
		)
	}

	if file.UserID !=
		identity.UserID {

		return status.Error(
			codes.PermissionDenied,
			"el archivo no pertenece al usuario autenticado",
		)
	}

	fileResponse, err :=
		s.fileClient.Get(
			stream.Context(),
			request.FileId,
		)

	if err != nil {

		return status.Errorf(
			codes.Internal,
			"error obteniendo archivo desde File Service: %v",
			err,
		)
	}

	if fileResponse == nil {

		return status.Error(
			codes.Internal,
			"File Service devolvió una respuesta vacía",
		)
	}

	if !fileResponse.Success {

		return status.Error(
			codes.FailedPrecondition,
			"File Service rechazó la descarga: "+
				fileResponse.Message,
		)
	}

	if fileResponse.File == nil {

		return status.Error(
			codes.Internal,
			"File Service no devolvió datos del archivo",
		)
	}

	if fileResponse.File.FileId !=
		request.FileId {

		return status.Error(
			codes.Internal,
			"File Service devolvió un archivo diferente al solicitado",
		)
	}

	if fileResponse.File.UserId !=
		identity.UserID {

		return status.Error(
			codes.PermissionDenied,
			"el propietario devuelto por File Service no coincide con el usuario autenticado",
		)
	}

	content :=
		fileResponse.File.Content

	// =====================================
	// METADATA
	// =====================================

	err =
		stream.Send(
			&pb.DownloadResponse{
				Data: &pb.DownloadResponse_Metadata{
					Metadata: toProtoFile(
						file,
					),
				},
			},
		)

	if err != nil {

		return status.Errorf(
			codes.Internal,
			"error enviando metadata: %v",
			err,
		)
	}

	// =====================================
	// CHUNKS
	// =====================================

	for start := 0; start < len(content); start += downloadChunkSize {

		end :=
			start +
				downloadChunkSize

		if end >
			len(content) {

			end =
				len(content)
		}

		chunk :=
			content[start:end]

		err =
			stream.Send(
				&pb.DownloadResponse{
					Data: &pb.DownloadResponse_Chunk{
						Chunk: chunk,
					},
				},
			)

		if err != nil {

			return status.Errorf(
				codes.Internal,
				"error enviando chunk: %v",
				err,
			)
		}
	}

	return nil
}

// =====================================
// UPDATE
// =====================================
//
// Un Update modifica el contenido del recurso,
// pero conserva:
//
// - file ID
// - nombre lógico
// - relative_path
// - propietario
//
// Flujo:
//
// Cliente
//    ↓
// Sync Service
//    ↓ valida ownership
// File Service
//    ↓ modifica contenido físico
// Sync Repository
//    ↓ incrementa versión
// sync_change
//    ↓ FILE_CHANGED
//
// =====================================

func (s *Server) UpdateFile(
	ctx context.Context,
	request *pb.UpdateFileRequest,
) (*pb.UpdateFileResponse, error) {

	// =====================================
	// IDENTIDAD
	// =====================================

	identity, err :=
		authenticatedIdentity(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	// =====================================
	// DEPENDENCIAS
	// =====================================

	if s == nil ||
		s.syncService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"Sync Service no inicializado",
			)
	}

	if s.fileClient == nil {

		return nil,
			status.Error(
				codes.Internal,
				"File Service Client no inicializado",
			)
	}

	// =====================================
	// REQUEST
	// =====================================

	if request == nil {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"la solicitud es obligatoria",
			)
	}

	request.FileId =
		strings.TrimSpace(
			request.FileId,
		)

	request.DeviceId =
		strings.TrimSpace(
			request.DeviceId,
		)

	if request.FileId == "" {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"el file ID es obligatorio",
			)
	}

	if request.DeviceId == "" {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"el device ID es obligatorio",
			)
	}

	// =====================================
	// ESTADO ACTUAL DE SYNC
	// =====================================

	currentFile, err :=
		s.syncService.GetFile(
			request.FileId,
		)

	if err != nil {

		return &pb.UpdateFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// =====================================
	// ANTI-SPOOF / OWNERSHIP
	// =====================================
	//
	// request.UserId NO es confiable.
	//
	// identity.UserID fue obtenido del token
	// autenticado por el interceptor.
	//
	// =====================================

	if currentFile.UserID !=
		identity.UserID {

		return nil,
			status.Error(
				codes.PermissionDenied,
				"el archivo no pertenece al usuario autenticado",
			)
	}

	// =====================================
	// ACTUALIZAR FILE SERVICE
	// =====================================

	fileResponse, err :=
		s.fileClient.Update(
			ctx,
			request.FileId,
			request.Content,
		)

	if err != nil {

		return nil,
			status.Errorf(
				codes.Internal,
				"error actualizando archivo en File Service: %v",
				err,
			)
	}

	if fileResponse == nil {

		return nil,
			status.Error(
				codes.Internal,
				"File Service devolvió una respuesta vacía",
			)
	}

	if !fileResponse.Success {

		return &pb.UpdateFileResponse{
			Success: false,
			Message: "File Service rechazó actualización: " +
				fileResponse.Message,
		}, nil
	}

	// =====================================
	// ACTUALIZAR ESTADO SYNC
	// =====================================
	//
	// SyncService:
	//
	// version++
	// SaveFile()
	// AddChange(FILE_CHANGED)
	//
	// =====================================

	updatedFile, err :=
		s.syncService.UpdateFile(
			identity.UserID,
			request.DeviceId,
			request.FileId,
			currentFile.FileName,
			currentFile.FileType,
			int64(
				len(
					request.Content,
				),
			),
		)

	if err != nil {

		// File Service ya modificó el contenido.
		//
		// Por tanto aquí existe una posible ventana
		// de inconsistencia si PostgreSQL Sync falla.
		//
		// La estrategia de compensación distribuida
		// se abordará después de validar el flujo
		// funcional de SYNC-5.

		return nil,
			status.Errorf(
				codes.Internal,
				"archivo actualizado en File Service pero falló actualización del estado Sync: %v",
				err,
			)
	}

	// =====================================
	// RESPONSE
	// =====================================

	return &pb.UpdateFileResponse{
		Success: true,

		Message: "Archivo actualizado correctamente",

		FileId: updatedFile.FileID,

		Size: updatedFile.Size,

		Version: updatedFile.Version,
	}, nil
}

// =====================================
// DELETE
// =====================================

func (s *Server) DeleteFile(
	ctx context.Context,
	request *pb.DeleteFileRequest,
) (*pb.DeleteFileResponse, error) {

	identity, err :=
		authenticatedIdentity(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	if s == nil ||
		s.syncService == nil {

		return nil,
			status.Error(
				codes.Internal,
				"Sync Service no inicializado",
			)
	}

	if s.fileClient == nil {

		return nil,
			status.Error(
				codes.Internal,
				"File Service Client no inicializado",
			)
	}

	if request == nil {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"la solicitud es obligatoria",
			)
	}

	request.FileId =
		strings.TrimSpace(
			request.FileId,
		)

	request.DeviceId =
		strings.TrimSpace(
			request.DeviceId,
		)

	if request.FileId == "" {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"el file ID es obligatorio",
			)
	}

	if request.DeviceId == "" {

		return nil,
			status.Error(
				codes.InvalidArgument,
				"el device ID es obligatorio",
			)
	}

	// request.UserId se ignora deliberadamente.

	file, err :=
		s.syncService.GetFile(
			request.FileId,
		)

	if err != nil {

		return &pb.DeleteFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// =====================================
	// OWNERSHIP
	// =====================================

	if file.UserID !=
		identity.UserID {

		return nil,
			status.Error(
				codes.PermissionDenied,
				"el archivo no pertenece al usuario autenticado",
			)
	}

	// =====================================
	// FILE SERVICE
	// =====================================

	fileResponse, err :=
		s.fileClient.Delete(
			ctx,
			request.FileId,
		)

	if err != nil {

		return nil,
			status.Errorf(
				codes.Internal,
				"error eliminando archivo en File Service: %v",
				err,
			)
	}

	if fileResponse == nil {

		return nil,
			status.Error(
				codes.Internal,
				"File Service devolvió una respuesta vacía",
			)
	}

	if !fileResponse.Success {

		return &pb.DeleteFileResponse{
			Success: false,
			Message: "File Service rechazó eliminación: " +
				fileResponse.Message,
		}, nil
	}

	// =====================================
	// ESTADO SYNC
	// =====================================
	//
	// Esta operación NO elimina físicamente
	// sync_file.
	//
	// El repository realiza eliminación lógica:
	//
	// eliminado = true
	// version++
	//
	// y posteriormente registra FILE_DELETED.
	//
	// =====================================

	err =
		s.syncService.DeleteFile(
			identity.UserID,
			request.DeviceId,
			request.FileId,
		)

	if err != nil {

		return nil,
			status.Errorf(
				codes.Internal,
				"archivo eliminado en File Service pero falló actualización del estado Sync: %v",
				err,
			)
	}

	return &pb.DeleteFileResponse{
		Success: true,
		Message: "Archivo eliminado correctamente del Home central",
	}, nil
}

// =====================================
// WATCH CHANGES
// =====================================
//
// WatchChanges utiliza el mismo mecanismo
// persistente de Sync.
//
// Cada llamada a Sync() consulta solamente
// eventos posteriores al cursor del dispositivo.
//
// =====================================

func (s *Server) WatchChanges(
	request *pb.WatchChangesRequest,
	stream grpc.ServerStreamingServer[pb.FileChange],
) error {

	identity, err :=
		authenticatedIdentity(
			stream.Context(),
		)

	if err != nil {
		return err
	}

	if s == nil ||
		s.syncService == nil {

		return status.Error(
			codes.Internal,
			"Sync Service no inicializado",
		)
	}

	if request == nil {

		return status.Error(
			codes.InvalidArgument,
			"la solicitud es obligatoria",
		)
	}

	request.DeviceId =
		strings.TrimSpace(
			request.DeviceId,
		)

	if request.DeviceId == "" {

		return status.Error(
			codes.InvalidArgument,
			"el device ID es obligatorio",
		)
	}

	// request.UserId se ignora deliberadamente.

	ticker :=
		time.NewTicker(
			2 * time.Second,
		)

	defer ticker.Stop()

	for {

		select {

		case <-stream.Context().Done():

			return nil

		case <-ticker.C:

			changes, syncErr :=
				s.syncService.Sync(
					identity.UserID,
					request.DeviceId,
				)

			if syncErr != nil {

				return status.Errorf(
					codes.Internal,
					"error consultando cambios: %v",
					syncErr,
				)
			}

			for _, change := range changes {

				if change == nil {
					continue
				}

				err =
					stream.Send(
						toProtoChange(
							change,
						),
					)

				if err != nil {
					return err
				}
			}
		}
	}
}

// =====================================
// CONVERSIONES
// =====================================

func toProtoFile(
	file *repository.FileMetadata,
) *pb.FileMetadata {

	if file == nil {
		return nil
	}

	return &pb.FileMetadata{
		FileId: file.FileID,

		UserId: file.UserID,

		FileName: file.FileName,

		FileType: file.FileType,

		Size: file.Size,

		CreatedAt: file.CreatedAt.Unix(),

		UpdatedAt: file.UpdatedAt.Unix(),

		RelativePath: file.RelativePath,

		Version: file.Version,
	}
}

func toProtoChange(
	change *repository.FileChange,
) *pb.FileChange {

	if change == nil {
		return nil
	}

	return &pb.FileChange{
		Type: change.Type,

		FileId: change.FileID,

		FileName: change.FileName,

		OriginDeviceId: change.OriginDeviceID,

		Timestamp: change.Timestamp.Unix(),

		RelativePath: change.RelativePath,

		Version: change.Version,
	}
}
