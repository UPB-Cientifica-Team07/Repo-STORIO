package monitoring

import (
	"context"
	"fmt"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/generated"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.MonitoringServiceClient
}

func NewClient(
	address string,
) (*Client, error) {

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		return nil, fmt.Errorf(
			"error conectando con Monitoring Service: %w",
			err,
		)
	}

	client :=
		pb.NewMonitoringServiceClient(
			conn,
		)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c *Client) Close() error {

	if c.conn == nil {
		return nil
	}

	return c.conn.Close()
}

// =====================================
// REPORT STATUS
// =====================================

func (c *Client) ReportStatus(
	componentID string,
	componentName string,
	statusValue string,
	message string,
) error {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			3*time.Second,
		)

	defer cancel()

	response, err :=
		c.client.ReportStatus(
			ctx,
			&pb.StatusRequest{
				ComponentId:   componentID,
				ComponentName: componentName,
				Status:        statusValue,
				Message:       message,
				Timestamp:     time.Now().Unix(),
			},
		)

	if err != nil {
		return fmt.Errorf(
			"error reportando estado: %w",
			err,
		)
	}

	if !response.Success {
		return fmt.Errorf(
			"Monitoring Service rechazó estado: %s",
			response.Message,
		)
	}

	return nil
}

// =====================================
// REPORT METRICS
// =====================================

func (c *Client) ReportMetrics(
	componentID string,
	componentName string,
	cpuUsage float64,
	memoryUsage float64,
	storageUsage int64,
	activeConnections int64,
	totalRequests int64,
) error {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			3*time.Second,
		)

	defer cancel()

	response, err :=
		c.client.ReportMetrics(
			ctx,
			&pb.MetricsRequest{
				ComponentId:       componentID,
				ComponentName:     componentName,
				CpuUsage:          cpuUsage,
				MemoryUsage:       memoryUsage,
				StorageUsage:      storageUsage,
				ActiveConnections: activeConnections,
				TotalRequests:     totalRequests,
				Timestamp:         time.Now().Unix(),
			},
		)

	if err != nil {
		return fmt.Errorf(
			"error reportando métricas: %w",
			err,
		)
	}

	if !response.Success {
		return fmt.Errorf(
			"Monitoring Service rechazó métricas: %s",
			response.Message,
		)
	}

	return nil
}
