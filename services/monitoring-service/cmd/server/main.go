package main

import (
	"log"
	"net"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/generated"
	monitoringgrpc "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/grpc"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/repository"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/service"

	"google.golang.org/grpc"
)

const port = ":50051"

func main() {

	// Capa de persistencia temporal en memoria.
	monitoringRepository :=
		repository.NewInMemoryMonitoringRepository()

	// Capa de lógica de negocio.
	monitoringService :=
		service.NewMonitoringService(monitoringRepository)

	// Servidor gRPC.
	monitoringServer :=
		monitoringgrpc.NewMonitoringServer(monitoringService)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("No se pudo iniciar el servidor: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterMonitoringServiceServer(
		grpcServer,
		monitoringServer,
	)

	log.Println("===================================")
	log.Println(" MONITORING SERVICE")
	log.Println(" Protocolo: gRPC")
	log.Println(" Puerto: 50051")
	log.Println(" Estado: ACTIVO")
	log.Println("===================================")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error en el servidor gRPC: %v", err)
	}
}
