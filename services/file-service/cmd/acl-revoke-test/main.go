package main

import (
	"context"
	"log"
	"os"
	"time"

	grpcClient "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"
)

func main() {
	token := os.Getenv("TOKEN_OWNER")

	if token == "" {
		log.Fatal("TOKEN_OWNER es obligatorio")
	}

	const fileID = "231a6faa-04e6-4751-a079-b28f6d83175b"

	connection, err := grpcClient.NewClient(
		"localhost:50053",
		grpcClient.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		log.Fatalf(
			"No se pudo conectar con File Service: %v",
			err,
		)
	}

	defer connection.Close()

	client := pb.NewFileServiceClient(
		connection,
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()

	ctx = metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs(
			"authorization",
			"Bearer "+token,
		),
	)

	log.Println("===================================")
	log.Println(" REVOKE FILE ACCESS TEST")
	log.Println(" Owner: user-001")
	log.Println(" Target: user-003")
	log.Printf(" File ID: %s", fileID)
	log.Println("===================================")

	response, err := client.RevokeFileAccess(
		ctx,
		&pb.RevokeFileAccessRequest{
			FileId:       fileID,
			TargetUserId: "user-003",
		},
	)

	if err != nil {
		log.Fatalf(
			"RevokeFileAccess falló: %v",
			err,
		)
	}

	log.Printf(
		"Success: %t",
		response.Success,
	)

	log.Printf(
		"Mensaje: %s",
		response.Message,
	)

	if !response.Success {
		log.Fatal(
			"RevokeFileAccess respondió success=false",
		)
	}

	log.Println("===================================")
	log.Println(" REVOKE COMPLETADO")
	log.Println("===================================")
}
