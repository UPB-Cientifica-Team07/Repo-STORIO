package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	gridpb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/generated"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/internal/auth"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/internal/database"
	gridgrpc "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/internal/grpc"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/internal/service"

	"google.golang.org/grpc"
)

func main() {

	bindAddress :=
		getEnv(
			"GRID_BIND_ADDRESS",
			"127.0.0.1:50056",
		)

	authService :=
		getEnv(
			"GRID_AUTH_SERVICE",
			"https://127.0.0.1:8081",
		)

	watchdogInterval :=
		getDurationEnv(
			"GRID_WATCHDOG_INTERVAL",
			5*time.Second,
		)

	resourceTimeout :=
		getDurationEnv(
			"GRID_RESOURCE_TIMEOUT",
			15*time.Second,
		)

	log.Println(
		"===================================",
	)

	log.Println(
		" GRID SERVICE",
	)

	log.Printf(
		" gRPC: %s",
		bindAddress,
	)

	log.Printf(
		" Auth: %s",
		authService,
	)

	log.Printf(
		" Watchdog interval: %s",
		watchdogInterval,
	)

	log.Printf(
		" Resource timeout: %s",
		resourceTimeout,
	)

	log.Println(
		" Estado: INICIANDO",
	)

	log.Println(
		"===================================",
	)

	db, err :=
		database.Open()

	if err != nil {

		log.Fatalf(
			"PostgreSQL Grid no disponible: %v",
			err,
		)
	}

	defer func() {

		if err :=
			db.Close(); err != nil {

			log.Printf(
				"error cerrando PostgreSQL Grid: %v",
				err,
			)
		}
	}()

	resourceRepository :=
		repository.NewResourceRepository(
			db,
		)

	gridService :=
		service.NewGridService(
			resourceRepository,
		)

	authClient :=
		auth.NewClient(
			authService,
		)

	listener, err :=
		net.Listen(
			"tcp",
			bindAddress,
		)

	if err != nil {

		log.Fatalf(
			"no se pudo escuchar en %s: %v",
			bindAddress,
			err,
		)
	}

	grpcServer :=
		grpc.NewServer(
			grpc.UnaryInterceptor(
				auth.UnaryServerInterceptor(
					authClient,
				),
			),
		)

	gridpb.RegisterGridServiceServer(
		grpcServer,
		gridgrpc.NewGridServer(
			gridService,
		),
	)

	watchdogContext,
		cancelWatchdog :=
		context.WithCancel(
			context.Background(),
		)

	defer cancelWatchdog()

	go service.RunWatchdog(
		watchdogContext,
		gridService,
		watchdogInterval,
		resourceTimeout,
	)

	go service.RunHpcResourceBridge(
		watchdogContext,
		resourceRepository,
		watchdogInterval,
		resourceTimeout,
	)

	serverErrors :=
		make(
			chan error,
			1,
		)

	go func() {

		log.Println(
			"===================================",
		)

		log.Println(
			" GRID SERVICE ACTIVO",
		)

		log.Printf(
			" Puerto: %s",
			bindAddress,
		)

		log.Println(
			" Estado: ACTIVE",
		)

		log.Println(
			"===================================",
		)

		serverErrors <- grpcServer.Serve(
			listener,
		)
	}()

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

	select {

	case receivedSignal :=
		<-signals:

		log.Printf(
			"señal recibida: %v",
			receivedSignal,
		)

		cancelWatchdog()

		done :=
			make(
				chan struct{},
			)

		go func() {

			grpcServer.GracefulStop()

			close(
				done,
			)
		}()

		select {

		case <-done:

			log.Println(
				"Grid Service detenido correctamente",
			)

		case <-time.After(
			5 * time.Second,
		):

			log.Println(
				"timeout de apagado; forzando stop",
			)

			grpcServer.Stop()
		}

	case err :=
		<-serverErrors:

		if err != nil {

			log.Fatalf(
				"Grid Service finalizó con error: %v",
				err,
			)
		}
	}
}

func getEnv(
	name string,
	fallback string,
) string {

	value :=
		strings.TrimSpace(
			os.Getenv(
				name,
			),
		)

	if value == "" {
		return fallback
	}

	return value
}

func getDurationEnv(
	name string,
	fallback time.Duration,
) time.Duration {

	raw :=
		strings.TrimSpace(
			os.Getenv(
				name,
			),
		)

	if raw == "" {
		return fallback
	}

	value, err :=
		time.ParseDuration(
			raw,
		)

	if err != nil ||
		value <= 0 {

		log.Fatalf(
			"%s inválido: %q",
			name,
			raw,
		)
	}

	return value
}
