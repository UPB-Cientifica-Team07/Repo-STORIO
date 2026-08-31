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
		os.Getenv("TOKEN")

	if token == "" {
		log.Fatal(
			"TOKEN no definido",
		)
	}

	// =====================================
	// CONEXIÓN
	// =====================================

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
	// METADATA AUTH
	// =====================================

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

	// =====================================
	// INTENTO DE SUPLANTACIÓN
	// =====================================

	log.Println(
		"===================================",
	)

	log.Println(
		" PRUEBA DE SUPLANTACIÓN",
	)

	log.Println(
		"===================================",
	)

	log.Println(
		"Token real: user-001",
	)

	log.Println(
		"UserId enviado falsamente: user-999",
	)

	response,
		err :=
		client.UploadFile(
			ctx,
			&pb.UploadFileRequest{
				UserId: "user-999",

				FileName: "spoof-test.txt",

				FileType: "text/plain",

				Content: []byte(
					"Prueba de suplantacion de identidad.",
				),
			},
		)

	if err != nil {
		log.Fatalf(
			"Upload RPC error: %v",
			err,
		)
	}

	if !response.Success {
		log.Fatalf(
			"Upload rechazado: %s",
			response.Message,
		)
	}

	log.Printf(
		"Upload Success: %t",
		response.Success,
	)

	log.Printf(
		"File ID: %s",
		response.FileId,
	)

	// =====================================
	// RECUPERAR ARCHIVO
	// =====================================

	getResponse,
		err :=
		client.GetFile(
			ctx,
			&pb.GetFileRequest{
				FileId: response.FileId,
			},
		)

	if err != nil {
		log.Fatalf(
			"GetFile RPC error: %v",
			err,
		)
	}

	if !getResponse.Success ||
		getResponse.File == nil {

		log.Fatalf(
			"No se pudo recuperar archivo: %s",
			getResponse.Message,
		)
	}

	log.Println(
		"===================================",
	)

	log.Println(
		" RESULTADO",
	)

	log.Println(
		"===================================",
	)

	log.Printf(
		"UserId enviado: user-999",
	)

	log.Printf(
		"Owner REAL almacenado: %s",
		getResponse.File.UserId,
	)

	if getResponse.File.UserId !=
		"user-001" {

		log.Fatalf(
			"FALLO DE SEGURIDAD: se aceptó identidad falsificada",
		)
	}

	log.Println(
		"SUPLANTACIÓN BLOQUEADA CORRECTAMENTE",
	)
}
