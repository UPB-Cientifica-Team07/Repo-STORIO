package repository

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/monitoring-service/internal/model"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type HpcRepository struct {
	db *sql.DB
}

// =====================================
// CONSTRUCTOR
// =====================================

func NewHpcRepository() (
	*HpcRepository,
	error,
) {

	host :=
		getEnv(
			"MONITORING_DB_HOST",
			"127.0.0.1",
		)

	port :=
		getEnv(
			"MONITORING_DB_PORT",
			"5434",
		)

	name :=
		getEnv(
			"MONITORING_DB_NAME",
			"upb_cientifica",
		)

	user :=
		getEnv(
			"MONITORING_DB_USER",
			"upb_app",
		)

	password :=
		os.Getenv(
			"MONITORING_DB_PASSWORD",
		)

	if password == "" {

		return nil,
			fmt.Errorf(
				"MONITORING_DB_PASSWORD es obligatoria",
			)
	}

	dsnURL :=
		&url.URL{
			Scheme: "postgres",
			User: url.UserPassword(
				user,
				password,
			),
			Host: fmt.Sprintf(
				"%s:%s",
				host,
				port,
			),
			Path: name,
		}

	query :=
		dsnURL.Query()

	query.Set(
		"sslmode",
		"disable",
	)

	dsnURL.RawQuery =
		query.Encode()

	dsn :=
		dsnURL.String()

	db, err :=
		sql.Open(
			"pgx",
			dsn,
		)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(
		30 * time.Minute,
	)

	if err :=
		db.Ping(); err != nil {

		db.Close()

		return nil,
			fmt.Errorf(
				"PostgreSQL HPC no disponible: %w",
				err,
			)
	}

	return &HpcRepository{
		db: db,
	}, nil
}

// =====================================
// CERRAR
// =====================================

func (r *HpcRepository) Close() error {

	if r == nil ||
		r.db == nil {

		return nil
	}

	return r.db.Close()
}

// =====================================
// CONSULTAR NODO
// =====================================

func (r *HpcRepository) GetNode(
	nodeID string,
) (*model.HpcNode, error) {

	if nodeID == "" {

		return nil,
			fmt.Errorf(
				"node_id es obligatorio",
			)
	}

	query := `
		SELECT
			id_nodo::text,
			hostname,
			estado,
			cpu,
			memoria_mb,
			COALESCE(ip, ''),
			COALESCE(ubicacion, '')
		FROM nodo_hpc
		WHERE id_nodo::text = $1
		   OR hostname = $1
		LIMIT 1
	`

	var node model.HpcNode

	err :=
		r.db.QueryRow(
			query,
			nodeID,
		).Scan(
			&node.NodeID,
			&node.Hostname,
			&node.Status,
			&node.CPUCores,
			&node.MemoryMB,
			&node.IP,
			&node.Location,
		)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {

		return nil,
			fmt.Errorf(
				"error consultando nodo HPC: %w",
				err,
			)
	}

	/*
		CPUUsage y MemoryUsage todavía no existen
		como telemetría dinámica en nodo_hpc.
		No se inventan valores.
	*/
	node.CPUUsage = 0
	node.MemoryUsage = 0
	/*
		La tabla nodo_hpc actualmente no conserva
		un timestamp de último heartbeat.
		LastUpdated permanece en su valor cero.
	*/
	return &node, nil
}

// =====================================
// RESUMEN DE JOBS + NODOS
// =====================================

func (r *HpcRepository) GetSummary() (
	model.HpcSummary,
	error,
) {

	var summary model.HpcSummary

	jobQuery := `
		SELECT
			COUNT(*)::bigint,

			COUNT(*) FILTER (
				WHERE estado = 'PENDIENTE'
			)::bigint,

			COUNT(*) FILTER (
				WHERE estado = 'EJECUTANDO'
			)::bigint,

			COUNT(*) FILTER (
				WHERE estado = 'FINALIZADO'
			)::bigint,

			COUNT(*) FILTER (
				WHERE estado = 'ERROR'
			)::bigint,

			COUNT(*) FILTER (
				WHERE estado = 'CANCELADO'
			)::bigint,

			COALESCE(
				AVG(
					EXTRACT(
						EPOCH FROM (fin - inicio)
					) * 1000
				) FILTER (
					WHERE inicio IS NOT NULL
					  AND fin IS NOT NULL
				),
				0
			)::double precision

		FROM trabajo_hpc
	`

	err :=
		r.db.QueryRow(
			jobQuery,
		).Scan(
			&summary.TotalJobs,
			&summary.PendingJobs,
			&summary.RunningJobs,
			&summary.CompletedJobs,
			&summary.FailedJobs,
			&summary.CancelledJobs,
			&summary.AverageExecutionMs,
		)

	if err != nil {

		return model.HpcSummary{},
			fmt.Errorf(
				"error consultando trabajos HPC: %w",
				err,
			)
	}

	nodeQuery := `
		SELECT
			COUNT(*)::bigint,

			COUNT(*) FILTER (
				WHERE estado = 'DISPONIBLE'
			)::bigint,

			COUNT(*) FILTER (
				WHERE estado = 'OCUPADO'
			)::bigint,

			COUNT(*) FILTER (
				WHERE estado = 'INACTIVO'
			)::bigint

		FROM nodo_hpc
	`

	err =
		r.db.QueryRow(
			nodeQuery,
		).Scan(
			&summary.TotalNodes,
			&summary.AvailableNodes,
			&summary.BusyNodes,
			&summary.InactiveNodes,
		)

	if err != nil {

		return model.HpcSummary{},
			fmt.Errorf(
				"error consultando nodos HPC: %w",
				err,
			)
	}

	summary.Timestamp =
		time.Now()

	return summary, nil
}

// =====================================
// ENV
// =====================================

func getEnv(
	name string,
	defaultValue string,
) string {

	value :=
		os.Getenv(
			name,
		)

	if value == "" {
		return defaultValue
	}

	return value
}
