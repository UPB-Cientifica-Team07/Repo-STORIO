package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	authgrpc "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/auth-services/internal/grpc"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/auth-services/internal/monitoring"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/auth-services/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/auth-services/internal/service"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/auth-services/proto"

	"google.golang.org/grpc"
)

const authServicePort = ":50052"

func main() {

	monitoringClient, err := monitoring.NewClient(
		"auth-service-01",
		"Auth Service",
	)

	if err != nil {
		log.Fatalf(
			"No se pudo crear el cliente de Monitoring Service: %v",
			err,
		)
	}

	defer monitoringClient.Close()

	userRepository := repository.NewUserRepository()

	authService := service.NewAuthService(
		userRepository,
	)

	authServer := authgrpc.NewServer(
		authService,
	)

	listener, err := net.Listen(
		"tcp",
		authServicePort,
	)

	if err != nil {
		log.Fatalf(
			"No se pudo iniciar Auth Service: %v",
			err,
		)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAuthServiceServer(
		grpcServer,
		authServer,
	)

	log.Println("===================================")
	log.Println(" AUTH SERVICE")
	log.Println(" Protocolo: gRPC")
	log.Println(" Puerto: 50052")
	log.Println(" Estado: INICIANDO")
	log.Println("===================================")

	err = monitoringClient.ReportStatus(
		"ACTIVE",
		"Auth Service iniciado correctamente",
	)

	if err != nil {
		log.Printf(
			"No se pudo reportar el estado al Monitoring Service: %v",
			err,
		)
	} else {
		log.Println(
			"Estado ACTIVE enviado correctamente",
		)
	}

	go reportMetrics(
		monitoringClient,
	)

	go func() {

		log.Println("===================================")
		log.Println(" AUTH SERVICE ACTIVO")
		log.Println(" Escuchando en :50052")
		log.Println("===================================")

		if err := grpcServer.Serve(listener); err != nil {
			log.Printf(
				"Error en servidor gRPC: %v",
				err,
			)
		}
	}()

	signals := make(
		chan os.Signal,
		1,
	)

	signal.Notify(
		signals,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-signals

	log.Println("===================================")
	log.Println(" DETENIENDO AUTH SERVICE")
	log.Println("===================================")

	grpcServer.GracefulStop()

	err = monitoringClient.ReportStatus(
		"INACTIVE",
		"Auth Service detenido correctamente",
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
}

func reportMetrics(
	monitoringClient *monitoring.Client,
) {

	ticker := time.NewTicker(
		5 * time.Second,
	)

	defer ticker.Stop()

	var (
		cpuUsage          float64 = 15.0
		memoryUsage       float64 = 25.0
		activeConnections int64   = 0
		totalRequests     int64   = 0
	)

	for range ticker.C {

		cpuUsage += 1.2
		memoryUsage += 0.7
		activeConnections = 1
		totalRequests += 5

		err := monitoringClient.ReportMetrics(
			cpuUsage,
			memoryUsage,
			activeConnections,
			totalRequests,
		)

		if err != nil {
			log.Printf(
				"Error reportando métricas: %v",
				err,
			)

			continue
		}

		log.Printf(
			"Métricas reportadas | CPU: %.2f%% | Memoria: %.2f%% | Conexiones: %d | Solicitudes: %d",
			cpuUsage,
			memoryUsage,
			activeConnections,
			totalRequests,
		)
	}
}
