package main

import (
	"context"
	"log"
	"os"
	"time"

	grpcClient "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"
)

func main() {

	// =====================================
	// TOKEN
	// =====================================

	token := os.Getenv("TOKEN")

	if token == "" {
		log.Fatal(
			"La variable de entorno TOKEN es obligatoria",
		)
	}

	// =====================================
	// CONECTAR CON FILE SERVICE
	// =====================================

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

	// =====================================
	// CONTEXTO + AUTH METADATA
	// =====================================

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)

	defer cancel()

	ctx = metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs(
			"authorization",
			"Bearer "+token,
		),
	)

	// =====================================
	// 1. SUBIR ARCHIVO
	// =====================================

	log.Println("===================================")
	log.Println(" SUBIENDO ARCHIVO AL HOME")
	log.Println("===================================")

	content := []byte(
		"Archivo almacenado dentro del Home de user-001.",
	)

	uploadResponse, err := client.UploadFile(
		ctx,
		&pb.UploadFileRequest{
			// Aunque enviemos otro valor aquí,
			// el servidor debe usar la identidad
			// obtenida del token.
			UserId: "user-001",

			FileName: "home-test.txt",

			FileType: "text/plain",

			Content: content,
		},
	)

	if err != nil {
		log.Fatalf(
			"Error subiendo archivo: %v",
			err,
		)
	}

	log.Printf(
		"Success: %t",
		uploadResponse.Success,
	)

	log.Printf(
		"Mensaje: %s",
		uploadResponse.Message,
	)

	log.Printf(
		"File ID: %s",
		uploadResponse.FileId,
	)

	fileID := uploadResponse.FileId

	if fileID == "" {
		log.Fatal(
			"UploadFile no devolvió un File ID",
		)
	}

	// =====================================
	// 2. OBTENER ARCHIVO
	// =====================================

	log.Println("===================================")
	log.Println(" OBTENIENDO ARCHIVO")
	log.Println("===================================")

	getResponse, err := client.GetFile(
		ctx,
		&pb.GetFileRequest{
			FileId: fileID,
		},
	)

	if err != nil {
		log.Fatalf(
			"Error obteniendo archivo: %v",
			err,
		)
	}

	log.Printf(
		"Success: %t",
		getResponse.Success,
	)

	log.Printf(
		"Mensaje: %s",
		getResponse.Message,
	)

	if getResponse.File != nil {

		log.Printf(
			"Archivo: %s",
			getResponse.File.FileName,
		)

		log.Printf(
			"Tipo: %s",
			getResponse.File.FileType,
		)

		log.Printf(
			"Contenido: %s",
			string(
				getResponse.File.Content,
			),
		)

		log.Printf(
			"Tamaño: %d bytes",
			getResponse.File.Size,
		)
	}

	// =====================================
	// 3. LISTAR ARCHIVOS
	// =====================================

	log.Println("===================================")
	log.Println(" LISTANDO ARCHIVOS DE USER-001")
	log.Println("===================================")

	listResponse, err := client.ListFiles(
		ctx,
		&pb.ListFilesRequest{
			UserId: "user-001",
		},
	)

	if err != nil {
		log.Fatalf(
			"Error listando archivos: %v",
			err,
		)
	}

	log.Printf(
		"Success: %t",
		listResponse.Success,
	)

	log.Printf(
		"Mensaje: %s",
		listResponse.Message,
	)

	log.Printf(
		"Cantidad de archivos: %d",
		len(
			listResponse.Files,
		),
	)

	for _, file := range listResponse.Files {

		log.Printf(
			"- %s | ID: %s | Tamaño: %d bytes",
			file.FileName,
			file.FileId,
			file.Size,
		)
	}

	// =====================================
	// NO ELIMINAR TODAVÍA
	// =====================================
	//
	// Queremos inspeccionar físicamente:
	//
	// homes/user-001/documentos/
	//
	// y comprobar Monitoring.
	//
	// =====================================

	log.Println("===================================")
	log.Println(" PRUEBA DE HOME COMPLETADA")
	log.Println(" Archivo conservado para inspección")
	log.Println("===================================")
}
