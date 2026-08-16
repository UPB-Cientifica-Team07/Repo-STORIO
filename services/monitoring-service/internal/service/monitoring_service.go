package service

import (
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/model"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/repository"
)

type MonitoringService struct {
	repository repository.MonitoringRepository
}

func NewMonitoringService(
	repository repository.MonitoringRepository,
) *MonitoringService {
	return &MonitoringService{
		repository: repository,
	}
}

// ==============================
// MÉTRICAS
// ==============================

func (s *MonitoringService) RegisterMetric(
	metric model.Metric,
) error {

	return s.repository.SaveMetric(metric)
}

func (s *MonitoringService) GetMetrics() (
	[]model.Metric,
	error,
) {

	return s.repository.GetMetrics()
}

// ==============================
// ESTADO DE SERVICIOS
// ==============================

func (s *MonitoringService) RegisterServiceStatus(
	serviceStatus model.ServiceStatus,
) error {

	return s.repository.SaveServiceStatus(serviceStatus)
}

func (s *MonitoringService) GetServiceStatus(
	componentName string,
) (*model.ServiceStatus, error) {

	return s.repository.GetServiceStatus(componentName)
}
