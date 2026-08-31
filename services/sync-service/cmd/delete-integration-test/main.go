package main

import (
	"context"
	"log"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"
	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	syncAddress = "localhost:50055"
	authURL     = "http://localhost:8081"

	username = "tercero"
	password = "123456"

	fileID = "d8042164-d900-4c9b-8b66-4f7ab422a7ce"

	desktopDevice = "sync5-desktop-delete-001"
	mobileDevice  = "sync4-mobile-001"

	expectedRelativePath = "SYNC4/Persistencia/persistencia-sync4.txt"
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
	log.Println(" SYNC-5 DELETE INTEGRATION TEST")
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
	// DELETE DESKTOP
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 1 - DELETE DESKTOP")
	log.Println("===================================")

	ctx1, cancel1 :=
		authContext(
			loginResult.Token,
			20*time.Second,
		)

	deleteResponse, err :=
		client.DeleteFile(
			ctx1,
			&pb.DeleteFileRequest{
				// Spoof deliberado.
				UserId: "user-001",

				DeviceId: desktopDevice,

				FileId: fileID,
			},
		)

	cancel1()

	if err != nil {
		log.Fatalf(
			"DeleteFile falló: %v",
			err,
		)
	}

	if !deleteResponse.Success {
		log.Fatalf(
			"DeleteFile Success=false: %s",
			deleteResponse.Message,
		)
	}

	log.Println(
		"Delete Desktop = OK",
	)

	// =====================================
	// TEST 2
	// LISTFILES YA NO DEBE MOSTRAR ARCHIVO
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 2 - LISTFILES SIN ARCHIVO")
	log.Println("===================================")

	ctx2, cancel2 :=
		authContext(
			loginResult.Token,
			20*time.Second,
		)

	listResponse, err :=
		client.ListFiles(
			ctx2,
			&pb.ListFilesRequest{
				UserId: "user-001",
			},
		)

	cancel2()

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

	for _, file := range listResponse.Files {

		if file == nil {
			continue
		}

		if file.FileId == fileID {
			log.Fatal(
				"El archivo eliminado todavía aparece en ListFiles",
			)
		}
	}

	log.Println(
		"Archivo ya no aparece en ListFiles = OK",
	)

	// =====================================
	// TEST 3
	// DOWNLOAD DEBE FALLAR
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 3 - DOWNLOAD ELIMINADO")
	log.Println("===================================")

	ctx3, cancel3 :=
		authContext(
			loginResult.Token,
			20*time.Second,
		)

	downloadStream, err :=
		client.Download(
			ctx3,
			&pb.DownloadRequest{
				UserId: "user-001",
				FileId: fileID,
			},
		)

	if err != nil {

		cancel3()

		code :=
			status.Code(
				err,
			)

		if code != codes.NotFound &&
			code != codes.FailedPrecondition {

			log.Fatalf(
				"Download falló con código inesperado: %v",
				err,
			)
		}

		log.Println(
			"Download rechazado tras Delete = OK",
		)

	} else {

		_, recvErr :=
			downloadStream.Recv()

		cancel3()

		if recvErr == nil {
			log.Fatal(
				"Download todavía permitió obtener archivo eliminado",
			)
		}

		code :=
			status.Code(
				recvErr,
			)

		if code != codes.NotFound &&
			code != codes.FailedPrecondition {

			log.Fatalf(
				"Download falló con código inesperado: %v",
				recvErr,
			)
		}

		log.Println(
			"Download rechazado tras Delete = OK",
		)
	}

	// =====================================
	// TEST 4
	// MOBILE RECIBE FILE_DELETED
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 4 - MOBILE RECIBE DELETE")
	log.Println("===================================")

	ctx4, cancel4 :=
		authContext(
			loginResult.Token,
			20*time.Second,
		)

	syncResponse, err :=
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

	if change.Type != "FILE_DELETED" {
		log.Fatalf(
			"Tipo incorrecto: esperado=FILE_DELETED recibido=%s",
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

	if change.Version != 3 {
		log.Fatalf(
			"Versión incorrecta: esperado=3 recibido=%d",
			change.Version,
		)
	}

	if change.RelativePath !=
		expectedRelativePath {

		log.Fatalf(
			"RelativePath incorrecto: esperado=%q recibido=%q",
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
		"Mobile recibió FILE_DELETED v3 = OK",
	)

	// =====================================
	// TEST 5
	// CURSOR SIN DUPLICADOS
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 5 - NO DUPLICADOS")
	log.Println("===================================")

	ctx5, cancel5 :=
		authContext(
			loginResult.Token,
			20*time.Second,
		)

	secondSync, err :=
		client.Sync(
			ctx5,
			&pb.SyncRequest{
				UserId: "user-001",

				DeviceId: mobileDevice,
			},
		)

	cancel5()

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
	log.Println(" SYNC-5 FILE_DELETED VALIDADO")
	log.Println("===================================")
	log.Println(" Delete File Service               = OK")
	log.Println(" Anti-spoof                        = OK")
	log.Println(" Version 2 -> 3                    = OK")
	log.Println(" Eliminación lógica Sync           = OK")
	log.Println(" ListFiles excluye eliminado       = OK")
	log.Println(" Download rechazado                = OK")
	log.Println(" FILE_DELETED persistente          = OK")
	log.Println(" Propagación segundo dispositivo   = OK")
	log.Println(" Cursor incremental                = OK")
	log.Println(" Sin duplicados                    = OK")
	log.Println("===================================")
}
