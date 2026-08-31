package main

import (
	"context"
	"log"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/proto"

	grpcClient "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	serverAddress = "localhost:50055"
	authURL       = "http://localhost:8081"

	realUserID = "user-003"

	spoofedUserID = "user-001"

	deviceID = "sync-security-test-device"
)

func main() {

	log.Println("===================================")
	log.Println(" SYNC SECURITY TEST")
	log.Println("===================================")

	// =====================================
	// LOGIN REAL: USER-003
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

	if loginResult.UserID != realUserID {
		log.Fatalf(
			"Usuario autenticado inesperado: esperado=%s obtenido=%s",
			realUserID,
			loginResult.UserID,
		)
	}

	log.Printf(
		"Login real: %s",
		loginResult.UserID,
	)

	// =====================================
	// CONEXIÓN GRPC
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
			"No se pudo conectar con Sync Service: %v",
			err,
		)
	}

	defer connection.Close()

	client :=
		pb.NewSyncServiceClient(
			connection,
		)

	// =====================================
	// TEST 1:
	// SIN TOKEN DEBE FALLAR
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 1 - TOKEN OBLIGATORIO")
	log.Println("===================================")

	noAuthCtx, noAuthCancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	_, err =
		client.ListFiles(
			noAuthCtx,
			&pb.ListFilesRequest{
				UserId: spoofedUserID,
			},
		)

	noAuthCancel()

	if err == nil {
		log.Fatal(
			"ERROR: ListFiles sin token debía ser rechazado",
		)
	}

	if status.Code(err) !=
		codes.Unauthenticated {

		log.Fatalf(
			"Código inesperado sin token: %s | %v",
			status.Code(err),
			err,
		)
	}

	log.Printf(
		"Sin token → %s",
		status.Code(err),
	)

	log.Println(
		"TOKEN OBLIGATORIO VALIDADO",
	)

	// =====================================
	// CONTEXTO AUTENTICADO
	// =====================================

	authenticatedCtx :=
		func() (
			context.Context,
			context.CancelFunc,
		) {

			ctx, cancel :=
				context.WithTimeout(
					context.Background(),
					15*time.Second,
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
	// TEST 2:
	// UPLOAD CON USER_ID FALSIFICADO
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 2 - UPLOAD ANTI-SPOOF")
	log.Println("===================================")
	log.Printf(
		"Token real: %s",
		realUserID,
	)
	log.Printf(
		"UserId falsificado enviado: %s",
		spoofedUserID,
	)

	content :=
		[]byte(
			"Sync anti-spoof test.",
		)

	uploadCtx, uploadCancel :=
		authenticatedCtx()

	uploadStream, err :=
		client.Upload(
			uploadCtx,
		)

	if err != nil {
		uploadCancel()

		log.Fatalf(
			"No se pudo iniciar Upload: %v",
			err,
		)
	}

	// =====================================
	// METADATA CON SPOOF DELIBERADO
	// =====================================

	err =
		uploadStream.Send(
			&pb.UploadRequest{
				Data: &pb.UploadRequest_Metadata{
					Metadata: &pb.UploadMetadata{
						UserId:   spoofedUserID,
						DeviceId: deviceID,
						FileName: "sync-security-test.txt",
						FileType: "text/plain",
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
			"No se pudo enviar metadata: %v",
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
			"No se pudo enviar contenido: %v",
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

	log.Printf(
		"Upload Success: %t",
		uploadResponse.Success,
	)

	log.Printf(
		"File ID: %s",
		fileID,
	)

	// =====================================
	// TEST 3:
	// LISTFILES TAMBIÉN IGNORA SPOOF
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 3 - LIST ANTI-SPOOF")
	log.Println("===================================")

	listCtx, listCancel :=
		authenticatedCtx()

	listResponse, err :=
		client.ListFiles(
			listCtx,
			&pb.ListFilesRequest{
				// Intentamos volver a decirle al
				// servidor que somos user-001.
				UserId: spoofedUserID,
			},
		)

	listCancel()

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

	found :=
		false

	for _, file := range listResponse.Files {

		if file == nil {
			continue
		}

		if file.FileId !=
			fileID {

			continue
		}

		found =
			true

		log.Printf(
			"Archivo encontrado: %s",
			file.FileName,
		)

		log.Printf(
			"Owner registrado: %s",
			file.UserId,
		)

		if file.UserId !=
			realUserID {

			log.Fatalf(
				"VULNERABILIDAD: owner esperado=%s obtenido=%s",
				realUserID,
				file.UserId,
			)
		}
	}

	if !found {
		log.Fatalf(
			"El archivo recién cargado no apareció para %s",
			realUserID,
		)
	}

	log.Println(
		"ANTI-SPOOF VALIDADO: request.UserId fue ignorado",
	)

	// =====================================
	// TEST 4:
	// DELETE CON USER_ID FALSIFICADO
	// =====================================

	log.Println("===================================")
	log.Println(" TEST 4 - DELETE ANTI-SPOOF")
	log.Println("===================================")

	deleteCtx, deleteCancel :=
		authenticatedCtx()

	deleteResponse, err :=
		client.DeleteFile(
			deleteCtx,
			&pb.DeleteFileRequest{
				UserId:   spoofedUserID,
				DeviceId: deviceID,
				FileId:   fileID,
			},
		)

	deleteCancel()

	if err != nil {
		log.Fatalf(
			"DeleteFile falló: %v",
			err,
		)
	}

	if !deleteResponse.Success {
		log.Fatalf(
			"DeleteFile respondió Success=false: %s",
			deleteResponse.Message,
		)
	}

	log.Println(
		"Archivo de prueba eliminado correctamente",
	)

	log.Println("===================================")
	log.Println(" SYNC SECURITY VALIDADO")
	log.Println("===================================")
	log.Println(" Token obligatorio       = OK")
	log.Println(" Unary interceptor        = OK")
	log.Println(" Stream interceptor       = OK")
	log.Println(" Upload anti-spoof        = OK")
	log.Println(" ListFiles anti-spoof     = OK")
	log.Println(" DeleteFile anti-spoof    = OK")
	log.Println("===================================")
}
