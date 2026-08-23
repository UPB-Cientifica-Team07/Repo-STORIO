package grpc

import (
	"context"
	"io"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/service"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/storage"

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
	storage     *storage.Storage
	authClient  *auth.Client
}

func NewServer(
	syncService *service.SyncService,
	fileStorage *storage.Storage,
	authClient *auth.Client,
) *Server {

	return &Server{
		syncService: syncService,
		storage:     fileStorage,
		authClient:  authClient,
	}
}

// =====================================
// AUTHENTICATE
// =====================================

func (s *Server) Authenticate(
	ctx context.Context,
	request *pb.AuthenticateRequest,
) (*pb.AuthenticateResponse, error) {

	if request.Token == "" {
		return &pb.AuthenticateResponse{
			Success: false,
			Message: "El token es obligatorio",
		}, nil
	}

	if request.DeviceId == "" {
		return &pb.AuthenticateResponse{
			Success: false,
			Message: "El device ID es obligatorio",
		}, nil
	}

	result, err := s.authClient.ValidateToken(
		request.Token,
	)

	if err != nil {
		return &pb.AuthenticateResponse{
			Success: false,
			Message: "No fue posible validar el token: " + err.Error(),
		}, nil
	}

	if !result.Valid {
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

	changes, err := s.syncService.Sync(
		request.UserId,
		request.DeviceId,
	)

	if err != nil {
		return &pb.SyncResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	responseChanges := make(
		[]*pb.FileChange,
		0,
		len(changes),
	)

	for _, change := range changes {
		responseChanges = append(
			responseChanges,
			toProtoChange(change),
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

	files, err := s.syncService.ListFiles(
		request.UserId,
	)

	if err != nil {
		return &pb.ListFilesResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	responseFiles := make(
		[]*pb.FileMetadata,
		0,
		len(files),
	)

	for _, file := range files {
		responseFiles = append(
			responseFiles,
			toProtoFile(file),
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

func (s *Server) Upload(
	stream grpc.ClientStreamingServer[
		pb.UploadRequest,
		pb.UploadResponse,
	],
) error {

	var metadata *pb.UploadMetadata
	var content []byte
	var bytesReceived int64

	firstMessage := true

	for {

		request, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			return status.Errorf(
				codes.Internal,
				"error recibiendo upload: %v",
				err,
			)
		}

		if firstMessage {

			metadata = request.GetMetadata()

			if metadata == nil {
				return status.Error(
					codes.InvalidArgument,
					"el primer mensaje debe contener metadata",
				)
			}

			if metadata.UserId == "" {
				return status.Error(
					codes.InvalidArgument,
					"el user ID es obligatorio",
				)
			}

			if metadata.DeviceId == "" {
				return status.Error(
					codes.InvalidArgument,
					"el device ID es obligatorio",
				)
			}

			if metadata.FileName == "" {
				return status.Error(
					codes.InvalidArgument,
					"el nombre del archivo es obligatorio",
				)
			}

			if metadata.Size < 0 {
				return status.Error(
					codes.InvalidArgument,
					"el tamaño del archivo no puede ser negativo",
				)
			}

			firstMessage = false
			continue
		}

		chunk := request.GetChunk()

		if chunk == nil {
			return status.Error(
				codes.InvalidArgument,
				"se esperaba un chunk de contenido",
			)
		}

		content = append(
			content,
			chunk...,
		)

		bytesReceived += int64(
			len(chunk),
		)

		if bytesReceived > metadata.Size {
			return status.Error(
				codes.InvalidArgument,
				"se recibieron más bytes de los indicados en metadata",
			)
		}
	}

	if metadata == nil {
		return status.Error(
			codes.InvalidArgument,
			"no se recibieron metadatos",
		)
	}

	if bytesReceived != metadata.Size {
		return status.Errorf(
			codes.InvalidArgument,
			"tamaño incorrecto: esperados=%d recibidos=%d",
			metadata.Size,
			bytesReceived,
		)
	}

	file, err := s.syncService.RegisterFile(
		metadata.UserId,
		metadata.DeviceId,
		metadata.FileName,
		metadata.FileType,
		bytesReceived,
	)

	if err != nil {
		return status.Errorf(
			codes.Internal,
			"error registrando archivo: %v",
			err,
		)
	}

	err = s.storage.Save(
		file.FileID,
		content,
	)

	if err != nil {
		return status.Errorf(
			codes.Internal,
			"error almacenando archivo: %v",
			err,
		)
	}

	return stream.SendAndClose(
		&pb.UploadResponse{
			Success:       true,
			Message:       "Archivo cargado correctamente",
			FileId:        file.FileID,
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

	if request.UserId == "" {
		return status.Error(
			codes.InvalidArgument,
			"el user ID es obligatorio",
		)
	}

	if request.FileId == "" {
		return status.Error(
			codes.InvalidArgument,
			"el file ID es obligatorio",
		)
	}

	file, err := s.syncService.GetFile(
		request.FileId,
	)

	if err != nil {
		return status.Errorf(
			codes.NotFound,
			"archivo no encontrado: %v",
			err,
		)
	}

	if file.UserID != request.UserId {
		return status.Error(
			codes.PermissionDenied,
			"el archivo no pertenece al usuario",
		)
	}

	content, err := s.storage.Read(
		file.FileID,
	)

	if err != nil {
		return status.Errorf(
			codes.NotFound,
			"no fue posible leer el archivo físico: %v",
			err,
		)
	}

	err = stream.Send(
		&pb.DownloadResponse{
			Data: &pb.DownloadResponse_Metadata{
				Metadata: toProtoFile(file),
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

	for start := 0; start < len(content); start += downloadChunkSize {

		end := start + downloadChunkSize

		if end > len(content) {
			end = len(content)
		}

		chunk := content[start:end]

		err = stream.Send(
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
// DELETE FILE
// =====================================

func (s *Server) DeleteFile(
	ctx context.Context,
	request *pb.DeleteFileRequest,
) (*pb.DeleteFileResponse, error) {

	file, err := s.syncService.GetFile(
		request.FileId,
	)

	if err != nil {
		return &pb.DeleteFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	if file.UserID != request.UserId {
		return &pb.DeleteFileResponse{
			Success: false,
			Message: "El archivo no pertenece al usuario",
		}, nil
	}

	err = s.storage.Delete(
		request.FileId,
	)

	if err != nil {
		return &pb.DeleteFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	err = s.syncService.DeleteFile(
		request.UserId,
		request.DeviceId,
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
// WATCH CHANGES
// =====================================

func (s *Server) WatchChanges(
	request *pb.WatchChangesRequest,
	stream grpc.ServerStreamingServer[pb.FileChange],
) error {

	if request.UserId == "" {
		return status.Error(
			codes.InvalidArgument,
			"el user ID es obligatorio",
		)
	}

	if request.DeviceId == "" {
		return status.Error(
			codes.InvalidArgument,
			"el device ID es obligatorio",
		)
	}

	sent := make(
		map[string]bool,
	)

	ticker := time.NewTicker(
		2 * time.Second,
	)

	defer ticker.Stop()

	for {

		select {

		case <-stream.Context().Done():
			return nil

		case <-ticker.C:

			changes, err := s.syncService.Sync(
				request.UserId,
				request.DeviceId,
			)

			if err != nil {
				return status.Errorf(
					codes.Internal,
					"error consultando cambios: %v",
					err,
				)
			}

			for _, change := range changes {

				key := change.Type +
					":" +
					change.FileID +
					":" +
					change.Timestamp.String()

				if sent[key] {
					continue
				}

				err = stream.Send(
					toProtoChange(change),
				)

				if err != nil {
					return err
				}

				sent[key] = true
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

	return &pb.FileMetadata{
		FileId:    file.FileID,
		UserId:    file.UserID,
		FileName:  file.FileName,
		FileType:  file.FileType,
		Size:      file.Size,
		CreatedAt: file.CreatedAt.Unix(),
		UpdatedAt: file.UpdatedAt.Unix(),
	}
}

func toProtoChange(
	change *repository.FileChange,
) *pb.FileChange {

	return &pb.FileChange{
		Type:           change.Type,
		FileId:         change.FileID,
		FileName:       change.FileName,
		OriginDeviceId: change.OriginDeviceID,
		Timestamp:      change.Timestamp.Unix(),
	}
}
