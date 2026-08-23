package repository

import (
	"errors"
	"sync"
	"time"
)

type Analysis struct {
	ID        string
	UserID    string
	FileID    string
	Summary   string
	CreatedAt time.Time
}

type AnalysisRepository struct {
	mu       sync.RWMutex
	analyses map[string]*Analysis
}

func NewAnalysisRepository() *AnalysisRepository {
	return &AnalysisRepository{
		analyses: make(map[string]*Analysis),
	}
}

func (r *AnalysisRepository) Save(
	analysis *Analysis,
) error {

	if analysis == nil {
		return errors.New("el análisis no puede ser nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.analyses[analysis.ID] = analysis

	return nil
}

func (r *AnalysisRepository) FindByID(
	analysisID string,
) (*Analysis, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	analysis, exists := r.analyses[analysisID]

	if !exists {
		return nil, errors.New("análisis no encontrado")
	}

	return analysis, nil
}

func (r *AnalysisRepository) FindByUserID(
	userID string,
) []*Analysis {

	r.mu.RLock()
	defer r.mu.RUnlock()

	var analyses []*Analysis

	for _, analysis := range r.analyses {

		if analysis.UserID == userID {
			analyses = append(
				analyses,
				analysis,
			)
		}
	}

	return analyses
}
