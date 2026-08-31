package grpc

import (
	"context"
	"log"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/generated"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/model"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MonitoringServer struct {
	pb.UnimplementedMonitoringServiceServer

	service *service.MonitoringService
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewMonitoringServer(
	service *service.MonitoringService,
) *MonitoringServer {

	return &MonitoringServer{
		service: service,
	}
}

// =====================================
// REPORTAR MÉTRICAS
// =====================================

func (s *MonitoringServer) ReportMetrics(
	ctx context.Context,
	req *pb.MetricsRequest,
) (*pb.MetricsResponse, error) {

	metric :=
		model.Metric{
			ComponentID: req.GetComponentId(),

			ComponentName: req.GetComponentName(),

			CPUUsage: req.GetCpuUsage(),

			MemoryUsage: req.GetMemoryUsage(),

			StorageUsage: req.GetStorageUsage(),

			ActiveConnections: req.GetActiveConnections(),

			TotalRequests: req.GetTotalRequests(),

			Timestamp: time.Unix(
				req.GetTimestamp(),
				0,
			),
		}

	if err := s.service.RegisterMetric(
		metric,
	); err != nil {

		return &pb.MetricsResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Println(
		"=== MÉTRICA REGISTRADA ===",
	)

	log.Printf(
		"Componente: %s",
		metric.ComponentName,
	)

	log.Printf(
		"ID: %s",
		metric.ComponentID,
	)

	log.Printf(
		"CPU: %.2f",
		metric.CPUUsage,
	)

	log.Printf(
		"Memoria: %.2f MB",
		metric.MemoryUsage,
	)

	log.Printf(
		"Almacenamiento: %d bytes",
		metric.StorageUsage,
	)

	log.Printf(
		"Conexiones activas: %d",
		metric.ActiveConnections,
	)

	log.Printf(
		"Total solicitudes: %d",
		metric.TotalRequests,
	)

	log.Printf(
		"Timestamp: %v",
		metric.Timestamp,
	)

	return &pb.MetricsResponse{
		Success: true,
		Message: "Métrica registrada correctamente",
	}, nil
}

// =====================================
// CONSULTAR MÉTRICAS
// =====================================

func (s *MonitoringServer) GetMetrics(
	ctx context.Context,
	req *pb.GetMetricsRequest,
) (*pb.GetMetricsResponse, error) {

	metrics, err :=
		s.service.GetMetrics()

	if err != nil {

		return nil,
			status.Error(
				codes.Internal,
				err.Error(),
			)
	}

	response :=
		&pb.GetMetricsResponse{
			Metrics: make(
				[]*pb.Metric,
				0,
				len(metrics),
			),
		}

	for _, metric := range metrics {

		response.Metrics =
			append(
				response.Metrics,
				&pb.Metric{
					ComponentId: metric.ComponentID,

					ComponentName: metric.ComponentName,

					CpuUsage: metric.CPUUsage,

					MemoryUsage: metric.MemoryUsage,

					StorageUsage: metric.StorageUsage,

					ActiveConnections: metric.ActiveConnections,

					TotalRequests: metric.TotalRequests,

					Timestamp: metric.Timestamp.Unix(),
				},
			)
	}

	log.Printf(
		"Consulta de métricas realizada. Total: %d",
		len(metrics),
	)

	return response, nil
}

// =====================================
// REPORTAR ESTADO
// =====================================

func (s *MonitoringServer) ReportStatus(
	ctx context.Context,
	req *pb.StatusRequest,
) (*pb.StatusResponse, error) {

	serviceStatus :=
		model.ServiceStatus{
			ComponentID: req.GetComponentId(),

			ComponentName: req.GetComponentName(),

			Status: req.GetStatus(),

			Message: req.GetMessage(),

			LastUpdated: time.Unix(
				req.GetTimestamp(),
				0,
			),
		}

	if err := s.service.RegisterServiceStatus(
		serviceStatus,
	); err != nil {

		return &pb.StatusResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Println(
		"=== ESTADO DE SERVICIO REGISTRADO ===",
	)

	log.Printf(
		"Componente: %s",
		serviceStatus.ComponentName,
	)

	log.Printf(
		"ID: %s",
		serviceStatus.ComponentID,
	)

	log.Printf(
		"Estado: %s",
		serviceStatus.Status,
	)

	log.Printf(
		"Mensaje: %s",
		serviceStatus.Message,
	)

	log.Printf(
		"Última actualización: %v",
		serviceStatus.LastUpdated,
	)

	return &pb.StatusResponse{
		Success: true,
		Message: "Estado registrado correctamente",
	}, nil
}

// =====================================
// CONSULTAR ESTADO
// =====================================

func (s *MonitoringServer) GetServiceStatus(
	ctx context.Context,
	req *pb.ServiceRequest,
) (*pb.ServiceResponse, error) {

	serviceStatus, err :=
		s.service.GetServiceStatus(
			req.GetComponentName(),
		)

	if err != nil {

		return nil,
			status.Error(
				codes.Internal,
				err.Error(),
			)
	}

	if serviceStatus == nil {

		return &pb.ServiceResponse{
			ComponentName: req.GetComponentName(),

			Status: "NOT_FOUND",

			Message: "No existe información para este servicio",
		}, nil
	}

	log.Printf(
		"Consulta de estado realizada: %s",
		serviceStatus.ComponentName,
	)

	return &pb.ServiceResponse{
		ComponentId: serviceStatus.ComponentID,

		ComponentName: serviceStatus.ComponentName,

		Status: serviceStatus.Status,

		Message: serviceStatus.Message,

		LastUpdated: serviceStatus.LastUpdated.Unix(),
	}, nil
}

// =====================================
// CONSULTAR NODO HPC
// =====================================

func (s *MonitoringServer) GetNodeStatus(
	ctx context.Context,
	req *pb.NodeRequest,
) (*pb.NodeResponse, error) {

	return &pb.NodeResponse{
		NodeId: req.GetNodeId(),

		Status: "NOT_IMPLEMENTED",
	}, nil
}

// =====================================
// CREAR REGLA DE ALERTA
// =====================================

func (s *MonitoringServer) CreateAlertRule(
	ctx context.Context,
	req *pb.CreateAlertRuleRequest,
) (*pb.CreateAlertRuleResponse, error) {

	rule :=
		model.AlertRule{
			ID: req.GetId(),

			Name: req.GetName(),

			Metric: req.GetMetric(),

			Operator: req.GetOperator(),

			Threshold: req.GetThreshold(),

			ComponentName: req.GetComponentName(),

			Enabled: req.GetEnabled(),
		}

	createdRule, err :=
		s.service.CreateAlertRule(
			rule,
		)

	if err != nil {

		return &pb.CreateAlertRuleResponse{
			Success: false,
			Message: err.Error(),
			RuleId:  rule.ID,
		}, nil
	}

	log.Println(
		"=== REGLA DE ALERTA CREADA ===",
	)

	log.Printf(
		"ID: %s",
		createdRule.ID,
	)

	log.Printf(
		"Nombre: %s",
		createdRule.Name,
	)

	log.Printf(
		"Métrica: %s",
		createdRule.Metric,
	)

	log.Printf(
		"Operador: %s",
		createdRule.Operator,
	)

	log.Printf(
		"Umbral: %.2f",
		createdRule.Threshold,
	)

	log.Printf(
		"Componente: %s",
		createdRule.ComponentName,
	)

	log.Printf(
		"Habilitada: %t",
		createdRule.Enabled,
	)

	return &pb.CreateAlertRuleResponse{
		Success: true,
		Message: "Regla de alerta creada correctamente",
		RuleId:  createdRule.ID,
	}, nil
}

// =====================================
// ACTUALIZAR REGLA DE ALERTA
// =====================================

func (s *MonitoringServer) UpdateAlertRule(
	ctx context.Context,
	req *pb.UpdateAlertRuleRequest,
) (*pb.UpdateAlertRuleResponse, error) {

	rule :=
		model.AlertRule{
			ID: req.GetId(),

			Name: req.GetName(),

			Metric: req.GetMetric(),

			Operator: req.GetOperator(),

			Threshold: req.GetThreshold(),

			ComponentName: req.GetComponentName(),

			Enabled: req.GetEnabled(),
		}

	updatedRule, err :=
		s.service.UpdateAlertRule(
			rule,
		)

	if err != nil {

		return &pb.UpdateAlertRuleResponse{
			Success: false,
			Message: err.Error(),
			Rule:    nil,
		}, nil
	}

	log.Println(
		"=== REGLA DE ALERTA ACTUALIZADA ===",
	)

	log.Printf(
		"ID: %s",
		updatedRule.ID,
	)

	log.Printf(
		"Nombre: %s",
		updatedRule.Name,
	)

	log.Printf(
		"Métrica: %s",
		updatedRule.Metric,
	)

	log.Printf(
		"Operador: %s",
		updatedRule.Operator,
	)

	log.Printf(
		"Umbral: %.2f",
		updatedRule.Threshold,
	)

	log.Printf(
		"Componente: %s",
		updatedRule.ComponentName,
	)

	log.Printf(
		"Habilitada: %t",
		updatedRule.Enabled,
	)

	return &pb.UpdateAlertRuleResponse{
		Success: true,

		Message: "Regla de alerta actualizada correctamente",

		Rule: &pb.AlertRule{
			Id: updatedRule.ID,

			Name: updatedRule.Name,

			Metric: updatedRule.Metric,

			Operator: updatedRule.Operator,

			Threshold: updatedRule.Threshold,

			ComponentName: updatedRule.ComponentName,

			Enabled: updatedRule.Enabled,
		},
	}, nil
}

// =====================================
// ELIMINAR REGLA DE ALERTA
// =====================================

func (s *MonitoringServer) DeleteAlertRule(
	ctx context.Context,
	req *pb.DeleteAlertRuleRequest,
) (*pb.DeleteAlertRuleResponse, error) {

	ruleID :=
		req.GetId()

	if err := s.service.DeleteAlertRule(
		ruleID,
	); err != nil {

		return &pb.DeleteAlertRuleResponse{
			Success: false,
			Message: err.Error(),
			RuleId:  ruleID,
		}, nil
	}

	log.Println(
		"=== REGLA DE ALERTA ELIMINADA ===",
	)

	log.Printf(
		"ID: %s",
		ruleID,
	)

	return &pb.DeleteAlertRuleResponse{
		Success: true,
		Message: "Regla de alerta eliminada correctamente",
		RuleId:  ruleID,
	}, nil
}

// =====================================
// CONSULTAR REGLAS
// =====================================

func (s *MonitoringServer) GetAlertRules(
	ctx context.Context,
	req *pb.GetAlertRulesRequest,
) (*pb.GetAlertRulesResponse, error) {

	rules, err :=
		s.service.GetAlertRules()

	if err != nil {

		return nil,
			status.Error(
				codes.Internal,
				err.Error(),
			)
	}

	response :=
		&pb.GetAlertRulesResponse{
			Rules: make(
				[]*pb.AlertRule,
				0,
				len(rules),
			),
		}

	for _, rule := range rules {

		response.Rules =
			append(
				response.Rules,
				&pb.AlertRule{
					Id: rule.ID,

					Name: rule.Name,

					Metric: rule.Metric,

					Operator: rule.Operator,

					Threshold: rule.Threshold,

					ComponentName: rule.ComponentName,

					Enabled: rule.Enabled,
				},
			)
	}

	log.Printf(
		"Consulta de reglas realizada. Total: %d",
		len(rules),
	)

	return response, nil
}

// =====================================
// CONSULTAR ALERTAS
// =====================================

func (s *MonitoringServer) GetAlerts(
	ctx context.Context,
	req *pb.GetAlertsRequest,
) (*pb.GetAlertsResponse, error) {

	alerts, err :=
		s.service.GetAlerts()

	if err != nil {

		return nil,
			status.Error(
				codes.Internal,
				err.Error(),
			)
	}

	response :=
		&pb.GetAlertsResponse{
			Alerts: make(
				[]*pb.Alert,
				0,
				len(alerts),
			),
		}

	for _, alert := range alerts {

		var resolvedAt int64

		if alert.ResolvedAt != nil {

			resolvedAt =
				alert.ResolvedAt.Unix()
		}

		response.Alerts =
			append(
				response.Alerts,
				&pb.Alert{
					Id: alert.ID,

					RuleId: alert.RuleID,

					RuleName: alert.RuleName,

					ComponentId: alert.ComponentID,

					ComponentName: alert.ComponentName,

					Metric: alert.Metric,

					CurrentValue: alert.CurrentValue,

					Threshold: alert.Threshold,

					Message: alert.Message,

					CreatedAt: alert.CreatedAt.Unix(),

					Active: alert.Active,

					ResolvedAt: resolvedAt,
				},
			)
	}

	log.Printf(
		"Consulta de alertas realizada. Total: %d",
		len(alerts),
	)

	return response, nil
}
