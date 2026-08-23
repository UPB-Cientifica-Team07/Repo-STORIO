package main

import (
	"context"
	"log"
	"time"

	analysispb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/proto"
	filepb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"

	grpcClient "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	// =====================================
	// CONECTAR CON FILE SERVICE
	// =====================================

	fileConnection, err := grpcClient.NewClient(
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

	defer fileConnection.Close()

	fileClient := filepb.NewFileServiceClient(
		fileConnection,
	)

	// =====================================
	// CONECTAR CON ANALYSIS SERVICE
	// =====================================

	analysisConnection, err := grpcClient.NewClient(
		"localhost:50054",
		grpcClient.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		log.Fatalf(
			"No se pudo conectar con Analysis Service: %v",
			err,
		)
	}

	defer analysisConnection.Close()

	analysisClient := analysispb.NewAnalysisServiceClient(
		analysisConnection,
	)

	// =====================================
	// 1. SUBIR ARCHIVO
	// =====================================

	log.Println("===================================")
	log.Println(" 1. SUBIENDO ARCHIVO")
	log.Println("===================================")

	uploadResponse, err := fileClient.UploadFile(
		ctx,
		&filepb.UploadFileRequest{
			UserId:   "user-001",
			FileName: "sistemas-distribuidos.txt",
			FileType: "text/plain",
			Content: []byte(
				"Los sistemas distribuidos permiten que múltiples servicios " +
					"se comuniquen mediante una red. gRPC permite implementar " +
					"comunicación eficiente entre microservicios.",
			),
		},
	)

	if err != nil {
		log.Fatalf(
			"Error subiendo archivo: %v",
			err,
		)
	}

	if !uploadResponse.Success {
		log.Fatalf(
			"File Service rechazó el archivo: %s",
			uploadResponse.Message,
		)
	}

	fileID := uploadResponse.FileId

	log.Printf("Archivo subido correctamente")
	log.Printf("File ID: %s", fileID)

	// =====================================
	// 2. ANALIZAR ARCHIVO
	// =====================================

	log.Println("===================================")
	log.Println(" 2. ENVIANDO ARCHIVO A ANALYSIS SERVICE")
	log.Println("===================================")

	analysisResponse, err := analysisClient.AnalyzeFile(
		ctx,
		&analysispb.AnalyzeFileRequest{
			UserId: "user-001",
			FileId: fileID,
		},
	)

	if err != nil {
		log.Fatalf(
			"Error analizando archivo: %v",
			err,
		)
	}

	if !analysisResponse.Success {
		log.Fatalf(
			"Analysis Service rechazó el análisis: %s",
			analysisResponse.Message,
		)
	}

	analysisID := analysisResponse.AnalysisId

	log.Printf("Archivo analizado correctamente")
	log.Printf("Analysis ID: %s", analysisID)
	log.Printf("Resumen: %s", analysisResponse.Summary)

	// =====================================
	// 3. CONSULTAR ANÁLISIS
	// =====================================

	log.Println("===================================")
	log.Println(" 3. CONSULTANDO ANÁLISIS")
	log.Println("===================================")

	getResponse, err := analysisClient.GetAnalysis(
		ctx,
		&analysispb.GetAnalysisRequest{
			AnalysisId: analysisID,
		},
	)

	if err != nil {
		log.Fatalf(
			"Error obteniendo análisis: %v",
			err,
		)
	}

	log.Printf("Success: %t", getResponse.Success)
	log.Printf("Mensaje: %s", getResponse.Message)

	if getResponse.Analysis != nil {
		log.Printf(
			"Analysis ID: %s",
			getResponse.Analysis.AnalysisId,
		)

		log.Printf(
			"Archivo: %s",
			getResponse.Analysis.FileId,
		)

		log.Printf(
			"Resumen: %s",
			getResponse.Analysis.Summary,
		)
	}

	// =====================================
	// 4. LISTAR ANÁLISIS
	// =====================================

	log.Println("===================================")
	log.Println(" 4. LISTANDO ANÁLISIS DEL USUARIO")
	log.Println("===================================")

	listResponse, err := analysisClient.ListAnalyses(
		ctx,
		&analysispb.ListAnalysesRequest{
			UserId: "user-001",
		},
	)

	if err != nil {
		log.Fatalf(
			"Error listando análisis: %v",
			err,
		)
	}

	log.Printf(
		"Cantidad de análisis: %d",
		len(listResponse.Analyses),
	)

	for _, analysis := range listResponse.Analyses {
		log.Printf(
			"- Analysis ID: %s | File ID: %s",
			analysis.AnalysisId,
			analysis.FileId,
		)

		log.Printf(
			"  Resumen: %s",
			analysis.Summary,
		)
	}

	log.Println("===================================")
	log.Println(" PRUEBA DE INTEGRACIÓN COMPLETADA")
	log.Println("===================================")
}
