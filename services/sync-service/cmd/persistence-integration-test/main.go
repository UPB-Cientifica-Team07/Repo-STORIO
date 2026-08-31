package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"
	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	syncAddress = "localhost:50055"
	authURL     = "http://localhost:8081"

	username = "tercero"
	password = "123456"

	originDevice = "sync4-desktop-001"

	fileName = "persistencia-sync4.txt"
	fileType = "text/plain"

	relativePath = "SYNC4/Persistencia/persistencia-sync4.txt"
)

var content = []byte(
	"SYNC-4 POSTGRESQL PERSISTENCE TEST\n",
)

func main() {

	log.Println("===================================")
	log.Println(" SYNC-4 PERSISTENCE CREATE TEST")
	log.Println("===================================")

	// =====================================
	// AUTH
	// =====================================

	authClient :=
		auth.NewClient(
			authURL,
		)

	loginResult, err :=
		authClient.Login(
			username,
			password,
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

	log.Printf(
		"Dispositivo origen: %s",
		originDevice,
	)

	log.Printf(
		"RelativePath: %s",
		relativePath,
	)

	log.Printf(
		"Tamaño: %d bytes",
		len(content),
	)

	// =====================================
	// GRPC CONNECTION
	// =====================================

	connection, err :=
		grpc.NewClient(
			syncAddress,
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

	defer connection.Close()

	client :=
		pb.NewSyncServiceClient(
			connection,
		)

	// =====================================
	// TEST 1
	// UPLOAD
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 1 - UPLOAD")
	log.Println("===================================")

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
		)

	defer cancel()

	ctx =
		metadata.NewOutgoingContext(
			ctx,
			metadata.Pairs(
				"authorization",
				"Bearer "+loginResult.Token,
			),
		)

	stream, err :=
		client.Upload(
			ctx,
		)

	if err != nil {

		log.Fatalf(
			"No se pudo abrir Upload: %v",
			err,
		)
	}

	err =
		stream.Send(
			&pb.UploadRequest{
				Data: &pb.UploadRequest_Metadata{
					Metadata: &pb.UploadMetadata{
						// Spoof deliberado.
						// Debe prevalecer el token user-003.
						UserId: "user-001",

						DeviceId: originDevice,

						FileName: fileName,
						FileType: fileType,

						Size: int64(
							len(content),
						),

						RelativePath: relativePath,
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
		stream.Send(
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
		stream.CloseAndRecv()

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
			"Upload no devolvió File ID",
		)
	}

	if uploadResponse.BytesReceived !=
		int64(
			len(content),
		) {

		log.Fatalf(
			"Tamaño incorrecto: esperado=%d recibido=%d",
			len(content),
			uploadResponse.BytesReceived,
		)
	}

	fileID :=
		uploadResponse.FileId

	log.Println(
		"Upload = OK",
	)

	log.Printf(
		"FILE_ID=%s",
		fileID,
	)

	// =====================================
	// TEST 2
	// LIST FILES ANTES DEL REINICIO
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 2 - LISTFILES ANTES DEL REINICIO")
	log.Println("===================================")

	listCtx, listCancel :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)

	defer listCancel()

	listCtx =
		metadata.NewOutgoingContext(
			listCtx,
			metadata.Pairs(
				"authorization",
				"Bearer "+loginResult.Token,
			),
		)

	listResponse, err :=
		client.ListFiles(
			listCtx,
			&pb.ListFilesRequest{
				UserId: "user-001",
			},
		)

	if err != nil {

		log.Fatalf(
			"ListFiles falló: %v",
			err,
		)
	}

	if !listResponse.Success {

		log.Fatalf(
			"ListFiles Success=false: %s",
			listResponse.Message,
		)
	}

	var found *pb.FileMetadata

	for _, file := range listResponse.Files {

		if file == nil {
			continue
		}

		if file.FileId == fileID {

			found =
				file

			break
		}
	}

	if found == nil {

		log.Fatal(
			"El archivo no apareció en ListFiles",
		)
	}

	if found.UserId !=
		loginResult.UserID {

		log.Fatalf(
			"Owner incorrecto: esperado=%s recibido=%s",
			loginResult.UserID,
			found.UserId,
		)
	}

	if found.RelativePath !=
		relativePath {

		log.Fatalf(
			"RelativePath incorrecto: esperado=%q recibido=%q",
			relativePath,
			found.RelativePath,
		)
	}

	log.Println(
		"ListFiles antes de reinicio = OK",
	)

	// =====================================
	// TEST 3
	// DOWNLOAD ANTES DEL REINICIO
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 3 - DOWNLOAD ANTES DEL REINICIO")
	log.Println("===================================")

	downloadCtx, downloadCancel :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)

	defer downloadCancel()

	downloadCtx =
		metadata.NewOutgoingContext(
			downloadCtx,
			metadata.Pairs(
				"authorization",
				"Bearer "+loginResult.Token,
			),
		)

	downloadStream, err :=
		client.Download(
			downloadCtx,
			&pb.DownloadRequest{
				UserId: "user-001",
				FileId: fileID,
			},
		)

	if err != nil {

		log.Fatalf(
			"Download falló al abrir stream: %v",
			err,
		)
	}

	var downloadedMetadata *pb.FileMetadata
	var downloadedContent []byte

	for {

		response, recvErr :=
			downloadStream.Recv()

		if recvErr == io.EOF {
			break
		}

		if recvErr != nil {

			log.Fatalf(
				"Download falló: %v",
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

	if downloadedMetadata == nil {

		log.Fatal(
			"Download no devolvió metadata",
		)
	}

	if downloadedMetadata.RelativePath !=
		relativePath {

		log.Fatalf(
			"RelativePath Download incorrecto: esperado=%q recibido=%q",
			relativePath,
			downloadedMetadata.RelativePath,
		)
	}

	if !bytes.Equal(
		downloadedContent,
		content,
	) {

		log.Fatalf(
			"Contenido incorrecto: esperado=%q recibido=%q",
			string(content),
			string(downloadedContent),
		)
	}

	log.Println(
		"Download antes de reinicio = OK",
	)

	// =====================================
	// RESULTADO
	// =====================================

	log.Println("===================================")
	log.Println(" SYNC-4 FASE 1 VALIDADA")
	log.Println("===================================")
	log.Println(" Upload                     = OK")
	log.Println(" Anti-spoof                 = OK")
	log.Println(" RelativePath               = OK")
	log.Println(" ListFiles                  = OK")
	log.Println(" Download                   = OK")
	log.Println(" Contenido                  = OK")
	log.Println(" Archivo queda persistido")
	log.Println(" NO ELIMINAR EL ARCHIVO TODAVÍA")
	log.Println("===================================")

	log.Printf(
		"FILE_ID=%s",
		fileID,
	)
}
