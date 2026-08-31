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

	desktopDevice = "sync5-desktop-001"
	mobileDevice  = "sync4-mobile-001"

	expectedRelativePath = "SYNC4/Persistencia/persistencia-sync4.txt"
)

var newContent = []byte(
	"SYNC-5 FILE_CHANGED VERSION 2\n",
)

func authContext(
	token string,
	timeout time.Duration,
) (
	context.Context,
	context.CancelFunc,
) {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			timeout,
		)

	ctx =
		metadata.NewOutgoingContext(
			ctx,
			metadata.Pairs(
				"authorization",
				"Bearer "+token,
			),
		)

	return ctx, cancel
}

func main() {

	log.Println("===================================")
	log.Println(" SYNC-5 UPDATE INTEGRATION TEST")
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
			"Error conectando con Sync: %v",
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
	// UPDATE DESKTOP
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 1 - UPDATE DESKTOP")
	log.Println("===================================")

	ctx1, cancel1 :=
		authContext(
			loginResult.Token,
			20*time.Second,
		)

	updateResponse, err :=
		client.UpdateFile(
			ctx1,
			&pb.UpdateFileRequest{
				// Spoof deliberado.
				UserId: "user-001",

				DeviceId: desktopDevice,

				FileId: fileID,

				Content: newContent,
			},
		)

	cancel1()

	if err != nil {
		log.Fatalf(
			"UpdateFile falló: %v",
			err,
		)
	}

	if !updateResponse.Success {
		log.Fatalf(
			"UpdateFile Success=false: %s",
			updateResponse.Message,
		)
	}

	if updateResponse.FileId != fileID {
		log.Fatalf(
			"FileID incorrecto: esperado=%s recibido=%s",
			fileID,
			updateResponse.FileId,
		)
	}

	if updateResponse.Version != 2 {
		log.Fatalf(
			"Versión incorrecta: esperado=2 recibido=%d",
			updateResponse.Version,
		)
	}

	if updateResponse.Size != int64(len(newContent)) {
		log.Fatalf(
			"Tamaño incorrecto: esperado=%d recibido=%d",
			len(newContent),
			updateResponse.Size,
		)
	}

	log.Println(
		"Update Desktop = OK",
	)

	log.Printf(
		"Version=%d Size=%d",
		updateResponse.Version,
		updateResponse.Size,
	)

	// =====================================
	// TEST 2
	// DOWNLOAD CONTENIDO NUEVO
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 2 - DOWNLOAD ACTUALIZADO")
	log.Println("===================================")

	ctx2, cancel2 :=
		authContext(
			loginResult.Token,
			20*time.Second,
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
		cancel2()

		log.Fatalf(
			"Download falló al abrir: %v",
			err,
		)
	}

	var (
		downloadedMetadata *pb.FileMetadata
		downloadedContent  []byte
	)

	for {

		response, recvErr :=
			downloadStream.Recv()

		if recvErr == io.EOF {
			break
		}

		if recvErr != nil {
			cancel2()

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

	cancel2()

	if downloadedMetadata == nil {
		log.Fatal(
			"Download no devolvió metadata",
		)
	}

	if downloadedMetadata.Version != 2 {
		log.Fatalf(
			"Versión Download incorrecta: esperado=2 recibido=%d",
			downloadedMetadata.Version,
		)
	}

	if downloadedMetadata.RelativePath !=
		expectedRelativePath {

		log.Fatalf(
			"RelativePath incorrecto: esperado=%q recibido=%q",
			expectedRelativePath,
			downloadedMetadata.RelativePath,
		)
	}

	if !bytes.Equal(
		downloadedContent,
		newContent,
	) {

		log.Fatalf(
			"Contenido incorrecto: esperado=%q recibido=%q",
			string(newContent),
			string(downloadedContent),
		)
	}

	log.Println(
		"Download contenido actualizado = OK",
	)

	// =====================================
	// TEST 3
	// MOBILE RECIBE FILE_CHANGED
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 3 - MOBILE RECIBE CAMBIO")
	log.Println("===================================")

	ctx3, cancel3 :=
		authContext(
			loginResult.Token,
			20*time.Second,
		)

	syncResponse, err :=
		client.Sync(
			ctx3,
			&pb.SyncRequest{
				UserId: "user-001",

				DeviceId: mobileDevice,
			},
		)

	cancel3()

	if err != nil {
		log.Fatalf(
			"Sync mobile falló: %v",
			err,
		)
	}

	if !syncResponse.Success {
		log.Fatalf(
			"Sync mobile Success=false: %s",
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

	if change.Type != "FILE_CHANGED" {
		log.Fatalf(
			"Tipo incorrecto: esperado=FILE_CHANGED recibido=%s",
			change.Type,
		)
	}

	if change.FileId != fileID {
		log.Fatalf(
			"FileID incorrecto: esperado=%s recibido=%s",
			fileID,
			change.FileId,
		)
	}

	if change.Version != 2 {
		log.Fatalf(
			"Versión cambio incorrecta: esperado=2 recibido=%d",
			change.Version,
		)
	}

	if change.RelativePath !=
		expectedRelativePath {

		log.Fatalf(
			"RelativePath cambio incorrecto: esperado=%q recibido=%q",
			expectedRelativePath,
			change.RelativePath,
		)
	}

	if change.OriginDeviceId !=
		desktopDevice {

		log.Fatalf(
			"OriginDevice incorrecto: esperado=%s recibido=%s",
			desktopDevice,
			change.OriginDeviceId,
		)
	}

	log.Println(
		"Mobile recibió FILE_CHANGED v2 = OK",
	)

	// =====================================
	// TEST 4
	// CURSOR NO DUPLICA
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 4 - NO DUPLICADOS")
	log.Println("===================================")

	ctx4, cancel4 :=
		authContext(
			loginResult.Token,
			20*time.Second,
		)

	secondSync, err :=
		client.Sync(
			ctx4,
			&pb.SyncRequest{
				UserId: "user-001",

				DeviceId: mobileDevice,
			},
		)

	cancel4()

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
			"Se esperaban 0 cambios duplicados; recibidos=%d",
			len(secondSync.Changes),
		)
	}

	log.Println(
		"Cursor evita duplicados = OK",
	)

	// =====================================
	// RESULTADO
	// =====================================

	log.Println("===================================")
	log.Println(" SYNC-5 FILE_CHANGED VALIDADO")
	log.Println("===================================")
	log.Println(" Update File Service              = OK")
	log.Println(" Anti-spoof                       = OK")
	log.Println(" Version 1 -> 2                   = OK")
	log.Println(" Contenido actualizado            = OK")
	log.Println(" RelativePath preservado          = OK")
	log.Println(" FILE_CHANGED persistente         = OK")
	log.Println(" Propagación a segundo dispositivo= OK")
	log.Println(" Cursor incremental               = OK")
	log.Println(" Sin duplicados                   = OK")
	log.Println("===================================")
}
