package main

import (
	"context"
	"log"
	"time"

	grpcClient "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"
)

func main() {

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

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()

	// =====================================
	// 1. SUBIR ARCHIVO
	// =====================================

	log.Println("===================================")
	log.Println(" SUBIENDO ARCHIVO")
	log.Println("===================================")

	uploadResponse, err := client.UploadFile(
		ctx,
		&pb.UploadFileRequest{
			UserId:   "user-001",
			FileName: "documento.txt",
			FileType: "text/plain",
			Content: []byte(
				"Este es el contenido del archivo de prueba.",
			),
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
			string(getResponse.File.Content),
		)

		log.Printf(
			"Tamaño: %d bytes",
			getResponse.File.Size,
		)
	}

	// =====================================
	// 3. LISTAR ARCHIVOS DEL USUARIO
	// =====================================

	log.Println("===================================")
	log.Println(" LISTANDO ARCHIVOS")
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
		len(listResponse.Files),
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
	// 4. ELIMINAR ARCHIVO
	// =====================================
	//
	//log.Println("===================================")
	//log.Println(" ELIMINANDO ARCHIVO")
	//log.Println("===================================")

	//deleteResponse, err := client.DeleteFile(
	//	ctx,
	//	&pb.DeleteFileRequest{
	//		FileId: fileID,
	//	},
	//)

	//if err != nil {
	//	log.Fatalf(
	//		"Error eliminando archivo: %v",
	//		err,
	//	)
	//}

	//log.Printf(
	//	"Success: %t",
	//	deleteResponse.Success,
	//)

	//log.Printf(
	//	"Mensaje: %s",
	//	deleteResponse.Message,
	//)

	// =====================================
	// 5. VERIFICAR QUE FUE ELIMINADO
	// =====================================

	log.Println("===================================")
	log.Println(" VERIFICANDO ELIMINACIÓN")
	log.Println("===================================")

	verifyResponse, err := client.GetFile(
		ctx,
		&pb.GetFileRequest{
			FileId: fileID,
		},
	)

	if err != nil {
		log.Fatalf(
			"Error verificando archivo: %v",
			err,
		)
	}

	log.Printf(
		"Success: %t",
		verifyResponse.Success,
	)

	log.Printf(
		"Mensaje: %s",
		verifyResponse.Message,
	)

	log.Println("===================================")
	log.Println(" PRUEBA COMPLETADA")
	log.Println("===================================")
}
