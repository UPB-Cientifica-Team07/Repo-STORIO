package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	grpcClient "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"
)

type result struct {
	request string
	success bool
	message string
	fileID  string
	err     error
}

func main() {
	token := os.Getenv("TOKEN_USER2")

	if token == "" {
		log.Fatal("TOKEN_USER2 es obligatorio")
	}

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

	client := pb.NewFileServiceClient(connection)

	log.Println("===================================")
	log.Println(" PRUEBA DE CUOTA CONCURRENTE")
	log.Println("===================================")
	log.Println("Cuota total:       100 bytes")
	log.Println("Uso inicial:        60 bytes")
	log.Println("Disponible:         40 bytes")
	log.Println("Upload concurrente: 30 + 30 bytes")
	log.Println("Esperado: solo uno puede ingresar")
	log.Println("===================================")

	results := make(
		chan result,
		2,
	)

	var wg sync.WaitGroup

	wg.Add(2)

	start := make(
		chan struct{},
	)

	upload := func(
		requestName string,
		fileName string,
		character string,
	) {
		defer wg.Done()

		// Ambos goroutines esperan esta señal.
		<-start

		ctx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

		defer cancel()

		ctx = metadata.NewOutgoingContext(
			ctx,
			metadata.Pairs(
				"authorization",
				"Bearer "+token,
			),
		)

		content := []byte(
			strings.Repeat(
				character,
				30,
			),
		)

		response, err := client.UploadFile(
			ctx,
			&pb.UploadFileRequest{
				UserId:   "user-002",
				FileName: fileName,
				FileType: "text/plain",
				Content:  content,
			},
		)

		if err != nil {
			results <- result{
				request: requestName,
				err:     err,
			}

			return
		}

		results <- result{
			request: requestName,
			success: response.Success,
			message: response.Message,
			fileID:  response.FileId,
		}
	}

	go upload(
		"A",
		"concurrent-a.txt",
		"A",
	)

	go upload(
		"B",
		"concurrent-b.txt",
		"B",
	)

	// Libera las dos solicitudes casi simultáneamente.
	close(start)

	wg.Wait()

	close(results)

	successCount := 0
	rejectedCount := 0

	for currentResult := range results {
		log.Println("-----------------------------------")

		log.Printf(
			"Solicitud: %s",
			currentResult.request,
		)

		if currentResult.err != nil {
			log.Printf(
				"Error gRPC: %v",
				currentResult.err,
			)

			continue
		}

		log.Printf(
			"Success: %t",
			currentResult.success,
		)

		log.Printf(
			"Mensaje: %s",
			currentResult.message,
		)

		log.Printf(
			"File ID: %s",
			currentResult.fileID,
		)

		if currentResult.success {
			successCount++
		} else {
			rejectedCount++
		}
	}

	log.Println("===================================")
	log.Printf(
		"Uploads exitosos: %d",
		successCount,
	)

	log.Printf(
		"Uploads rechazados: %d",
		rejectedCount,
	)

	if successCount != 1 ||
		rejectedCount != 1 {

		log.Fatal(
			fmt.Sprintf(
				"FALLO DE CONCURRENCIA: se esperaba 1 éxito y 1 rechazo; obtenidos éxitos=%d rechazos=%d",
				successCount,
				rejectedCount,
			),
		)
	}

	log.Println("===================================")
	log.Println(" RESERVA ATÓMICA VALIDADA")
	log.Println(" Solo una solicitud obtuvo cuota")
	log.Println("===================================")
}
