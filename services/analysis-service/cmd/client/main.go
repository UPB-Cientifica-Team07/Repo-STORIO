package main

import (
	"context"
	"log"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	connection, err := grpc.NewClient(
		"localhost:50054",
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		log.Fatalf(
			"No se pudo conectar al Analysis Service: %v",
			err,
		)
	}

	defer connection.Close()

	client := pb.NewAnalysisServiceClient(
		connection,
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()

	// =====================================
	// ANALIZAR ARCHIVO
	// =====================================

	log.Println("===================================")
	log.Println(" ANALIZANDO ARCHIVO")
	log.Println("===================================")

	analysisResponse, err := client.AnalyzeFile(
		ctx,
		&pb.AnalyzeFileRequest{
			UserId: "user-001",
			FileId: "b876028b-172b-46bb-9115-7ab0176bad6e",
		},
	)

	if err != nil {
		log.Fatalf(
			"Error analizando archivo: %v",
			err,
		)
	}

	log.Printf(
		"Success: %v",
		analysisResponse.Success,
	)

	log.Printf(
		"Mensaje: %s",
		analysisResponse.Message,
	)

	log.Printf(
		"Analysis ID: %s",
		analysisResponse.AnalysisId,
	)

	log.Printf(
		"Resumen: %s",
		analysisResponse.Summary,
	)

	analysisID := analysisResponse.AnalysisId

	// =====================================
	// OBTENER ANÁLISIS
	// =====================================

	log.Println("===================================")
	log.Println(" OBTENIENDO ANÁLISIS")
	log.Println("===================================")

	getResponse, err := client.GetAnalysis(
		ctx,
		&pb.GetAnalysisRequest{
			AnalysisId: analysisID,
		},
	)

	if err != nil {
		log.Fatalf(
			"Error obteniendo análisis: %v",
			err,
		)
	}

	log.Printf(
		"Success: %v",
		getResponse.Success,
	)

	log.Printf(
		"Mensaje: %s",
		getResponse.Message,
	)

	if getResponse.Analysis != nil {

		log.Printf(
			"Analysis ID: %s",
			getResponse.Analysis.AnalysisId,
		)

		log.Printf(
			"Usuario: %s",
			getResponse.Analysis.UserId,
		)

		log.Printf(
			"Archivo: %s",
			getResponse.Analysis.FileId,
		)

		log.Printf(
			"Resumen: %s",
			getResponse.Analysis.Summary,
		)

		log.Printf(
			"Creado: %s",
			getResponse.Analysis.CreatedAt,
		)
	}

	// =====================================
	// LISTAR ANÁLISIS DEL USUARIO
	// =====================================

	log.Println("===================================")
	log.Println(" LISTANDO ANÁLISIS")
	log.Println("===================================")

	listResponse, err := client.ListAnalyses(
		ctx,
		&pb.ListAnalysesRequest{
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
		"Success: %v",
		listResponse.Success,
	)

	log.Printf(
		"Mensaje: %s",
		listResponse.Message,
	)

	log.Printf(
		"Cantidad de análisis: %d",
		len(listResponse.Analyses),
	)

	for _, analysis := range listResponse.Analyses {

		log.Printf(
			"- ID: %s | Archivo: %s",
			analysis.AnalysisId,
			analysis.FileId,
		)

		log.Printf(
			"  Resumen: %s",
			analysis.Summary,
		)
	}

	log.Println("===================================")
	log.Println(" PRUEBA COMPLETADA")
	log.Println("===================================")
}
