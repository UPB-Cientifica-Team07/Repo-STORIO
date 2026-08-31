package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"
	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	syncAddress = "localhost:50055"
	authURL     = "http://localhost:8081"

	username = "tercero"
	password = "123456"

	deviceID = "sync-tree-security-device"
)

type testCase struct {
	name         string
	fileName     string
	relativePath string
}

func main() {

	log.Println("===================================")
	log.Println(" SYNC TREE SECURITY TEST")
	log.Println("===================================")

	authClient :=
		auth.NewClient(
			authURL,
		)

	loginResult, err :=
		authClient.Login(
			username,
			password,
		)

	if err != nil {
		log.Fatalf(
			"Login falló: %v",
			err,
		)
	}

	if !loginResult.Success {
		log.Fatalf(
			"Login rechazado: %s",
			loginResult.Message,
		)
	}

	log.Printf(
		"Usuario autenticado: %s",
		loginResult.UserID,
	)

	connection, err :=
		grpc.NewClient(
			syncAddress,
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		)

	if err != nil {
		log.Fatalf(
			"Error conectando con Sync: %v",
			err,
		)
	}

	defer connection.Close()

	client :=
		pb.NewSyncServiceClient(
			connection,
		)

	tests :=
		[]testCase{
			{
				name:         "Traversal ../",
				fileName:     "main.go",
				relativePath: "../otro-home/main.go",
			},
			{
				name:         "Ruta absoluta Unix",
				fileName:     "main.go",
				relativePath: "/etc/main.go",
			},
			{
				name:         "Ruta absoluta Windows",
				fileName:     "main.go",
				relativePath: `C:\Windows\main.go`,
			},
			{
				name:         "FileName inconsistente",
				fileName:     "main.go",
				relativePath: "proyecto/otro.go",
			},
			{
				name:         "Traversal intermedio",
				fileName:     "main.go",
				relativePath: "proyecto/../../user-001/main.go",
			},
		}

	for index, test := range tests {

		log.Println("===================================")

		log.Printf(
			" TEST %d - %s",
			index+1,
			test.name,
		)

		log.Println("===================================")

		err :=
			runRejectedUpload(
				client,
				loginResult.Token,
				test.fileName,
				test.relativePath,
			)

		if err != nil {
			log.Fatalf(
				"%s: %v",
				test.name,
				err,
			)
		}

		log.Printf(
			"%s = RECHAZADO CORRECTAMENTE",
			test.name,
		)
	}

	log.Println("===================================")
	log.Println(" SYNC-3 SECURITY VALIDADO")
	log.Println("===================================")
	log.Println(" Traversal ../             = BLOQUEADO")
	log.Println(" Ruta absoluta Unix        = BLOQUEADA")
	log.Println(" Ruta absoluta Windows     = BLOQUEADA")
	log.Println(" FileName inconsistente    = BLOQUEADO")
	log.Println(" Traversal intermedio      = BLOQUEADO")
	log.Println("===================================")
}

func runRejectedUpload(
	client pb.SyncServiceClient,
	token string,
	fileName string,
	relativePath string,
) error {

	content :=
		[]byte(
			"SECURITY TEST",
		)

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)

	defer cancel()

	ctx =
		metadata.NewOutgoingContext(
			ctx,
			metadata.Pairs(
				"authorization",
				"Bearer "+token,
			),
		)

	stream, err :=
		client.Upload(
			ctx,
		)

	if err != nil {
		return err
	}

	err =
		stream.Send(
			&pb.UploadRequest{
				Data: &pb.UploadRequest_Metadata{
					Metadata: &pb.UploadMetadata{
						UserId:       "user-001",
						DeviceId:     deviceID,
						FileName:     fileName,
						FileType:     "text/plain",
						Size:         int64(len(content)),
						RelativePath: relativePath,
					},
				},
			},
		)

	if err != nil {
		return err
	}

	err =
		stream.Send(
			&pb.UploadRequest{
				Data: &pb.UploadRequest_Chunk{
					Chunk: content,
				},
			},
		)

	if err != nil {
		return err
	}

	_, err =
		stream.CloseAndRecv()

	if err == nil {
		return status.Error(
			codes.Internal,
			"el servidor aceptó una ruta que debía ser rechazada",
		)
	}

	grpcStatus, ok :=
		status.FromError(
			err,
		)

	if !ok {
		return err
	}

	if grpcStatus.Code() !=
		codes.InvalidArgument {

		return status.Errorf(
			codes.Internal,
			"se esperaba InvalidArgument pero se recibió %s: %s",
			grpcStatus.Code(),
			grpcStatus.Message(),
		)
	}

	message :=
		strings.TrimSpace(
			grpcStatus.Message(),
		)

	if message == "" {
		return status.Error(
			codes.Internal,
			"el servidor rechazó la petición sin mensaje",
		)
	}

	log.Printf(
		"Respuesta esperada: %s",
		message,
	)

	return nil
}
