package main

import (
	"context"
	"log"
	"os"
	"strings"
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
		15*time.Second,
	)

	defer cancel()

	ctx = metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs(
			"authorization",
			"Bearer "+token,
		),
	)

	// =====================================
	// ARCHIVO A = 60 BYTES
	// =====================================

	contentA := []byte(
		strings.Repeat("A", 60),
	)

	log.Println("===================================")
	log.Println(" UPLOAD A - 60 BYTES")
	log.Println("===================================")

	responseA, err := client.UploadFile(
		ctx,
		&pb.UploadFileRequest{
			UserId:   "user-002",
			FileName: "quota-a.txt",
			FileType: "text/plain",
			Content:  contentA,
		},
	)

	if err != nil {
		log.Fatalf(
			"Error gRPC en Upload A: %v",
			err,
		)
	}

	log.Printf("Success: %t", responseA.Success)
	log.Printf("Mensaje: %s", responseA.Message)
	log.Printf("File ID: %s", responseA.FileId)

	if !responseA.Success {
		log.Fatal("Upload A debía ser permitido")
	}

	// =====================================
	// ARCHIVO B = 50 BYTES
	// =====================================

	contentB := []byte(
		strings.Repeat("B", 50),
	)

	log.Println("===================================")
	log.Println(" UPLOAD B - 50 BYTES")
	log.Println("===================================")

	responseB, err := client.UploadFile(
		ctx,
		&pb.UploadFileRequest{
			UserId:   "user-002",
			FileName: "quota-b.txt",
			FileType: "text/plain",
			Content:  contentB,
		},
	)

	if err != nil {
		log.Fatalf(
			"Error gRPC en Upload B: %v",
			err,
		)
	}

	log.Printf("Success: %t", responseB.Success)
	log.Printf("Mensaje: %s", responseB.Message)
	log.Printf("File ID: %s", responseB.FileId)

	if responseB.Success {
		log.Fatal(
			"ERROR: Upload B debía ser rechazado por cuota",
		)
	}

	log.Println("===================================")
	log.Println(" CUOTA BLOQUEADA CORRECTAMENTE")
	log.Println("===================================")
}
