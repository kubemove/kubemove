package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

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

const (
	MoveEngineActive  = "active"
	MoveEngineStandby = "standby"
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
		"daemonsets": {
			groupResource: []string{"extensions", "apps"},
			seen:          false,
		},

		"deployments": {
			groupResource: []string{"extensions", "apps"},
			seen:          false,
		},

		"ingresses": {
			groupResource: []string{"extensions", "networking.k8s.io"},
			seen:          false,
		},

		"networkpolicies": {
			groupResource: []string{"extensions", "networking.k8s.io"},
			seen:          false,
		},

		"podsecuritypolicies": {
			groupResource: []string{"extensions", "policy"},
			seen:          false,
		},

		"replicasets": {
			groupResource: []string{"extensions", "apps"},
			seen:          false,
		},
	}

}
func NewMoveEngineAction(log logr.Logger, c client.Client, h helper.Helper, me *v1alpha1.MoveEngine) (*MoveEngineAction, error) {
	resourcesMap := make(map[schema.GroupVersionResource]NamedObj)
	syncedMap := make(map[MResources]v1alpha1.ResourceStatus)
	exposedResourceMap := make(map[MResources]unstructured.Unstructured)
	volumeMap := make(map[MResources]unstructured.Unstructured)
	syncedVolume := make(map[MResources]v1alpha1.VolumeStatus)
	stsVolumeMap := make(map[MResources]unstructured.Unstructured)

	opt := &MoveEngineAction{
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
	if me == nil {
		return nil, fmt.Errorf("failed to cretate MoveEngine clients. Reason: MoveEngine instance is nil")
	}
	// Add remote cluster specific options
	err := opt.setRemoteOptions(me.Spec.MovePair)
	if err != nil {
		return nil, err
	}
	return opt, nil
}

func (m *MoveEngineAction) ParseResourceEngine(mov *v1alpha1.MoveEngine) error {
	err := m.UpdateSyncResourceList()
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
func (m *MoveEngineAction) setRemoteOptions(mpairName string) error {
	mpair, err := pair.Get(
		client.ObjectKey{
			Namespace: os.Getenv("WATCH_NAMESPACE"),
			Name:      mpairName,
		},
		m.client)
	if err != nil {
		m.log.Error(err, "Failed to fetch movePair %v.. %v\n", mpairName)
		return err
	}

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

// SetMoveEngineState set status of the respective MoveEngine CR
func (m *MoveEngineAction) SetMoveEngineState(state v1alpha1.MoveEngineState) error {
	return m.UpdateMoveEngineStatus(&v1alpha1.MoveEngineStatus{Status: state}, "")
}

func (m *MoveEngineAction) UpdateMoveEngineStatus(newStatus *v1alpha1.MoveEngineStatus, ds string) error {
	lastStatus := m.MEngine.Status

	// If status has been updated, then set the new status and update lastStatus accordingly
	if newStatus != nil &&
		newStatus.Status != "" &&
		newStatus.Status != lastStatus.Status {
		m.MEngine.Status.Status = newStatus.Status
		m.MEngine.Status.LastStatus = lastStatus.Status
	}

	// If a new sync schedule has appeared, then set the synced time and update last syncTime accordingly
	if ds != "" {
		// Update the latest DataSync CR name
		m.MEngine.Status.DataSync = ds
		// Update Sync time
		m.MEngine.Status.SyncedTime = metav1.Time{Time: time.Now()}
		m.MEngine.Status.LastSyncedTime = lastStatus.SyncedTime
	}
	return m.client.Update(context.TODO(), &m.MEngine)
}

// CreateStandbyMoveEngine creates MoveEngine in the destination cluster with standby mode.
func (m *MoveEngineAction) CreateStandbyMoveEngine() error {
	standbyMoveEngine := &v1alpha1.MoveEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.MEngine.Name,
			Namespace: m.MEngine.Spec.RemoteNamespace,
		},
		Spec: m.MEngine.Spec,
	}
	// MoveEngine in the destination cluster should be in "standby" mode.
	m.MEngine.Spec.Mode = MoveEngineStandby

	// Create the MoveEngine in the destination cluster
	return m.remoteClient.Create(context.TODO(), standbyMoveEngine)
}

// GetStandbyMoveEngineStatus returns the status of standby MoveEngine from remote cluster
func (m *MoveEngineAction) GetStandbyMoveEngineStatus() (v1alpha1.MoveEngineState, error) {
	standbyMoveEngine := &v1alpha1.MoveEngine{}
	key := client.ObjectKey{Name: m.MEngine.Name, Namespace: m.MEngine.Spec.RemoteNamespace}

	// Get the standby MoveEngine from remote cluster
	err := m.remoteClient.Get(context.TODO(), key, standbyMoveEngine)
	if err != nil {
		return "", err
	}
	return standbyMoveEngine.Status.Status, nil
}
