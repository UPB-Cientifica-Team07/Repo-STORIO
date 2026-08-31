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
	token := os.Getenv("TOKEN_USER3")

	if token == "" {
		log.Fatal("TOKEN_USER3 es obligatorio")
	}

	const fileID = "231a6faa-04e6-4751-a079-b28f6d83175b"

	connection, err := grpcClient.NewClient(
		"localhost:50053",
		grpcClient.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		log.Fatalf("No se pudo conectar: %v", err)
	}

	defer connection.Close()

	client := pb.NewFileServiceClient(connection)

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
	log.Println(" USER-003 READ TEST")
	log.Println("===================================")

	response, err := client.GetFile(
		ctx,
		&pb.GetFileRequest{
			FileId: fileID,
		},
	)

	if err != nil {
		log.Fatalf(
			"GetFile falló: %v",
			err,
		)
	}

	log.Printf("Success: %t", response.Success)
	log.Printf("Mensaje: %s", response.Message)

	if response.File != nil {
		log.Printf("Nombre: %s", response.File.FileName)
		log.Printf("Contenido: %s", string(response.File.Content))
	}

	if !response.Success {
		log.Fatal(
			"user-003 debía poder leer",
		)
	}

	log.Println("===================================")
	log.Println(" USER-003 LECTURA VALIDADA")
	log.Println("===================================")
}
