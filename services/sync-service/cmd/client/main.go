package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	serverAddress = "localhost:50055"
	authURL       = "http://localhost:8081"
)

func main() {

	// =====================================
	// CLIENTE AUTH
	// =====================================

	authClient :=
		auth.NewClient(
			authURL,
		)

	// =====================================
	// ENCABEZADO
	// =====================================

	fmt.Println("===================================")
	fmt.Println(" CLIENTE SYNC SERVICE")
	fmt.Println("===================================")

	// =====================================
	// 0. LOGIN
	// =====================================

	fmt.Println()
	fmt.Println("0. LOGIN")

	loginResult, err :=
		authClient.Login(
			"samuel",
			"123456",
		)

	if err != nil {
		log.Fatalf(
			"Error Login: %v",
			err,
		)
	}

	fmt.Println(
		"Success:",
		loginResult.Success,
	)

	fmt.Println(
		"Mensaje:",
		loginResult.Message,
	)

	fmt.Println(
		"User ID:",
		loginResult.UserID,
	)

	fmt.Println(
		"Rol:",
		loginResult.Role,
	)

	if !loginResult.Success {
		return
	}

	token :=
		loginResult.Token

	// =====================================
	// CONEXIÓN GRPC
	// =====================================

	conn, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		log.Fatalf(
			"Error conectando con Sync Service: %v",
			err,
		)
	}

	defer conn.Close()

	client :=
		pb.NewSyncServiceClient(
			conn,
		)

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			40*time.Second,
		)

	defer cancel()

	// =====================================
	// 1. AUTHENTICATE
	// =====================================

	fmt.Println()
	fmt.Println("1. AUTHENTICATE")

	authResponse, err :=
		client.Authenticate(
			ctx,
			&pb.AuthenticateRequest{
				Token:    token,
				DeviceId: "desktop-001",
			},
		)

	if err != nil {
		log.Fatalf(
			"Error Authenticate: %v",
			err,
		)
	}

	fmt.Println(
		"Success:",
		authResponse.Success,
	)

	fmt.Println(
		"Mensaje:",
		authResponse.Message,
	)

	fmt.Println(
		"User ID:",
		authResponse.UserId,
	)

	if !authResponse.Success {
		return
	}

	userID :=
		authResponse.UserId

	deviceID :=
		"desktop-001"

	// =====================================
	// 2. UPLOAD
	// =====================================

	fmt.Println()
	fmt.Println("2. UPLOAD")

	content :=
		[]byte(
			"Archivo de prueba para UPB-CIENTIFICA y Sync Service.",
		)

	uploadStream, err :=
		client.Upload(
			ctx,
		)

	if err != nil {
		log.Fatalf(
			"Error iniciando Upload: %v",
			err,
		)
	}

	err =
		uploadStream.Send(
			&pb.UploadRequest{
				Data: &pb.UploadRequest_Metadata{
					Metadata: &pb.UploadMetadata{
						UserId:   userID,
						DeviceId: deviceID,
						FileName: "prueba-sync.txt",
						FileType: "DOCUMENTO",
						Size: int64(
							len(content),
						),
					},
				},
			},
		)

	if err != nil {
		log.Fatalf(
			"Error enviando metadata: %v",
			err,
		)
	}

	err =
		uploadStream.Send(
			&pb.UploadRequest{
				Data: &pb.UploadRequest_Chunk{
					Chunk: content,
				},
			},
		)

	if err != nil {
		log.Fatalf(
			"Error enviando contenido: %v",
			err,
		)
	}

	uploadResponse, err :=
		uploadStream.CloseAndRecv()

	if err != nil {
		log.Fatalf(
			"Error finalizando Upload: %v",
			err,
		)
	}

	fmt.Println(
		"Success:",
		uploadResponse.Success,
	)

	fmt.Println(
		"Mensaje:",
		uploadResponse.Message,
	)

	fmt.Println(
		"File ID:",
		uploadResponse.FileId,
	)

	fmt.Println(
		"Bytes:",
		uploadResponse.BytesReceived,
	)

	if !uploadResponse.Success {
		return
	}

	fileID :=
		uploadResponse.FileId

	// =====================================
	// 3. LIST FILES
	// =====================================

	fmt.Println()
	fmt.Println("3. LIST FILES")

	listResponse, err :=
		client.ListFiles(
			ctx,
			&pb.ListFilesRequest{
				UserId: userID,
			},
		)

	if err != nil {
		log.Fatalf(
			"Error ListFiles: %v",
			err,
		)
	}

	fmt.Println(
		"Success:",
		listResponse.Success,
	)

	fmt.Println(
		"Cantidad:",
		len(listResponse.Files),
	)

	for _, file := range listResponse.Files {

		fmt.Printf(
			"- %s | %s | %d bytes\n",
			file.FileId,
			file.FileName,
			file.Size,
		)
	}

	// =====================================
	// 4. DOWNLOAD
	// =====================================

	fmt.Println()
	fmt.Println("4. DOWNLOAD")

	downloadStream, err :=
		client.Download(
			ctx,
			&pb.DownloadRequest{
				UserId: userID,
				FileId: fileID,
			},
		)

	if err != nil {
		log.Fatalf(
			"Error iniciando Download: %v",
			err,
		)
	}

	var downloadedContent []byte

	for {

		response, err :=
			downloadStream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf(
				"Error recibiendo Download: %v",
				err,
			)
		}

		if metadata :=
			response.GetMetadata(); metadata != nil {

			fmt.Println(
				"Archivo:",
				metadata.FileName,
			)

			fmt.Println(
				"Tamaño:",
				metadata.Size,
			)

			continue
		}

		chunk :=
			response.GetChunk()

		if chunk != nil {

			downloadedContent =
				append(
					downloadedContent,
					chunk...,
				)
		}
	}

	fmt.Println(
		"Contenido descargado:",
		string(downloadedContent),
	)

	// =====================================
	// 5. SYNC DESDE OTRO DISPOSITIVO
	// =====================================

	fmt.Println()
	fmt.Println(
		"5. SYNC DESDE mobile-001",
	)

	syncResponse, err :=
		client.Sync(
			ctx,
			&pb.SyncRequest{
				UserId:   userID,
				DeviceId: "mobile-001",
			},
		)

	if err != nil {
		log.Fatalf(
			"Error Sync: %v",
			err,
		)
	}

	fmt.Println(
		"Success:",
		syncResponse.Success,
	)

	fmt.Println(
		"Cambios:",
		len(syncResponse.Changes),
	)

	for _, change := range syncResponse.Changes {

		fmt.Printf(
			"- %s | %s | origen=%s\n",
			change.Type,
			change.FileName,
			change.OriginDeviceId,
		)
	}

	// =====================================
	// PAUSA PARA MONITOREO DE STORAGE
	// =====================================

	fmt.Println()
	fmt.Println(
		"Esperando 10 segundos antes de eliminar...",
	)

	fmt.Println(
		"Durante este tiempo Monitoring debería detectar 53 bytes.",
	)

	time.Sleep(
		10 * time.Second,
	)

	// =====================================
	// 6. DELETE
	// =====================================

	fmt.Println()
	fmt.Println("6. DELETE")

	deleteResponse, err :=
		client.DeleteFile(
			ctx,
			&pb.DeleteFileRequest{
				UserId:   userID,
				DeviceId: deviceID,
				FileId:   fileID,
			},
		)

	if err != nil {
		log.Fatalf(
			"Error DeleteFile: %v",
			err,
		)
	}

	fmt.Println(
		"Success:",
		deleteResponse.Success,
	)

	fmt.Println(
		"Mensaje:",
		deleteResponse.Message,
	)

	// =====================================
	// PAUSA DESPUÉS DEL DELETE
	// =====================================

	fmt.Println()
	fmt.Println(
		"Esperando 6 segundos después del DELETE...",
	)

	fmt.Println(
		"Monitoring debería volver a detectar 0 bytes.",
	)

	time.Sleep(
		6 * time.Second,
	)

	// =====================================
	// 7. LIST FILES FINAL
	// =====================================

	fmt.Println()
	fmt.Println(
		"7. LIST FILES FINAL",
	)

	finalList, err :=
		client.ListFiles(
			ctx,
			&pb.ListFilesRequest{
				UserId: userID,
			},
		)

	if err != nil {
		log.Fatalf(
			"Error ListFiles final: %v",
			err,
		)
	}

	fmt.Println(
		"Cantidad final:",
		len(finalList.Files),
	)

	fmt.Println()
	fmt.Println(
		"===================================",
	)

	fmt.Println(
		" PRUEBA SYNC COMPLETADA",
	)

	fmt.Println(
		"===================================",
	)
}
