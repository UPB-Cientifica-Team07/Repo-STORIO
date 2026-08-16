package repository

import (
	"sync"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/model"
)

type MonitoringRepository interface {
	SaveMetric(metric model.Metric) error
	GetMetrics() ([]model.Metric, error)

	SaveServiceStatus(serviceStatus model.ServiceStatus) error
	GetServiceStatus(componentName string) (*model.ServiceStatus, error)
}

type InMemoryMonitoringRepository struct {
	metrics       []model.Metric
	serviceStatus map[string]model.ServiceStatus
	mutex         sync.RWMutex
}

func NewInMemoryMonitoringRepository() *InMemoryMonitoringRepository {
	return &InMemoryMonitoringRepository{
		metrics:       make([]model.Metric, 0),
		serviceStatus: make(map[string]model.ServiceStatus),
	}
}

// ==============================
// MÉTRICAS
// ==============================

func (r *InMemoryMonitoringRepository) SaveMetric(
	metric model.Metric,
) error {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.metrics = append(r.metrics, metric)

	return nil
}

func (r *InMemoryMonitoringRepository) GetMetrics() (
	[]model.Metric,
	error,
) {

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	metrics := make([]model.Metric, len(r.metrics))

	copy(metrics, r.metrics)

	return metrics, nil
}

// ==============================
// ESTADO DE SERVICIOS
// ==============================

func (r *InMemoryMonitoringRepository) SaveServiceStatus(
	serviceStatus model.ServiceStatus,
) error {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.serviceStatus[serviceStatus.ComponentName] = serviceStatus

	return nil
}

func (r *InMemoryMonitoringRepository) GetServiceStatus(
	componentName string,
) (*model.ServiceStatus, error) {

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	serviceStatus, exists :=
		r.serviceStatus[componentName]

	if !exists {
		return nil, nil
	}

	return &serviceStatus, nil
}
