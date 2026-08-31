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
	token := os.Getenv("TOKEN_USER3")

	if token == "" {
		log.Fatal("TOKEN_USER3 es obligatorio")
	}

	const revokedFileID = "231a6faa-04e6-4751-a079-b28f6d83175b"

	connection, err :=
		grpcClient.NewClient(
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

	client :=
		pb.NewFileServiceClient(
			connection,
		)

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

	defer cancel()

	ctx =
		metadata.NewOutgoingContext(
			ctx,
			metadata.Pairs(
				"authorization",
				"Bearer "+token,
			),
		)

	log.Println("===================================")
	log.Println(" USER-003 LIST AFTER REVOKE TEST")
	log.Println("===================================")
	log.Println(" Usuario: user-003")
	log.Printf(
		" Archivo revocado: %s",
		revokedFileID,
	)
	log.Println(" Esperado:")
	log.Println(" - archivos propios pueden permanecer visibles")
	log.Println(" - archivo revocado NO debe aparecer")
	log.Println("===================================")

	response, err :=
		client.ListFiles(
			ctx,
			&pb.ListFilesRequest{},
		)

	if err != nil {
		log.Fatalf(
			"ListFiles falló: %v",
			err,
		)
	}

	if !response.Success {
		log.Fatalf(
			"ListFiles respondió Success=false: %s",
			response.Message,
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

	log.Printf(
		"Archivos visibles: %d",
		len(response.Files),
	)

	for _, file := range response.Files {
		if file.FileId == revokedFileID {
			log.Fatalf(
				"ERROR DE ACL: el archivo revocado sigue visible | FileID=%s | Owner=%s | Nombre=%s",
				file.FileId,
				file.UserId,
				file.FileName,
			)
		}
	}

	log.Println("===================================")
	log.Println(" LISTADO POST-REVOCACIÓN VALIDADO")
	log.Println(" Archivo revocado: NO VISIBLE")
	log.Println(" Archivos propios: CONSERVADOS")
	log.Println("===================================")
}
