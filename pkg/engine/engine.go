package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/kubemove/kubemove/pkg/apis/kubemove/v1alpha1"
	pair "github.com/kubemove/kubemove/pkg/pair"
	"github.com/pkg/errors"
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
	groupResource []string //nolint
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
		MEngine:            *me,
	}
	if me == nil {
		return nil, fmt.Errorf("failed to cretate MoveEngine clients. Reason: MoveEngine instance is nil")
	}
	// Add remote cluster specific options. This should run only in the source cluster.
	if me.Spec.Mode == MoveEngineActive {
		err := opt.setRemoteOptions(me.Spec.MovePair)
		if err != nil {
			return nil, err
		}
	}
	return opt, nil
}

func (m *MoveEngineAction) ParseResourceEngine(mov *v1alpha1.MoveEngine) error {
	if err := m.setRemoteOptions(mov.Spec.MovePair); err != nil {
		return errors.Errorf("Failed to initialize client.. %v", err)
	}
	m.MEngine = *mov

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

//nolint:unused,deadcode
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
		m.log.Error(err, "Failed to fetch DiscoveryClient")
		return err
	}

	if m.remoteClient, err = pair.FetchPairClient(mpair); err != nil {
		m.log.Error(err, "Failed to fetch PairClient")
		return err
	}

	if m.remotedClient, err = pair.FetchPairDiscoveryClient(mpair); err != nil {
		m.log.Error(err, "Failed to fetch PairDiscoveryClient")
		return err
	}

	rgr, err := restmapper.GetAPIGroupResources(m.remotedClient)
	if err != nil {
		m.log.Error(err, "Failed to get APIGroupResources")
		return err
	}
	m.remoteMapper = restmapper.NewDiscoveryRESTMapper(rgr)

	if m.remotedyClient, err = pair.FetchPairDynamicClient(mpair); err != nil {
		m.log.Error(err, "Failed to fetch PairDynamicClient")
		return err
	}
	return nil
}

// SetMoveEngineState set status of the respective MoveEngine CR
func (m *MoveEngineAction) SetMoveEngineState(state v1alpha1.MoveEngineState) error {
	return m.UpdateMoveEngineStatus(&v1alpha1.MoveEngineStatus{Status: state})
}

func (m *MoveEngineAction) UpdateMoveEngineStatus(newStatus *v1alpha1.MoveEngineStatus) error {
	if newStatus == nil {
		return fmt.Errorf("failed to update MoveEngine %s/%s Status. Reason: newStatus is nil", m.MEngine.Namespace, m.MEngine.Name)
	}

	if newStatus.Status != "" && newStatus.Status != m.MEngine.Status.Status {
		m.MEngine.Status.LastStatus = m.MEngine.Status.Status
		m.MEngine.Status.Status = newStatus.Status
	}

	if newStatus.SyncedTime != nil {
		m.MEngine.Status.LastSyncedTime = m.MEngine.Status.SyncedTime
		m.MEngine.Status.SyncedTime = newStatus.SyncedTime
	}

	if newStatus.DataSync != "" {
		m.MEngine.Status.DataSync = newStatus.DataSync
	}

	if newStatus.DataSyncStatus != "" && newStatus.DataSyncStatus != m.MEngine.Status.DataSyncStatus {
		m.MEngine.Status.DataSyncStatus = newStatus.DataSyncStatus
	}
	return m.client.Update(context.TODO(), &m.MEngine)
}

func (m *MoveEngineAction) UpdateStandbyMoveEngineStatus(newStatus *v1alpha1.MoveEngineStatus) error {
	if newStatus == nil {
		return fmt.Errorf("failed to update MoveEngine %s/%s Status. Reason: newStatus is nil", m.MEngine.Namespace, m.MEngine.Name)
	}

	standbyMoveEngine, err := m.getStandbyMoveEngine()
	if err != nil {
		return err
	}

	if newStatus.Status != "" && newStatus.Status != standbyMoveEngine.Status.Status {
		standbyMoveEngine.Status.LastStatus = standbyMoveEngine.Status.Status
		standbyMoveEngine.Status.Status = newStatus.Status
	}

	if newStatus.SyncedTime != nil {
		standbyMoveEngine.Status.LastSyncedTime = standbyMoveEngine.Status.SyncedTime
		standbyMoveEngine.Status.SyncedTime = newStatus.SyncedTime
	}

	if newStatus.DataSync != "" {
		standbyMoveEngine.Status.DataSync = newStatus.DataSync
	}

	if newStatus.DataSyncStatus != "" && newStatus.DataSyncStatus != standbyMoveEngine.Status.DataSyncStatus {
		standbyMoveEngine.Status.DataSyncStatus = newStatus.DataSyncStatus
	}
	return m.remoteClient.Update(context.TODO(), standbyMoveEngine)
}

func (m *MoveEngineAction) SyncDataSyncStatus() (bool, error) {
	ds, err := GetDataSync(m.client, m.MEngine.Status.DataSync, m.MEngine.Namespace)
	if err != nil {
		return false, err
	}

	requeue := false
	var syncStatus v1alpha1.DataSyncPhase
	switch ds.Status.Status {
	case v1alpha1.SyncStateRunning:
		syncStatus = v1alpha1.SyncPhaseRunning
		requeue = true
	case v1alpha1.SyncStateCompleted:
		syncStatus = v1alpha1.SyncPhaseCompleted
	case v1alpha1.SyncStateFailed:
		syncStatus = v1alpha1.SyncPhaseFailed
	default:
		return requeue, fmt.Errorf("invalid DataSync Status for DataSync: %s/%s", ds.Namespace, ds.Name)
	}
	return requeue, m.UpdateMoveEngineStatus(&v1alpha1.MoveEngineStatus{DataSyncStatus: syncStatus})
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
	standbyMoveEngine.Spec.Mode = MoveEngineStandby

	// Create the MoveEngine in the destination cluster
	return m.remoteClient.Create(context.TODO(), standbyMoveEngine)
}

// GetStandbyMoveEngineStatus returns the status of standby MoveEngine from remote cluster
func (m *MoveEngineAction) GetStandbyMoveEngineStatus() (v1alpha1.MoveEngineStatus, error) {
	standbyMoveEngine, err := m.getStandbyMoveEngine()
	if err != nil {
		return v1alpha1.MoveEngineStatus{}, err
	}
	return standbyMoveEngine.Status, nil
}

func (m *MoveEngineAction) getStandbyMoveEngine() (*v1alpha1.MoveEngine, error) {
	standbyMoveEngine := &v1alpha1.MoveEngine{}
	key := client.ObjectKey{Name: m.MEngine.Name, Namespace: m.MEngine.Spec.RemoteNamespace}

	// Get the standby MoveEngine from remote cluster
	err := m.remoteClient.Get(context.TODO(), key, standbyMoveEngine)
	if err != nil {
		return standbyMoveEngine, err
	}
	return standbyMoveEngine, nil
}
