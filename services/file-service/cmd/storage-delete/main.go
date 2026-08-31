package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {

	token :=
		os.Getenv(
			"TOKEN_ADMIN",
		)

	fileID :=
		os.Getenv(
			"STORAGE_TEST_FILE_ID",
		)

	if token == "" {
		log.Fatal(
			"TOKEN_ADMIN no definido",
		)
	}

	if fileID == "" {
		log.Fatal(
			"STORAGE_TEST_FILE_ID no definido",
		)
	}

	conn,
		err :=
		grpc.NewClient(
			"localhost:50053",
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		)

	if err != nil {
		log.Fatalf(
			"Error conectando a File Service: %v",
			err,
		)
	}

	defer conn.Close()

	client :=
		pb.NewFileServiceClient(
			conn,
		)

	md :=
		metadata.New(
			map[string]string{
				"authorization": "Bearer " + token,
			},
		)

	ctx :=
		metadata.NewOutgoingContext(
			context.Background(),
			md,
		)

	ctx,
		cancel :=
		context.WithTimeout(
			ctx,
			10*time.Second,
		)

	defer cancel()

	response,
		err :=
		client.DeleteFile(
			ctx,
			&pb.DeleteFileRequest{
				FileId: fileID,
			},
		)

	if err != nil {
		log.Fatalf(
			"DeleteFile RPC error: %v",
			err,
		)
	}

	log.Println(
		"===================================",
	)

	log.Println(
		" STORAGE DELETE TEST",
	)

	log.Println(
		"===================================",
	)

	log.Printf(
		"Success: %t",
		response.Success,
	)

	log.Printf(
		"Message: %s",
		response.Message,
	)

	log.Printf(
		"File ID eliminado: %s",
		fileID,
	)

	log.Println(
		"===================================",
	)
}
