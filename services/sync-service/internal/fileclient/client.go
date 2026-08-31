package fileclient

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	connection *grpc.ClientConn
	client     pb.FileServiceClient
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewClient(
	address string,
) (*Client, error) {

	address =
		strings.TrimSpace(
			address,
		)

	if address == "" {

		return nil,
			errors.New(
				"la dirección de File Service es obligatoria",
			)
	}

	connection, err :=
		grpc.NewClient(
			address,
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"error conectando con File Service: %w",
				err,
			)
	}

	return &Client{
		connection: connection,
		client: pb.NewFileServiceClient(
			connection,
		),
	}, nil
}

// =====================================
// CLOSE
// =====================================

func (c *Client) Close() error {

	if c == nil ||
		c.connection == nil {

		return nil
	}

	return c.connection.Close()
}

// =====================================
// PROPAGAR AUTHORIZATION
// =====================================
//
// Sync recibe Authorization como metadata
// entrante.
//
// gRPC no propaga automáticamente esa metadata
// cuando Sync realiza una llamada hacia otro
// servicio.
//
// Por eso copiamos explícitamente:
//
// incoming Authorization
//          ↓
// outgoing Authorization
//
// =====================================

func authenticatedOutgoingContext(
	ctx context.Context,
) (context.Context, error) {

	if ctx == nil {

		return nil,
			errors.New(
				"el contexto es obligatorio",
			)
	}

	incomingMetadata, ok :=
		metadata.FromIncomingContext(
			ctx,
		)

	if !ok {

		return nil,
			errors.New(
				"metadata de autenticación no encontrada",
			)
	}

	values :=
		incomingMetadata.Get(
			"authorization",
		)

	if len(values) == 0 {

		return nil,
			errors.New(
				"Authorization no encontrado",
			)
	}

	authorization :=
		strings.TrimSpace(
			values[0],
		)

	if authorization == "" {

		return nil,
			errors.New(
				"Authorization vacío",
			)
	}

	outgoingMetadata :=
		metadata.Pairs(
			"authorization",
			authorization,
		)

	return metadata.NewOutgoingContext(
		ctx,
		outgoingMetadata,
	), nil
}

// =====================================
// UPLOAD
// =====================================
//
// relativeDirectory:
//
// "":
//   Shared File mantiene clasificación
//   documentos / imagenes / videos.
//
// Con valor:
//
//   File Sync conserva el árbol.
//
// Ejemplo:
//
//   proyecto/codigo
//
// =====================================

func (c *Client) Upload(
	ctx context.Context,
	fileName string,
	fileType string,
	content []byte,
	relativeDirectory string,
) (*pb.UploadFileResponse, error) {

	if c == nil ||
		c.client == nil {

		return nil,
			errors.New(
				"File Service Client no inicializado",
			)
	}

	fileName =
		strings.TrimSpace(
			fileName,
		)

	if fileName == "" {

		return nil,
			errors.New(
				"el nombre del archivo es obligatorio",
			)
	}

	relativeDirectory =
		strings.TrimSpace(
			relativeDirectory,
		)

	authenticatedCtx, err :=
		authenticatedOutgoingContext(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	callCtx, cancel :=
		context.WithTimeout(
			authenticatedCtx,
			30*time.Second,
		)

	defer cancel()

	// =====================================
	// USER ID VACÍO DELIBERADAMENTE
	// =====================================
	//
	// File Service obtiene el propietario
	// real desde el Bearer token.
	//
	// =====================================

	response, err :=
		c.client.UploadFile(
			callCtx,
			&pb.UploadFileRequest{
				UserId:            "",
				FileName:          fileName,
				FileType:          fileType,
				Content:           content,
				RelativeDirectory: relativeDirectory,
			},
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"UploadFile en File Service falló: %w",
				err,
			)
	}

	return response, nil
}

// =====================================
// GET
// =====================================

func (c *Client) Get(
	ctx context.Context,
	fileID string,
) (*pb.GetFileResponse, error) {

	if c == nil ||
		c.client == nil {

		return nil,
			errors.New(
				"File Service Client no inicializado",
			)
	}

	fileID =
		strings.TrimSpace(
			fileID,
		)

	if fileID == "" {

		return nil,
			errors.New(
				"el file ID es obligatorio",
			)
	}

	authenticatedCtx, err :=
		authenticatedOutgoingContext(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	callCtx, cancel :=
		context.WithTimeout(
			authenticatedCtx,
			30*time.Second,
		)

	defer cancel()

	response, err :=
		c.client.GetFile(
			callCtx,
			&pb.GetFileRequest{
				FileId: fileID,
			},
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"GetFile en File Service falló: %w",
				err,
			)
	}

	return response, nil
}

// =====================================
// UPDATE
// =====================================

func (c *Client) Update(
	ctx context.Context,
	fileID string,
	content []byte,
) (*pb.UpdateFileResponse, error) {

	if c == nil ||
		c.client == nil {

		return nil,
			errors.New(
				"File Service Client no inicializado",
			)
	}

	fileID =
		strings.TrimSpace(
			fileID,
		)

	if fileID == "" {

		return nil,
			errors.New(
				"el file ID es obligatorio",
			)
	}

	authenticatedCtx, err :=
		authenticatedOutgoingContext(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	callCtx, cancel :=
		context.WithTimeout(
			authenticatedCtx,
			30*time.Second,
		)

	defer cancel()

	response, err :=
		c.client.UpdateFile(
			callCtx,
			&pb.UpdateFileRequest{
				FileId:  fileID,
				Content: content,
			},
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"UpdateFile en File Service falló: %w",
				err,
			)
	}

	return response, nil
}

// =====================================
// DELETE
// =====================================

func (c *Client) Delete(
	ctx context.Context,
	fileID string,
) (*pb.DeleteFileResponse, error) {

	if c == nil ||
		c.client == nil {

		return nil,
			errors.New(
				"File Service Client no inicializado",
			)
	}

	fileID =
		strings.TrimSpace(
			fileID,
		)

	if fileID == "" {

		return nil,
			errors.New(
				"el file ID es obligatorio",
			)
	}

	authenticatedCtx, err :=
		authenticatedOutgoingContext(
			ctx,
		)

	if err != nil {
		return nil, err
	}

	callCtx, cancel :=
		context.WithTimeout(
			authenticatedCtx,
			30*time.Second,
		)

	defer cancel()

	response, err :=
		c.client.DeleteFile(
			callCtx,
			&pb.DeleteFileRequest{
				FileId: fileID,
			},
		)

	if err != nil {

		return nil,
			fmt.Errorf(
				"DeleteFile en File Service falló: %w",
				err,
			)
	}

	return response, nil
}
