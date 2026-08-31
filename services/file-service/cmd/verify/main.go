package main

import (
	"context"
	"log"
	"os"
	"time"

	grpcClient "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"
)

func main() {

	fileID := os.Getenv("FILE_ID")

	if fileID == "" {
		log.Fatal(
			"Debe definir la variable FILE_ID",
		)
	}

	// =====================================
	// CONEXIÓN
	// =====================================

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

	ctx,
		cancel :=
		context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

	defer cancel()

	// =====================================
	// CONSULTAR ARCHIVO EXISTENTE
	// =====================================

	log.Println("===================================")
	log.Println(" VERIFICANDO ARCHIVO PERSISTENTE")
	log.Println("===================================")

	response,
		err :=
		client.GetFile(
			ctx,
			&pb.GetFileRequest{
				FileId: fileID,
			},
		)

	if err != nil {

		log.Fatalf(
			"Error consultando archivo: %v",
			err,
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

	if !response.Success ||
		response.File == nil {

		log.Fatal(
			"El archivo no pudo recuperarse",
		)
	}

	log.Printf(
		"File ID: %s",
		response.File.FileId,
	)

	log.Printf(
		"Usuario: %s",
		response.File.UserId,
	)

	log.Printf(
		"Nombre: %s",
		response.File.FileName,
	)

	log.Printf(
		"Tipo: %s",
		response.File.FileType,
	)

	log.Printf(
		"Tamaño: %d bytes",
		response.File.Size,
	)

	log.Printf(
		"Contenido: %s",
		string(
			response.File.Content,
		),
	)

	log.Println("===================================")
	log.Println(" PERSISTENCIA VERIFICADA")
	log.Println("===================================")
}
