package grpc

import (
	"context"
	"log"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/generated"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/model"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/service"
)

type MonitoringServer struct {
	pb.UnimplementedMonitoringServiceServer
	monitoringService *service.MonitoringService
}

func NewMonitoringServer(
	monitoringService *service.MonitoringService,
) *MonitoringServer {
	return &MonitoringServer{
		monitoringService: monitoringService,
	}
}

// ==============================
// REGISTRAR MÉTRICAS
// ==============================

func (s *MonitoringServer) ReportMetrics(
	ctx context.Context,
	req *pb.MetricsRequest,
) (*pb.MetricsResponse, error) {

	metric := model.Metric{
		ComponentID:       req.GetComponentId(),
		ComponentName:     req.GetComponentName(),
		CPUUsage:          req.GetCpuUsage(),
		MemoryUsage:       req.GetMemoryUsage(),
		ActiveConnections: req.GetActiveConnections(),
		TotalRequests:     req.GetTotalRequests(),
		Timestamp:         time.Unix(req.GetTimestamp(), 0),
	}

	err := s.monitoringService.RegisterMetric(metric)

	if err != nil {
		log.Printf("Error registrando métrica: %v", err)

		return &pb.MetricsResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Println("=== MÉTRICA REGISTRADA ===")
	log.Println("Componente:", metric.ComponentName)
	log.Println("ID:", metric.ComponentID)
	log.Printf("CPU: %.2f", metric.CPUUsage)
	log.Printf("Memoria: %.2f", metric.MemoryUsage)
	log.Println("Conexiones activas:", metric.ActiveConnections)
	log.Println("Total solicitudes:", metric.TotalRequests)
	log.Println("Timestamp:", metric.Timestamp)

	return &pb.MetricsResponse{
		Success: true,
		Message: "Métrica registrada correctamente",
	}, nil
}

// ==============================
// CONSULTAR MÉTRICAS
// ==============================

func (s *MonitoringServer) GetMetrics(
	ctx context.Context,
	req *pb.GetMetricsRequest,
) (*pb.GetMetricsResponse, error) {

	metrics, err := s.monitoringService.GetMetrics()

	if err != nil {
		log.Printf("Error consultando métricas: %v", err)
		return nil, err
	}

	response := &pb.GetMetricsResponse{
		Metrics: make([]*pb.Metric, 0, len(metrics)),
	}

	for _, metric := range metrics {
		response.Metrics = append(
			response.Metrics,
			&pb.Metric{
				ComponentId:       metric.ComponentID,
				ComponentName:     metric.ComponentName,
				CpuUsage:          metric.CPUUsage,
				MemoryUsage:       metric.MemoryUsage,
				ActiveConnections: metric.ActiveConnections,
				TotalRequests:     metric.TotalRequests,
				Timestamp:         metric.Timestamp.Unix(),
			},
		)
	}

	log.Printf(
		"Consulta de métricas realizada. Total: %d",
		len(response.Metrics),
	)

	return response, nil
}

// ==============================
// REGISTRAR ESTADO DE SERVICIO
// ==============================

func (s *MonitoringServer) ReportStatus(
	ctx context.Context,
	req *pb.StatusRequest,
) (*pb.StatusResponse, error) {

	serviceStatus := model.ServiceStatus{
		ComponentID:   req.GetComponentId(),
		ComponentName: req.GetComponentName(),
		Status:        req.GetStatus(),
		Message:       req.GetMessage(),
		LastUpdated:   time.Unix(req.GetTimestamp(), 0),
	}

	err := s.monitoringService.RegisterServiceStatus(
		serviceStatus,
	)

	if err != nil {
		log.Printf(
			"Error registrando estado del servicio: %v",
			err,
		)

		return &pb.StatusResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	log.Println("=== ESTADO DE SERVICIO REGISTRADO ===")
	log.Println("Componente:", serviceStatus.ComponentName)
	log.Println("ID:", serviceStatus.ComponentID)
	log.Println("Estado:", serviceStatus.Status)
	log.Println("Mensaje:", serviceStatus.Message)
	log.Println("Última actualización:", serviceStatus.LastUpdated)

	return &pb.StatusResponse{
		Success: true,
		Message: "Estado del servicio registrado correctamente",
	}, nil
}

// ==============================
// CONSULTAR ESTADO DE SERVICIO
// ==============================

func (s *MonitoringServer) GetServiceStatus(
	ctx context.Context,
	req *pb.ServiceRequest,
) (*pb.ServiceResponse, error) {

	serviceStatus, err := s.monitoringService.GetServiceStatus(
		req.GetComponentName(),
	)

	if err != nil {
		log.Printf(
			"Error consultando estado del servicio: %v",
			err,
		)

		return nil, err
	}

	if serviceStatus == nil {
		return &pb.ServiceResponse{
			ComponentName: req.GetComponentName(),
			Status:        "NOT_FOUND",
			Message:       "No existe información para este servicio",
		}, nil
	}

	log.Printf(
		"Consulta de estado realizada: %s",
		serviceStatus.ComponentName,
	)

	return &pb.ServiceResponse{
		ComponentId:   serviceStatus.ComponentID,
		ComponentName: serviceStatus.ComponentName,
		Status:        serviceStatus.Status,
		Message:       serviceStatus.Message,
		LastUpdated:   serviceStatus.LastUpdated.Unix(),
	}, nil
}
