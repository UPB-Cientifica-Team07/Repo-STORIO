package fileclient

import (
	"context"
	"errors"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	connection *grpc.ClientConn
	client     pb.FileServiceClient
}

func NewClient() (*Client, error) {

	connection, err := grpc.NewClient(
		"localhost:50053",
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		return nil, err
	}

	client := pb.NewFileServiceClient(
		connection,
	)

	return &Client{
		connection: connection,
		client:     client,
	}, nil
}

func (c *Client) GetFileContent(
	fileID string,
) (string, error) {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)

	defer cancel()

	response, err := c.client.GetFile(
		ctx,
		&pb.GetFileRequest{
			FileId: fileID,
		},
	)

	if err != nil {
		return "", err
	}

	if !response.Success {
		return "", errors.New(response.Message)
	}

	if response.File == nil {
		return "", errors.New(
			"el File Service no devolvió información del archivo",
		)
	}

	return string(response.File.Content), nil
}

func (c *Client) Close() error {

	return c.connection.Close()
}
