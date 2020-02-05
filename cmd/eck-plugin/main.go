package main

import (
	"github.com/go-logr/logr"
	client "github.com/kubemove/kubemove/pkg/plugin/ddm/plugin"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type eckDDM struct {
	log logr.Logger
	p   client.Plugin
}

var _ client.Plugin = (*eckDDM)(nil)

func main() {
	var log = logf.Log.WithName("ECK Plugin")

	err := client.Register("eck-plugin",
		&eckDDM{
			log: log,
		})
	if err != nil {
		log.Error(err, "Failed to create a plugin")
		return
	}
}

func (d *eckDDM) Init(params map[string]string) error {
	// Register Snapshot Repository in the source Elasticsearch
	d.log.Info("Registering Snapshot Repository in Elasticsearch %s/%s of Cluster: %s\n")
	err := registerSnapshotRepository(params)
	if err != nil {
		return err
	}

	// Register Snapshot Repository in the destination Elasticsearch
	d.log.Info("Registering Snapshot Repository in Elasticsearch %s/%s of Cluster: %s\n")
	return registerSnapshotRepository(params)
}

func (d *eckDDM) Sync(syncPhase string, whyThisField bool, params map[string]string, vol []*client.Volume) (string, error) {
	// trigger a snapshot in the source ES in backup phase
	if SyncPhase(syncPhase) == SyncPhaseBackup {
		return triggerSnapshot(params)
	}

	// trigger a restore in the destination ES in restore phase
	if SyncPhase(syncPhase) == SyncPhaseRestore {
		return triggerRestore(params)
	}

	return "what the use of this?", nil
}

func (d *eckDDM) Status(statusType string) (int32, error) {

	// send snapshot state on backup status request
	if StatusType(statusType) == BackupStatus {
		return 0, retrieveSnapshotState("should send Snapshot name")
	}

	// send restore state on restore status request
	if StatusType(statusType) == RestoreStatus {
		return 0, retrieveRestoreState("should send Snapshot name")
	}

	return 0, nil
}
