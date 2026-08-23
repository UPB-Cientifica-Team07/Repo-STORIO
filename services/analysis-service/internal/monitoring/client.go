package monitoring

import (
	"context"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/generated"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	connection *grpc.ClientConn
	client     pb.MonitoringServiceClient

	componentID   string
	componentName string
}

func NewClient(
	componentID string,
	componentName string,
) (*Client, error) {

	connection, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		return nil, err
	}

	client := pb.NewMonitoringServiceClient(
		connection,
	)

	return &Client{
		connection:    connection,
		client:        client,
		componentID:   componentID,
		componentName: componentName,
	}, nil
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

	_, err := c.client.ReportStatus(
		ctx,
		&pb.StatusRequest{
			ComponentId:   c.componentID,
			ComponentName: c.componentName,
			Status:        status,
			Message:       message,
			Timestamp:     time.Now().Unix(),
		},
	)

	return err
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

	_, err := c.client.ReportMetrics(
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

	return err
}

func (c *Client) Close() error {

	return c.connection.Close()
}
