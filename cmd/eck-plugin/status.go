package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/kubemove/kubemove/pkg/plugin/ddm/plugin"
)

const (
	SnapshotSucceeded  = "SUCCESS"
	SnapshotInProgress = "IN_PROGRESS"
	SnapshotFailed     = "FAILED"
)

type SnapshotGetResponse struct {
	Snasphots []SnapshotStatus `json:"snapshots"`
}

type SnapshotStatus struct {
	Snapshot string `json:"snapshot"`
	State    string `json:"state"`
}

type IndexRecoveryResponse struct {

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
	resp, err := snapshotGetRequest.Do(context.Background(), esClient)
	if err != nil {
		return plugin.Errored, err
	}
	defer resp.Body.Close()

	fmt.Println("Response: ", resp.String())

	if resp.StatusCode != http.StatusOK {
		// TODO: parse response and return failure case
		return plugin.Errored, fmt.Errorf("failed to create snapshot. Reason: ")
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

	for _, s := range statusResponse.Snasphots {
		fmt.Println("Snapshot: ", s.Snapshot, " State: ", s.State)
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

	// configure cat recovery request
	indexRecoveryRequest := esapi.IndicesRecoveryRequest{
		Index:      nil,
		Pretty:     true,
	}

	// get recover status
	resp, err := indexRecoveryRequest.Do(context.Background(), esClient)
	if err != nil {
		return plugin.Errored, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// TODO: parse response and return failure case
		return plugin.Errored, fmt.Errorf("failed to create snapshot. Reason: ")
	}

	fmt.Println(resp.String())
	return plugin.Completed, nil
}
