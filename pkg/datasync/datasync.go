package datasync

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	server "github.com/kubemove/kubemove/pkg/plugin/ddm/server"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type DataSync struct {
	ds     *v1alpha1.DataSync
	ddm    server.DDM
	log    logr.Logger
	client client.Client
}

func NewDataSyncer(client client.Client, ds *v1alpha1.DataSync, ddm server.DDM) *DataSync {
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
		client: client,
		ds:     ds,
		ddm:    ddm,
		log:    log,
	}
}

func (d *DataSync) UpdateDataSyncStatus(state v1alpha1.SyncState, failureReason string) error {
	// Set the state to new state
	d.ds.Status.Status = state
	// if sync is completed, set the completion time
	if state == v1alpha1.SyncStateCompleted {
		d.ds.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	}
	// if sync has failed, set the respective failure reason
	if state == v1alpha1.SyncStateFailed {
		d.ds.Status.Reason = failureReason
	}
	return d.client.Update(context.TODO(), d.ds)
}
