package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/auth-services/internal/repository"
)

type AuthService struct {
	userRepository *repository.UserRepository
	tokens         map[string]string
}

func NewAuthService(
	userRepository *repository.UserRepository,
) *AuthService {

	return &AuthService{
		userRepository: userRepository,
		tokens:         make(map[string]string),
	}
}

func (s *AuthService) Register(
	username string,
	email string,
	password string,
) (string, error) {

	if username == "" {
		return "", errors.New("username requerido")
	}

	if email == "" {
		return "", errors.New("email requerido")
	}

	if password == "" {
		return "", errors.New("password requerido")
	}

	userID := uuid.New().String()

	user := &repository.User{
		ID:       userID,
		Username: username,
		Email:    email,
		Password: password,
	}

	err := s.userRepository.Create(user)

	if err != nil {
		return "", err
	}

	return userID, nil
}

func (s *AuthService) Login(
	email string,
	password string,
) (string, string, error) {

	user, err := s.userRepository.FindByEmail(email)

	if err != nil {
		return "", "", err
	}

	if user.Password != password {
		return "", "", errors.New("credenciales inválidas")
	}

	token := fmt.Sprintf(
		"token-%s-%d",
		user.ID,
		time.Now().UnixNano(),
	)

	s.tokens[token] = user.ID

	return token, user.ID, nil
}

func (s *AuthService) ValidateToken(
	token string,
) (string, error) {

	userID, exists := s.tokens[token]

	if !exists {
		return "", errors.New("token inválido")
	}

	return userID, nil
}
