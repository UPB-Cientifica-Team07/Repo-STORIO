package grpc

import (
	"context"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/internal/fileclient"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/internal/service"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/proto"
)

type Server struct {
	pb.UnimplementedAnalysisServiceServer

	analysisService *service.AnalysisService
	fileClient      *fileclient.Client
}

func NewServer(
	analysisService *service.AnalysisService,
	fileClient *fileclient.Client,
) *Server {

	return &Server{
		analysisService: analysisService,
		fileClient:      fileClient,
	}
}

func (s *Server) AnalyzeFile(
	ctx context.Context,
	request *pb.AnalyzeFileRequest,
) (*pb.AnalyzeFileResponse, error) {

	// El Analysis Service solicita el contenido real
	// del archivo al File Service.
	content, err := s.fileClient.GetFileContent(
		request.FileId,
	)

	if err != nil {

		return &pb.AnalyzeFileResponse{
			Success: false,
			Message: "No se pudo obtener el archivo: " + err.Error(),
		}, nil
	}

	// Una vez obtenido el contenido, se realiza
	// el análisis interno.
	analysis, err := s.analysisService.AnalyzeFile(
		request.UserId,
		request.FileId,
		content,
	)

	if err != nil {

		return &pb.AnalyzeFileResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.AnalyzeFileResponse{
		Success:    true,
		Message:    "Archivo analizado correctamente",
		AnalysisId: analysis.ID,
		Summary:    analysis.Summary,
	}, nil
}

func (s *Server) GetAnalysis(
	ctx context.Context,
	request *pb.GetAnalysisRequest,
) (*pb.GetAnalysisResponse, error) {

	analysis, err := s.analysisService.GetAnalysis(
		request.AnalysisId,
	)

	if err != nil {

		return &pb.GetAnalysisResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.GetAnalysisResponse{
		Success:  true,
		Message:  "Análisis encontrado",
		Analysis: toProtoAnalysis(analysis),
	}, nil
}

func (s *Server) ListAnalyses(
	ctx context.Context,
	request *pb.ListAnalysesRequest,
) (*pb.ListAnalysesResponse, error) {

	// ListAnalyses devuelve:
	// analyses, error
	analyses, err := s.analysisService.ListAnalyses(
		request.UserId,
	)

	if err != nil {

		return &pb.ListAnalysesResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	responseAnalyses := make(
		[]*pb.AnalysisData,
		0,
		len(analyses),
	)

	for _, analysis := range analyses {

		responseAnalyses = append(
			responseAnalyses,
			toProtoAnalysis(analysis),
		)
	}

	return &pb.ListAnalysesResponse{
		Success:  true,
		Message:  "Análisis encontrados correctamente",
		Analyses: responseAnalyses,
	}, nil
}

func toProtoAnalysis(
	analysis *repository.Analysis,
) *pb.AnalysisData {

	return &pb.AnalysisData{
		AnalysisId: analysis.ID,
		UserId:     analysis.UserID,
		FileId:     analysis.FileID,
		Summary:    analysis.Summary,
		CreatedAt: analysis.CreatedAt.Format(
			"2006-01-02 15:04:05",
		),
	}
}
