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
	log.Println(" ACL ESCALATION DENIED TEST")
	log.Println(" Delegador: user-002")
	log.Println(" Target: user-003")
	log.Println(" Esperado: PermissionDenied")
	log.Println(" Intento:")
	log.Println("   read  = true")
	log.Println("   write = true")
	log.Println("   share = false")
	log.Println("===================================")

	response, err := client.ShareFile(
		ctx,
		&pb.ShareFileRequest{
			FileId:       fileID,
			TargetUserId: "user-003",
			CanRead:      true,
			CanWrite:     true,
			CanShare:     false,
		},
	)

	if err == nil {
		log.Printf(
			"Respuesta inesperada: success=%t message=%s",
			response.Success,
			response.Message,
		)

		log.Fatal(
			"ERROR: user-002 logró conceder write sin poseerlo",
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
	log.Println(" ESCALAMIENTO BLOQUEADO")
	log.Println("===================================")
}
