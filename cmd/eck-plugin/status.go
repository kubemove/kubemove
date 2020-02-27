package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/client-go/kubernetes"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/kubemove/kubemove/pkg/plugin/ddm/plugin"
)

const (
	SnapshotSucceeded  = "SUCCESS"
	SnapshotInProgress = "IN_PROGRESS"
	SnapshotFailed     = "FAILED"

	RecoveryTypeSnapshot = "SNAPSHOT"
	RecoveryStageDone    = "DONE"
)

// SnapshotGetResponse is used to unmarshal the response body of SnapshotGetRequest
type SnapshotGetResponse struct {
	Snapshots []SnapshotStatus `json:"snapshots"`
}
type SnapshotStatus struct {
	Snapshot string `json:"snapshot"`
	State    string `json:"state"`
}

// IndexesRecoveryStatus is used to unmarshal the indexes recovery status from IndicesRecoveryRequest response
type IndexesRecoveryStatus struct {
	Shards []ShardRecoveryStatus `json:"shards"`
}
type ShardRecoveryStatus struct {
	Type   string         `json:"type"`
	Stage  string         `json:"stage"`
	Source RecoverySource `json:"source"`
}
type RecoverySource struct {
	Snapshot string `json:"snapshot,omitempty"`
}

// retrieveSnapshotState hits the Snapshot API of Elasticsearch to retrieve the backup status of a Snapshot
func retrieveBackupState(k8sClient kubernetes.Interface, params PluginParameters, snapshotName string) (int32, error) {
	// crate an Elasticsearch client
	esClient, err := newElasticsearchClient(k8sClient, params.Elasticsearch)
	if err != nil {
		return plugin.Errored, err
	}

	// configure snapshot get request
	snapshotGetRequest := esapi.SnapshotGetRequest{
		Repository: params.Repository.Name,
		Snapshot:   []string{snapshotName},
		Pretty:     true,
	}

	// get snapshot status
	fmt.Println("Requesting for snapshot status.....")
	resp, err := snapshotGetRequest.Do(context.Background(), esClient)
	if err != nil {
		return plugin.Errored, err
	}
	defer resp.Body.Close()

	fmt.Println("Response: ", resp.String())

	if resp.StatusCode != http.StatusOK {
		// TODO: parse response and return failure case
		return plugin.Errored, fmt.Errorf("failed to create snapshot. Reason: <TO DO>")
	}

	var statusResponse SnapshotGetResponse
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return plugin.Errored, err
	}

	err = json.Unmarshal(data, &statusResponse)
	if err != nil {
		return plugin.Errored, err
	}

	for _, s := range statusResponse.Snapshots {
		if s.Snapshot == snapshotName {
			switch s.State {
			case SnapshotSucceeded:
				return plugin.Completed, nil
			case SnapshotInProgress:
				return plugin.InProgress, nil
			case SnapshotFailed:
				return plugin.Failed, nil // TODO: return failure reason.
			default:
				return plugin.Unknown, fmt.Errorf("snapshot status unknown")
			}
		}
	}
	return plugin.Invalid, fmt.Errorf("no snapshot found with named %s", snapshotName)
}

// retrieveRestoreState hits the Recovery API of Elasticsearch to retrieve the restore status of a ES.
func retrieveRestoreState(k8sClient kubernetes.Interface, params PluginParameters, snapshotName string) (int32, error) {
	// crate an Elasticsearch client
	esClient, err := newElasticsearchClient(k8sClient, params.Elasticsearch)
	if err != nil {
		return plugin.Errored, err
	}

	// configure indexes recovery request
	indexRecoveryRequest := esapi.IndicesRecoveryRequest{
		Index:  nil,
		Pretty: true,
	}

	// get recovery status
	fmt.Println("Requesting for recovery status........")
	resp, err := indexRecoveryRequest.Do(context.Background(), esClient)
	if err != nil {
		return plugin.Errored, err
	}
	defer resp.Body.Close()

	fmt.Println("Response: ", resp.String())

	if resp.StatusCode != http.StatusOK {
		// TODO: parse response and return failure case
		return plugin.Errored, fmt.Errorf("failed to create snapshot. Reason: <TO DO>")
	}

	// parse recovery response
	var recoveryStats map[string]IndexesRecoveryStatus
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return plugin.Errored, err
	}

	err = json.Unmarshal(data, &recoveryStats)
	if err != nil {
		return plugin.Errored, err
	}

	recoveryInitiated := false
	for _, index := range recoveryStats {
		recoveryInitiated = true
		for _, shard := range index.Shards {
			if shard.Type == RecoveryTypeSnapshot && shard.Source.Snapshot != snapshotName {
				// not recovering from the current snapshot. so, this is outside of current scope.
				continue
			}
			if shard.Stage != RecoveryStageDone {
				return plugin.InProgress, nil
			}
		}
	}
	if !recoveryInitiated {
		return plugin.InProgress, nil
	}
	return plugin.Completed, nil
}
