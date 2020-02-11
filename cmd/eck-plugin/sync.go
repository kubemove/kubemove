package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type SyncPhase string
type StatusType string

const (
	SyncPhaseBackup  SyncPhase = "backup"
	SyncPhaseRestore SyncPhase = "restore"

	BackupStatus  StatusType = "backup-status"
	RestoreStatus StatusType = "restore-status"
)

// triggerSnapshot hits the Snapshot API of Elasticsearch to trigger a snapshot
func (d *eckDDM) triggerSnapshot(params PluginParameters, snapshotName string) (string, error) {
	// crate an Elasticsearch client
	esClient, err := d.newElasticsearchClient(params.Elasticsearch)
	if err != nil {
		return "", err
	}

	// configure snapshot create request
	snapshotRequest := esapi.SnapshotCreateRequest{
		Repository: params.Repository.Name,
		Snapshot:   snapshotName,
	}

	// create snapshot
	resp, err := snapshotRequest.Do(context.Background(), esClient)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		defer resp.Body.Close()
		// TODO: parse response and return failure case
		return "", fmt.Errorf("failed to create snapshot. Reason: ")
	}

	return "", nil
}

// triggerRestore hits the Recovery API of Elasticsearch to trigger a restore process
func (d *eckDDM) triggerRestore(params PluginParameters, snapshotName string) (string, error) {
	// crate an Elasticsearch client
	esClient, err := d.newElasticsearchClient(params.Elasticsearch)
	if err != nil {
		return "", err
	}

	// configure snapshot restore request
	restoreRequest := esapi.SnapshotRestoreRequest{
		Repository: params.Repository.Name,
		Snapshot:   snapshotName,
	}

	// request to restore
	resp, err := restoreRequest.Do(context.Background(), esClient)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusAccepted {
		defer resp.Body.Close()
		// TODO: parse response and return failure case
		return "", fmt.Errorf("failed to restore snapshot. Reason: ")
	}

	return "", nil
}
