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

	deviceID = "sync-tree-integration-device"

	fileName = "main.go"
	fileType = "text/plain"

	relativePath = "Universidad/SistemasDistribuidos/proyecto/main.go"
)

var content = []byte(
	"package main\n\nfunc main() {}\n",
)

func main() {

	log.Println("===================================")
	log.Println(" SYNC TREE INTEGRATION TEST")
	log.Println("===================================")

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
		"Relative path esperado: %s",
		relativePath,
	)

	log.Printf(
		"Tamaño esperado: %d bytes",
		len(content),
	)

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
	// UPLOAD CON ÁRBOL
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 1 - UPLOAD TREE")
	log.Println("===================================")

	uploadCtx, uploadCancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
		)

	defer uploadCancel()

	uploadCtx =
		metadata.NewOutgoingContext(
			uploadCtx,
			metadata.Pairs(
				"authorization",
				"Bearer "+loginResult.Token,
			),
		)

	uploadStream, err :=
		client.Upload(
			uploadCtx,
		)

	if err != nil {
		log.Fatalf(
			"No se pudo crear stream Upload: %v",
			err,
		)
	}

	err =
		uploadStream.Send(
			&pb.UploadRequest{
				Data: &pb.UploadRequest_Metadata{
					Metadata: &pb.UploadMetadata{
						// Spoof deliberado.
						// El token debe seguir ganando.
						UserId: "user-001",

						DeviceId: deviceID,

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
			"Bytes incorrectos: esperados=%d recibidos=%d",
			len(content),
			uploadResponse.BytesReceived,
		)
	}

	fileID :=
		uploadResponse.FileId

	log.Println(
		"Upload con árbol correcto",
	)

	log.Printf(
		"FILE_ID=%s",
		fileID,
	)

	// =====================================
	// TEST 2
	// LIST FILES
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 2 - LIST TREE METADATA")
	log.Println("===================================")

	listCtx, listCancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
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
				// Spoof deliberado.
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
			"ListFiles respondió Success=false: %s",
			listResponse.Message,
		)
	}

	var listedFile *pb.FileMetadata

	for _, file := range listResponse.Files {

		if file == nil {
			continue
		}

		if file.FileId == fileID {

			listedFile =
				file

			break
		}
	}

	if listedFile == nil {
		log.Fatal(
			"El archivo recién cargado no apareció en ListFiles",
		)
	}

	if listedFile.UserId !=
		loginResult.UserID {

		log.Fatalf(
			"Owner incorrecto: esperado=%s recibido=%s",
			loginResult.UserID,
			listedFile.UserId,
		)
	}

	if listedFile.RelativePath !=
		relativePath {

		log.Fatalf(
			"RelativePath incorrecto en ListFiles: esperado=%q recibido=%q",
			relativePath,
			listedFile.RelativePath,
		)
	}

	if listedFile.FileName !=
		fileName {

		log.Fatalf(
			"FileName incorrecto: esperado=%q recibido=%q",
			fileName,
			listedFile.FileName,
		)
	}

	log.Println(
		"ListFiles preserva RelativePath",
	)

	// =====================================
	// TEST 3
	// DOWNLOAD
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 3 - DOWNLOAD TREE")
	log.Println("===================================")

	downloadCtx, downloadCancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
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
				// Spoof deliberado.
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

	if downloadedMetadata == nil {
		log.Fatal(
			"Download no devolvió metadata",
		)
	}

	if downloadedMetadata.FileId !=
		fileID {

		log.Fatalf(
			"File ID de Download incorrecto: esperado=%s recibido=%s",
			fileID,
			downloadedMetadata.FileId,
		)
	}

	if downloadedMetadata.UserId !=
		loginResult.UserID {

		log.Fatalf(
			"Owner Download incorrecto: esperado=%s recibido=%s",
			loginResult.UserID,
			downloadedMetadata.UserId,
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
			"Contenido descargado no coincide: esperado=%q recibido=%q",
			string(content),
			string(downloadedContent),
		)
	}

	log.Println(
		"Download preserva metadata y contenido",
	)

	// =====================================
	// RESULTADO
	// =====================================

	log.Println("===================================")
	log.Println(" ESTADO INTERMEDIO SYNC-3 VALIDADO")
	log.Println("===================================")
	log.Println(" Upload árbol              = OK")
	log.Println(" Anti-spoof                = OK")
	log.Println(" RelativePath en Sync      = OK")
	log.Println(" ListFiles                 = OK")
	log.Println(" Download                  = OK")
	log.Println(" Contenido                 = OK")
	log.Println(" Archivo queda creado para inspección")
	log.Println("===================================")

	log.Printf(
		"FILE_ID=%s",
		fileID,
	)
}
