package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const protectedFileID = "f875809f-0dd2-4a1c-ae65-835b25f83df3"

func authenticatedContext(
	token string,
) (context.Context, context.CancelFunc) {

	md :=
		metadata.New(
			map[string]string{
				"authorization": "Bearer " + token,
			},
		)

	ctx :=
		metadata.NewOutgoingContext(
			context.Background(),
			md,
		)

	return context.WithTimeout(
		ctx,
		10*time.Second,
	)
}

func main() {

	tokenUser :=
		os.Getenv(
			"TOKEN_USER",
		)

	tokenAdmin :=
		os.Getenv(
			"TOKEN_ADMIN",
		)

	if tokenUser == "" {
		log.Fatal(
			"TOKEN_USER no definido",
		)
	}

	if tokenAdmin == "" {
		log.Fatal(
			"TOKEN_ADMIN no definido",
		)
	}

	conn,
		err :=
		grpc.NewClient(
			"localhost:50053",
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		)

	if err != nil {
		log.Fatalf(
			"No se pudo conectar: %v",
			err,
		)
	}

	defer conn.Close()

	client :=
		pb.NewFileServiceClient(
			conn,
		)

	log.Println(
		"===================================",
	)

	log.Println(
		" PRUEBA DE AUTORIZACIÓN",
	)

	log.Println(
		"===================================",
	)

	log.Println(
		"Archivo protegido:",
		protectedFileID,
	)

	log.Println(
		"Propietario esperado: user-001",
	)

	log.Println(
		"Atacante autenticado: user-002 / USER",
	)

	// =====================================
	// PRUEBA 1: GET AJENO
	// =====================================

	log.Println()
	log.Println(
		"-----------------------------------",
	)

	log.Println(
		"1. GET DE ARCHIVO AJENO",
	)

	log.Println(
		"-----------------------------------",
	)

	ctxGet,
		cancelGet :=
		authenticatedContext(
			tokenUser,
		)

	_,
		err =
		client.GetFile(
			ctxGet,
			&pb.GetFileRequest{
				FileId: protectedFileID,
			},
		)

	cancelGet()

	if err == nil {

		log.Fatal(
			"FALLO DE SEGURIDAD: user-002 pudo leer archivo de user-001",
		)
	}

	if status.Code(err) !=
		codes.PermissionDenied {

		log.Fatalf(
			"Se esperaba PermissionDenied en GetFile, se obtuvo: %v",
			err,
		)
	}

	log.Println(
		"GET bloqueado correctamente:",
		err,
	)

	// =====================================
	// PRUEBA 2: DELETE AJENO
	// =====================================

	log.Println()
	log.Println(
		"-----------------------------------",
	)

	log.Println(
		"2. DELETE DE ARCHIVO AJENO",
	)

	log.Println(
		"-----------------------------------",
	)

	ctxDelete,
		cancelDelete :=
		authenticatedContext(
			tokenUser,
		)

	_,
		err =
		client.DeleteFile(
			ctxDelete,
			&pb.DeleteFileRequest{
				FileId: protectedFileID,
			},
		)

	cancelDelete()

	if err == nil {

		log.Fatal(
			"FALLO DE SEGURIDAD: user-002 pudo borrar archivo de user-001",
		)
	}

	if status.Code(err) !=
		codes.PermissionDenied {

		log.Fatalf(
			"Se esperaba PermissionDenied en DeleteFile, se obtuvo: %v",
			err,
		)
	}

	log.Println(
		"DELETE bloqueado correctamente:",
		err,
	)

	// =====================================
	// PRUEBA 3: LIST CON USER_ID FALSIFICADO
	// =====================================

	log.Println()
	log.Println(
		"-----------------------------------",
	)

	log.Println(
		"3. LISTFILES CON USER_ID FALSIFICADO",
	)

	log.Println(
		"-----------------------------------",
	)

	ctxList,
		cancelList :=
		authenticatedContext(
			tokenUser,
		)

	listResponse,
		err :=
		client.ListFiles(
			ctxList,
			&pb.ListFilesRequest{
				// user-002 intenta pedir
				// explícitamente archivos
				// de user-001.
				UserId: "user-001",
			},
		)

	cancelList()

	if err != nil {

		log.Fatalf(
			"ListFiles falló: %v",
			err,
		)
	}

	log.Printf(
		"Cantidad de archivos devueltos: %d",
		len(
			listResponse.Files,
		),
	)

	for _, file := range listResponse.Files {

		log.Printf(
			"Archivo listado | ID=%s | Owner=%s | Nombre=%s",
			file.FileId,
			file.UserId,
			file.FileName,
		)

		if file.UserId !=
			"user-002" {

			log.Fatalf(
				"FALLO DE SEGURIDAD: ListFiles expuso archivo de %s",
				file.UserId,
			)
		}
	}

	log.Println(
		"ListFiles ignoró correctamente user_id=user-001",
	)

	// =====================================
	// PRUEBA 4: ADMIN CONFIRMA QUE ARCHIVO
	// TODAVÍA EXISTE
	// =====================================

	log.Println()
	log.Println(
		"-----------------------------------",
	)

	log.Println(
		"4. VERIFICACIÓN FINAL COMO ADMIN",
	)

	log.Println(
		"-----------------------------------",
	)

	ctxAdmin,
		cancelAdmin :=
		authenticatedContext(
			tokenAdmin,
		)

	adminResponse,
		err :=
		client.GetFile(
			ctxAdmin,
			&pb.GetFileRequest{
				FileId: protectedFileID,
			},
		)

	cancelAdmin()

	if err != nil {

		log.Fatalf(
			"Admin no pudo verificar archivo: %v",
			err,
		)
	}

	if !adminResponse.Success ||
		adminResponse.File == nil {

		log.Fatal(
			"El archivo protegido desapareció",
		)
	}

	log.Printf(
		"Archivo sigue existiendo | FileID=%s | Owner=%s",
		adminResponse.File.FileId,
		adminResponse.File.UserId,
	)

	if adminResponse.File.UserId !=
		"user-001" {

		log.Fatalf(
			"Owner inesperado: %s",
			adminResponse.File.UserId,
		)
	}

	log.Println()
	log.Println(
		"===================================",
	)

	log.Println(
		" AUTORIZACIÓN VERIFICADA",
	)

	log.Println(
		"===================================",
	)

	log.Println(
		"GET ajeno: BLOQUEADO",
	)

	log.Println(
		"DELETE ajeno: BLOQUEADO",
	)

	log.Println(
		"ListFiles falsificado: BLOQUEADO",
	)

	log.Println(
		"Archivo original: CONSERVADO",
	)

	log.Println(
		"===================================",
	)
}
