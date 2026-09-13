package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/internal/model"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/internal/repository"
)

type GridService struct {
	repository *repository.ResourceRepository
}

func NewGridService(
	repository *repository.ResourceRepository,
) *GridService {

	return &GridService{
		repository: repository,
	}
}

func (s *GridService) RegisterResource(
	resource model.Resource,
) (*model.Resource, error) {

	resource.ResourceID =
		strings.TrimSpace(
			resource.ResourceID,
		)

	resource.Name =
		strings.TrimSpace(
			resource.Name,
		)

	resource.DeviceType =
		strings.ToUpper(
			strings.TrimSpace(
				resource.DeviceType,
			),
		)

	resource.IP =
		strings.TrimSpace(
			resource.IP,
		)

	resource.Capabilities =
		repository.NormalizeCapabilities(
			resource.Capabilities,
		)

	if resource.ResourceID == "" {
		return nil,
			fmt.Errorf(
				"resource_id es obligatorio",
			)
	}

	if resource.Name == "" {
		return nil,
			fmt.Errorf(
				"name es obligatorio",
			)
	}

	if err :=
		validateDeviceType(
			resource.DeviceType,
		); err != nil {

		return nil, err
	}

	if resource.CPUCores <= 0 {
		return nil,
			fmt.Errorf(
				"cpu_cores debe ser mayor que cero",
			)
	}

	if resource.MemoryMB < 0 {
		return nil,
			fmt.Errorf(
				"memory_mb no puede ser negativo",
			)
	}

	if resource.MPICapable &&
		resource.DeviceType !=
			model.DeviceTypeCompute {

		return nil,
			fmt.Errorf(
				"solo los recursos COMPUTE pueden ser mpi_capable",
			)
	}

	resource.Status =
		model.StatusAvailable

	return s.repository.Upsert(
		resource,
	)
}

func (s *GridService) Heartbeat(
	resourceID string,
	status string,
) (*model.Resource, error) {

	resourceID =
		strings.TrimSpace(
			resourceID,
		)

	status =
		strings.ToUpper(
			strings.TrimSpace(
				status,
			),
		)

	if resourceID == "" {
		return nil,
			fmt.Errorf(
				"resource_id es obligatorio",
			)
	}

	if status == "" {
		status =
			model.StatusAvailable
	}

	if err :=
		validateHeartbeatStatus(
			status,
		); err != nil {

		return nil, err
	}

	resource, err :=
		s.repository.Heartbeat(
			resourceID,
			status,
		)

	if err != nil {
		return nil, err
	}

	if resource == nil {
		return nil,
			fmt.Errorf(
				"recurso Grid no encontrado",
			)
	}

	return resource, nil
}

func (s *GridService) GetResource(
	resourceID string,
) (*model.Resource, error) {

	resourceID =
		strings.TrimSpace(
			resourceID,
		)

	if resourceID == "" {
		return nil,
			fmt.Errorf(
				"resource_id es obligatorio",
			)
	}

	resource, err :=
		s.repository.GetByResourceID(
			resourceID,
		)

	if err != nil {
		return nil, err
	}

	if resource == nil {
		return nil,
			fmt.Errorf(
				"recurso Grid no encontrado",
			)
	}

	return resource, nil
}

func (s *GridService) ListResources(
	deviceType string,
	status string,
) ([]model.Resource, error) {

	deviceType =
		strings.ToUpper(
			strings.TrimSpace(
				deviceType,
			),
		)

	status =
		strings.ToUpper(
			strings.TrimSpace(
				status,
			),
		)

	if deviceType != "" {

		if err :=
			validateDeviceType(
				deviceType,
			); err != nil {

			return nil, err
		}
	}

	if status != "" {

		if err :=
			validateStatus(
				status,
			); err != nil {

			return nil, err
		}
	}

	return s.repository.List(
		deviceType,
		status,
	)
}

func (s *GridService) ListComputeResources(
	onlyAvailable bool,
) ([]model.Resource, error) {

	return s.repository.ListCompute(
		onlyAvailable,
	)
}

func (s *GridService) CountResources() (
	int64,
	error,
) {

	return s.repository.Count()
}

func (s *GridService) MarkInactive(
	timeout time.Duration,
) (int64, error) {

	if timeout <= 0 {
		return 0,
			fmt.Errorf(
				"timeout debe ser mayor que cero",
			)
	}

	cutoff :=
		time.Now().
			Add(
				-timeout,
			)

	return s.repository.MarkInactiveBefore(
		cutoff,
	)
}

func validateDeviceType(
	deviceType string,
) error {

	switch deviceType {

	case model.DeviceTypeCompute,
		model.DeviceTypeMobile,
		model.DeviceTypeIoT:

		return nil

	default:

		return fmt.Errorf(
			"device_type inválido: %s",
			deviceType,
		)
	}
}

func validateStatus(
	status string,
) error {

	switch status {

	case model.StatusAvailable,
		model.StatusBusy,
		model.StatusInactive:

		return nil

	default:

		return fmt.Errorf(
			"status inválido: %s",
			status,
		)
	}
}

func validateHeartbeatStatus(
	status string,
) error {

	switch status {

	case model.StatusAvailable,
		model.StatusBusy:

		return nil

	default:

		return fmt.Errorf(
			"heartbeat status inválido: %s",
			status,
		)
	}
}
