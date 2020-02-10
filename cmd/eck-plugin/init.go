package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type RepoCreateRequestBodyTemplate struct {
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
func (d *eckDDM) registerSnapshotRepository(params PluginParameters, mode string) error {

	bodyTemplate := RepoCreateRequestBodyTemplate{
		Type: params.Repository.Type,
	}

	if params.Repository.Prefix != "" {
		bodyTemplate.Settings.BasePath = params.Repository.Prefix
	}
	if params.Repository.Endpoint != "" {
		bodyTemplate.Settings.Endpoint = params.Repository.Endpoint
	}
	if params.Repository.Scheme != "" {
		bodyTemplate.Settings.Protocol = params.Repository.Scheme
	}
	if mode != "active" {
		bodyTemplate.Settings.ReadOnly = "true"
	}

	body, err := json.Marshal(bodyTemplate)
	if err != nil {
		return err
	}

	repo := esapi.SnapshotCreateRepositoryRequest{
		Repository: params.Repository.Name,
		Body:       bytes.NewReader(body),
	}

	// TODO: Add Authorization Headers
	// TODO: Add TLS Certificates

	resp, err := repo.Do(context.Background(), nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var x RepoCreateRequestRespone
	err = json.NewDecoder(resp.Body).Decode(&x)
	if err != nil {
		return err
	}
	if !x.Accpeted {
		return fmt.Errorf("failed to create Repository")
	}
	return nil
}
