package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	ResourceKindMoveEngine = "MoveEngine"
	ResourceSingularMoveEngine = "moveengine"
	ResourcePluralMoveEngine = "moveengines"
)

// MoveEngineSpec defines the desired state of MoveEngine
// +k8s:openapi-gen=true
type MoveEngineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	MovePair         string                `json:"movePair"`
	Namespace        string                `json:"namespace"`
	RemoteNamespace  string                `json:"remoteNamespace"`
	Selectors        *metav1.LabelSelector `json:"selectors"`
	SyncPeriod       string                `json:"syncPeriod"`
	Mode             string                `json:"mode"`
	PluginProvider   string                `json:"plugin"`
	IncludeResources bool                  `json:"includeResources"`
	PluginParameters *runtime.RawExtension `json:"pluginParameters,omitempty"`
}

// MoveEngineStatus defines the observed state of MoveEngine
// +k8s:openapi-gen=true
type MoveEngineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Status         string            `json:"Status"`
	LastStatus     string            `json:"LastStatus"`
	SyncedTime     metav1.Time       `json:"SyncedTime"`
	LastSyncedTime metav1.Time       `json:"LastSyncedTime"`
	DataSync       string            `json:"DataSync"`
	DataSyncStatus string            `json:"DataSyncStatus"`
	Volumes        []*VolumeStatus   `json:"Volumes"`
	Resources      []*ResourceStatus `json:"Resources"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MoveEngine is the Schema for the moveengines API
// +k8s:openapi-gen=true
type MoveEngine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MoveEngineSpec   `json:"spec,omitempty"`
	Status MoveEngineStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MoveEngineList contains a list of MoveEngine
type MoveEngineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MoveEngine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MoveEngine{}, &MoveEngineList{})
}

// VolumeStatus sync status of volumes
type VolumeStatus struct {
	Namespace       string      `json:"namespace"`
	RemoteNamespace string      `json:"remoteNamespace"`
	PVC             string      `json:"pvc"`
	Status          string      `json:"Status"`
	SyncedTime      metav1.Time `json:"Synced"`
	LastStatus      string      `json:"lastStatus"`
	LastSyncedTime  metav1.Time `json:"lastSyncedTime"`
	Reason          string      `json:"reason"`
	Volume          string      `json:"Volume"`
	RemoteVolume    string      `json:"RemoteVolume"`
}

// ResourceStatus sync status of resource
type ResourceStatus struct {
	Kind       string      `json:"kind"`
	Name       string      `json:"name"`
	Phase      string      `json:"phase"`
	Status     string      `json:"status"`
	Reason     string      `json:"reason"`
	SyncedTime metav1.Time `json:"Synced"`
}
