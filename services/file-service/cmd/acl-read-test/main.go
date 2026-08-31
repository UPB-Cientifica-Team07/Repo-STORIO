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
	token := os.Getenv("TOKEN_USER2")

	if token == "" {
		log.Fatal("TOKEN_USER2 es obligatorio")
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
	log.Println(" ACL READ TEST")
	log.Println(" Usuario: user-002")
	log.Printf(" Archivo: %s", fileID)
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
		log.Printf("Tipo: %s", response.File.FileType)
		log.Printf("Tamaño: %d bytes", len(response.File.Content))
		log.Printf("Contenido: %s", string(response.File.Content))
	}

	if !response.Success {
		log.Fatal(
			"user-002 debía poder leer el archivo compartido",
		)
	}

	log.Println("===================================")
	log.Println(" ACL DE LECTURA VALIDADA")
	log.Println("===================================")
}
