package main

import (
	"context"
	"log"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/generated"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const serverAddress = "localhost:50051"

func main() {

	conn, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		log.Fatalf(
			"No se pudo conectar al Monitoring Service: %v",
			err,
		)
	}

	defer conn.Close()

	client := pb.NewMonitoringServiceClient(conn)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	// =====================================
	// CONSULTAR MÉTRICAS
	// =====================================

	log.Println("===================================")
	log.Println(" CONSULTANDO MÉTRICAS")
	log.Println("===================================")

	metricsResponse, err := client.GetMetrics(
		ctx,
		&pb.GetMetricsRequest{},
	)

	if err != nil {
		log.Fatalf(
			"Error consultando métricas: %v",
			err,
		)
	}

	metrics := metricsResponse.GetMetrics()

	log.Printf(
		"Total de métricas registradas: %d",
		len(metrics),
	)

	for index, metric := range metrics {

		log.Println("-----------------------------------")
		log.Printf("Métrica #%d", index+1)

		log.Printf(
			"Componente: %s",
			metric.GetComponentName(),
		)

		log.Printf(
			"ID: %s",
			metric.GetComponentId(),
		)

		log.Printf(
			"CPU: %.2f%%",
			metric.GetCpuUsage(),
		)

		log.Printf(
			"Memoria: %.2f%%",
			metric.GetMemoryUsage(),
		)

		log.Printf(
			"Conexiones activas: %d",
			metric.GetActiveConnections(),
		)

		log.Printf(
			"Total solicitudes: %d",
			metric.GetTotalRequests(),
		)

		log.Printf(
			"Timestamp: %s",
			time.Unix(
				metric.GetTimestamp(),
				0,
			).Format("2006-01-02 15:04:05"),
		)
	}

	// =====================================
	// CONSULTAR ESTADO DEL FILE SERVICE
	// =====================================

	log.Println("")
	log.Println("===================================")
	log.Println(" CONSULTANDO ESTADO DEL FILE SERVICE")
	log.Println("===================================")

	statusResponse, err := client.GetServiceStatus(
		ctx,
		&pb.ServiceRequest{
			ComponentName: "File Service",
		},
	)

	if err != nil {
		log.Fatalf(
			"Error consultando estado del servicio: %v",
			err,
		)
	}

	log.Println("-----------------------------------")

	log.Printf(
		"Componente: %s",
		statusResponse.GetComponentName(),
	)

	log.Printf(
		"ID: %s",
		statusResponse.GetComponentId(),
	)

	log.Printf(
		"Estado: %s",
		statusResponse.GetStatus(),
	)

	log.Printf(
		"Mensaje: %s",
		statusResponse.GetMessage(),
	)

	if statusResponse.GetLastUpdated() != 0 {

		log.Printf(
			"Última actualización: %s",
			time.Unix(
				statusResponse.GetLastUpdated(),
				0,
			).Format("2006-01-02 15:04:05"),
		)
	}
}
