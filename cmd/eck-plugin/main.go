package main

import (
	"fmt"

	"github.com/appscode/go/crypto/rand"
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
	Port        int32  `json:"port"`
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

	err = client.Register(rand.WithUniqSuffix("eck-plugin"),
		&eckDDM{
			log:       log,
			k8sClient: kubernetes.NewForConfigOrDie(config),
			dmClient:  dynamic.NewForConfigOrDie(config),
		})
	if err != nil {
		panic(err)
	}
}

func (d *eckDDM) Init(params map[string]string) error {
	fmt.Println("\n\nHandling INIT request..................")
	// extract the plugin parameters from the respective MoveEngine
	pluginParameters, mode, err := extractPluginParameters(d.dmClient, params)
	if err != nil {
		return err
	}

	// Register Snapshot Repository in the Elasticsearch
	return registerSnapshotRepository(d.k8sClient, pluginParameters, mode)
}

func (d *eckDDM) Sync(params map[string]string, vol []*client.Volume) (string, error) {
	fmt.Println("\n\nHandling SYNC request..................")
	// extract the plugin parameters from the respective MoveEngine
	pluginParameters, mode, err := extractPluginParameters(d.dmClient, params)
	if err != nil {
		return "", err
	}

	snapshotName, found := params[KeySnapshotName]
	if !found {
		return "", fmt.Errorf("snapshot name not found in the parameters")
	}

	// if it is source cluster, then trigger a snapshot
	if mode == EngineModeActive {
		fmt.Println("Triggering Snapshot: ",snapshotName)
		return triggerSnapshot(d.k8sClient, pluginParameters, snapshotName)
	} else {
		fmt.Println("Triggering Restore from Snapshot: ",snapshotName)
		// if it is destination cluster, then trigger a restore
		return triggerRestore(d.k8sClient, pluginParameters, snapshotName)
	}
}

func (d *eckDDM) Status(params map[string]string) (int32, error) {
	fmt.Println("\n\nHandling STATUS request..................")
	// extract the plugin parameters from the respective MoveEngine
	pluginParameters, mode, err := extractPluginParameters(d.dmClient, params)
	if err != nil {
		return client.Errored, err
	}

	snapshotName, found := params[KeySnapshotName]
	if !found {
		return client.Errored, fmt.Errorf("snapshot name not found in the parameters")
	}

	// if it is active cluster, then send the backup state
	if mode == EngineModeActive {
		fmt.Println("Checking Backup Status........")
		return retrieveBackupState(d.k8sClient, pluginParameters, snapshotName)
	} else {
		fmt.Println("Checking Restore Status........")
		// if it is destination cluster, then send restore status
		return retrieveRestoreState(d.k8sClient, pluginParameters, snapshotName)
	}
}
