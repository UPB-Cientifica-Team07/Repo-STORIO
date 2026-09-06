package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/generated"
	monitoringgrpc "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/grpc"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/service"

	"google.golang.org/grpc"
)

const (
	port = ":50051"

	watchdogInterval = 5 * time.Second
)

func main() {

	log.Println("===================================")
	log.Println(" MONITORING SERVICE")
	log.Println(" Protocolo: gRPC")
	log.Println(" Puerto: 50051")
	log.Println(" Persistencia: JSON en disco")
	log.Println(" Watchdog: 5 segundos")
	log.Println(" Timeout servicios: 15 segundos")
	log.Println(" Estado: INICIANDO")
	log.Println("===================================")

	// =====================================
	// REPOSITORY
	// =====================================

	monitoringRepository :=
		repository.NewMonitoringRepository()

	// =====================================
	// HPC POSTGRESQL REPOSITORY
	// =====================================

	hpcRepository, err :=
		repository.NewHpcRepository()

	if err != nil {

		log.Fatalf(
			"No fue posible conectar Monitoring con PostgreSQL HPC: %v",
			err,
		)
	}

	defer func() {

		if err := hpcRepository.Close(); err != nil {

			log.Printf(
				"Error cerrando PostgreSQL HPC: %v",
				err,
			)
		}
	}()

	// =====================================
	// SERVICE
	// =====================================

	monitoringService :=
		service.NewMonitoringService(
			monitoringRepository,
			hpcRepository,
		)

	// =====================================
	// GRPC IMPLEMENTATION
	// =====================================

	monitoringServer :=
		monitoringgrpc.NewMonitoringServer(
			monitoringService,
		)

	// =====================================
	// TCP LISTENER
	// =====================================

	listener, err :=
		net.Listen(
			"tcp",
			port,
		)

	if err != nil {

		log.Fatalf(
			"No se pudo iniciar Monitoring Service en %s: %v",
			port,
			err,
		)
	}

	// =====================================
	// GRPC SERVER
	// =====================================

	grpcServer :=
		grpc.NewServer()

	pb.RegisterMonitoringServiceServer(
		grpcServer,
		monitoringServer,
	)

	// =====================================
	// CONTEXTO DEL WATCHDOG
	// =====================================

	watchdogContext, cancelWatchdog :=
		context.WithCancel(
			context.Background(),
		)

	defer cancelWatchdog()

	// =====================================
	// INICIAR WATCHDOG
	// =====================================

	go runAvailabilityWatchdog(
		watchdogContext,
		monitoringService,
	)

	// =====================================
	// SERVIDOR GRPC
	// =====================================

	serverErrors :=
		make(
			chan error,
			1,
		)

	go func() {

		log.Println("===================================")
		log.Println(" MONITORING SERVICE ACTIVO")
		log.Println(" Protocolo: gRPC")
		log.Println(" Puerto: 50051")
		log.Println(" Persistencia: data/monitoring")
		log.Println(" Watchdog: ACTIVO")
		log.Println(" Intervalo watchdog: 5s")
		log.Println(" Timeout heartbeat: 15s")
		log.Println(" Estado: ACTIVO")
		log.Println("===================================")

		serverErrors <- grpcServer.Serve(
			listener,
		)
	}()

	// =====================================
	// SEÑALES DEL SISTEMA
	// =====================================

	signals :=
		make(
			chan os.Signal,
			1,
		)

	signal.Notify(
		signals,
		os.Interrupt,
		syscall.SIGTERM,
	)

	// =====================================
	// ESPERAR CIERRE O ERROR
	// =====================================

	select {

	case receivedSignal :=
		<-signals:

		log.Printf(
			"Señal recibida: %v",
			receivedSignal,
		)

	case err :=
		<-serverErrors:

		if err != nil {

			log.Printf(
				"Servidor gRPC finalizado con error: %v",
				err,
			)
		}
	}

	// =====================================
	// DETENER WATCHDOG
	// =====================================

	cancelWatchdog()

	// =====================================
	// DETENER GRPC
	// =====================================

	log.Println(
		"Deteniendo Monitoring Service...",
	)

	grpcServer.GracefulStop()

	if err := listener.Close(); err != nil {

		log.Printf(
			"Error cerrando listener: %v",
			err,
		)
	}

	log.Println(
		"Monitoring Service detenido correctamente",
	)
}

// =====================================
// WATCHDOG DE DISPONIBILIDAD
// =====================================

func runAvailabilityWatchdog(
	ctx context.Context,
	monitoringService *service.MonitoringService,
) {

	ticker :=
		time.NewTicker(
			watchdogInterval,
		)

	defer ticker.Stop()

	log.Println(
		"Watchdog de disponibilidad iniciado",
	)

	// Primera revisión al iniciar.
	if err :=
		monitoringService.CheckAllServiceAvailability(); err != nil {

		log.Printf(
			"Error en revisión inicial del watchdog: %v",
			err,
		)
	}

	for {

		select {

		case <-ctx.Done():

			log.Println(
				"Watchdog de disponibilidad detenido",
			)

			return

		case <-ticker.C:

			if err :=
				monitoringService.CheckAllServiceAvailability(); err != nil {

				log.Printf(
					"Error ejecutando watchdog: %v",
					err,
				)
			}
		}
	}
}
