package engine

import (
	"encoding/json"
	"fmt"
	"os"

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
	mov                v1alpha1.MoveEngine
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
	sr := make(map[schema.GroupVersionResource]NamedObj)
	syncedMap := make(map[MResources]v1alpha1.ResourceStatus)
	er := make(map[MResources]unstructured.Unstructured)
	v := make(map[MResources]unstructured.Unstructured)
	sv := make(map[MResources]unstructured.Unstructured)
	return &MoveEngineAction{
		log:                log,
		client:             c,
		resourcesMap:       sr,
		exposedResourceMap: er,
		volMap:             v,
		stsVolMap:          sv,
		multiAPIResources:  NewMultiAPIResources(),
		syncedResourceMap:  syncedMap,
		discoveryHelper:    h,
	}
}

func (m *MoveEngineAction) ParseResourceEngine(mov *v1alpha1.MoveEngine) error {
	pairObj, err := pair.Get(
		client.ObjectKey{
			Namespace: os.Getenv("WATCH_NAMESPACE"),
			Name:      mov.Spec.MovePair},
		m.client)
	if err != nil {
		fmt.Printf("Failed to fetch movePair %v.. %v\n", mov.Spec.MovePair, err)
		return err
	}

	if err := m.updateClient(pairObj); err != nil {
		return errors.Errorf("Failed to initialize client.. %v", err)
	}

	ls, err := metav1.LabelSelectorAsSelector(mov.Spec.Selectors)
	if err != nil {
		fmt.Printf("Failed to parse label selector %v\n", err)
		return err
	}

	m.mov = *mov
	m.selector = ls

	err = m.UpdateSyncResourceList()
	if err != nil {
		return err
	}

	err = m.syncResourceList()
	if err != nil {
		return err
	}

	m.resetMultiAPIResources()
	return nil
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
