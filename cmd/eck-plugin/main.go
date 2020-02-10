package main

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	client "github.com/kubemove/kubemove/pkg/plugin/ddm/plugin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	PluginVersion       = "7.x"
	MoveEngineName      = "engine-name"
	MoveEngineNamespace = "engine-namespace"

	RepositoryParameters   = "repository"
	ElasticsearchParameter = "elasticsearch"
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
	TlsSecret   string `json:"tlsSecret,omitempty"`
}

var _ client.Plugin = (*eckDDM)(nil)

func main() {
	var log = logf.Log.WithName(fmt.Sprintf("ECK Plugin %s", PluginVersion))

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, "failed to build kubernetes client")
		utilruntime.Must(err)
	}

	err := client.Register("eck-plugin",
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

func (d *eckDDM) Init(params map[string]string) error {
	// Register Snapshot Repository in the Elasticsearch
	engineName, ok := params[MoveEngineName]
	if !ok {
		return fmt.Errorf("MoveEngine name not found in the paramters")
	}

	engineNamespace, ok := params[MoveEngineNamespace]
	if !ok {
		return fmt.Errorf("MoveEngine namespace not found in the paramters")
	}

	pluginParameters, err := d.getPluginParameters(engineName, engineNamespace)
	if err != nil {
		return err
	}

	return d.registerSnapshotRepository(pluginParameters)
}

func (d *eckDDM) Sync(engine string, isSourceCluster bool, params map[string]string, vol []*client.Volume) (string, error) {
	// trigger a snapshot in the source ES in backup phase
	if isSourceCluster {
		return d.triggerSnapshot(params)
	} else {
		// trigger a restore in the destination ES in restore phase
		return d.triggerRestore(params)

	}

	return "what the use of this?", nil
}

func (d *eckDDM) Status(engine, snapshotId string, isSourceCluster bool) (int32, error) {

	// send snapshot state on backup status request
	if isSourceCluster {
		return 0, d.retrieveSnapshotState("should send Snapshot name")
	} else {
		// send restore state on restore status request
		return 0, d.retrieveRestoreState("should send Snapshot name")
	}

	return 0, nil
}

func (d *eckDDM) getPluginParameters(engineName, engineNamespace string) (PluginParameters, error) {
	parameters := PluginParameters{}

	moveEngineGVR := v1alpha1.SchemeGroupVersion.WithResource(v1alpha1.ResourcePluralMoveEngine)

	resp, err := d.dmClient.Resource(moveEngineGVR).Namespace(engineNamespace).Get(engineName, metav1.GetOptions{})
	if err != nil {
		return parameters, err
	}

	var moveEngine v1alpha1.MoveEngine
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), &moveEngine)
	if err != nil {
		return parameters, err
	}

	repoParams, ok := moveEngine.Spec.PluginParameters[RepositoryParameters]
	if !ok {
		return parameters, fmt.Errorf("repository parameters not found in pluginParameters")
	}

	parameters.Repository, ok = repoParams.(RepositoryOptions)
	if !ok {
		return parameters, fmt.Errorf("failed to cast repository prameters into RepositoryOptions")
	}

	esParams, ok := moveEngine.Spec.PluginParameters[ElasticsearchParameter]
	if !ok {
		return parameters, fmt.Errorf("elasticsearch parameters not found in pluginParameters")
	}
	parameters.Elasticsearch, ok = esParams.(ElasticsearchOptions)
	if !ok {
		return parameters, fmt.Errorf("failed to cast elasticsearch prameters into ElasticsearchOptions")
	}

	return parameters, nil
}
