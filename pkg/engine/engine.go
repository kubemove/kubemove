package engine

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/go-logr/logr"
	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	pair "github.com/kubemove/kubemove/pkg/pair"
	helper "github.com/vmware-tanzu/velero/pkg/discovery"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MResources struct {
	Name       string
	Kind       string
	APIVersion string
}

type MultiResource struct {
	groupResource []string
	seen          bool
}

type NamedObj map[string]unstructured.Unstructured

type MoveEngineAction struct {
	MEngine            v1alpha1.MoveEngine
	selector           labels.Selector
	client             client.Client
	dclient            *discovery.DiscoveryClient
	remoteClient       client.Client
	remotedClient      *discovery.DiscoveryClient
	remotedyClient     dynamic.Interface
	remoteMapper       meta.RESTMapper
	discoveryHelper    helper.Helper
	multiAPIResources  map[string]*MultiResource
	resourcesMap       map[schema.GroupVersionResource]NamedObj
	resourceList       []unstructured.Unstructured
	exposedResourceMap map[MResources]unstructured.Unstructured
	volMap             map[MResources]unstructured.Unstructured
	stsVolMap          map[MResources]unstructured.Unstructured
	syncedResourceMap  map[MResources]v1alpha1.ResourceStatus
	syncedVolMap       map[MResources]v1alpha1.VolumeStatus
	namespace          string
	plugin             string
	log                logr.Logger
}

// If any API resource have multiple APIVersion then add it here
func NewMultiAPIResources() map[string]*MultiResource {
	return map[string]*MultiResource{
		"daemonsets": &MultiResource{
			groupResource: []string{"extensions", "apps"},
			seen:          false,
		},

		"deployments": &MultiResource{
			groupResource: []string{"extensions", "apps"},
			seen:          false,
		},

		"ingresses": &MultiResource{
			groupResource: []string{"extensions", "networking.k8s.io"},
			seen:          false,
		},

		"networkpolicies": &MultiResource{
			groupResource: []string{"extensions", "networking.k8s.io"},
			seen:          false,
		},

		"podsecuritypolicies": &MultiResource{
			groupResource: []string{"extensions", "policy"},
			seen:          false,
		},

		"replicasets": &MultiResource{
			groupResource: []string{"extensions", "apps"},
			seen:          false,
		},
	}

}
func NewMoveEngineAction(log logr.Logger, c client.Client, h helper.Helper) *MoveEngineAction {
	resourcesMap := make(map[schema.GroupVersionResource]NamedObj)
	syncedMap := make(map[MResources]v1alpha1.ResourceStatus)
	exposedResourceMap := make(map[MResources]unstructured.Unstructured)
	volumeMap := make(map[MResources]unstructured.Unstructured)
	syncedVolume := make(map[MResources]v1alpha1.VolumeStatus)
	stsVolumeMap := make(map[MResources]unstructured.Unstructured)

	return &MoveEngineAction{
		log:                log,
		client:             c,
		resourcesMap:       resourcesMap,
		exposedResourceMap: exposedResourceMap,
		volMap:             volumeMap,
		stsVolMap:          stsVolumeMap,
		multiAPIResources:  NewMultiAPIResources(),
		syncedResourceMap:  syncedMap,
		discoveryHelper:    h,
		syncedVolMap:       syncedVolume,
	}
}

func (m *MoveEngineAction) ParseResourceEngine(mov *v1alpha1.MoveEngine) error {
	pairObj, err := pair.Get(
		client.ObjectKey{
			Namespace: os.Getenv("WATCH_NAMESPACE"),
			Name:      mov.Spec.MovePair},
		m.client)
	if err != nil {
		m.log.Error(err, "Failed to fetch movePair %v.. %v\n", mov.Spec.MovePair)
		return err
	}

	if err := m.updateClient(pairObj); err != nil {
		return errors.Errorf("Failed to initialize client.. %v", err)
	}

	ls, err := metav1.LabelSelectorAsSelector(mov.Spec.Selectors)
	if err != nil {
		m.log.Error(err, "Failed to parse label selector")
		return err
	}

	m.MEngine = *mov
	m.selector = ls

	err = m.UpdateSyncResourceList()
	if err != nil {
		return err
	}

	m.resetMultiAPIResources()

	err = m.syncResourceList()
	return err
}

func (m *MoveEngineAction) resetMultiAPIResources() {
	for _, v := range m.multiAPIResources {
		v.seen = false
	}
}

func convertJSONToUnstructured(data []byte) (unstructured.Unstructured, error) {
	obj := unstructured.Unstructured{}

	om := make(map[string]interface{})
	err := json.Unmarshal(data, &om)
	if err != nil {
		return obj, err
	}

	obj.Object = om
	return obj, nil
}

//TODO improve error reporting
func (m *MoveEngineAction) updateClient(mpair *v1alpha1.MovePair) error {
	var err error

	if m.dclient, err = pair.FetchDiscoveryClient(); err != nil {
		return err
	}

	if m.remoteClient, err = pair.FetchPairClient(mpair); err != nil {
		return err
	}

	if m.remotedClient, err = pair.FetchPairDiscoveryClient(mpair); err != nil {
		return err
	}

	rgr, err := restmapper.GetAPIGroupResources(m.remotedClient)
	if err != nil {
		return err
	}
	m.remoteMapper = restmapper.NewDiscoveryRESTMapper(rgr)

	if m.remotedyClient, err = pair.FetchPairDynamicClient(mpair); err != nil {
		return err
	}
	return nil
}

func (m *MoveEngineAction) UpdateMoveEngineStatus(err error, ds string) error {
	lastStatus := m.MEngine.Status
	newStatus := v1alpha1.MoveEngineStatus{}

	if len(ds) == 0 {
		return errors.New("Empty dataSync")
	}

	if err != nil {
		newStatus.Status = "Errored"
	} else {
		newStatus.Status = "Synced"
	}

	newStatus.SyncedTime = metav1.Time{Time: time.Now()}
	newStatus.LastSyncedTime = lastStatus.SyncedTime
	newStatus.LastStatus = lastStatus.Status

	// resource status
	for _, l := range m.syncedResourceMap {
		r := l
		newStatus.Resources = append(newStatus.Resources, &r)
	}

	// volume status
	for _, l := range m.syncedVolMap {
		r := l
		newStatus.Volumes = append(newStatus.Volumes, &r)
	}

	newStatus.DataSync = ds

	m.MEngine.Status = newStatus
	return m.client.Update(context.TODO(), &m.MEngine)
}
