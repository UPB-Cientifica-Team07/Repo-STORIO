package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func authenticatedContext(
	token string,
) (context.Context, context.CancelFunc) {

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

	return context.WithTimeout(
		ctx,
		10*time.Second,
	)
}

func main() {

	token :=
		os.Getenv(
			"TOKEN_ADMIN",
		)

	if token == "" {

		log.Fatal(
			"TOKEN_ADMIN no definido",
		)
	}

	filePath :=
		"/tmp/storage-metric-test.bin"

	content,
		err :=
		os.ReadFile(
			filePath,
		)

	if err != nil {

		log.Fatalf(
			"No se pudo leer %s: %v",
			filePath,
			err,
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
			"No se pudo conectar: %v",
			err,
		)
	}

	defer conn.Close()

	client :=
		pb.NewFileServiceClient(
			conn,
		)

	// =====================================
	// UPLOAD
	// =====================================

	ctxUpload,
		cancelUpload :=
		authenticatedContext(
			token,
		)

	uploadResponse,
		err :=
		client.UploadFile(
			ctxUpload,
			&pb.UploadFileRequest{
				UserId: "user-001",

				FileName: "storage-metric-test.bin",

				FileType: "application/octet-stream",

				Content: content,
			},
		)

	cancelUpload()

	if err != nil {

		log.Fatalf(
			"UploadFile error: %v",
			err,
		)
	}

	if !uploadResponse.Success {

		log.Fatalf(
			"Upload rechazado: %s",
			uploadResponse.Message,
		)
	}

	fmt.Println(
		"===================================",
	)

	fmt.Println(
		" STORAGE TEST",
	)

	fmt.Println(
		"===================================",
	)

	fmt.Printf(
		"File ID: %s\n",
		uploadResponse.FileId,
	)

	fmt.Printf(
		"Bytes subidos: %d\n",
		len(content),
	)

	fmt.Println(
		"Archivo almacenado correctamente",
	)

	fmt.Println(
		"===================================",
	)

	fmt.Println()
	fmt.Println(
		"Para eliminarlo después:",
	)

	fmt.Printf(
		"export STORAGE_TEST_FILE_ID=\"%s\"\n",
		uploadResponse.FileId,
	)
}
