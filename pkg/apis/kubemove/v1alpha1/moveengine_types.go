package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

//======================================== Developer Note ===================================================
// After updating any fields of this file you must run the code generator.
// To prepare code generator/developer environment, run: make dev-image REGISTRY=<your docker registry>
// Then, each time you modify any field, run: make gen REGISTRY=<your docker registry>
// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
//============================================================================================================

const (
	ResourceKindMoveEngine     = "MoveEngine"
	ResourceSingularMoveEngine = "moveengine"
	ResourcePluralMoveEngine   = "moveengines"
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
	SyncPeriod       string                `json:"syncPeriod"`
	Mode             string                `json:"mode"`
	PluginProvider   string                `json:"plugin"`
	IncludeResources bool                  `json:"includeResources"`
	PluginParameters *runtime.RawExtension `json:"pluginParameters,omitempty"`
}

type MoveEngineState string

const (
	MoveEngineInitializing         MoveEngineState = "Initializing"
	MoveEngineInitialized          MoveEngineState = "Initialized"
	MoveEngineReady                MoveEngineState = "Ready"
	MoveEngineInitializationFailed MoveEngineState = "InitializationFailed"
	MoveEngineInvalid              MoveEngineState = "Invalid"
)

type DataSyncPhase string

const (
	SyncPhaseRunning   DataSyncPhase = "Running"
	SyncPhaseCompleted DataSyncPhase = "Completed"
	SyncPhaseFailed    DataSyncPhase = "Failed"
	SyncPhaseSynced    DataSyncPhase = "Synced"
)

// MoveEngineStatus defines the observed state of MoveEngine
// +k8s:openapi-gen=true
type MoveEngineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Status         MoveEngineState `json:"status,omitempty"`
	LastStatus     MoveEngineState `json:"lastStatus,omitempty"`
	SyncedTime     *metav1.Time    `json:"syncedTime,omitempty"`
	LastSyncedTime *metav1.Time    `json:"lastSyncedTime,omitempty"`
	DataSync       string          `json:"dataSync,omitempty"`
	DataSyncStatus DataSyncPhase   `json:"dataSyncStatus,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MoveEngine is the Schema for the moveengines API
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Mode",type="string",JSONPath=".spec.mode"
// +kubebuilder:printcolumn:name="Sync-Period",type="string",JSONPath=".spec.syncPeriod"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
// +kubebuilder:printcolumn:name="Data-Sync",type="string",JSONPath=".status.dataSync"
// +kubebuilder:printcolumn:name="Sync-Status",type="string",JSONPath=".status.dataSyncStatus"
// +kubebuilder:printcolumn:name="Synced-Time",type="date",JSONPath=".status.syncedTime"
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
	Status          string      `json:"status"`
	SyncedTime      metav1.Time `json:"synced"`
	LastStatus      string      `json:"lastStatus"`
	LastSyncedTime  metav1.Time `json:"lastSyncedTime"`
	Reason          string      `json:"reason"`
	Volume          string      `json:"volume"`
	RemoteVolume    string      `json:"remoteVolume"`
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
