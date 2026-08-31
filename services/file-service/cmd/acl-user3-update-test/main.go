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

	const fileID = "231a6faa-04e6-4751-a079-b28f6d83175b"

	// Contenido reducido para validar:
	//
	// - UpdateFile con can_write=true
	// - disminución del tamaño físico
	// - liberación de cuota
	// - actualización de metadata.json
	// - actualización de PostgreSQL
	newContent := []byte(
		"Contenido reducido a 30 bytes.",
	)

	expectedSize := int64(30)

	if int64(len(newContent)) != expectedSize {
		log.Fatalf(
			"El contenido de prueba debía medir %d bytes, pero mide %d",
			expectedSize,
			len(newContent),
		)
	}

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
	log.Println(" USER-003 UPDATE REDUCTION TEST")
	log.Println("===================================")
	log.Println(" Usuario: user-003")
	log.Printf(" File ID: %s", fileID)
	log.Printf(" Nuevo tamaño esperado: %d bytes", expectedSize)
	log.Printf(" Nuevo tamaño real: %d bytes", len(newContent))
	log.Printf(" Nuevo contenido: %s", string(newContent))
	log.Println("===================================")

	response, err := client.UpdateFile(
		ctx,
		&pb.UpdateFileRequest{
			FileId:  fileID,
			Content: newContent,
		},
	)

	if err != nil {
		log.Fatalf(
			"UpdateFile falló: %v",
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

	if !response.Success {
		log.Fatal(
			"UpdateFile debía ser exitoso",
		)
	}

	if response.File == nil {
		log.Fatal(
			"respuesta sin FileData",
		)
	}

	// =====================================
	// VALIDAR IDENTIDAD DEL ARCHIVO
	// =====================================

	if response.File.FileId != fileID {
		log.Fatalf(
			"File ID inesperado: esperado=%s obtenido=%s",
			fileID,
			response.File.FileId,
		)
	}

	if response.File.UserId != "user-001" {
		log.Fatalf(
			"Owner inesperado: esperado=user-001 obtenido=%s",
			response.File.UserId,
		)
	}

	// =====================================
	// VALIDAR TAMAÑO
	// =====================================

	if response.File.Size != expectedSize {
		log.Fatalf(
			"Tamaño incorrecto: esperado=%d obtenido=%d",
			expectedSize,
			response.File.Size,
		)
	}

	// =====================================
	// VALIDAR CONTENIDO
	// =====================================

	if string(response.File.Content) != string(newContent) {
		log.Fatalf(
			"Contenido devuelto por el servidor no coincide con el enviado",
		)
	}

	log.Println("===================================")
	log.Println(" RESPUESTA DEL SERVIDOR")
	log.Println("===================================")

	log.Printf(
		"File ID: %s",
		response.File.FileId,
	)

	log.Printf(
		"Owner: %s",
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
		string(response.File.Content),
	)

	log.Println("===================================")
	log.Println(" USER-003 UPDATE REDUCTION VALIDADO")
	log.Println("===================================")
}
