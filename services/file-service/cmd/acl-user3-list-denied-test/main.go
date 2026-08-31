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
	log.Println(" USER-003 LIST DENIED TEST")
	log.Println("===================================")
	log.Println(" Usuario: user-003")
	log.Println(" Esperado:")
	log.Println(" - sin archivos propios")
	log.Println(" - sin archivos compartidos")
	log.Println(" - total = 0")
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
		"Archivos recibidos: %d",
		len(response.Files),
	)

	if len(response.Files) != 0 {
		log.Fatalf(
			"ERROR: se esperaban 0 archivos, pero se recibieron %d",
			len(response.Files),
		)
	}

	log.Println("===================================")
	log.Println(" LISTADO VACÍO VALIDADO")
	log.Println(" ACL REVOCADA CORRECTAMENTE")
	log.Println("===================================")
}
