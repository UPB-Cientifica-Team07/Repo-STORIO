package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"

	grpcClient "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
	log.Println(" USER-003 UPDATE DENIED TEST")
	log.Println(" Esperado: PermissionDenied")
	log.Println("===================================")

	response, err := client.UpdateFile(
		ctx,
		&pb.UpdateFileRequest{
			FileId: fileID,
			Content: []byte(
				"Intento de modificación no autorizado por user-003.",
			),
		},
	)

	if err == nil {
		log.Printf(
			"Respuesta inesperada: success=%t message=%s",
			response.Success,
			response.Message,
		)

		log.Fatal(
			"ERROR: user-003 pudo modificar sin can_write",
		)
	}

	code := status.Code(err)

	log.Printf("Código gRPC: %s", code)
	log.Printf("Error: %v", err)

	if code != codes.PermissionDenied {
		log.Fatalf(
			"Se esperaba PermissionDenied, se obtuvo %s",
			code,
		)
	}

	log.Println("===================================")
	log.Println(" UPDATE BLOQUEADO CORRECTAMENTE")
	log.Println("===================================")
}
