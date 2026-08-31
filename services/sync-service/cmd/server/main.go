package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/auth"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/database"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/fileclient"
	syncgrpc "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/grpc"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/monitoring"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/internal/service"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/sync-service/proto"

	"google.golang.org/grpc"
)

const (
	port = ":50055"

	authURL = "http://localhost:8081"

	fileServiceAddress = "localhost:50053"

	monitoringAddress = "localhost:50051"

	componentID = "sync-service"

	componentName = "Sync Service"
)

func main() {

	log.Println("===================================")
	log.Println(" SYNC SERVICE")
	log.Println(" Tecnología: Go + gRPC")
	log.Println(" Puerto:", port)
	log.Println(" Auth Service:", authURL)
	log.Println(" File Service:", fileServiceAddress)
	log.Println(" PostgreSQL: localhost:5434")
	log.Println(" Monitoring:", monitoringAddress)
	log.Println(" Estado: INICIANDO")
	log.Println("===================================")

	// =====================================
	// POSTGRESQL
	// =====================================
	//
	// Sync Service mantiene en PostgreSQL:
	//
	// - metadata lógica de archivos sincronizados
	// - relative_path
	// - versiones
	// - historial de cambios
	// - cursores por dispositivo
	//
	// El contenido físico sigue perteneciendo
	// exclusivamente a File Service.
	//
	// =====================================

	db, err :=
		database.Open()

	if err != nil {

		log.Fatalf(
			"No se pudo inicializar PostgreSQL: %v",
			err,
		)
	}

	defer func() {

		if closeErr :=
			db.Close(); closeErr != nil {

			log.Printf(
				"Advertencia cerrando PostgreSQL: %v",
				closeErr,
			)
		}
	}()

	log.Println(
		"PostgreSQL Sync: CONECTADO",
	)

	// =====================================
	// SYNC REPOSITORY PERSISTENTE
	// =====================================
	//
	// PostgreSQL reemplaza los antiguos maps
	// en memoria.
	//
	// Tablas utilizadas:
	//
	// - sync_file
	// - sync_change
	// - sync_device_cursor
	//
	// =====================================

	repo :=
		repository.NewSyncRepository(
			db,
		)

	log.Println(
		"Sync Repository PostgreSQL: ACTIVO",
	)

	// =====================================
	// DOMAIN SERVICE
	// =====================================

	syncService :=
		service.NewSyncService(
			repo,
		)

	log.Println(
		"Sync Domain Service: ACTIVO",
	)

	// =====================================
	// AUTH CLIENT
	// =====================================

	authClient :=
		auth.NewClient(
			authURL,
		)

	// =====================================
	// FILE SERVICE CLIENT
	// =====================================
	//
	// File Service continúa siendo propietario
	// exclusivo de:
	//
	// - almacenamiento físico
	// - Home
	// - cuotas
	// - metadata física PostgreSQL
	// - permisos Unix
	// - ACL
	//
	// Sync Service coordina:
	//
	// - árbol lógico
	// - historial de cambios
	// - versiones
	// - cursores
	//
	// =====================================

	fileServiceClient, err :=
		fileclient.NewClient(
			fileServiceAddress,
		)

	if err != nil {

		log.Fatalf(
			"Error inicializando File Service Client: %v",
			err,
		)
	}

	defer func() {

		if closeErr :=
			fileServiceClient.Close(); closeErr != nil {

			log.Printf(
				"Advertencia cerrando File Service Client: %v",
				closeErr,
			)
		}
	}()

	log.Println(
		"File Service Client: ACTIVO",
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

		monitoringClient =
			nil
	}

	if monitoringClient != nil {

		defer func() {

			if closeErr :=
				monitoringClient.Close(); closeErr != nil {

				log.Printf(
					"Advertencia cerrando Monitoring Client: %v",
					closeErr,
				)
			}
		}()

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
	// =====================================

	runtimeMetrics :=
		monitoring.NewRuntimeMetrics()

	connectionStats :=
		monitoring.NewConnectionStatsHandler(
			runtimeMetrics,
		)

	// =====================================
	// SYSTEM METRICS
	// =====================================

	systemMetrics :=
		monitoring.NewSystemMetrics()

	// =====================================
	// SYNC GRPC IMPLEMENTATION
	// =====================================
	//
	// El servidor gRPC recibe:
	//
	// - lógica de sincronización
	// - cliente hacia File Service
	// - cliente hacia Auth Service
	//
	// El repository PostgreSQL se utiliza
	// desde SyncService.
	//
	// =====================================

	grpcServerImpl :=
		syncgrpc.NewServer(
			syncService,
			fileServiceClient,
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

	defer func() {

		if closeErr :=
			listener.Close(); closeErr != nil {

			log.Printf(
				"Advertencia cerrando listener: %v",
				closeErr,
			)
		}
	}()

	// =====================================
	// GRPC SERVER
	// =====================================

	grpcServer :=
		grpc.NewServer(

			grpc.ChainUnaryInterceptor(
				runtimeMetrics.UnaryInterceptor,
				auth.UnaryServerInterceptor(
					authClient,
				),
			),

			grpc.ChainStreamInterceptor(
				runtimeMetrics.StreamInterceptor,
				auth.StreamServerInterceptor(
					authClient,
				),
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
		make(
			chan struct{},
		)

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

						cpuUsage =
							0
					}

					// =====================================
					// RAM REAL
					// =====================================

					memoryUsage :=
						systemMetrics.MemoryUsageMB()

					// =====================================
					// STORAGE
					// =====================================
					//
					// Sync Service no almacena contenido
					// físico.
					//
					// PostgreSQL contiene metadata de Sync,
					// pero no se contabiliza aquí como Home.
					//
					// =====================================

					var storageBytes int64 = 0

					// =====================================
					// REQUESTS / CONNECTIONS
					// =====================================

					activeConnections :=
						runtimeMetrics.ActiveConnections()

					totalRequests :=
						runtimeMetrics.TotalRequests()

					// =====================================
					// LOG LOCAL
					// =====================================

					log.Printf(
						"Métricas reales | CPU: %.2f%% | RAM: %.2f MB | Storage propio: %d bytes | Connections: %d | Requests: %d",
						cpuUsage,
						memoryUsage,
						storageBytes,
						activeConnections,
						totalRequests,
					)

					// =====================================
					// REPORTAR A MONITORING
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
		log.Println(" File Service:", fileServiceAddress)
		log.Println(" Auth Bridge:", authURL)
		log.Println(" PostgreSQL Sync: ACTIVO")
		log.Println(" Monitoring:", monitoringAddress)
		log.Println(" Storage local: DESHABILITADO")
		log.Println(" Home central: gestionado por File Service")
		log.Println(" Metadata Sync: PostgreSQL")
		log.Println(" Historial Sync: PERSISTENTE")
		log.Println(" Cursores dispositivo: PERSISTENTES")
		log.Println("===================================")

		serveErr :=
			grpcServer.Serve(
				listener,
			)

		if serveErr != nil {

			log.Fatalf(
				"Error ejecutando servidor gRPC: %v",
				serveErr,
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

	// Evita seguir recibiendo nuevas señales
	// sobre el canal durante el cierre.

	signal.Stop(
		stop,
	)

	// =====================================
	// DETENER LOOP DE MÉTRICAS
	// =====================================

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
		"Servidor gRPC detenido",
	)

	log.Println(
		"Sync Service detenido correctamente",
	)
}
