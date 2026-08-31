package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"

	grpcClient "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	token := os.Getenv("TOKEN_USER3")

	if token == "" {
		log.Fatal("TOKEN_USER3 es obligatorio")
	}

	const fileID = "231a6faa-04e6-4751-a079-b28f6d83175b"

	newContent := []byte(
		strings.Repeat("X", 60),
	)

	if len(newContent) != 60 {
		log.Fatalf(
			"contenido de prueba inválido: %d bytes",
			len(newContent),
		)
	}

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
	log.Println(" UPDATE QUOTA DENIED TEST")
	log.Println("===================================")
	log.Printf("Contenido solicitado: %d bytes", len(newContent))
	log.Println("Esperado: rechazo por cuota")
	log.Println("===================================")

	response, err := client.UpdateFile(
		ctx,
		&pb.UpdateFileRequest{
			FileId:  fileID,
			Content: newContent,
		},
	)

	if err != nil {
		log.Fatalf(
			"El servidor devolvió error gRPC inesperado: %v",
			err,
		)
	}

	log.Printf("Success: %t", response.Success)
	log.Printf("Mensaje: %s", response.Message)

	if response.Success {
		log.Fatal(
			"ERROR: UpdateFile debía fallar por cuota insuficiente",
		)
	}

	log.Println("===================================")
	log.Println(" UPDATE RECHAZADO POR CUOTA")
	log.Println("===================================")
}
