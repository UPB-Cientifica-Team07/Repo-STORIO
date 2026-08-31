package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"
	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/proto"

	grpcClient "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	serverAddress = "localhost:50055"
	authURL       = "http://localhost:8081"

	deviceID = "sync-home-integration-device"
)

func main() {

	log.Println("===================================")
	log.Println(" SYNC HOME INTEGRATION TEST")
	log.Println("===================================")

	// =====================================
	// LOGIN
	// =====================================

	authClient :=
		auth.NewClient(
			authURL,
		)

	loginResult, err :=
		authClient.Login(
			"tercero",
			"123456",
		)

	if err != nil {
		log.Fatalf(
			"Login falló: %v",
			err,
		)
	}

	if !loginResult.Success {
		log.Fatalf(
			"Login rechazado: %s",
			loginResult.Message,
		)
	}

	log.Printf(
		"Usuario autenticado: %s",
		loginResult.UserID,
	)

	// =====================================
	// GRPC
	// =====================================

	connection, err :=
		grpcClient.NewClient(
			serverAddress,
			grpcClient.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		)

	if err != nil {
		log.Fatalf(
			"Error conectando con Sync Service: %v",
			err,
		)
	}

	defer connection.Close()

	client :=
		pb.NewSyncServiceClient(
			connection,
		)

	newAuthenticatedContext :=
		func() (
			context.Context,
			context.CancelFunc,
		) {

			ctx, cancel :=
				context.WithTimeout(
					context.Background(),
					30*time.Second,
				)

			ctx =
				metadata.NewOutgoingContext(
					ctx,
					metadata.Pairs(
						"authorization",
						"Bearer "+loginResult.Token,
					),
				)

			return ctx, cancel
		}

	// =====================================
	// CONTENIDO
	// =====================================

	content :=
		[]byte(
			"SYNC HOME CENTRAL INTEGRATION TEST",
		)

	fileName :=
		"sync-home-central-test.txt"

	fileType :=
		"text/plain"

	log.Printf(
		"Tamaño esperado: %d bytes",
		len(content),
	)

	// =====================================
	// UPLOAD VIA SYNC
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 1 - UPLOAD VIA SYNC")
	log.Println("===================================")

	uploadCtx, uploadCancel :=
		newAuthenticatedContext()

	uploadStream, err :=
		client.Upload(
			uploadCtx,
		)

	if err != nil {
		uploadCancel()

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
						UserId:   "user-001",
						DeviceId: deviceID,
						FileName: fileName,
						FileType: fileType,
						Size: int64(
							len(content),
						),
					},
				},
			},
		)

	if err != nil {
		uploadCancel()

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
		uploadCancel()

		log.Fatalf(
			"Error enviando contenido: %v",
			err,
		)
	}

	uploadResponse, err :=
		uploadStream.CloseAndRecv()

	uploadCancel()

	if err != nil {
		log.Fatalf(
			"Upload falló: %v",
			err,
		)
	}

	if !uploadResponse.Success {
		log.Fatalf(
			"Upload respondió Success=false: %s",
			uploadResponse.Message,
		)
	}

	if uploadResponse.FileId == "" {
		log.Fatal(
			"Upload no devolvió file_id",
		)
	}

	fileID :=
		uploadResponse.FileId

	log.Println(
		"Upload correcto",
	)

	log.Printf(
		"File ID oficial: %s",
		fileID,
	)

	log.Printf(
		"Bytes recibidos: %d",
		uploadResponse.BytesReceived,
	)

	// =====================================
	// LISTFILES VIA SYNC
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 2 - LIST VIA SYNC")
	log.Println("===================================")

	listCtx, listCancel :=
		newAuthenticatedContext()

	listResponse, err :=
		client.ListFiles(
			listCtx,
			&pb.ListFilesRequest{
				UserId: "user-001",
			},
		)

	listCancel()

	if err != nil {
		log.Fatalf(
			"ListFiles falló: %v",
			err,
		)
	}

	found :=
		false

	for _, file := range listResponse.Files {

		if file == nil ||
			file.FileId != fileID {

			continue
		}

		found =
			true

		log.Printf(
			"Archivo listado: %s",
			file.FileName,
		)

		log.Printf(
			"Owner Sync: %s",
			file.UserId,
		)

		if file.UserId !=
			loginResult.UserID {

			log.Fatalf(
				"Owner incorrecto: esperado=%s obtenido=%s",
				loginResult.UserID,
				file.UserId,
			)
		}
	}

	if !found {
		log.Fatal(
			"El archivo no apareció en ListFiles",
		)
	}

	log.Println(
		"ListFiles validado",
	)

	// =====================================
	// DOWNLOAD VIA SYNC
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 3 - DOWNLOAD VIA SYNC")
	log.Println("===================================")

	downloadCtx, downloadCancel :=
		newAuthenticatedContext()

	downloadStream, err :=
		client.Download(
			downloadCtx,
			&pb.DownloadRequest{
				UserId: "user-001",
				FileId: fileID,
			},
		)

	if err != nil {
		downloadCancel()

		log.Fatalf(
			"Error iniciando Download: %v",
			err,
		)
	}

	var downloadedContent []byte
	var downloadedMetadata *pb.FileMetadata

	for {

		response, recvErr :=
			downloadStream.Recv()

		if recvErr == io.EOF {
			break
		}

		if recvErr != nil {
			downloadCancel()

			log.Fatalf(
				"Error recibiendo Download: %v",
				recvErr,
			)
		}

		if response.GetMetadata() != nil {

			downloadedMetadata =
				response.GetMetadata()

			continue
		}

		if response.GetChunk() != nil {

			downloadedContent =
				append(
					downloadedContent,
					response.GetChunk()...,
				)
		}
	}

	downloadCancel()

	if downloadedMetadata == nil {
		log.Fatal(
			"Download no devolvió metadata",
		)
	}

	if downloadedMetadata.FileId !=
		fileID {

		log.Fatalf(
			"File ID incorrecto en Download: %s",
			downloadedMetadata.FileId,
		)
	}

	if downloadedMetadata.UserId !=
		loginResult.UserID {

		log.Fatalf(
			"Owner incorrecto en Download: esperado=%s obtenido=%s",
			loginResult.UserID,
			downloadedMetadata.UserId,
		)
	}

	if !bytes.Equal(
		downloadedContent,
		content,
	) {

		log.Fatalf(
			"Contenido descargado no coincide: esperado=%q obtenido=%q",
			string(content),
			string(downloadedContent),
		)
	}

	log.Printf(
		"Contenido descargado: %q",
		string(downloadedContent),
	)

	log.Println(
		"Download validado",
	)

	// =====================================
	// NO ELIMINAMOS TODAVÍA
	// =====================================
	//
	// Se deja el archivo creado para poder
	// comprobar filesystem, PostgreSQL,
	// cuota y permisos desde terminal.
	//
	// =====================================

	log.Println("===================================")
	log.Println(" ESTADO INTERMEDIO VALIDADO")
	log.Println("===================================")
	log.Println(" Upload Sync -> File Service = OK")
	log.Println(" File ID único              = OK")
	log.Println(" Anti-spoof                 = OK")
	log.Println(" ListFiles                  = OK")
	log.Println(" Download                   = OK")
	log.Println(" Archivo se deja creado para inspección")
	log.Println("===================================")

	log.Printf(
		"FILE_ID=%s",
		fileID,
	)
}
