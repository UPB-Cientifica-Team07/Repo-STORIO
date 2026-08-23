package service

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/internal/repository"
)

type AnalysisService struct {
	repository *repository.AnalysisRepository
}

func NewAnalysisService(
	repository *repository.AnalysisRepository,
) *AnalysisService {

	return &AnalysisService{
		repository: repository,
	}
}

func (s *AnalysisService) AnalyzeFile(
	userID string,
	fileID string,
	content string,
) (*repository.Analysis, error) {

	if userID == "" {
		return nil, errors.New("el user ID es obligatorio")
	}

	if fileID == "" {
		return nil, errors.New("el file ID es obligatorio")
	}

	if strings.TrimSpace(content) == "" {
		return nil, errors.New("el contenido del archivo está vacío")
	}

	summary := generateSummary(
		content,
	)

	analysis := &repository.Analysis{
		ID:        uuid.New().String(),
		UserID:    userID,
		FileID:    fileID,
		Summary:   summary,
		CreatedAt: time.Now(),
	}

	err := s.repository.Save(
		analysis,
	)

	if err != nil {
		return nil, err
	}

	return analysis, nil
}

func (s *AnalysisService) GetAnalysis(
	analysisID string,
) (*repository.Analysis, error) {

	if analysisID == "" {
		return nil, errors.New("el analysis ID es obligatorio")
	}

	return s.repository.FindByID(
		analysisID,
	)
}

func (s *AnalysisService) ListAnalyses(
	userID string,
) ([]*repository.Analysis, error) {

	if userID == "" {
		return nil, errors.New("el user ID es obligatorio")
	}

	return s.repository.FindByUserID(
		userID,
	), nil
}

func generateSummary(
	content string,
) string {

	content = strings.TrimSpace(content)

	sentences := strings.FieldsFunc(
		content,
		func(r rune) bool {
			return r == '.' ||
				r == '!' ||
				r == '?'
		},
	)

	if len(sentences) == 0 {
		return content
	}

	maxSentences := 3

	if len(sentences) < maxSentences {
		maxSentences = len(sentences)
	}

	var summary []string

	for i := 0; i < maxSentences; i++ {

		sentence := strings.TrimSpace(
			sentences[i],
		)

		if sentence != "" {
			summary = append(
				summary,
				sentence+".",
			)
		}
	}

	return strings.Join(
		summary,
		" ",
	)
}
