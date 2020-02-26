package main

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/client-go/kubernetes"

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
func triggerSnapshot(k8sClient kubernetes.Interface, params PluginParameters, snapshotName string) (string, error) {
	// crate an Elasticsearch client
	esClient, err := newElasticsearchClient(k8sClient, params.Elasticsearch)
	if err != nil {
		return "", err
	}

	// configure snapshot create request
	snapshotRequest := esapi.SnapshotCreateRequest{
		Repository: params.Repository.Name,
		Snapshot:   snapshotName,
		Pretty:     true,
	}

	// create snapshot
	resp, err := snapshotRequest.Do(context.Background(), esClient)
	if err != nil {
		return "", err
	}

	fmt.Println("Response: ", resp.String())

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		// TODO: parse response and return failure case
		return "", fmt.Errorf("failed to create snapshot. Reason: ")
	}

	return "", nil
}

// triggerRestore hits the Recovery API of Elasticsearch to trigger a restore process
func triggerRestore(k8sClient kubernetes.Interface, params PluginParameters, snapshotName string) (string, error) {
	// crate an Elasticsearch client
	esClient, err := newElasticsearchClient(k8sClient, params.Elasticsearch)
	if err != nil {
		return "", err
	}

	// close all indexes
	closeRequest := esapi.IndicesCloseRequest{
		Index:  []string{"_all"},
		Pretty: true,
	}
	closeResp, err := closeRequest.Do(context.Background(), esClient)
	if err != nil {
		return "", err
	}
	fmt.Println("Close Response: ", closeResp.String())

	// trigger restore
	restoreRequest := esapi.SnapshotRestoreRequest{
		Repository: params.Repository.Name,
		Snapshot:   snapshotName,
		Pretty:     true,
	}
	resp, err := restoreRequest.Do(context.Background(), esClient)
	if err != nil {
		return "", err
	}
	fmt.Println("Restore Response: ", resp.String())

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		// TODO: parse response and return failure case
		return "", fmt.Errorf("failed to restore snapshot. Reason: ")
	}

	return "", nil
}
