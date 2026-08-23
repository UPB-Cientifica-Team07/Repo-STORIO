package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcServer "google.golang.org/grpc"

	filegrpc "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/grpc"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/monitoring"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/service"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"
)

func main() {

	// =====================================
	// CREAR REPOSITORIO Y SERVICIO
	// =====================================

	fileRepository := repository.NewFileRepository()

	fileService := service.NewFileService(
		fileRepository,
	)

	// =====================================
	// CREAR LISTENER gRPC
	// =====================================

	listener, err := net.Listen(
		"tcp",
		":50053",
	)

	if err != nil {
		log.Fatalf(
			"No se pudo iniciar File Service: %v",
			err,
		)
	}

	// =====================================
	// CREAR SERVIDOR gRPC
	// =====================================

	grpcSrv := grpcServer.NewServer()

	fileGRPCServer := filegrpc.NewServer(
		fileService,
	)

	pb.RegisterFileServiceServer(
		grpcSrv,
		fileGRPCServer,
	)

	// =====================================
	// INICIAR SERVIDOR gRPC
	// =====================================

	go func() {

		if err := grpcSrv.Serve(listener); err != nil {
			log.Printf(
				"Error ejecutando servidor gRPC: %v",
				err,
			)
		}
	}()

	// =====================================
	// CREAR CLIENTE DE MONITOREO
	// =====================================

	monitoringClient, err := monitoring.NewClient(
		"file-service-01",
		"File Service",
	)

	if err != nil {
		log.Fatalf(
			"No se pudo crear el cliente de Monitoring Service: %v",
			err,
		)
	}

	defer monitoringClient.Close()

	log.Println("===================================")
	log.Println(" FILE SERVICE")
	log.Println(" Protocolo: gRPC")
	log.Println(" Puerto: 50053")
	log.Println(" Estado: INICIANDO")
	log.Println("===================================")

	// =====================================
	// REPORTAR ESTADO INICIAL
	// =====================================

	err = monitoringClient.ReportStatus(
		"ACTIVE",
		"File Service iniciado correctamente",
	)

	if err != nil {
		log.Printf(
			"No se pudo reportar el estado al Monitoring Service: %v",
			err,
		)
	} else {
		log.Println(
			"Estado enviado correctamente al Monitoring Service",
		)
	}

	log.Println("===================================")
	log.Println(" FILE SERVICE ACTIVO")
	log.Println(" Escuchando en :50053")
	log.Println("===================================")

	// =====================================
	// MÉTRICAS INICIALES
	// =====================================

	var cpuUsage float64 = 25.0
	var memoryUsage float64 = 40.0
	var activeConnections int64 = 1
	var totalRequests int64 = 0

	// =====================================
	// ENVIAR MÉTRICAS CADA 5 SEGUNDOS
	// =====================================

	ticker := time.NewTicker(
		5 * time.Second,
	)

	defer ticker.Stop()

	// =====================================
	// DETECTAR CTRL + C
	// =====================================

	signals := make(
		chan os.Signal,
		1,
	)

	signal.Notify(
		signals,
		os.Interrupt,
		syscall.SIGTERM,
	)

	for {
		select {

		case <-ticker.C:

			cpuUsage += 1.5
			memoryUsage += 0.8
			totalRequests += 10

			if cpuUsage > 90 {
				cpuUsage = 25.0
			}

			if memoryUsage > 85 {
				memoryUsage = 40.0
			}

			err := monitoringClient.ReportMetrics(
				cpuUsage,
				memoryUsage,
				activeConnections,
				totalRequests,
			)

			if err != nil {
				log.Printf(
					"Error enviando métricas: %v",
					err,
				)
			}

		case <-signals:

			log.Println("===================================")
			log.Println(" DETENIENDO FILE SERVICE")
			log.Println("===================================")

			err := monitoringClient.ReportStatus(
				"INACTIVE",
				"File Service detenido correctamente",
			)

			if err != nil {
				log.Printf(
					"No se pudo reportar el estado INACTIVE: %v",
					err,
				)
			} else {
				log.Println(
					"Estado INACTIVE enviado correctamente",
				)
			}

			grpcSrv.GracefulStop()

			if err := listener.Close(); err != nil {
				log.Printf(
					"Error cerrando listener: %v",
					err,
				)
			}

			return
		}
	}
}
