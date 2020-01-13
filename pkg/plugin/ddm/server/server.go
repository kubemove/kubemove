package server

import (
	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	"github.com/kubemove/kubemove/pkg/plugin/framework/server"
	pb "github.com/kubemove/kubemove/pkg/plugin/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type ddm struct {
	server server.Server
}

type plugin struct {
	pb.DataSyncerClient
}

func NewDDMServer() (DDM, error) {
	s, err := server.NewServer(
		func(s *grpc.ClientConn) interface{} {
			return pb.NewDataSyncerClient(s)
		},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create server")
	}

	go s.Serve()

	return &ddm{
		server: s,
	}, nil
}

func (d *ddm) Init(pluginName string, config map[string]string) error {
	cpl := d.server.Get(pluginName, true)
	defer d.server.Put(pluginName)
	if cpl == nil {
		return errors.New("Plugin is not available")
	}

	pl, ok := cpl.(pb.DataSyncerClient)
	if !ok {
		return errors.New("Plugin is not valid")
	}

	s := &plugin{
		DataSyncerClient: pl,
	}
	return s.init(config)
}

func (d *ddm) SyncData(pluginName string, ds v1alpha1.DataSync) (string, error) {
	cpl := d.server.Get(pluginName, true)
	defer d.server.Put(pluginName)
	if cpl == nil {
		return "", errors.New("Plugin is not available")
	}

	pl, ok := cpl.(pb.DataSyncerClient)
	if !ok {
		return "", errors.New("Plugin is not valid")
	}

	s := &plugin{
		DataSyncerClient: pl,
	}
	return s.syncData(ds)
}

func (d *ddm) SyncStatus(pluginName, id string) (int32, error) {
	cpl := d.server.Get(pluginName, true)
	defer d.server.Put(pluginName)
	if cpl == nil {
		return ERR, errors.New("Plugin is not available")
	}

	pl, ok := cpl.(pb.DataSyncerClient)
	if !ok {
		return ERR, errors.New("Plugin is not valid")
	}

	s := &plugin{
		DataSyncerClient: pl,
	}
	return s.syncStatus(id)
}
