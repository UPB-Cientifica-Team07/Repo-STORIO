package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/file-service/internal/monitoring"
)

func main() {

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

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// =====================================
	// DETECTAR CTRL + C
	// =====================================

	signals := make(chan os.Signal, 1)

	signal.Notify(
		signals,
		os.Interrupt,
		syscall.SIGTERM,
	)

	for {
		select {

		case <-ticker.C:

			// Simulación temporal de métricas.
			cpuUsage += 1.5
			memoryUsage += 0.8
			totalRequests += 10

			// Evitar valores irreales.
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

			return
		}
	}
}
