package main

import (
	"context"
	"log"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"
	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	serverAddress = "localhost:50055"
	authURL       = "http://localhost:8081"

	fileID = "d3ef9c22-8300-44de-900f-e5806b581274"

	deviceID = "sync-home-delete-device"
)

func main() {

	log.Println("===================================")
	log.Println(" SYNC HOME DELETE TEST")
	log.Println("===================================")

	authClient :=
		auth.NewClient(
			authURL,
		)

	loginResult, err :=
		authClient.Login(
			"tercero",
			"123456",
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

	conn, err :=
		grpc.NewClient(
			serverAddress,
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		)

	if err != nil {
		log.Fatalf(
			"Error conectando con Sync Service: %v",
			err,
		)
	}

	defer conn.Close()

	client :=
		pb.NewSyncServiceClient(
			conn,
		)

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
		)

	defer cancel()

	ctx =
		metadata.NewOutgoingContext(
			ctx,
			metadata.Pairs(
				"authorization",
				"Bearer "+loginResult.Token,
			),
		)

	response, err :=
		client.DeleteFile(
			ctx,
			&pb.DeleteFileRequest{
				UserId:   "user-001",
				DeviceId: deviceID,
				FileId:   fileID,
			},
		)

	if err != nil {
		log.Fatalf(
			"DeleteFile falló: %v",
			err,
		)
	}

	if !response.Success {
		log.Fatalf(
			"DeleteFile respondió Success=false: %s",
			response.Message,
		)
	}

	log.Println(
		"Archivo eliminado correctamente mediante Sync",
	)

	log.Println("===================================")
	log.Println(" DELETE END-TO-END = OK")
	log.Println("===================================")
}
