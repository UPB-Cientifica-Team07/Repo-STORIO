package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/model"
)

const (
	defaultDataDirectory     = "data/monitoring"
	defaultMetricsFile       = "metrics.json"
	defaultServiceStatusFile = "service_status.json"
	defaultAlertRulesFile    = "alert_rules.json"
	defaultAlertsFile        = "alerts.json"
)

type MonitoringRepository struct {
	mu sync.RWMutex

	metrics         []model.Metric
	serviceStatuses map[string]model.ServiceStatus
	alertRules      []model.AlertRule
	alerts          []model.Alert

	dataDirectory  string
	metricsFile    string
	statusFile     string
	alertRulesFile string
	alertsFile     string
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewMonitoringRepository() *MonitoringRepository {

	repo := &MonitoringRepository{
		metrics:         make([]model.Metric, 0),
		serviceStatuses: make(map[string]model.ServiceStatus),
		alertRules:      make([]model.AlertRule, 0),
		alerts:          make([]model.Alert, 0),

		dataDirectory: defaultDataDirectory,

		metricsFile: filepath.Join(
			defaultDataDirectory,
			defaultMetricsFile,
		),

		statusFile: filepath.Join(
			defaultDataDirectory,
			defaultServiceStatusFile,
		),

		alertRulesFile: filepath.Join(
			defaultDataDirectory,
			defaultAlertRulesFile,
		),

		alertsFile: filepath.Join(
			defaultDataDirectory,
			defaultAlertsFile,
		),
	}

	if err := repo.initializeStorage(); err != nil {

		fmt.Printf(
			"Advertencia inicializando persistencia de Monitoring: %v\n",
			err,
		)
	}

	return repo
}

// =====================================
// INICIALIZAR PERSISTENCIA
// =====================================

func (r *MonitoringRepository) initializeStorage() error {

	if err := os.MkdirAll(
		r.dataDirectory,
		0755,
	); err != nil {

		return fmt.Errorf(
			"error creando directorio de monitoreo: %w",
			err,
		)
	}

	if err := r.loadMetrics(); err != nil {

		return fmt.Errorf(
			"error cargando métricas: %w",
			err,
		)
	}

	if err := r.loadServiceStatuses(); err != nil {

		return fmt.Errorf(
			"error cargando estados de servicios: %w",
			err,
		)
	}

	if err := r.loadAlertRules(); err != nil {

		return fmt.Errorf(
			"error cargando reglas de alerta: %w",
			err,
		)
	}

	if err := r.loadAlerts(); err != nil {

		return fmt.Errorf(
			"error cargando alertas: %w",
			err,
		)
	}

	return nil
}

// =====================================
// GUARDAR MÉTRICA
// =====================================

func (r *MonitoringRepository) SaveMetric(
	metric model.Metric,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	r.metrics = append(
		r.metrics,
		metric,
	)

	if err := r.persistMetricsLocked(); err != nil {

		r.metrics =
			r.metrics[:len(r.metrics)-1]

		return fmt.Errorf(
			"error persistiendo métrica: %w",
			err,
		)
	}

	return nil
}

// =====================================
// CONSULTAR MÉTRICAS
// =====================================

func (r *MonitoringRepository) GetMetrics() (
	[]model.Metric,
	error,
) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(
		[]model.Metric,
		len(r.metrics),
	)

	copy(
		result,
		r.metrics,
	)

	return result, nil
}

// =====================================
// GUARDAR ESTADO DE SERVICIO
// =====================================

func (r *MonitoringRepository) SaveServiceStatus(
	status model.ServiceStatus,
) error {

	if status.ComponentName == "" {

		return errors.New(
			"component_name es obligatorio",
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	previousStatus, existed :=
		r.serviceStatuses[status.ComponentName]

	r.serviceStatuses[status.ComponentName] =
		status

	if err := r.persistServiceStatusesLocked(); err != nil {

		if existed {

			r.serviceStatuses[status.ComponentName] =
				previousStatus

		} else {

			delete(
				r.serviceStatuses,
				status.ComponentName,
			)
		}

		return fmt.Errorf(
			"error persistiendo estado: %w",
			err,
		)
	}

	return nil
}

// =====================================
// CONSULTAR ESTADO
// =====================================

func (r *MonitoringRepository) GetServiceStatus(
	componentName string,
) (*model.ServiceStatus, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	status, exists :=
		r.serviceStatuses[componentName]

	if !exists {
		return nil, nil
	}

	statusCopy :=
		status

	return &statusCopy, nil
}

// =====================================
// CONSULTAR TODOS LOS ESTADOS
// =====================================

func (r *MonitoringRepository) GetServiceStatuses() (
	[]model.ServiceStatus,
	error,
) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	result :=
		make(
			[]model.ServiceStatus,
			0,
			len(r.serviceStatuses),
		)

	for _, serviceStatus := range r.serviceStatuses {

		result =
			append(
				result,
				serviceStatus,
			)
	}

	return result, nil
}

// =====================================
// GUARDAR REGLA DE ALERTA
// =====================================

func (r *MonitoringRepository) SaveAlertRule(
	rule model.AlertRule,
) error {

	if rule.ID == "" {

		return errors.New(
			"id de regla obligatorio",
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existingRule := range r.alertRules {

		if existingRule.ID == rule.ID {

			return fmt.Errorf(
				"ya existe una regla de alerta con id: %s",
				rule.ID,
			)
		}
	}

	r.alertRules = append(
		r.alertRules,
		rule,
	)

	if err := r.persistAlertRulesLocked(); err != nil {

		r.alertRules =
			r.alertRules[:len(r.alertRules)-1]

		return fmt.Errorf(
			"error persistiendo regla de alerta: %w",
			err,
		)
	}

	return nil
}

// =====================================
// ACTUALIZAR REGLA DE ALERTA
// =====================================

func (r *MonitoringRepository) UpdateAlertRule(
	rule model.AlertRule,
) (model.AlertRule, error) {

	if rule.ID == "" {

		return model.AlertRule{},
			errors.New(
				"id de regla obligatorio",
			)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.alertRules {

		if r.alertRules[i].ID != rule.ID {
			continue
		}

		previousRule :=
			r.alertRules[i]

		r.alertRules[i] =
			rule

		if err := r.persistAlertRulesLocked(); err != nil {

			r.alertRules[i] =
				previousRule

			return model.AlertRule{},
				fmt.Errorf(
					"error persistiendo actualización de regla: %w",
					err,
				)
		}

		return rule, nil
	}

	return model.AlertRule{},
		fmt.Errorf(
			"regla de alerta no encontrada: %s",
			rule.ID,
		)
}

// =====================================
// ELIMINAR REGLA DE ALERTA
// =====================================

func (r *MonitoringRepository) DeleteAlertRule(
	ruleID string,
) error {

	if ruleID == "" {

		return errors.New(
			"id de regla obligatorio",
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.alertRules {

		if r.alertRules[i].ID != ruleID {
			continue
		}

		previousRules :=
			make(
				[]model.AlertRule,
				len(r.alertRules),
			)

		copy(
			previousRules,
			r.alertRules,
		)

		r.alertRules =
			append(
				r.alertRules[:i],
				r.alertRules[i+1:]...,
			)

		if err := r.persistAlertRulesLocked(); err != nil {

			r.alertRules =
				previousRules

			return fmt.Errorf(
				"error persistiendo eliminación de regla: %w",
				err,
			)
		}

		return nil
	}

	return fmt.Errorf(
		"regla de alerta no encontrada: %s",
		ruleID,
	)
}

// =====================================
// CONSULTAR REGLAS
// =====================================

func (r *MonitoringRepository) GetAlertRules() (
	[]model.AlertRule,
	error,
) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(
		[]model.AlertRule,
		len(r.alertRules),
	)

	copy(
		result,
		r.alertRules,
	)

	return result, nil
}

// =====================================
// GUARDAR ALERTA
// =====================================

func (r *MonitoringRepository) SaveAlert(
	alert model.Alert,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	r.alerts = append(
		r.alerts,
		alert,
	)

	if err := r.persistAlertsLocked(); err != nil {

		r.alerts =
			r.alerts[:len(r.alerts)-1]

		return fmt.Errorf(
			"error persistiendo alerta: %w",
			err,
		)
	}

	return nil
}

// =====================================
// CONSULTAR ALERTAS
// =====================================

func (r *MonitoringRepository) GetAlerts() (
	[]model.Alert,
	error,
) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(
		[]model.Alert,
		len(r.alerts),
	)

	copy(
		result,
		r.alerts,
	)

	return result, nil
}

// =====================================
// BUSCAR ALERTA ACTIVA
// =====================================

func (r *MonitoringRepository) GetActiveAlert(
	ruleID string,
	componentID string,
) (*model.Alert, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	for i := range r.alerts {

		alert :=
			r.alerts[i]

		if alert.RuleID == ruleID &&
			alert.ComponentID == componentID &&
			alert.Active {

			alertCopy :=
				alert

			return &alertCopy, nil
		}
	}

	return nil, nil
}

// =====================================
// RESOLVER ALERTA
// =====================================

func (r *MonitoringRepository) ResolveAlert(
	alertID string,
	currentValue float64,
	resolvedAt time.Time,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.alerts {

		if r.alerts[i].ID != alertID {
			continue
		}

		previousAlert :=
			r.alerts[i]

		r.alerts[i].Active =
			false

		r.alerts[i].CurrentValue =
			currentValue

		r.alerts[i].ResolvedAt =
			&resolvedAt

		r.alerts[i].Message =
			fmt.Sprintf(
				"Alerta '%s' resuelta: %s = %.2f",
				r.alerts[i].RuleName,
				r.alerts[i].Metric,
				currentValue,
			)

		if err := r.persistAlertsLocked(); err != nil {

			r.alerts[i] =
				previousAlert

			return fmt.Errorf(
				"error persistiendo resolución de alerta: %w",
				err,
			)
		}

		return nil
	}

	return fmt.Errorf(
		"alerta no encontrada: %s",
		alertID,
	)
}

// =====================================
// CARGAR MÉTRICAS
// =====================================

func (r *MonitoringRepository) loadMetrics() error {

	data, err :=
		os.ReadFile(
			r.metricsFile,
		)

	if err != nil {

		if os.IsNotExist(err) {

			r.metrics =
				make(
					[]model.Metric,
					0,
				)

			return nil
		}

		return err
	}

	if len(data) == 0 {

		r.metrics =
			make(
				[]model.Metric,
				0,
			)

		return nil
	}

	var metrics []model.Metric

	if err := json.Unmarshal(
		data,
		&metrics,
	); err != nil {

		return fmt.Errorf(
			"JSON de métricas inválido: %w",
			err,
		)
	}

	if metrics == nil {

		metrics =
			make(
				[]model.Metric,
				0,
			)
	}

	r.metrics =
		metrics

	fmt.Printf(
		"Histórico cargado: %d métricas\n",
		len(r.metrics),
	)

	return nil
}

// =====================================
// CARGAR ESTADOS
// =====================================

func (r *MonitoringRepository) loadServiceStatuses() error {

	data, err :=
		os.ReadFile(
			r.statusFile,
		)

	if err != nil {

		if os.IsNotExist(err) {

			r.serviceStatuses =
				make(
					map[string]model.ServiceStatus,
				)

			return nil
		}

		return err
	}

	if len(data) == 0 {

		r.serviceStatuses =
			make(
				map[string]model.ServiceStatus,
			)

		return nil
	}

	var statuses map[string]model.ServiceStatus

	if err := json.Unmarshal(
		data,
		&statuses,
	); err != nil {

		return fmt.Errorf(
			"JSON de estados inválido: %w",
			err,
		)
	}

	if statuses == nil {

		statuses =
			make(
				map[string]model.ServiceStatus,
			)
	}

	r.serviceStatuses =
		statuses

	fmt.Printf(
		"Estados cargados: %d\n",
		len(r.serviceStatuses),
	)

	return nil
}

// =====================================
// CARGAR REGLAS DE ALERTA
// =====================================

func (r *MonitoringRepository) loadAlertRules() error {

	data, err :=
		os.ReadFile(
			r.alertRulesFile,
		)

	if err != nil {

		if os.IsNotExist(err) {

			r.alertRules =
				make(
					[]model.AlertRule,
					0,
				)

			return nil
		}

		return err
	}

	if len(data) == 0 {

		r.alertRules =
			make(
				[]model.AlertRule,
				0,
			)

		return nil
	}

	var rules []model.AlertRule

	if err := json.Unmarshal(
		data,
		&rules,
	); err != nil {

		return fmt.Errorf(
			"JSON de reglas de alerta inválido: %w",
			err,
		)
	}

	if rules == nil {

		rules =
			make(
				[]model.AlertRule,
				0,
			)
	}

	r.alertRules =
		rules

	fmt.Printf(
		"Reglas de alerta cargadas: %d\n",
		len(r.alertRules),
	)

	return nil
}

// =====================================
// CARGAR ALERTAS
// =====================================

func (r *MonitoringRepository) loadAlerts() error {

	data, err :=
		os.ReadFile(
			r.alertsFile,
		)

	if err != nil {

		if os.IsNotExist(err) {

			r.alerts =
				make(
					[]model.Alert,
					0,
				)

			return nil
		}

		return err
	}

	if len(data) == 0 {

		r.alerts =
			make(
				[]model.Alert,
				0,
			)

		return nil
	}

	var alerts []model.Alert

	if err := json.Unmarshal(
		data,
		&alerts,
	); err != nil {

		return fmt.Errorf(
			"JSON de alertas inválido: %w",
			err,
		)
	}

	if alerts == nil {

		alerts =
			make(
				[]model.Alert,
				0,
			)
	}

	r.alerts =
		alerts

	fmt.Printf(
		"Alertas cargadas: %d\n",
		len(r.alerts),
	)

	return nil
}

// =====================================
// PERSISTIR MÉTRICAS
// =====================================

func (r *MonitoringRepository) persistMetricsLocked() error {

	data, err :=
		json.MarshalIndent(
			r.metrics,
			"",
			"  ",
		)

	if err != nil {

		return fmt.Errorf(
			"error serializando métricas: %w",
			err,
		)
	}

	return atomicWriteFile(
		r.metricsFile,
		data,
	)
}

// =====================================
// PERSISTIR ESTADOS
// =====================================

func (r *MonitoringRepository) persistServiceStatusesLocked() error {

	data, err :=
		json.MarshalIndent(
			r.serviceStatuses,
			"",
			"  ",
		)

	if err != nil {

		return fmt.Errorf(
			"error serializando estados: %w",
			err,
		)
	}

	return atomicWriteFile(
		r.statusFile,
		data,
	)
}

// =====================================
// PERSISTIR REGLAS
// =====================================

func (r *MonitoringRepository) persistAlertRulesLocked() error {

	data, err :=
		json.MarshalIndent(
			r.alertRules,
			"",
			"  ",
		)

	if err != nil {

		return fmt.Errorf(
			"error serializando reglas de alerta: %w",
			err,
		)
	}

	return atomicWriteFile(
		r.alertRulesFile,
		data,
	)
}

// =====================================
// PERSISTIR ALERTAS
// =====================================

func (r *MonitoringRepository) persistAlertsLocked() error {

	data, err :=
		json.MarshalIndent(
			r.alerts,
			"",
			"  ",
		)

	if err != nil {

		return fmt.Errorf(
			"error serializando alertas: %w",
			err,
		)
	}

	return atomicWriteFile(
		r.alertsFile,
		data,
	)
}

// =====================================
// ESCRITURA ATÓMICA
// =====================================

func atomicWriteFile(
	filePath string,
	data []byte,
) error {

	directory :=
		filepath.Dir(
			filePath,
		)

	if err := os.MkdirAll(
		directory,
		0755,
	); err != nil {

		return fmt.Errorf(
			"error creando directorio %s: %w",
			directory,
			err,
		)
	}

	tempFile :=
		filePath + ".tmp"

	if err := os.WriteFile(
		tempFile,
		data,
		0644,
	); err != nil {

		return fmt.Errorf(
			"error escribiendo archivo temporal: %w",
			err,
		)
	}

	if err := os.Rename(
		tempFile,
		filePath,
	); err != nil {

		_ =
			os.Remove(
				tempFile,
			)

		return fmt.Errorf(
			"error reemplazando archivo persistente: %w",
			err,
		)
	}

	return nil
}
