package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/model"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/repository"
)

const (
	serviceAvailabilityTimeout = 15 * time.Second

	availabilityMetric = "service_availability"
)

type MonitoringService struct {
	repository *repository.MonitoringRepository
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewMonitoringService(
	repository *repository.MonitoringRepository,
) *MonitoringService {

	return &MonitoringService{
		repository: repository,
	}
}

// =====================================
// REGISTRAR MÉTRICA
// =====================================

func (s *MonitoringService) RegisterMetric(
	metric model.Metric,
) error {

	if metric.ComponentID == "" {

		return errors.New(
			"component_id es obligatorio",
		)
	}

	if metric.ComponentName == "" {

		return errors.New(
			"component_name es obligatorio",
		)
	}

	if metric.Timestamp.IsZero() {

		return errors.New(
			"timestamp de métrica es obligatorio",
		)
	}

	if err := s.repository.SaveMetric(
		metric,
	); err != nil {

		return err
	}

	// =====================================
	// HEARTBEAT
	// =====================================

	currentStatus, err :=
		s.repository.GetServiceStatus(
			metric.ComponentName,
		)

	if err != nil {

		return fmt.Errorf(
			"error consultando estado para heartbeat: %w",
			err,
		)
	}

	if currentStatus != nil {

		wasUnavailable :=
			currentStatus.Status == "UNAVAILABLE"

		updatedStatus :=
			*currentStatus

		updatedStatus.LastUpdated =
			metric.Timestamp

		if wasUnavailable {

			updatedStatus.Status =
				"ACTIVE"

			updatedStatus.Message =
				"Servicio disponible nuevamente"
		}

		if err := s.repository.SaveServiceStatus(
			updatedStatus,
		); err != nil {

			return fmt.Errorf(
				"error actualizando heartbeat del servicio: %w",
				err,
			)
		}

		if wasUnavailable {

			if err := s.resolveAvailabilityAlert(
				updatedStatus,
			); err != nil {

				fmt.Printf(
					"Error resolviendo alerta de disponibilidad: %v\n",
					err,
				)
			}

			fmt.Printf(
				"SERVICIO RECUPERADO | Componente=%s\n",
				updatedStatus.ComponentName,
			)
		}
	}

	s.evaluateMetricAlerts(
		metric,
	)

	return nil
}

// =====================================
// CONSULTAR MÉTRICAS
// =====================================

func (s *MonitoringService) GetMetrics() (
	[]model.Metric,
	error,
) {

	return s.repository.GetMetrics()
}

// =====================================
// REGISTRAR ESTADO
// =====================================

func (s *MonitoringService) RegisterServiceStatus(
	serviceStatus model.ServiceStatus,
) error {

	if serviceStatus.ComponentID == "" {

		return errors.New(
			"component_id es obligatorio",
		)
	}

	if serviceStatus.ComponentName == "" {

		return errors.New(
			"component_name es obligatorio",
		)
	}

	if serviceStatus.Status == "" {

		return errors.New(
			"status es obligatorio",
		)
	}

	if serviceStatus.LastUpdated.IsZero() {

		return errors.New(
			"last_updated es obligatorio",
		)
	}

	previousStatus, err :=
		s.repository.GetServiceStatus(
			serviceStatus.ComponentName,
		)

	if err != nil {
		return err
	}

	if err := s.repository.SaveServiceStatus(
		serviceStatus,
	); err != nil {

		return err
	}

	// Si estaba caído y reporta ACTIVE explícitamente,
	// resolvemos también la alerta.
	if previousStatus != nil &&
		previousStatus.Status == "UNAVAILABLE" &&
		serviceStatus.Status == "ACTIVE" {

		if err := s.resolveAvailabilityAlert(
			serviceStatus,
		); err != nil {

			fmt.Printf(
				"Error resolviendo alerta de disponibilidad: %v\n",
				err,
			)
		}

		fmt.Printf(
			"SERVICIO RECUPERADO | Componente=%s\n",
			serviceStatus.ComponentName,
		)
	}

	return nil
}

// =====================================
// CONSULTAR ESTADO
// =====================================

func (s *MonitoringService) GetServiceStatus(
	componentName string,
) (*model.ServiceStatus, error) {

	if componentName == "" {

		return nil,
			errors.New(
				"component_name es obligatorio",
			)
	}

	serviceStatus, err :=
		s.repository.GetServiceStatus(
			componentName,
		)

	if err != nil {

		return nil, err
	}

	if serviceStatus == nil {

		return nil, nil
	}

	return serviceStatus, nil
}

// =====================================
// WATCHDOG DE DISPONIBILIDAD
// =====================================

func (s *MonitoringService) CheckAllServiceAvailability() error {

	serviceStatuses, err :=
		s.repository.GetServiceStatuses()

	if err != nil {

		return fmt.Errorf(
			"error consultando estados para watchdog: %w",
			err,
		)
	}

	now :=
		time.Now()

	for _, snapshot := range serviceStatuses {

		// Volvemos a consultar para trabajar
		// con el estado más reciente posible.
		currentStatus, err :=
			s.repository.GetServiceStatus(
				snapshot.ComponentName,
			)

		if err != nil {

			fmt.Printf(
				"Error consultando %s en watchdog: %v\n",
				snapshot.ComponentName,
				err,
			)

			continue
		}

		if currentStatus == nil {
			continue
		}

		// Una parada ordenada ya fue informada.
		// No se considera una caída inesperada.
		if currentStatus.Status == "INACTIVE" {
			continue
		}

		elapsed :=
			now.Sub(
				currentStatus.LastUpdated,
			)

		// =====================================
		// SERVICIO DISPONIBLE
		// =====================================

		if elapsed <= serviceAvailabilityTimeout {
			continue
		}

		// Ya fue detectado anteriormente.
		if currentStatus.Status == "UNAVAILABLE" {
			continue
		}

		// =====================================
		// SERVICIO NO DISPONIBLE
		// =====================================

		unavailableStatus :=
			*currentStatus

		unavailableStatus.Status =
			"UNAVAILABLE"

		unavailableStatus.Message =
			fmt.Sprintf(
				"Servicio sin heartbeat durante %s",
				elapsed.Round(
					time.Second,
				),
			)

		// No modificamos LastUpdated.
		// Debe conservar el último heartbeat real.
		if err := s.repository.SaveServiceStatus(
			unavailableStatus,
		); err != nil {

			fmt.Printf(
				"Error marcando %s como UNAVAILABLE: %v\n",
				currentStatus.ComponentName,
				err,
			)

			continue
		}

		if err := s.createAvailabilityAlert(
			unavailableStatus,
		); err != nil {

			fmt.Printf(
				"Error creando alerta de disponibilidad: %v\n",
				err,
			)

			continue
		}

		fmt.Printf(
			"SERVICIO NO DISPONIBLE | Componente=%s | Sin heartbeat=%s\n",
			unavailableStatus.ComponentName,
			elapsed.Round(
				time.Second,
			),
		)
	}

	return nil
}

// =====================================
// CREAR ALERTA DE DISPONIBILIDAD
// =====================================

func (s *MonitoringService) createAvailabilityAlert(
	serviceStatus model.ServiceStatus,
) error {

	ruleID :=
		availabilityRuleID(
			serviceStatus.ComponentID,
		)

	activeAlert, err :=
		s.repository.GetActiveAlert(
			ruleID,
			serviceStatus.ComponentID,
		)

	if err != nil {

		return err
	}

	if activeAlert != nil {

		return nil
	}

	alert :=
		model.Alert{
			ID: uuid.NewString(),

			RuleID: ruleID,

			RuleName: "Disponibilidad de servicio",

			ComponentID: serviceStatus.ComponentID,

			ComponentName: serviceStatus.ComponentName,

			Metric: availabilityMetric,

			CurrentValue: 0,

			Threshold: 1,

			Message: fmt.Sprintf(
				"Servicio '%s' no disponible: no se recibe heartbeat",
				serviceStatus.ComponentName,
			),

			CreatedAt: time.Now(),

			Active: true,

			ResolvedAt: nil,
		}

	return s.repository.SaveAlert(
		alert,
	)
}

// =====================================
// RESOLVER ALERTA DE DISPONIBILIDAD
// =====================================

func (s *MonitoringService) resolveAvailabilityAlert(
	serviceStatus model.ServiceStatus,
) error {

	ruleID :=
		availabilityRuleID(
			serviceStatus.ComponentID,
		)

	activeAlert, err :=
		s.repository.GetActiveAlert(
			ruleID,
			serviceStatus.ComponentID,
		)

	if err != nil {

		return err
	}

	if activeAlert == nil {

		return nil
	}

	return s.repository.ResolveAlert(
		activeAlert.ID,
		1,
		time.Now(),
	)
}

// =====================================
// ID DE REGLA DE DISPONIBILIDAD
// =====================================

func availabilityRuleID(
	componentID string,
) string {

	return fmt.Sprintf(
		"availability-%s",
		componentID,
	)
}

// =====================================
// CREAR REGLA DE ALERTA
// =====================================

func (s *MonitoringService) CreateAlertRule(
	rule model.AlertRule,
) (model.AlertRule, error) {

	if rule.ID == "" {

		rule.ID =
			uuid.NewString()
	}

	if rule.Name == "" {

		return model.AlertRule{},
			errors.New(
				"name de regla obligatorio",
			)
	}

	if rule.Metric == "" {

		return model.AlertRule{},
			errors.New(
				"metric es obligatorio",
			)
	}

	if rule.Operator == "" {

		return model.AlertRule{},
			errors.New(
				"operator es obligatorio",
			)
	}

	if !isValidOperator(
		rule.Operator,
	) {

		return model.AlertRule{},
			fmt.Errorf(
				"operador no soportado: %s",
				rule.Operator,
			)
	}

	if !isValidMetric(
		rule.Metric,
	) {

		return model.AlertRule{},
			fmt.Errorf(
				"métrica no soportada: %s",
				rule.Metric,
			)
	}

	if err := s.repository.SaveAlertRule(
		rule,
	); err != nil {

		return model.AlertRule{},
			err
	}

	return rule, nil
}

// =====================================
// ACTUALIZAR REGLA DE ALERTA
// =====================================

func (s *MonitoringService) UpdateAlertRule(
	rule model.AlertRule,
) (model.AlertRule, error) {

	if rule.ID == "" {

		return model.AlertRule{},
			errors.New(
				"id de regla obligatorio",
			)
	}

	if rule.Name == "" {

		return model.AlertRule{},
			errors.New(
				"name de regla obligatorio",
			)
	}

	if rule.Metric == "" {

		return model.AlertRule{},
			errors.New(
				"metric es obligatorio",
			)
	}

	if rule.Operator == "" {

		return model.AlertRule{},
			errors.New(
				"operator es obligatorio",
			)
	}

	if !isValidOperator(
		rule.Operator,
	) {

		return model.AlertRule{},
			fmt.Errorf(
				"operador no soportado: %s",
				rule.Operator,
			)
	}

	if !isValidMetric(
		rule.Metric,
	) {

		return model.AlertRule{},
			fmt.Errorf(
				"métrica no soportada: %s",
				rule.Metric,
			)
	}

	return s.repository.UpdateAlertRule(
		rule,
	)
}

// =====================================
// ELIMINAR REGLA DE ALERTA
// =====================================

func (s *MonitoringService) DeleteAlertRule(
	ruleID string,
) error {

	if ruleID == "" {

		return errors.New(
			"id de regla obligatorio",
		)
	}

	return s.repository.DeleteAlertRule(
		ruleID,
	)
}

// =====================================
// CONSULTAR REGLAS
// =====================================

func (s *MonitoringService) GetAlertRules() (
	[]model.AlertRule,
	error,
) {

	return s.repository.GetAlertRules()
}

// =====================================
// CONSULTAR ALERTAS
// =====================================

func (s *MonitoringService) GetAlerts() (
	[]model.Alert,
	error,
) {

	return s.repository.GetAlerts()
}

// =====================================
// EVALUAR ALERTAS NUMÉRICAS
// =====================================

func (s *MonitoringService) evaluateMetricAlerts(
	metric model.Metric,
) {

	rules, err :=
		s.repository.GetAlertRules()

	if err != nil {

		fmt.Printf(
			"Error consultando reglas de alerta: %v\n",
			err,
		)

		return
	}

	for _, rule := range rules {

		if !rule.Enabled {
			continue
		}

		if rule.ComponentName != "" &&
			rule.ComponentName != metric.ComponentName {

			continue
		}

		value, valid :=
			getMetricValue(
				metric,
				rule.Metric,
			)

		if !valid {
			continue
		}

		triggered :=
			compareValue(
				value,
				rule.Operator,
				rule.Threshold,
			)

		activeAlert, err :=
			s.repository.GetActiveAlert(
				rule.ID,
				metric.ComponentID,
			)

		if err != nil {

			fmt.Printf(
				"Error buscando alerta activa: %v\n",
				err,
			)

			continue
		}

		if triggered &&
			activeAlert == nil {

			alert :=
				model.Alert{
					ID: uuid.NewString(),

					RuleID: rule.ID,

					RuleName: rule.Name,

					ComponentID: metric.ComponentID,

					ComponentName: metric.ComponentName,

					Metric: rule.Metric,

					CurrentValue: value,

					Threshold: rule.Threshold,

					Message: fmt.Sprintf(
						"Alerta '%s': %s = %.2f %s %.2f",
						rule.Name,
						rule.Metric,
						value,
						rule.Operator,
						rule.Threshold,
					),

					CreatedAt: time.Now(),

					Active: true,

					ResolvedAt: nil,
				}

			if err := s.repository.SaveAlert(
				alert,
			); err != nil {

				fmt.Printf(
					"Error guardando alerta: %v\n",
					err,
				)

				continue
			}

			fmt.Printf(
				"ALERTA ACTIVADA | Componente=%s | Regla=%s | Valor=%.2f | Umbral=%.2f\n",
				metric.ComponentName,
				rule.Name,
				value,
				rule.Threshold,
			)

			continue
		}

		if triggered &&
			activeAlert != nil {

			continue
		}

		if !triggered &&
			activeAlert != nil {

			resolvedAt :=
				time.Now()

			if err := s.repository.ResolveAlert(
				activeAlert.ID,
				value,
				resolvedAt,
			); err != nil {

				fmt.Printf(
					"Error resolviendo alerta: %v\n",
					err,
				)

				continue
			}

			fmt.Printf(
				"ALERTA RESUELTA | Componente=%s | Regla=%s | Valor=%.2f\n",
				metric.ComponentName,
				rule.Name,
				value,
			)
		}
	}
}

// =====================================
// OBTENER VALOR DE MÉTRICA
// =====================================

func getMetricValue(
	metric model.Metric,
	metricName string,
) (float64, bool) {

	switch metricName {

	case "cpu_usage":

		return metric.CPUUsage,
			true

	case "memory_usage":

		return metric.MemoryUsage,
			true

	case "storage_usage":

		return float64(
				metric.StorageUsage,
			),
			true

	case "active_connections":

		return float64(
				metric.ActiveConnections,
			),
			true

	case "total_requests":

		return float64(
				metric.TotalRequests,
			),
			true

	default:

		return 0,
			false
	}
}

// =====================================
// COMPARAR VALOR
// =====================================

func compareValue(
	value float64,
	operator string,
	threshold float64,
) bool {

	switch operator {

	case ">":

		return value >
			threshold

	case ">=":

		return value >=
			threshold

	case "<":

		return value <
			threshold

	case "<=":

		return value <=
			threshold

	case "==":

		return value ==
			threshold

	default:

		return false
	}
}

// =====================================
// VALIDAR OPERADOR
// =====================================

func isValidOperator(
	operator string,
) bool {

	switch operator {

	case ">",
		">=",
		"<",
		"<=",
		"==":

		return true

	default:

		return false
	}
}

// =====================================
// VALIDAR MÉTRICA
// =====================================

func isValidMetric(
	metric string,
) bool {

	switch metric {

	case "cpu_usage",
		"memory_usage",
		"storage_usage",
		"active_connections",
		"total_requests":

		return true

	default:

		return false
	}
}
