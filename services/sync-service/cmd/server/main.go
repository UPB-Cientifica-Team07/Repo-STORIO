package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"
	syncgrpc "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/grpc"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/monitoring"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/service"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/storage"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/proto"

	"google.golang.org/grpc"
)

const (
	port              = ":50055"
	storagePath       = "data/shared-storage"
	authURL           = "http://localhost:8081"
	monitoringAddress = "localhost:50051"

	componentID   = "sync-service"
	componentName = "Sync Service"
)

func main() {

	log.Println("===================================")
	log.Println(" SYNC SERVICE")
	log.Println(" Tecnología: Go + gRPC")
	log.Println(" Puerto:", port)
	log.Println(" Auth Service:", authURL)
	log.Println(" Monitoring:", monitoringAddress)
	log.Println(" Estado: INICIANDO")
	log.Println("===================================")

	// =====================================
	// REPOSITORY
	// =====================================

	repo :=
		repository.NewSyncRepository()

	// =====================================
	// DOMAIN SERVICE
	// =====================================

	syncService :=
		service.NewSyncService(
			repo,
		)

	// =====================================
	// STORAGE
	// =====================================

	fileStorage, err :=
		storage.NewStorage(
			storagePath,
		)

	if err != nil {
		log.Fatalf(
			"Error inicializando almacenamiento: %v",
			err,
		)
	}

	// =====================================
	// AUTH
	// =====================================

	authClient :=
		auth.NewClient(
			authURL,
		)

	// =====================================
	// MONITORING CLIENT
	// =====================================

	monitoringClient, err :=
		monitoring.NewClient(
			monitoringAddress,
		)

	if err != nil {

		log.Printf(
			"Advertencia: Monitoring Service no disponible: %v",
			err,
		)

		monitoringClient = nil
	}

	if monitoringClient != nil {

		defer monitoringClient.Close()

		err =
			monitoringClient.ReportStatus(
				componentID,
				componentName,
				"ACTIVE",
				"Sync Service iniciado correctamente",
			)

		if err != nil {
			log.Printf(
				"Advertencia reportando ACTIVE: %v",
				err,
			)
		}
	}

	// =====================================
	// RUNTIME METRICS
	// requests y conexiones reales
	// =====================================

	runtimeMetrics :=
		monitoring.NewRuntimeMetrics()

	connectionStats :=
		monitoring.NewConnectionStatsHandler(
			runtimeMetrics,
		)

	// =====================================
	// SYSTEM METRICS
	// CPU y RAM reales
	// =====================================

	systemMetrics :=
		monitoring.NewSystemMetrics()

	// =====================================
	// SYNC GRPC IMPLEMENTATION
	// =====================================

	grpcServerImpl :=
		syncgrpc.NewServer(
			syncService,
			fileStorage,
			authClient,
		)

	// =====================================
	// LISTENER
	// =====================================

	listener, err :=
		net.Listen(
			"tcp",
			port,
		)

	if err != nil {
		log.Fatalf(
			"Error abriendo puerto %s: %v",
			port,
			err,
		)
	}

	// =====================================
	// GRPC SERVER
	// =====================================

	grpcServer :=
		grpc.NewServer(

			grpc.UnaryInterceptor(
				runtimeMetrics.UnaryInterceptor,
			),

			grpc.StreamInterceptor(
				runtimeMetrics.StreamInterceptor,
			),

			grpc.StatsHandler(
				connectionStats,
			),
		)

	pb.RegisterSyncServiceServer(
		grpcServer,
		grpcServerImpl,
	)

	// =====================================
	// MONITORING LOOP
	// =====================================

	stopMetrics :=
		make(chan struct{})

	if monitoringClient != nil {

		go func() {

			ticker :=
				time.NewTicker(
					5 * time.Second,
				)

			defer ticker.Stop()

			for {

				select {

				case <-ticker.C:

					// =====================================
					// CPU REAL
					// =====================================

					cpuUsage, cpuErr :=
						systemMetrics.CPUUsagePercent()

					if cpuErr != nil {

						log.Printf(
							"Error obteniendo CPU real: %v",
							cpuErr,
						)

						cpuUsage = 0
					}

					// =====================================
					// RAM REAL
					// =====================================

					memoryUsage :=
						systemMetrics.MemoryUsageMB()

					// =====================================
					// STORAGE REAL
					// =====================================

					storageBytes, storageErr :=
						monitoring.StorageUsageBytes(
							storagePath,
						)

					if storageErr != nil {

						log.Printf(
							"Error obteniendo almacenamiento: %v",
							storageErr,
						)

						storageBytes = 0
					}

					// =====================================
					// REQUESTS Y CONEXIONES REALES
					// =====================================

					activeConnections :=
						runtimeMetrics.ActiveConnections()

					totalRequests :=
						runtimeMetrics.TotalRequests()

					// =====================================
					// LOG LOCAL
					// =====================================

					log.Printf(
						"Métricas reales | CPU: %.2f%% | RAM: %.2f MB | Storage: %d bytes | Connections: %d | Requests: %d",
						cpuUsage,
						memoryUsage,
						storageBytes,
						activeConnections,
						totalRequests,
					)

					// =====================================
					// ENVIAR A MONITORING SERVICE
					// =====================================

					reportErr :=
						monitoringClient.ReportMetrics(
							componentID,
							componentName,
							cpuUsage,
							memoryUsage,
							storageBytes,
							activeConnections,
							totalRequests,
						)

					if reportErr != nil {

						log.Printf(
							"Error reportando métricas: %v",
							reportErr,
						)
					}

				case <-stopMetrics:

					return
				}
			}
		}()
	}

	// =====================================
	// START SERVER
	// =====================================

	go func() {

		log.Println("===================================")
		log.Println(" SYNC SERVICE ACTIVO")
		log.Println(" Puerto:", port)
		log.Println(" Storage:", storagePath)
		log.Println(" Auth Bridge:", authURL)
		log.Println(" Monitoring:", monitoringAddress)
		log.Println("===================================")

		err :=
			grpcServer.Serve(
				listener,
			)

		if err != nil {

			log.Fatalf(
				"Error ejecutando servidor gRPC: %v",
				err,
			)
		}
	}()

	// =====================================
	// GRACEFUL SHUTDOWN
	// =====================================

	stop :=
		make(
			chan os.Signal,
			1,
		)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Println(
		"Deteniendo Sync Service...",
	)

	close(
		stopMetrics,
	)

	// =====================================
	// REPORTAR INACTIVE
	// =====================================

	if monitoringClient != nil {

		err =
			monitoringClient.ReportStatus(
				componentID,
				componentName,
				"INACTIVE",
				"Sync Service detenido",
			)

		if err != nil {

			log.Printf(
				"Error reportando INACTIVE: %v",
				err,
			)
		}
	}

	// =====================================
	// DETENER GRPC
	// =====================================

	grpcServer.GracefulStop()

	log.Println(
		"Sync Service detenido correctamente",
	)
}
