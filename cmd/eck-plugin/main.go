package main

import (
	"fmt"

	"github.com/go-logr/logr"
	client "github.com/kubemove/kubemove/pkg/plugin/ddm/plugin"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	PluginVersion       = "7.x"
	MoveEngineName      = "engineName"
	MoveEngineNamespace = "engineNamespace"

	RepositoryParameters   = "repository"
	ElasticsearchParameter = "elasticsearch"

	EngineModeActive = "active"
	ElasticUser      = "elastic"
	TLSCertKey       = "tls.crt"
	KeySnapshotName  = "snapshotName"
)

type eckDDM struct {
	log       logr.Logger
	p         client.Plugin
	k8sClient kubernetes.Interface
	dmClient  dynamic.Interface
}

type PluginParameters struct {
	Repository    RepositoryOptions    `json:"repository"`
	Elasticsearch ElasticsearchOptions `json:"elasticsearch"`
}

type RepositoryOptions struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Bucket      string `json:"bucket"`
	Prefix      string `json:"prefix,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
	Scheme      string `json:"scheme"`
	Credentials string `json:"credentials"`
}

type ElasticsearchOptions struct {
	ServiceName string `json:"serviceName"`
	Namespace   string `json:"namespace"`
	Scheme      string `json:"scheme"`
	AuthSecret  string `json:"authSecret"`
	TLSSecret   string `json:"tlsSecret,omitempty"`
}

var _ client.Plugin = (*eckDDM)(nil)

func main() {
	var log = logf.Log.WithName(fmt.Sprintf("ECK Plugin %s", PluginVersion))

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, "failed to build kubernetes client")
		utilruntime.Must(err)
	}

	err = client.Register("eck-plugin",
		&eckDDM{
			log:       log,
			k8sClient: kubernetes.NewForConfigOrDie(config),
			dmClient:  dynamic.NewForConfigOrDie(config),
		})
	if err != nil {
		log.Error(err, "Failed to create a plugin")
		return
	}
}

func (d *eckDDM) Init(params map[string]interface{}) error {
	// extract the plugin parameters from the respective MoveEngine
	pluginParameters, mode, err := d.extractPluginParameters(params)
	if err != nil {
		return err
	}

	// Register Snapshot Repository in the Elasticsearch
	return d.registerSnapshotRepository(pluginParameters, mode)
}

func (d *eckDDM) Sync(params map[string]interface{}, vol []*client.Volume) (string, error) {
	// extract the plugin parameters from the respective MoveEngine
	pluginParameters, mode, err := d.extractPluginParameters(params)
	if err != nil {
		return "", err
	}

	val, found := params[KeySnapshotName]
	if !found {
		return "", fmt.Errorf("snapshot name not found in the parameters")
	}
	snapshotName := val.(string)

	// if it is source cluster, then trigger a snapshot
	if mode == EngineModeActive {
		return d.triggerSnapshot(pluginParameters, snapshotName)
	} else {
		// if it is destination cluster, then trigger a restore
		return d.triggerRestore(pluginParameters, snapshotName)
	}
}

func (d *eckDDM) Status(params map[string]interface{}) (int32, error) {
	// extract the plugin parameters from the respective MoveEngine
	pluginParameters, mode, err := d.extractPluginParameters(params)
	if err != nil {
		return client.Errored, err
	}

	val, found := params[KeySnapshotName]
	if !found {
		return client.Errored, fmt.Errorf("snapshot name not found in the parameters")
	}
	snapshotName := val.(string)

	// if it is active cluster, then send the backup state
	if mode == EngineModeActive {
		return d.retrieveBackupState(pluginParameters, snapshotName)
	} else {
		// if it is destination cluster, then send restore status
		return d.retrieveRestoreState(pluginParameters, snapshotName)
	}
}
