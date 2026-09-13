package gridgrpc

import (
	"context"
	"time"

	gridpb "github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/generated"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/internal/model"
	"github.com/UPB-Cientifica-Team07/Repo-STORIO/services/grid-service/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GridServer struct {
	gridpb.UnimplementedGridServiceServer

	service *service.GridService
}

func NewGridServer(
	service *service.GridService,
) *GridServer {

	return &GridServer{
		service: service,
	}
}

func (s *GridServer) RegisterResource(
	ctx context.Context,
	req *gridpb.RegisterResourceRequest,
) (*gridpb.ResourceResponse, error) {

	resource, err :=
		s.service.RegisterResource(
			model.Resource{
				ResourceID: req.GetResourceId(),
				Name:       req.GetName(),
				DeviceType: req.GetDeviceType(),
				IP:         req.GetIp(),
				CPUCores:   req.GetCpuCores(),
				MemoryMB:   req.GetMemoryMb(),
				MPICapable: req.GetMpiCapable(),

				Capabilities: req.GetCapabilities(),
			},
		)

	if err != nil {
		return nil,
			status.Error(
				codes.InvalidArgument,
				err.Error(),
			)
	}

	return &gridpb.ResourceResponse{
		Resource: toProtoResource(
			resource,
		),
	}, nil
}

func (s *GridServer) Heartbeat(
	ctx context.Context,
	req *gridpb.HeartbeatRequest,
) (*gridpb.ResourceResponse, error) {

	resource, err :=
		s.service.Heartbeat(
			req.GetResourceId(),
			req.GetStatus(),
		)

	if err != nil {
		return nil,
			status.Error(
				codes.InvalidArgument,
				err.Error(),
			)
	}

	return &gridpb.ResourceResponse{
		Resource: toProtoResource(
			resource,
		),
	}, nil
}

func (s *GridServer) GetResource(
	ctx context.Context,
	req *gridpb.GetResourceRequest,
) (*gridpb.ResourceResponse, error) {

	resource, err :=
		s.service.GetResource(
			req.GetResourceId(),
		)

	if err != nil {
		return nil,
			status.Error(
				codes.NotFound,
				err.Error(),
			)
	}

	return &gridpb.ResourceResponse{
		Resource: toProtoResource(
			resource,
		),
	}, nil
}

func (s *GridServer) ListResources(
	ctx context.Context,
	req *gridpb.ListResourcesRequest,
) (*gridpb.ListResourcesResponse, error) {

	resources, err :=
		s.service.ListResources(
			req.GetDeviceType(),
			req.GetStatus(),
		)

	if err != nil {
		return nil,
			status.Error(
				codes.InvalidArgument,
				err.Error(),
			)
	}

	return &gridpb.ListResourcesResponse{
		Resources: toProtoResources(
			resources,
		),
	}, nil
}

func (s *GridServer) ListComputeResources(
	ctx context.Context,
	req *gridpb.ListComputeResourcesRequest,
) (*gridpb.ListResourcesResponse, error) {

	resources, err :=
		s.service.ListComputeResources(
			req.GetOnlyAvailable(),
		)

	if err != nil {
		return nil,
			status.Error(
				codes.Internal,
				err.Error(),
			)
	}

	return &gridpb.ListResourcesResponse{
		Resources: toProtoResources(
			resources,
		),
	}, nil
}

func (s *GridServer) Health(
	ctx context.Context,
	req *gridpb.HealthRequest,
) (*gridpb.HealthResponse, error) {

	total, err :=
		s.service.CountResources()

	if err != nil {
		return nil,
			status.Error(
				codes.Internal,
				err.Error(),
			)
	}

	return &gridpb.HealthResponse{
		Status:         "ACTIVE",
		TotalResources: total,
		Timestamp:      time.Now().Unix(),
	}, nil
}

func toProtoResource(
	resource *model.Resource,
) *gridpb.Resource {

	if resource == nil {
		return nil
	}

	return &gridpb.Resource{
		Id:           resource.ID,
		ResourceId:   resource.ResourceID,
		Name:         resource.Name,
		DeviceType:   resource.DeviceType,
		Ip:           resource.IP,
		CpuCores:     resource.CPUCores,
		MemoryMb:     resource.MemoryMB,
		MpiCapable:   resource.MPICapable,
		Capabilities: resource.Capabilities,
		Status:       resource.Status,

		LastHeartbeat: resource.LastHeartbeat.Unix(),
		CreatedAt:     resource.CreatedAt.Unix(),
		UpdatedAt:     resource.UpdatedAt.Unix(),
	}
}

func toProtoResources(
	resources []model.Resource,
) []*gridpb.Resource {

	result :=
		make(
			[]*gridpb.Resource,
			0,
			len(resources),
		)

	for index := range resources {

		result =
			append(
				result,
				toProtoResource(
					&resources[index],
				),
			)
	}

	return result
}
