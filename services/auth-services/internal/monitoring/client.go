package monitoring

import (
	"context"
	"log"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/generated"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const monitoringAddress = "localhost:50051"

type Client struct {
	conn             *grpc.ClientConn
	monitoringClient pb.MonitoringServiceClient
	componentID      string
	componentName    string
}

func NewClient(
	componentID string,
	componentName string,
) (*Client, error) {

	conn, err := grpc.NewClient(
		monitoringAddress,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		return nil, err
	}

	client := pb.NewMonitoringServiceClient(conn)

	return &Client{
		conn:             conn,
		monitoringClient: client,
		componentID:      componentID,
		componentName:    componentName,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) ReportStatus(
	status string,
	message string,
) error {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	response, err := c.monitoringClient.ReportStatus(
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
		return nil
	}

	log.Printf(
		"Estado reportado al Monitoring Service: %s - %s",
		status,
		message,
	)

	return nil
}

func (c *Client) ReportMetrics(
	cpuUsage float64,
	memoryUsage float64,
	activeConnections int64,
	totalRequests int64,
) error {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	response, err := c.monitoringClient.ReportMetrics(
		ctx,
		&pb.MetricsRequest{
			ComponentId:       c.componentID,
			ComponentName:     c.componentName,
			CpuUsage:          cpuUsage,
			MemoryUsage:       memoryUsage,
			ActiveConnections: activeConnections,
			TotalRequests:     totalRequests,
			Timestamp:         time.Now().Unix(),
		},
	)

	if err != nil {
		return err
	}

	if !response.GetSuccess() {
		return nil
	}

	return nil
}
