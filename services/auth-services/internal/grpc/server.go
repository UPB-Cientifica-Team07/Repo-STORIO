package grpc

import (
	"context"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/auth-services/proto"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/auth-services/internal/service"
)

type Server struct {
	pb.UnimplementedAuthServiceServer

	authService *service.AuthService
}

func NewServer(
	authService *service.AuthService,
) *Server {

	return &Server{
		authService: authService,
	}
}

func (s *Server) Register(
	ctx context.Context,
	request *pb.RegisterRequest,
) (*pb.RegisterResponse, error) {

	userID, err := s.authService.Register(
		request.GetUsername(),
		request.GetEmail(),
		request.GetPassword(),
	)

	if err != nil {
		return &pb.RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.RegisterResponse{
		Success: true,
		Message: "Usuario registrado correctamente",
		UserId:  userID,
	}, nil
}

func (s *Server) Login(
	ctx context.Context,
	request *pb.LoginRequest,
) (*pb.LoginResponse, error) {

	token, userID, err := s.authService.Login(
		request.GetEmail(),
		request.GetPassword(),
	)

	if err != nil {
		return &pb.LoginResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.LoginResponse{
		Success: true,
		Message: "Login correcto",
		Token:   token,
		UserId:  userID,
	}, nil
}

func (s *Server) ValidateToken(
	ctx context.Context,
	request *pb.TokenRequest,
) (*pb.TokenResponse, error) {

	userID, err := s.authService.ValidateToken(
		request.GetToken(),
	)

	if err != nil {
		return &pb.TokenResponse{
			Valid:   false,
			Message: err.Error(),
		}, nil
	}

	return &pb.TokenResponse{
		Valid:   true,
		Message: "Token válido",
		UserId:  userID,
	}, nil
}
