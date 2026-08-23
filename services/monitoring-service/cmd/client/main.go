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

	log.Println("===================================")
	log.Println(" MONITORING CLIENT")
	log.Println(" Servidor:", serverAddress)
	log.Println("===================================")

	// =====================================
	// CONEXIÓN GRPC
	// =====================================

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

	client :=
		pb.NewMonitoringServiceClient(
			conn,
		)

	// =====================================
	// CONSULTAR MÉTRICAS
	// =====================================

	log.Println("")
	log.Println("===================================")
	log.Println(" CONSULTANDO MÉTRICAS")
	log.Println("===================================")

	metricsCtx, metricsCancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	metricsResponse, err :=
		client.GetMetrics(
			metricsCtx,
			&pb.GetMetricsRequest{},
		)

	metricsCancel()

	if err != nil {

		log.Fatalf(
			"Error consultando métricas: %v",
			err,
		)
	}

	metrics :=
		metricsResponse.GetMetrics()

	log.Printf(
		"Total de métricas registradas: %d",
		len(metrics),
	)

	if len(metrics) == 0 {

		log.Println(
			"No existen métricas registradas.",
		)
	}

	for index, metric :=
		range metrics {

		log.Println(
			"-----------------------------------",
		)

		log.Printf(
			"Métrica #%d",
			index+1,
		)

		log.Printf(
			"Componente: %s",
			metric.GetComponentName(),
		)

		log.Printf(
			"ID: %s",
			metric.GetComponentId(),
		)

		// =====================================
		// CPU
		// =====================================

		log.Printf(
			"CPU: %.2f%%",
			metric.GetCpuUsage(),
		)

		// =====================================
		// MEMORIA
		// Actualmente reportada en MB
		// =====================================

		log.Printf(
			"Memoria: %.2f MB",
			metric.GetMemoryUsage(),
		)

		// =====================================
		// STORAGE
		// =====================================

		storageBytes :=
			metric.GetStorageUsage()

		log.Printf(
			"Almacenamiento: %d bytes",
			storageBytes,
		)

		log.Printf(
			"Almacenamiento: %.2f KB",
			float64(storageBytes)/1024,
		)

		log.Printf(
			"Almacenamiento: %.2f MB",
			float64(storageBytes)/(1024*1024),
		)

		// =====================================
		// CONEXIONES
		// =====================================

		log.Printf(
			"Conexiones activas: %d",
			metric.GetActiveConnections(),
		)

		// =====================================
		// SOLICITUDES
		// =====================================

		log.Printf(
			"Total solicitudes: %d",
			metric.GetTotalRequests(),
		)

		// =====================================
		// TIMESTAMP
		// =====================================

		if metric.GetTimestamp() != 0 {

			log.Printf(
				"Timestamp: %s",
				time.Unix(
					metric.GetTimestamp(),
					0,
				).Format(
					"2006-01-02 15:04:05",
				),
			)

		} else {

			log.Println(
				"Timestamp: no disponible",
			)
		}
	}

	// =====================================
	// CONSULTAR ESTADO DEL FILE SERVICE
	// =====================================

	log.Println("")
	log.Println("===================================")
	log.Println(" CONSULTANDO ESTADO DEL FILE SERVICE")
	log.Println("===================================")

	fileCtx, fileCancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	fileStatusResponse, err :=
		client.GetServiceStatus(
			fileCtx,
			&pb.ServiceRequest{
				ComponentName: "File Service",
			},
		)

	fileCancel()

	if err != nil {

		log.Printf(
			"Error consultando estado del File Service: %v",
			err,
		)

	} else {

		printServiceStatus(
			fileStatusResponse,
		)
	}

	// =====================================
	// CONSULTAR ESTADO DEL SYNC SERVICE
	// =====================================

	log.Println("")
	log.Println("===================================")
	log.Println(" CONSULTANDO ESTADO DEL SYNC SERVICE")
	log.Println("===================================")

	syncCtx, syncCancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	syncStatusResponse, err :=
		client.GetServiceStatus(
			syncCtx,
			&pb.ServiceRequest{
				ComponentName: "Sync Service",
			},
		)

	syncCancel()

	if err != nil {

		log.Printf(
			"Error consultando estado del Sync Service: %v",
			err,
		)

	} else {

		printServiceStatus(
			syncStatusResponse,
		)
	}

	log.Println("")
	log.Println("===================================")
	log.Println(" CONSULTA FINALIZADA")
	log.Println("===================================")
}

// =====================================
// IMPRIMIR ESTADO DE SERVICIO
// =====================================

func printServiceStatus(
	statusResponse *pb.ServiceResponse,
) {

	if statusResponse == nil {

		log.Println(
			"No se recibió información del servicio.",
		)

		return
	}

	log.Println(
		"-----------------------------------",
	)

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
			).Format(
				"2006-01-02 15:04:05",
			),
		)

	} else {

		log.Println(
			"Última actualización: no disponible",
		)
	}
}