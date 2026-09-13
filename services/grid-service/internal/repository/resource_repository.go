package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/internal/model"
)

type ResourceRepository struct {
	db *sql.DB
}

func NewResourceRepository(
	db *sql.DB,
) *ResourceRepository {

	return &ResourceRepository{
		db: db,
	}
}

func (r *ResourceRepository) Upsert(
	resource model.Resource,
) (*model.Resource, error) {

	capabilitiesJSON, err :=
		json.Marshal(
			resource.Capabilities,
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"no se pudieron serializar capabilities: %w",
				err,
			)
	}

	query := `
		INSERT INTO grid_resource (
			resource_id,
			nombre,
			tipo,
			ip,
			cpu_cores,
			memoria_mb,
			mpi_capable,
			capabilities,
			estado,
			last_heartbeat,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8::jsonb,
			$9,
			CURRENT_TIMESTAMP,
			CURRENT_TIMESTAMP
		)
		ON CONFLICT (resource_id)
		DO UPDATE SET
			nombre = EXCLUDED.nombre,
			tipo = EXCLUDED.tipo,
			ip = EXCLUDED.ip,
			cpu_cores = EXCLUDED.cpu_cores,
			memoria_mb = EXCLUDED.memoria_mb,
			mpi_capable = EXCLUDED.mpi_capable,
			capabilities = EXCLUDED.capabilities,
			estado = EXCLUDED.estado,
			last_heartbeat = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		RETURNING
			id_recurso::text,
			resource_id,
			nombre,
			tipo,
			COALESCE(ip, ''),
			cpu_cores,
			memoria_mb,
			mpi_capable,
			capabilities,
			estado,
			last_heartbeat,
			created_at,
			updated_at
	`

	return scanResource(
		r.db.QueryRow(
			query,
			resource.ResourceID,
			resource.Name,
			resource.DeviceType,
			resource.IP,
			resource.CPUCores,
			resource.MemoryMB,
			resource.MPICapable,
			string(capabilitiesJSON),
			resource.Status,
		),
	)
}

func (r *ResourceRepository) Heartbeat(
	resourceID string,
	status string,
) (*model.Resource, error) {

	query := `
		UPDATE grid_resource
		SET
			estado = $2,
			last_heartbeat = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE resource_id = $1
		RETURNING
			id_recurso::text,
			resource_id,
			nombre,
			tipo,
			COALESCE(ip, ''),
			cpu_cores,
			memoria_mb,
			mpi_capable,
			capabilities,
			estado,
			last_heartbeat,
			created_at,
			updated_at
	`

	resource, err :=
		scanResource(
			r.db.QueryRow(
				query,
				resourceID,
				status,
			),
		)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return nil, nil
	}

	return resource, err
}

func (r *ResourceRepository) GetByResourceID(
	resourceID string,
) (*model.Resource, error) {

	query := `
		SELECT
			id_recurso::text,
			resource_id,
			nombre,
			tipo,
			COALESCE(ip, ''),
			cpu_cores,
			memoria_mb,
			mpi_capable,
			capabilities,
			estado,
			last_heartbeat,
			created_at,
			updated_at
		FROM grid_resource
		WHERE resource_id = $1
		LIMIT 1
	`

	resource, err :=
		scanResource(
			r.db.QueryRow(
				query,
				resourceID,
			),
		)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return nil, nil
	}

	return resource, err
}

func (r *ResourceRepository) List(
	deviceType string,
	status string,
) ([]model.Resource, error) {

	query := `
		SELECT
			id_recurso::text,
			resource_id,
			nombre,
			tipo,
			COALESCE(ip, ''),
			cpu_cores,
			memoria_mb,
			mpi_capable,
			capabilities,
			estado,
			last_heartbeat,
			created_at,
			updated_at
		FROM grid_resource
		WHERE
			($1 = '' OR tipo = $1)
			AND
			($2 = '' OR estado = $2)
		ORDER BY
			tipo,
			resource_id
	`

	rows, err :=
		r.db.Query(
			query,
			deviceType,
			status,
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"error listando recursos Grid: %w",
				err,
			)
	}

	defer rows.Close()

	return scanResources(
		rows,
	)
}

func (r *ResourceRepository) ListCompute(
	onlyAvailable bool,
) ([]model.Resource, error) {

	query := `
		SELECT
			id_recurso::text,
			resource_id,
			nombre,
			tipo,
			COALESCE(ip, ''),
			cpu_cores,
			memoria_mb,
			mpi_capable,
			capabilities,
			estado,
			last_heartbeat,
			created_at,
			updated_at
		FROM grid_resource
		WHERE
			tipo = 'COMPUTE'
			AND mpi_capable = TRUE
			AND (
				$1 = FALSE
				OR estado = 'AVAILABLE'
			)
		ORDER BY
			resource_id
	`

	rows, err :=
		r.db.Query(
			query,
			onlyAvailable,
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"error listando recursos Compute Grid: %w",
				err,
			)
	}

	defer rows.Close()

	return scanResources(
		rows,
	)
}

func (r *ResourceRepository) MarkInactiveBefore(
	cutoff time.Time,
) (int64, error) {

	result, err :=
		r.db.Exec(
			`
				UPDATE grid_resource
				SET
					estado = 'INACTIVE',
					updated_at = CURRENT_TIMESTAMP
				WHERE
					last_heartbeat < $1
					AND estado <> 'INACTIVE'
			`,
			cutoff,
		)

	if err != nil {
		return 0,
			fmt.Errorf(
				"error marcando recursos Grid inactivos: %w",
				err,
			)
	}

	rowsAffected, err :=
		result.RowsAffected()

	if err != nil {
		return 0,
			fmt.Errorf(
				"error obteniendo recursos Grid actualizados: %w",
				err,
			)
	}

	return rowsAffected, nil
}

func (r *ResourceRepository) Count() (
	int64,
	error,
) {

	var total int64

	err :=
		r.db.QueryRow(
			`
				SELECT COUNT(*)::bigint
				FROM grid_resource
			`,
		).Scan(
			&total,
		)

	if err != nil {
		return 0,
			fmt.Errorf(
				"error contando recursos Grid: %w",
				err,
			)
	}

	return total, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanResource(
	row scanner,
) (*model.Resource, error) {

	var resource model.Resource

	var capabilitiesJSON []byte

	err :=
		row.Scan(
			&resource.ID,
			&resource.ResourceID,
			&resource.Name,
			&resource.DeviceType,
			&resource.IP,
			&resource.CPUCores,
			&resource.MemoryMB,
			&resource.MPICapable,
			&capabilitiesJSON,
			&resource.Status,
			&resource.LastHeartbeat,
			&resource.CreatedAt,
			&resource.UpdatedAt,
		)

	if err != nil {
		return nil, err
	}

	if len(capabilitiesJSON) == 0 {
		resource.Capabilities =
			[]string{}
	} else {

		if err :=
			json.Unmarshal(
				capabilitiesJSON,
				&resource.Capabilities,
			); err != nil {

			return nil,
				fmt.Errorf(
					"capabilities JSON inválido: %w",
					err,
				)
		}
	}

	return &resource, nil
}

func scanResources(
	rows *sql.Rows,
) ([]model.Resource, error) {

	resources :=
		make(
			[]model.Resource,
			0,
		)

	for rows.Next() {

		resource, err :=
			scanResource(
				rows,
			)

		if err != nil {
			return nil,
				fmt.Errorf(
					"error leyendo recurso Grid: %w",
					err,
				)
		}

		resources =
			append(
				resources,
				*resource,
			)
	}

	if err := rows.Err(); err != nil {
		return nil,
			fmt.Errorf(
				"error iterando recursos Grid: %w",
				err,
			)
	}

	return resources, nil
}

func NormalizeCapabilities(
	values []string,
) []string {

	seen :=
		make(
			map[string]struct{},
		)

	result :=
		make(
			[]string,
			0,
			len(values),
		)

	for _, value := range values {

		value =
			strings.TrimSpace(
				value,
			)

		if value == "" {
			continue
		}

		key :=
			strings.ToLower(
				value,
			)

		if _, exists :=
			seen[key]; exists {

			continue
		}

		seen[key] =
			struct{}{}

		result =
			append(
				result,
				value,
			)
	}

	return result
}
