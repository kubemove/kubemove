package main

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/client-go/kubernetes"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	client "github.com/kubemove/kubemove/pkg/plugin/ddm/plugin"
)

// retrieveSnapshotState hits the Snapshot API of Elasticsearch to retrieve the backup status of a Snapshot
func retrieveBackupState(k8sClient kubernetes.Interface, params PluginParameters, snapshotName string) (int32, error) {
	// crate an Elasticsearch client
	esClient, err := newElasticsearchClient(k8sClient, params.Elasticsearch)
	if err != nil {
		return client.Errored, err
	}

	// configure snapshot get request
	snapshotGetRequest := esapi.SnapshotGetRequest{
		Repository: params.Repository.Name,
		Snapshot:   []string{snapshotName},
	}

	// get snapshot status
	resp, err := snapshotGetRequest.Do(context.Background(), esClient)
	if err != nil {
		return client.Errored, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		// TODO: parse response and return failure case
		return client.Errored, fmt.Errorf("failed to create snapshot. Reason: ")
	}

	// TODO: Parse the response and return snapshot status
	return client.Completed, nil
}

// retrieveRestoreState hits the Recovery API of Elasticsearch to retrieve the restore status of a ES.
func retrieveRestoreState(k8sClient kubernetes.Interface, params PluginParameters, snapshotName string) (int32, error) {
	// crate an Elasticsearch client
	esClient, err := newElasticsearchClient(k8sClient, params.Elasticsearch)
	if err != nil {
		return client.Errored, err
	}

	// configure cat recovery request
	catRecoveryRequest := esapi.CatRecoveryRequest{
		H: []string{"i", "st", "snap"},
		V: esapi.BoolPtr(true),
	}

	// get recover status
	resp, err := catRecoveryRequest.Do(context.Background(), esClient)
	if err != nil {
		return client.Errored, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		// TODO: parse response and return failure case
		return client.Errored, fmt.Errorf("failed to create snapshot. Reason: ")
	}

	// TODO: Parse the response and return retsore status
	return client.Completed, nil
}
