package server

import (
	"context"

	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	pb "github.com/kubemove/kubemove/pkg/plugin/proto"
	"github.com/pkg/errors"
)

const (
	OK  = 0
	ERR = 1
)

func (p *plugin) init(config map[string]string) error {
	req := &pb.InitRequest{
		Config: config,
	}

	res, err := p.Init(context.Background(), req)
	if err != nil {
		return err
	}

	if res.Status != OK {
		return errors.Errorf("Failed to initialize plugin {%s}", res.Reason)
	}
	return nil
}

func getSyncVolList(vol []*v1alpha1.DataVolume) map[string]*pb.SyncRequest_SyncVolRequest {
	volList := make(map[string]*pb.SyncRequest_SyncVolRequest)

	for _, v := range vol {
		syncVol := &pb.SyncRequest_SyncVolRequest{
			VolumeName:        v.Name,
			VolumeClaim:       v.PVC,
			RemoteVolumeClaim: v.PVC,
			LocalNS:           v.Namespace,
			RemoteNS:          v.RemoteNamespace,
		}
		volList[v.Name] = syncVol
	}

	return volList
}

func (p *plugin) syncData(ds v1alpha1.DataSync) (string, error) {
	volList := getSyncVolList(ds.Spec.Volume)
	req := &pb.SyncRequest{
		Engine:     ds.Spec.MoveEngine,
		Restore:    ds.Spec.Restore,
		Params:     ds.Spec.Config,
		VolumeList: volList,
	}

	res, err := p.SyncData(context.Background(), req)
	if err != nil {
		return "", err
	}

	if len(res.SyncID) == 0 {
		return "", errors.Errorf("Failed to sync data {%v}", res.Reason)
	}

	return res.SyncID, nil
}

func (p *plugin) syncStatus(id string) (int32, error) {
	req := &pb.SyncStatusRequest{
		SyncID: id,
	}

	res, err := p.SyncStatus(context.Background(), req)
	if err != nil {
		return ERR, err
	}

	return res.Status, errors.New(res.Reason)
}
