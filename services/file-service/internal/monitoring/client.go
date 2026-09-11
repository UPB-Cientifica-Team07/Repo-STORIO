package monitoring

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/generated"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const monitoringAddress = "localhost:50051"

type Client struct {
	conn             *grpc.ClientConn
	monitoringClient pb.MonitoringServiceClient
	componentID      string
	componentName    string
	serviceToken     string
}

func NewClient(
	componentID string,
	componentName string,
) (*Client, error) {

	serviceToken :=
		strings.TrimSpace(
			os.Getenv("MONITORING_SERVICE_TOKEN"),
		)

	if serviceToken == "" {
		return nil,
			fmt.Errorf(
				"MONITORING_SERVICE_TOKEN es obligatorio",
			)
	}

	conn, err := grpc.NewClient(
		monitoringAddress,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		return nil, err
	}

	client :=
		pb.NewMonitoringServiceClient(
			conn,
		)

	return &Client{
		conn:             conn,
		monitoringClient: client,
		componentID:      componentID,
		componentName:    componentName,
		serviceToken:     serviceToken,
	}, nil
}

func (c *Client) Close() error {

	return c.conn.Close()
}

// =====================================
// REPORTAR ESTADO
// =====================================

func (c *Client) ReportStatus(
	status string,
	message string,
) error {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	ctx =
		metadata.AppendToOutgoingContext(
			ctx,
			"x-monitoring-service-token",
			c.serviceToken,
		)

	response, err :=
		c.monitoringClient.ReportStatus(
			ctx,
			&pb.StatusRequest{
				ComponentId:   c.componentID,
				ComponentName: c.componentName,
				Status:        status,
				Message:       message,
				Timestamp:     time.Now().Unix(),
			},
		)

	if err != nil {

		return err
	}

	if !response.GetSuccess() {

		return fmt.Errorf(
			"Monitoring Service rechazó el estado: %s",
			response.GetMessage(),
		)
	}

	log.Printf(
		"Estado reportado | Estado: %s | Mensaje: %s",
		status,
		message,
	)

	return nil
}

// =====================================
// REPORTAR MÉTRICAS
// =====================================

func (c *Client) ReportMetrics(
	cpuUsage float64,
	memoryUsage float64,
	storageUsage int64,
	activeConnections int64,
	totalRequests int64,
) error {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	ctx =
		metadata.AppendToOutgoingContext(
			ctx,
			"x-monitoring-service-token",
			c.serviceToken,
		)

	response, err :=
		c.monitoringClient.ReportMetrics(
			ctx,
			&pb.MetricsRequest{
				ComponentId:       c.componentID,
				ComponentName:     c.componentName,
				CpuUsage:          cpuUsage,
				MemoryUsage:       memoryUsage,
				StorageUsage:      storageUsage,
				ActiveConnections: activeConnections,
				TotalRequests:     totalRequests,
				Timestamp:         time.Now().Unix(),
			},
		)

	if err != nil {

		return err
	}

	if !response.GetSuccess() {

		return fmt.Errorf(
			"Monitoring Service rechazó las métricas: %s",
			response.GetMessage(),
		)
	}

	log.Printf(
		"Métricas reales | CPU: %.2f%% | RAM: %.2f MB | Storage: %d bytes | Connections: %d | Requests: %d",
		cpuUsage,
		memoryUsage,
		storageUsage,
		activeConnections,
		totalRequests,
	)

	return nil
}
