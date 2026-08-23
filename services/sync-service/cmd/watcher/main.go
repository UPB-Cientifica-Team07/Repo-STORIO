package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	serverAddress = "localhost:50055"
	authURL       = "http://localhost:8081"
)

func main() {

	fmt.Println("===================================")
	fmt.Println(" WATCH CLIENT")
	fmt.Println(" Device: mobile-001")
	fmt.Println("===================================")

	// =====================================
	// LOGIN
	// =====================================

	authClient := auth.NewClient(
		authURL,
	)

	loginResult, err := authClient.Login(
		"samuel",
		"123456",
	)

	if err != nil {
		log.Fatalf(
			"Error Login: %v",
			err,
		)
	}

	if !loginResult.Success {
		log.Fatalf(
			"Login rechazado: %s",
			loginResult.Message,
		)
	}

	fmt.Println(
		"Login correcto:",
		loginResult.UserID,
	)

	// =====================================
	// CONEXIÓN GRPC
	// =====================================

	conn, err := grpc.NewClient(
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

	client := pb.NewSyncServiceClient(
		conn,
	)

	ctx := context.Background()

	// =====================================
	// AUTHENTICATE
	// =====================================

	authResponse, err := client.Authenticate(
		ctx,
		&pb.AuthenticateRequest{
			Token:    loginResult.Token,
			DeviceId: "mobile-001",
		},
	)

	if err != nil {
		log.Fatalf(
			"Error Authenticate: %v",
			err,
		)
	}

	if !authResponse.Success {
		log.Fatalf(
			"Autenticación rechazada: %s",
			authResponse.Message,
		)
	}

	fmt.Println(
		"Autenticado como:",
		authResponse.UserId,
	)

	// =====================================
	// WATCH CHANGES
	// =====================================

	stream, err := client.WatchChanges(
		ctx,
		&pb.WatchChangesRequest{
			UserId:   authResponse.UserId,
			DeviceId: "mobile-001",
		},
	)

	if err != nil {
		log.Fatalf(
			"Error WatchChanges: %v",
			err,
		)
	}

	fmt.Println()
	fmt.Println(
		"Esperando cambios...",
	)

	for {

		change, err := stream.Recv()

		if err == io.EOF {
			fmt.Println(
				"Stream finalizado",
			)
			return
		}

		if err != nil {
			log.Fatalf(
				"Error recibiendo cambio: %v",
				err,
			)
		}

		fmt.Println()
		fmt.Println("CAMBIO RECIBIDO")

		fmt.Println(
			"Tipo:",
			change.Type,
		)

		fmt.Println(
			"Archivo:",
			change.FileName,
		)

		fmt.Println(
			"File ID:",
			change.FileId,
		)

		fmt.Println(
			"Origen:",
			change.OriginDeviceId,
		)
	}
}