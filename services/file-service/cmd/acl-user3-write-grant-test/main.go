package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"

	grpcClient "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	token := os.Getenv("TOKEN_OWNER")
	if token == "" {
		log.Fatal("TOKEN_OWNER es obligatorio")
	}

	const fileID = "231a6faa-04e6-4751-a079-b28f6d83175b"

	conn, err := grpcClient.NewClient(
		"localhost:50053",
		grpcClient.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	if err != nil {
		log.Fatalf("No se pudo conectar: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)

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
	log.Println(" SHARE WRITE TO USER-003 TEST")
	log.Println(" Owner: user-001")
	log.Println(" Target: user-003")
	log.Println(" READ=true WRITE=true SHARE=false")
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
	if err != nil {
		log.Fatalf("ShareFile falló: %v", err)
	}

	log.Printf("Success: %t", response.Success)
	log.Printf("Mensaje: %s", response.Message)

	if !response.Success {
		log.Fatal("No se pudo conceder WRITE a user-003")
	}

	log.Println("===================================")
	log.Println(" WRITE CONCEDIDO A USER-003")
	log.Println("===================================")
}
