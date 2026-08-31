package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"

	grpcClient "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	fileID := strings.TrimSpace(
		os.Getenv("FILE_ID"),
	)

	token := strings.TrimSpace(
		os.Getenv("TOKEN_USER"),
	)

	if fileID == "" {
		log.Fatal(
			"Debe definir la variable FILE_ID",
		)
	}

	if token == "" {
		log.Fatal(
			"Debe definir la variable TOKEN_USER",
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

	baseCtx,
		cancel :=
		context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

	defer cancel()

	// =====================================
	// AUTENTICACIÓN BEARER
	// =====================================

	ctx :=
		metadata.NewOutgoingContext(
			baseCtx,
			metadata.Pairs(
				"authorization",
				"Bearer "+token,
			),
		)

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
