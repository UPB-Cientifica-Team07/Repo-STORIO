package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/internal/fileclient"
	grpcserver "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/internal/grpc"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/internal/monitoring"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/internal/service"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/analysis-service/proto"

	"google.golang.org/grpc"
)

const (
	servicePort = ":50054"

	componentID   = "analysis-service-01"
	componentName = "Analysis Service"
)

func main() {

	log.Println("===================================")
	log.Println(" ANALYSIS SERVICE")
	log.Println(" Protocolo: gRPC")
	log.Println(" Puerto: 50054")
	log.Println(" Estado: INICIANDO")
	log.Println("===================================")

	// ==============================
	// REPOSITORIO
	// ==============================

	analysisRepository := repository.NewAnalysisRepository()

	// ==============================
	// SERVICIO DE ANÁLISIS
	// ==============================

	analysisService := service.NewAnalysisService(
		analysisRepository,
	)

	// ==============================
	// CLIENTE DEL FILE SERVICE
	// ==============================

	fileClient, err := fileclient.NewClient()

	if err != nil {
		log.Fatalf(
			"No se pudo conectar al File Service: %v",
			err,
		)
	}

	defer fileClient.Close()

	// ==============================
	// CLIENTE DE MONITOREO
	// ==============================

	monitoringClient, err := monitoring.NewClient(
		componentID,
		componentName,
	)

	if err != nil {
		log.Fatalf(
			"No se pudo conectar al Monitoring Service: %v",
			err,
		)
	}

	defer monitoringClient.Close()

	err = monitoringClient.ReportStatus(
		"ACTIVE",
		"Analysis Service iniciado correctamente",
	)

	if err != nil {
		log.Printf(
			"Error enviando estado al Monitoring Service: %v",
			err,
		)
	} else {
		log.Println(
			"Estado enviado correctamente al Monitoring Service",
		)
	}

	// ==============================
	// SERVIDOR gRPC
	// ==============================

	listener, err := net.Listen(
		"tcp",
		servicePort,
	)

	if err != nil {
		log.Fatalf(
			"No se pudo iniciar el servidor: %v",
			err,
		)
	}

	grpcServer := grpc.NewServer()

	analysisServer := grpcserver.NewServer(
		analysisService,
		fileClient,
	)

	pb.RegisterAnalysisServiceServer(
		grpcServer,
		analysisServer,
	)

	// ==============================
	// MANEJO DE SEÑALES
	// ==============================

	signalChannel := make(
		chan os.Signal,
		1,
	)

	signal.Notify(
		signalChannel,
		os.Interrupt,
		syscall.SIGTERM,
	)

	go func() {

		log.Println("===================================")
		log.Println(" ANALYSIS SERVICE ACTIVO")
		log.Println(" Escuchando en :50054")
		log.Println("===================================")

		if err := grpcServer.Serve(
			listener,
		); err != nil {

			log.Fatalf(
				"Error ejecutando servidor: %v",
				err,
			)
		}
	}()

	// Esperar señal de terminación
	<-signalChannel

	log.Println("===================================")
	log.Println(" DETENIENDO ANALYSIS SERVICE")
	log.Println("===================================")

	err = monitoringClient.ReportStatus(
		"INACTIVE",
		"Analysis Service detenido correctamente",
	)

	if err != nil {
		log.Printf(
			"Error enviando estado INACTIVE: %v",
			err,
		)
	} else {
		log.Println(
			"Estado INACTIVE enviado correctamente",
		)
	}

	grpcServer.GracefulStop()
}
