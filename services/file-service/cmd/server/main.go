package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	grpcServer "google.golang.org/grpc"
	"google.golang.org/grpc/stats"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/auth"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/database"
	filegrpc "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/grpc"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/metrics"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/monitoring"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/service"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/streaming"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/proto"
)

const (
	fileServicePort = ":50053"
	componentID     = "file-service-01"
	componentName   = "File Service"
	metricsInterval = 5 * time.Second
	authURL         = "http://localhost:8081"
)

// =====================================
// TELEMETRÍA DEL SERVIDOR GRPC
// =====================================

type serverTelemetry struct {
	activeConnections atomic.Int64
	totalRequests     atomic.Int64
}

// =====================================
// STATS HANDLER
// =====================================

func (t *serverTelemetry) TagRPC(
	ctx context.Context,
	info *stats.RPCTagInfo,
) context.Context {
	return ctx
}

func (t *serverTelemetry) HandleRPC(
	ctx context.Context,
	rpcStats stats.RPCStats,
) {
}

func (t *serverTelemetry) TagConn(
	ctx context.Context,
	info *stats.ConnTagInfo,
) context.Context {
	return ctx
}

func (t *serverTelemetry) HandleConn(
	ctx context.Context,
	connStats stats.ConnStats,
) {
	switch connStats.(type) {
	case *stats.ConnBegin:
		t.activeConnections.Add(1)

	case *stats.ConnEnd:
		t.activeConnections.Add(-1)
	}
}

// =====================================
// MAIN
// =====================================

func main() {
	log.Println("===================================")
	log.Println(" FILE SERVICE")
	log.Println(" Tecnología: Go + gRPC")
	log.Println(" Puerto: 50053")
	log.Println(" Auth Service:", authURL)
	log.Println(" PostgreSQL: localhost:5434")
	log.Println(" Monitoring: localhost:50051")
	log.Println(" Estado: INICIANDO")
	log.Println("===================================")

	// =====================================
	// STORAGE
	// =====================================

	storageDir := os.Getenv(
		"FILE_STORAGE_DIR",
	)

	if storageDir == "" {
		storageDir =
			"services/file-service/data/shared-storage"
	}

	storageDir =
		filepath.Clean(
			storageDir,
		)

	log.Printf(
		"Shared Storage: %s",
		storageDir,
	)

	// =====================================
	// STREAMING SERVICE
	// =====================================

	streamingURL :=
		os.Getenv(
			"STREAMING_SERVICE_URL",
		)

	if streamingURL == "" {
		streamingURL =
			"http://127.0.0.1:50054"
	}

	log.Printf(
		"Streaming Service: %s",
		streamingURL,
	)

	// =====================================
	// POSTGRESQL
	// =====================================

	db,
		err :=
		database.Open()

	if err != nil {
		log.Fatalf(
			"No se pudo inicializar PostgreSQL: %v",
			err,
		)
	}

	defer db.Close()

	log.Println(
		"PostgreSQL: CONECTADO",
	)

	// =====================================
	// FILE REPOSITORY
	// =====================================

	fileRepository,
		err :=
		repository.NewFileRepository(
			storageDir,
		)

	if err != nil {
		log.Fatalf(
			"No se pudo inicializar File Repository: %v",
			err,
		)
	}

	log.Println(
		"File Repository: ACTIVO",
	)

	// =====================================
	// HOME REPOSITORY
	// =====================================

	homeRepository :=
		repository.NewHomeRepository(
			db,
		)

	log.Println(
		"Home Repository: ACTIVO",
	)

	// =====================================
	// FILE METADATA REPOSITORY
	// =====================================

	metadataRepository :=
		repository.NewFileMetadataRepository(
			db,
		)

	log.Println(
		"File Metadata Repository: ACTIVO",
	)

	// =====================================
	// PERMISSION REPOSITORY
	// =====================================

	permissionRepository :=
		repository.NewPermissionRepository(
			db,
		)

	log.Println(
		"Permission Repository: ACTIVO",
	)

	// =====================================
	// PERMISSION SERVICE
	// =====================================

	permissionService :=
		service.NewPermissionService(
			permissionRepository,
		)

	log.Println(
		"Permission Service: ACTIVO",
	)

	// =====================================
	// RECONCILIAR USO FÍSICO DE HOMES
	// =====================================

	reconcileCtx,
		reconcileCancel :=
		context.WithTimeout(
			context.Background(),
			10*time.Second,
		)

	err =
		reconcileHomeUsage(
			reconcileCtx,
			storageDir,
			homeRepository,
		)

	reconcileCancel()

	if err != nil {
		log.Fatalf(
			"No se pudo reconciliar uso de Homes: %v",
			err,
		)
	}

	log.Println(
		"Cuotas de Homes: RECONCILIADAS",
	)

	// =====================================
	// FILE SERVICE
	// =====================================

	fileService :=
		service.NewFileService(
			fileRepository,
			homeRepository,
			metadataRepository,
		)

	// =====================================
	// STREAMING CLIENT
	// =====================================

	streamingClient :=
		streaming.NewClient(
			streamingURL,
		)

	fileService.SetStreamingNotifier(
		func(
			fileID string,
		) error {

			result,
				err :=
				streamingClient.RegisterVideo(
					fileID,
				)

			if err != nil {
				log.Printf(
					"Streaming registerVideo falló para file=%s: %v",
					fileID,
					err,
				)

				return err
			}

			log.Printf(
				"Streaming registrado: file=%s video=%s duracion=%ds calidad=%s formato=%s",
				result.FileID,
				result.VideoID,
				result.DurationSeconds,
				result.Quality,
				result.Format,
			)

			return nil
		},
	)

	// =====================================
	// AUTH CLIENT
	// =====================================

	authClient :=
		auth.NewClient(
			authURL,
		)

	// =====================================
	// TELEMETRÍA
	// =====================================

	telemetry :=
		&serverTelemetry{}

	// =====================================
	// LISTENER
	// =====================================

	listener,
		err :=
		net.Listen(
			"tcp",
			fileServicePort,
		)

	if err != nil {
		log.Fatalf(
			"No se pudo iniciar File Service: %v",
			err,
		)
	}

	// =====================================
	// INTERCEPTOR CONTADOR DE RPC
	// =====================================

	requestCounterInterceptor :=
		func(
			ctx context.Context,
			req interface{},
			info *grpcServer.UnaryServerInfo,
			handler grpcServer.UnaryHandler,
		) (interface{}, error) {

			telemetry.totalRequests.Add(1)

			return handler(
				ctx,
				req,
			)
		}

	// =====================================
	// GRPC SERVER
	// =====================================

	grpcSrv :=
		grpcServer.NewServer(
			grpcServer.StatsHandler(
				telemetry,
			),

			grpcServer.ChainUnaryInterceptor(
				auth.UnaryServerInterceptor(
					authClient,
				),

				requestCounterInterceptor,
			),
		)

	fileGRPCServer :=
		filegrpc.NewServer(
			fileService,
			permissionService,
		)

	pb.RegisterFileServiceServer(
		grpcSrv,
		fileGRPCServer,
	)

	// =====================================
	// MONITORING CLIENT
	// =====================================

	monitoringClient,
		err :=
		monitoring.NewClient(
			componentID,
			componentName,
		)

	if err != nil {
		_ = listener.Close()

		log.Fatalf(
			"No se pudo crear cliente de Monitoring: %v",
			err,
		)
	}

	defer monitoringClient.Close()

	// =====================================
	// METRICS COLLECTOR
	// =====================================

	metricsCollector :=
		metrics.NewCollector(
			storageDir,
		)

	// =====================================
	// CANAL DE ERRORES DEL SERVIDOR
	// =====================================

	serverErrors :=
		make(
			chan error,
			1,
		)

	go func() {
		serverErrors <- grpcSrv.Serve(
			listener,
		)
	}()

	// =====================================
	// REPORTAR ACTIVE
	// =====================================

	if err :=
		monitoringClient.ReportStatus(
			"ACTIVE",
			"File Service iniciado correctamente",
		); err != nil {

		log.Printf(
			"No se pudo reportar ACTIVE: %v",
			err,
		)

	} else {
		log.Println(
			"Estado ACTIVE enviado a Monitoring",
		)
	}

	// =====================================
	// ESTADO
	// =====================================

	log.Println("===================================")
	log.Println(" FILE SERVICE ACTIVO")
	log.Println(" Puerto: 50053")
	log.Printf(" Shared Storage: %s", storageDir)
	log.Println(" Persistencia física: ACTIVA")
	log.Println(" Storage legacy: COMPATIBLE")
	log.Println(" Homes: ACTIVOS")
	log.Println(" Cuotas: RESERVA ATÓMICA ACTIVA")
	log.Println(" Cuotas: RECONCILIACIÓN ACTIVA")
	log.Println(" Metadata PostgreSQL: ACTIVA")
	log.Println(" Tabla archivo: INTEGRADA")
	log.Println(" ACL Repository: ACTIVO")
	log.Println(" ACL Service: ACTIVO")
	log.Println(" PostgreSQL: CONECTADO")
	log.Println(" Auth: ACTIVO")
	log.Println(" Seguridad gRPC: TOKEN OBLIGATORIO")
	log.Println(" Monitoring: localhost:50051")
	log.Println(" Métricas reales: ACTIVO")
	log.Println(" Storage Monitoring: files + homes")
	log.Println(" Intervalo métricas: 5s")
	log.Println("===================================")

	// =====================================
	// TICKER
	// =====================================

	ticker :=
		time.NewTicker(
			metricsInterval,
		)

	defer ticker.Stop()

	// =====================================
	// SIGNALS
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

	defer signal.Stop(
		signals,
	)

	// =====================================
	// LOOP PRINCIPAL
	// =====================================

	for {
		select {

		// =====================================
		// MÉTRICAS
		// =====================================

		case <-ticker.C:

			cpuUsage,
				memoryUsage,
				storageUsage,
				err :=
				metricsCollector.Collect()

			if err != nil {
				log.Printf(
					"Error recolectando métricas: %v",
					err,
				)

				continue
			}

			activeConnections :=
				telemetry.
					activeConnections.
					Load()

			totalRequests :=
				telemetry.
					totalRequests.
					Load()

			err =
				monitoringClient.ReportMetrics(
					cpuUsage,
					memoryUsage,
					storageUsage,
					activeConnections,
					totalRequests,
				)

			if err != nil {
				log.Printf(
					"Error enviando métricas a Monitoring: %v",
					err,
				)
			}

			log.Printf(
				"Métricas reales | CPU: %.2f%% | RAM: %.2f MB | Storage: %d bytes | Connections: %d | Requests: %d",
				cpuUsage,
				memoryUsage,
				storageUsage,
				activeConnections,
				totalRequests,
			)

		// =====================================
		// SEÑAL DEL SISTEMA
		// =====================================

		case receivedSignal :=
			<-signals:

			log.Printf(
				"Señal recibida: %v",
				receivedSignal,
			)

			shutdown(
				grpcSrv,
				listener,
				monitoringClient,
			)

			return

		// =====================================
		// ERROR DEL SERVIDOR
		// =====================================

		case err :=
			<-serverErrors:

			if err != nil {
				log.Printf(
					"Servidor gRPC finalizado: %v",
					err,
				)
			}

			return
		}
	}
}

// =====================================
// RECONCILIAR USO DE TODOS LOS HOMES
// =====================================

func reconcileHomeUsage(
	ctx context.Context,
	storageDir string,
	homeRepository *repository.HomeRepository,
) error {

	if homeRepository == nil {
		return fmt.Errorf(
			"Home Repository no inicializado",
		)
	}

	homes,
		err :=
		homeRepository.ListActive(
			ctx,
		)

	if err != nil {
		return fmt.Errorf(
			"no se pudieron listar Homes: %w",
			err,
		)
	}

	log.Printf(
		"Reconciliando %d Home(s)",
		len(homes),
	)

	for _, home := range homes {

		if home == nil {
			continue
		}

		if home.DirectoryID == "" {
			return fmt.Errorf(
				"Home encontrado sin directorio_id",
			)
		}

		if home.BasePath == "" {
			return fmt.Errorf(
				"Home %s no tiene ruta_base",
				home.DirectoryID,
			)
		}

		cleanBasePath :=
			filepath.Clean(
				home.BasePath,
			)

		if filepath.IsAbs(
			cleanBasePath,
		) {
			return fmt.Errorf(
				"ruta_base absoluta no permitida para %s: %s",
				home.DirectoryID,
				home.BasePath,
			)
		}

		homePath :=
			filepath.Join(
				storageDir,
				cleanBasePath,
			)

		relativeToStorage,
			err :=
			filepath.Rel(
				storageDir,
				homePath,
			)

		if err != nil {
			return fmt.Errorf(
				"no se pudo validar ruta Home de %s: %w",
				home.DirectoryID,
				err,
			)
		}

		if relativeToStorage == ".." ||
			strings.HasPrefix(
				relativeToStorage,
				".."+string(
					os.PathSeparator,
				),
			) ||
			filepath.IsAbs(
				relativeToStorage,
			) {

			return fmt.Errorf(
				"ruta Home fuera de shared-storage para %s: %s",
				home.DirectoryID,
				home.BasePath,
			)
		}

		if err :=
			os.MkdirAll(
				homePath,
				0750,
			); err != nil {

			return fmt.Errorf(
				"no se pudo crear Home físico de %s: %w",
				home.DirectoryID,
				err,
			)
		}

		usedBytes,
			err :=
			calculateDirectoryUsage(
				homePath,
			)

		if err != nil {
			return fmt.Errorf(
				"no se pudo calcular uso del Home %s: %w",
				home.DirectoryID,
				err,
			)
		}

		if err :=
			homeRepository.SetUsage(
				ctx,
				home.DirectoryID,
				usedBytes,
			); err != nil {

			return fmt.Errorf(
				"no se pudo reconciliar Home %s: %w",
				home.DirectoryID,
				err,
			)
		}

		log.Printf(
			"Home reconciliado | Usuario: %s | Ruta: %s | Uso: %d/%d bytes",
			home.DirectoryID,
			cleanBasePath,
			usedBytes,
			home.QuotaBytes,
		)
	}

	return nil
}

// =====================================
// CALCULAR BYTES DE UN HOME
// =====================================

func calculateDirectoryUsage(
	root string,
) (
	int64,
	error,
) {

	var totalBytes int64

	err :=
		filepath.WalkDir(
			root,
			func(
				path string,
				entry fs.DirEntry,
				walkErr error,
			) error {

				if walkErr != nil {
					return walkErr
				}

				if entry.IsDir() {
					return nil
				}

				info,
					err :=
					entry.Info()

				if err != nil {
					return err
				}

				if !info.Mode().
					IsRegular() {

					return nil
				}

				totalBytes +=
					info.Size()

				return nil
			},
		)

	if err != nil {
		return 0,
			err
	}

	return totalBytes,
		nil
}

// =====================================
// CIERRE ORDENADO
// =====================================

func shutdown(
	grpcSrv *grpcServer.Server,
	listener net.Listener,
	monitoringClient *monitoring.Client,
) {

	log.Println("===================================")
	log.Println(" DETENIENDO FILE SERVICE")
	log.Println("===================================")

	err :=
		monitoringClient.ReportStatus(
			"INACTIVE",
			"File Service detenido correctamente",
		)

	if err != nil {
		log.Printf(
			"No se pudo reportar INACTIVE: %v",
			err,
		)

	} else {
		log.Println(
			"Estado INACTIVE enviado correctamente",
		)
	}

	grpcSrv.GracefulStop()

	if err :=
		listener.Close(); err != nil {

		log.Printf(
			"Listener cerrado/finalizado: %v",
			err,
		)
	}

	log.Println(
		"File Service detenido correctamente",
	)
}
