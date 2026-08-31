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

	fileID = "d8042164-d900-4c9b-8b66-4f7ab422a7ce"

	fileName = "persistencia-sync4.txt"

	relativePath = "SYNC4/Persistencia/persistencia-sync4.txt"

	secondDevice = "sync4-mobile-001"
)

var expectedContent = []byte(
	"SYNC-4 POSTGRESQL PERSISTENCE TEST\n",
)

func main() {

	log.Println("===================================")
	log.Println(" SYNC-4 RESTART PERSISTENCE TEST")
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

	if loginResult.UserID != "user-003" {
		log.Fatalf(
			"Usuario inesperado: %s",
			loginResult.UserID,
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
	// LISTFILES DESPUÉS DEL REINICIO
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 1 - LISTFILES POST-RESTART")
	log.Println("===================================")

	ctx1, cancel1 :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)

	defer cancel1()

	ctx1 =
		metadata.NewOutgoingContext(
			ctx1,
			metadata.Pairs(
				"authorization",
				"Bearer "+loginResult.Token,
			),
		)

	listResponse, err :=
		client.ListFiles(
			ctx1,
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

			found = file

			break
		}
	}

	if found == nil {
		log.Fatalf(
			"El archivo %s desapareció después del reinicio",
			fileID,
		)
	}

	if found.UserId != "user-003" {
		log.Fatalf(
			"Owner incorrecto: %s",
			found.UserId,
		)
	}

	if found.FileName != fileName {
		log.Fatalf(
			"Nombre incorrecto: esperado=%q recibido=%q",
			fileName,
			found.FileName,
		)
	}

	if found.RelativePath != relativePath {
		log.Fatalf(
			"RelativePath incorrecto: esperado=%q recibido=%q",
			relativePath,
			found.RelativePath,
		)
	}

	log.Println(
		"ListFiles sobrevivió al reinicio = OK",
	)

	// =====================================
	// TEST 2
	// DOWNLOAD DESPUÉS DEL REINICIO
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 2 - DOWNLOAD POST-RESTART")
	log.Println("===================================")

	ctx2, cancel2 :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)

	defer cancel2()

	ctx2 =
		metadata.NewOutgoingContext(
			ctx2,
			metadata.Pairs(
				"authorization",
				"Bearer "+loginResult.Token,
			),
		)

	downloadStream, err :=
		client.Download(
			ctx2,
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

	if downloadedMetadata.RelativePath != relativePath {
		log.Fatalf(
			"RelativePath incorrecto en Download: esperado=%q recibido=%q",
			relativePath,
			downloadedMetadata.RelativePath,
		)
	}

	if !bytes.Equal(
		downloadedContent,
		expectedContent,
	) {

		log.Fatalf(
			"Contenido incorrecto: esperado=%q recibido=%q",
			string(expectedContent),
			string(downloadedContent),
		)
	}

	log.Println(
		"Download sobrevivió al reinicio = OK",
	)

	// =====================================
	// TEST 3
	// SEGUNDO DISPOSITIVO RECIBE CREATED
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 3 - SYNC MOBILE PRIMERA VEZ")
	log.Println("===================================")

	ctx3, cancel3 :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)

	defer cancel3()

	ctx3 =
		metadata.NewOutgoingContext(
			ctx3,
			metadata.Pairs(
				"authorization",
				"Bearer "+loginResult.Token,
			),
		)

	syncResponse, err :=
		client.Sync(
			ctx3,
			&pb.SyncRequest{
				UserId:   "user-001",
				DeviceId: secondDevice,
			},
		)

	if err != nil {
		log.Fatalf(
			"Sync inicial mobile falló: %v",
			err,
		)
	}

	if !syncResponse.Success {
		log.Fatalf(
			"Sync inicial Success=false: %s",
			syncResponse.Message,
		)
	}

	if len(syncResponse.Changes) != 1 {
		log.Fatalf(
			"Se esperaba exactamente 1 cambio; recibidos=%d",
			len(syncResponse.Changes),
		)
	}

	change :=
		syncResponse.Changes[0]

	if change.FileId != fileID {
		log.Fatalf(
			"FileID del cambio incorrecto: esperado=%s recibido=%s",
			fileID,
			change.FileId,
		)
	}

	if change.Type != "FILE_CREATED" {
		log.Fatalf(
			"Tipo incorrecto: esperado=FILE_CREATED recibido=%s",
			change.Type,
		)
	}

	if change.OriginDeviceId != "sync4-desktop-001" {
		log.Fatalf(
			"OriginDevice incorrecto: %s",
			change.OriginDeviceId,
		)
	}

	log.Println(
		"Segundo dispositivo recibió FILE_CREATED = OK",
	)

	// =====================================
	// TEST 4
	// MISMO DISPOSITIVO NO DEBE RECIBIRLO
	// DE NUEVO
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 4 - CURSOR EVITA DUPLICADOS")
	log.Println("===================================")

	ctx4, cancel4 :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)

	defer cancel4()

	ctx4 =
		metadata.NewOutgoingContext(
			ctx4,
			metadata.Pairs(
				"authorization",
				"Bearer "+loginResult.Token,
			),
		)

	secondSync, err :=
		client.Sync(
			ctx4,
			&pb.SyncRequest{
				UserId:   "user-001",
				DeviceId: secondDevice,
			},
		)

	if err != nil {
		log.Fatalf(
			"Segundo Sync falló: %v",
			err,
		)
	}

	if !secondSync.Success {
		log.Fatalf(
			"Segundo Sync Success=false: %s",
			secondSync.Message,
		)
	}

	if len(secondSync.Changes) != 0 {
		log.Fatalf(
			"El cambio fue reenviado; se esperaban 0 cambios y llegaron %d",
			len(secondSync.Changes),
		)
	}

	log.Println(
		"Cursor evita reenvío = OK",
	)

	// =====================================
	// RESULTADO
	// =====================================

	log.Println("===================================")
	log.Println(" SYNC-4 RESTART VALIDADO")
	log.Println("===================================")
	log.Println(" Metadata sobrevive reinicio     = OK")
	log.Println(" ListFiles post-restart           = OK")
	log.Println(" Download post-restart            = OK")
	log.Println(" Anti-spoof                       = OK")
	log.Println(" Segundo dispositivo              = OK")
	log.Println(" FILE_CREATED persistente         = OK")
	log.Println(" Cursor persistente               = OK")
	log.Println(" Sin cambios duplicados           = OK")
	log.Println("===================================")
}
