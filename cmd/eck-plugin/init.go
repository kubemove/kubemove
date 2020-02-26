package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type RepoCreateRequestBody struct {
	Type     string             `json:"type"`
	Settings RepositorySettings `json:"settings"`
}

type RepositorySettings struct {
	Bucket   string `json:"bucket"`
	BasePath string `json:"base_path,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	ReadOnly string `json:"read_only,omitempty"`
}

type RepoCreateRequestRespone struct {
	Accpeted bool `json:"accpeted"`
}

// registerSnapshotRepository hits Snapshot API of Elasticsearch to register a repository
func registerSnapshotRepository(k8sClient kubernetes.Interface, params PluginParameters, mode string) error {
	// crate an Elasticsearch client
	esClient, err := newElasticsearchClient(k8sClient, params.Elasticsearch)
	if err != nil {
		return err
	}

	// configure request body
	body := RepoCreateRequestBody{
		Type: params.Repository.Type,
		Settings: RepositorySettings{
			Bucket: params.Repository.Bucket,
		},
	}
	if params.Repository.Prefix != "" {
		body.Settings.BasePath = params.Repository.Prefix
	}
	if params.Repository.Endpoint != "" {
		body.Settings.Endpoint = params.Repository.Endpoint
	}
	if params.Repository.Scheme != "" {
		body.Settings.Protocol = params.Repository.Scheme
	}
	if mode != EngineModeActive {
		body.Settings.ReadOnly = "true"
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// configure snapshot repository create request
	repoRequest := esapi.SnapshotCreateRepositoryRequest{
		Repository:    params.Repository.Name,
		Body:          bytes.NewReader(jsonBody),
	}

	fmt.Println("Registering repository..................")
	// register repository
	resp, err := repoRequest.Do(context.Background(), esClient)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Response: ", resp.String())

	if resp.StatusCode != http.StatusOK {
		// TODO: parse the response body and identify the failure cause
		return fmt.Errorf("failed to register Snapshot Repository to the Elasticsearch. Reason: ")
	}
	return nil
}
