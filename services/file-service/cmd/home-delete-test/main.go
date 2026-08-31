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

	token := os.Getenv("TOKEN")

	if token == "" {
		log.Fatal("La variable TOKEN es obligatoria")
	}

	const fileID = "315475e3-b462-4a72-89ca-a72bc20cdce1"

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
	log.Println(" ELIMINANDO ARCHIVO DEL HOME")
	log.Println("===================================")
	log.Printf("File ID: %s", fileID)

	response, err := client.DeleteFile(
		ctx,
		&pb.DeleteFileRequest{
			FileId: fileID,
		},
	)

	if err != nil {
		log.Fatalf(
			"Error gRPC eliminando archivo: %v",
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
			"DeleteFile respondió success=false",
		)
	}

	log.Println("===================================")
	log.Println(" DELETE COMPLETADO")
	log.Println("===================================")
}
