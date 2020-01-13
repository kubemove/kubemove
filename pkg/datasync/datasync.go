package datasync

import (
	"github.com/go-logr/logr"
	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	server "github.com/kubemove/kubemove/pkg/plugin/ddm/server"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type DataSync struct {
	ds  *v1alpha1.DataSync
	ddm server.DDM
	log logr.Logger
}

func NewDataSyncer(ds *v1alpha1.DataSync, ddm server.DDM) *DataSync {
	if ds == nil || ddm == nil {
		return nil
	}

	log := logf.Log.WithName("DataSync").
		WithValues("Request.Name", ds.Name)

	if len(ds.Spec.PluginProvider) == 0 ||
		len(ds.Spec.MoveEngine) == 0 {
		log.Error(nil, "DataSync having insufficient details")
		return nil
	}

	return &DataSync{
		ds:  ds,
		ddm: ddm,
		log: log,
	}
}
