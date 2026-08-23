package grpc

import (
	"context"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/service"
)

type Server struct {
	pb.UnimplementedFileServiceServer
	fileService *service.FileService
}

func NewServer(
	fileService *service.FileService,
) *Server {

	return &Server{
		fileService: fileService,
	}
}

func (s *Server) UploadFile(
	ctx context.Context,
	request *pb.UploadFileRequest,
) (*pb.UploadFileResponse, error) {

	file, err := s.fileService.UploadFile(
		request.UserId,
		request.FileName,
		request.FileType,
		request.Content,
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

func (s *Server) GetFile(
	ctx context.Context,
	request *pb.GetFileRequest,
) (*pb.GetFileResponse, error) {

	file, err := s.fileService.GetFile(
		request.FileId,
	)

	if err != nil {

		return &pb.GetFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.GetFileResponse{
		Success: true,
		Message: "Archivo encontrado",
		File:    toProtoFile(file),
	}, nil
}

func (s *Server) DeleteFile(
	ctx context.Context,
	request *pb.DeleteFileRequest,
) (*pb.DeleteFileResponse, error) {

	err := s.fileService.DeleteFile(
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

func (s *Server) ListFiles(
	ctx context.Context,
	request *pb.ListFilesRequest,
) (*pb.ListFilesResponse, error) {

	files := s.fileService.ListFiles(
		request.UserId,
	)

	responseFiles := make(
		[]*pb.FileData,
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

func toProtoFile(
	file *repository.File,
) *pb.FileData {

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
