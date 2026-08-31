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

	const expectedFileID = "231a6faa-04e6-4751-a079-b28f6d83175b"

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
	log.Println(" USER-003 LIST FILES TEST")
	log.Println("===================================")
	log.Println(" Usuario autenticado: user-003")
	log.Println(" Esperado:")
	log.Println(" - archivo compartido visible")
	log.Println(" - owner permanece user-001")
	log.Println(" - archivo: home-test.txt")
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

	found := false

	for index, file := range response.Files {

		if file == nil {
			continue
		}

		log.Println("-----------------------------------")
		log.Printf(
			"Archivo #%d",
			index+1,
		)
		log.Printf(
			"File ID: %s",
			file.FileId,
		)
		log.Printf(
			"Owner: %s",
			file.UserId,
		)
		log.Printf(
			"Nombre: %s",
			file.FileName,
		)
		log.Printf(
			"Tipo: %s",
			file.FileType,
		)
		log.Printf(
			"Tamaño: %d bytes",
			file.Size,
		)
		log.Printf(
			"Contenido: %s",
			string(file.Content),
		)

		if file.FileId ==
			expectedFileID {

			found = true

			if file.UserId !=
				"user-001" {

				log.Fatalf(
					"Owner incorrecto: esperado=user-001 obtenido=%s",
					file.UserId,
				)
			}

			if file.FileName !=
				"home-test.txt" {

				log.Fatalf(
					"Nombre incorrecto: esperado=home-test.txt obtenido=%s",
					file.FileName,
				)
			}

			if file.Size != 30 {

				log.Fatalf(
					"Tamaño incorrecto: esperado=30 obtenido=%d",
					file.Size,
				)
			}
		}
	}

	if !found {
		log.Fatalf(
			"ERROR: el archivo compartido %s no apareció en ListFiles",
			expectedFileID,
		)
	}

	log.Println("===================================")
	log.Println(" ARCHIVO COMPARTIDO ENCONTRADO")
	log.Println(" USER-003 LIST FILES VALIDADO")
	log.Println("===================================")
}
