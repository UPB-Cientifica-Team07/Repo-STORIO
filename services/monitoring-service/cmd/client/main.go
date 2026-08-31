package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/generated"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const serverAddress = "localhost:50051"

func main() {

	fmt.Println("===================================")
	fmt.Println(" CLIENTE MONITORING SERVICE")
	fmt.Println("===================================")

	conn, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {

		log.Fatalf(
			"Error creando conexión gRPC: %v",
			err,
		)
	}

	defer conn.Close()

	client :=
		pb.NewMonitoringServiceClient(
			conn,
		)

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			20*time.Second,
		)

	defer cancel()

	// =====================================
	// 1. CONSULTAR MÉTRICAS
	// =====================================

	consultMetrics(
		ctx,
		client,
	)

	// =====================================
	// 2. CONSULTAR ESTADOS
	// =====================================

	consultServiceStatus(
		ctx,
		client,
		"File Service",
	)

	consultServiceStatus(
		ctx,
		client,
		"Sync Service",
	)

	// =====================================
	// 3. CONSULTAR ALERTAS
	// =====================================

	consultAlerts(
		ctx,
		client,
	)

	// =====================================
	// 4. CONSULTAR REGLAS
	// =====================================

	consultAlertRules(
		ctx,
		client,
	)

	fmt.Println()
	fmt.Println("===================================")
	fmt.Println(" CONSULTA FINALIZADA")
	fmt.Println("===================================")
}

// =====================================
// CONSULTAR MÉTRICAS
// =====================================

func consultMetrics(
	ctx context.Context,
	client pb.MonitoringServiceClient,
) {

	fmt.Println()
	fmt.Println("===================================")
	fmt.Println(" CONSULTANDO MÉTRICAS")
	fmt.Println("===================================")

	response, err :=
		client.GetMetrics(
			ctx,
			&pb.GetMetricsRequest{},
		)

	if err != nil {

		log.Printf(
			"Error consultando métricas: %v",
			err,
		)

		return
	}

	metrics :=
		response.GetMetrics()

	fmt.Printf(
		"Total de métricas registradas: %d\n",
		len(metrics),
	)

	start :=
		0

	if len(metrics) > 10 {

		start =
			len(metrics) - 10
	}

	for index := start; index < len(metrics); index++ {

		metric :=
			metrics[index]

		fmt.Println(
			"-----------------------------------",
		)

		fmt.Printf(
			"Métrica #%d\n",
			index+1,
		)

		fmt.Printf(
			"Componente: %s\n",
			metric.GetComponentName(),
		)

		fmt.Printf(
			"ID: %s\n",
			metric.GetComponentId(),
		)

		fmt.Printf(
			"CPU: %.2f%%\n",
			metric.GetCpuUsage(),
		)

		fmt.Printf(
			"Memoria: %.2f MB\n",
			metric.GetMemoryUsage(),
		)

		fmt.Printf(
			"Almacenamiento: %d bytes\n",
			metric.GetStorageUsage(),
		)

		fmt.Printf(
			"Almacenamiento: %.2f KB\n",
			bytesToKB(
				metric.GetStorageUsage(),
			),
		)

		fmt.Printf(
			"Almacenamiento: %.2f MB\n",
			bytesToMB(
				metric.GetStorageUsage(),
			),
		)

		fmt.Printf(
			"Conexiones activas: %d\n",
			metric.GetActiveConnections(),
		)

		fmt.Printf(
			"Total solicitudes: %d\n",
			metric.GetTotalRequests(),
		)

		fmt.Printf(
			"Timestamp: %s\n",
			formatTimestamp(
				metric.GetTimestamp(),
			),
		)
	}

	if len(metrics) == 0 {

		return
	}

	last :=
		metrics[len(metrics)-1]

	fmt.Println()
	fmt.Println("===================================")
	fmt.Println(" ÚLTIMA MÉTRICA REGISTRADA")
	fmt.Println("===================================")

	fmt.Printf(
		"Componente: %s\n",
		last.GetComponentName(),
	)

	fmt.Printf(
		"ID: %s\n",
		last.GetComponentId(),
	)

	fmt.Printf(
		"CPU: %.2f%%\n",
		last.GetCpuUsage(),
	)

	fmt.Printf(
		"Memoria: %.2f MB\n",
		last.GetMemoryUsage(),
	)

	fmt.Printf(
		"Almacenamiento: %d bytes\n",
		last.GetStorageUsage(),
	)

	fmt.Printf(
		"Conexiones activas: %d\n",
		last.GetActiveConnections(),
	)

	fmt.Printf(
		"Total solicitudes: %d\n",
		last.GetTotalRequests(),
	)

	fmt.Printf(
		"Timestamp: %s\n",
		formatTimestamp(
			last.GetTimestamp(),
		),
	)
}

// =====================================
// CONSULTAR ESTADO
// =====================================

func consultServiceStatus(
	ctx context.Context,
	client pb.MonitoringServiceClient,
	componentName string,
) {

	fmt.Println()
	fmt.Println("===================================")

	fmt.Printf(
		" CONSULTANDO ESTADO: %s\n",
		componentName,
	)

	fmt.Println("===================================")

	response, err :=
		client.GetServiceStatus(
			ctx,
			&pb.ServiceRequest{
				ComponentName: componentName,
			},
		)

	if err != nil {

		log.Printf(
			"Error consultando estado: %v",
			err,
		)

		return
	}

	fmt.Println(
		"-----------------------------------",
	)

	fmt.Printf(
		"Componente: %s\n",
		response.GetComponentName(),
	)

	if response.GetComponentId() == "" {

		fmt.Println(
			"ID: no disponible",
		)

	} else {

		fmt.Printf(
			"ID: %s\n",
			response.GetComponentId(),
		)
	}

	fmt.Printf(
		"Estado: %s\n",
		response.GetStatus(),
	)

	fmt.Printf(
		"Mensaje: %s\n",
		response.GetMessage(),
	)

	if response.GetLastUpdated() == 0 {

		fmt.Println(
			"Última actualización: no disponible",
		)

	} else {

		fmt.Printf(
			"Última actualización: %s\n",
			formatTimestamp(
				response.GetLastUpdated(),
			),
		)
	}
}

// =====================================
// CONSULTAR ALERTAS
// =====================================

func consultAlerts(
	ctx context.Context,
	client pb.MonitoringServiceClient,
) {

	fmt.Println()
	fmt.Println("===================================")
	fmt.Println(" CONSULTANDO ALERTAS")
	fmt.Println("===================================")

	response, err :=
		client.GetAlerts(
			ctx,
			&pb.GetAlertsRequest{},
		)

	if err != nil {

		log.Printf(
			"Error consultando alertas: %v",
			err,
		)

		return
	}

	alerts :=
		response.GetAlerts()

	fmt.Printf(
		"Total alertas: %d\n",
		len(alerts),
	)

	if len(alerts) == 0 {

		fmt.Println(
			"No existen alertas registradas.",
		)

		return
	}

	for index, alert := range alerts {

		fmt.Println(
			"-----------------------------------",
		)

		fmt.Printf(
			"Alerta #%d\n",
			index+1,
		)

		fmt.Printf(
			"ID: %s\n",
			alert.GetId(),
		)

		fmt.Printf(
			"Regla: %s\n",
			alert.GetRuleName(),
		)

		fmt.Printf(
			"Rule ID: %s\n",
			alert.GetRuleId(),
		)

		fmt.Printf(
			"Componente: %s\n",
			alert.GetComponentName(),
		)

		fmt.Printf(
			"Métrica: %s\n",
			alert.GetMetric(),
		)

		fmt.Printf(
			"Valor actual: %.2f\n",
			alert.GetCurrentValue(),
		)

		fmt.Printf(
			"Umbral: %.2f\n",
			alert.GetThreshold(),
		)

		fmt.Printf(
			"Mensaje: %s\n",
			alert.GetMessage(),
		)

		fmt.Printf(
			"Creada: %s\n",
			formatTimestamp(
				alert.GetCreatedAt(),
			),
		)

		fmt.Printf(
			"Activa: %t\n",
			alert.GetActive(),
		)

		if alert.GetResolvedAt() == 0 {

			fmt.Println(
				"Resuelta: no",
			)

		} else {

			fmt.Printf(
				"Resuelta: %s\n",
				formatTimestamp(
					alert.GetResolvedAt(),
				),
			)
		}
	}
}

// =====================================
// CONSULTAR REGLAS DE ALERTA
// =====================================

func consultAlertRules(
	ctx context.Context,
	client pb.MonitoringServiceClient,
) {

	fmt.Println()
	fmt.Println("===================================")
	fmt.Println(" CONSULTANDO REGLAS DE ALERTA")
	fmt.Println("===================================")

	response, err :=
		client.GetAlertRules(
			ctx,
			&pb.GetAlertRulesRequest{},
		)

	if err != nil {

		log.Printf(
			"Error consultando reglas: %v",
			err,
		)

		return
	}

	rules :=
		response.GetRules()

	fmt.Printf(
		"Total reglas: %d\n",
		len(rules),
	)

	if len(rules) == 0 {

		fmt.Println(
			"No existen reglas de alerta.",
		)

		return
	}

	for index, rule := range rules {

		fmt.Println(
			"-----------------------------------",
		)

		fmt.Printf(
			"Regla #%d\n",
			index+1,
		)

		fmt.Printf(
			"ID: %s\n",
			rule.GetId(),
		)

		fmt.Printf(
			"Nombre: %s\n",
			rule.GetName(),
		)

		fmt.Printf(
			"Métrica: %s\n",
			rule.GetMetric(),
		)

		fmt.Printf(
			"Operador: %s\n",
			rule.GetOperator(),
		)

		fmt.Printf(
			"Umbral: %.2f\n",
			rule.GetThreshold(),
		)

		fmt.Printf(
			"Componente: %s\n",
			rule.GetComponentName(),
		)

		fmt.Printf(
			"Habilitada: %t\n",
			rule.GetEnabled(),
		)
	}
}

// =====================================
// BYTES -> KB
// =====================================

func bytesToKB(
	value int64,
) float64 {

	return float64(value) /
		1024
}

// =====================================
// BYTES -> MB
// =====================================

func bytesToMB(
	value int64,
) float64 {

	return float64(value) /
		(1024 * 1024)
}

// =====================================
// FORMATEAR TIMESTAMP
// =====================================

func formatTimestamp(
	timestamp int64,
) string {

	if timestamp == 0 {

		return "no disponible"
	}

	return time.Unix(
		timestamp,
		0,
	).Format(
		"2006-01-02 15:04:05",
	)
}
