package plugin

import (
	"context"

	kmove "github.com/kubemove/kubemove/pkg/plugin/framework/client"
	pb "github.com/kubemove/kubemove/pkg/plugin/proto"
	"google.golang.org/grpc"
)

type server struct {
	client Plugin
}

func Register(pluginName string, iface Plugin) error {
	return kmove.NewClient(pluginName,
		func(g *grpc.Server) {
			pb.RegisterDataSyncerServer(g,
				&server{
					client: iface,
				},
			)
		})
}

func (s *server) Init(ctx context.Context, req *pb.InitRequest) (*pb.InitResponse, error) {
	err := s.client.Init(req.Params)
	if err != nil {
		//TODO
		return nil, err
	}

	return &pb.InitResponse{}, nil
}

func (s *server) SyncData(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	id, err := s.client.Sync(req.Params)

	if err != nil {
		return nil, err
	}

	return &pb.SyncResponse{
		SyncID: id,
	}, nil
}

func (s *server) SyncStatus(ctx context.Context, req *pb.SyncStatusRequest) (*pb.SyncStatusResponse, error) {
	status, err := s.client.Status(req.Params)
	if err != nil {
		return nil, err
	}

	return &pb.SyncStatusResponse{
		Status: status,
	}, nil
}
