package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

//======================================== Developer Note ===================================================
// After updating any fields of this file you must run the code generator.
// To prepare code generator/developer environment, run: make dev-image REGISTRY=<your docker registry>
// Then, each time you modify any field, run: make gen REGISTRY=<your docker registry>
// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
//============================================================================================================

type SyncMode string

const (
	SyncModeBackup  SyncMode = "backup"
	SyncModeRestore SyncMode = "restore"
)

// DataSyncSpec defines the desired state of DataSync
// +k8s:openapi-gen=true
type DataSyncSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Namespace      string            `json:"namespace"`
	PluginProvider string            `json:"plugin"`
	MoveEngine     string            `json:"moveEngine"`
	Mode           SyncMode          `json:"mode"`
	Config         map[string]string `json:"config,omitempty"`
}

type SyncState string

const (
	SyncStateRunning   SyncState = "Running"
	SyncStateCompleted SyncState = "Completed"
	SyncStateFailed    SyncState = "Failed"
)

// DataSyncStatus defines the observed state of DataSync
// +k8s:openapi-gen=true
type DataSyncStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Stage          string       `json:"stage"`
	Status         SyncState    `json:"status"`
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	Reason         string       `json:"reason,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataSync is the Schema for the datasyncs API
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="MoveEngine",type="string",JSONPath=".spec.moveEngine"
// +kubebuilder:printcolumn:name="Mode",type="string",JSONPath=".spec.mode"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
// +kubebuilder:printcolumn:name="Completion-Time",type="date",JSONPath=".status.completionTime"
type DataSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataSyncSpec   `json:"spec,omitempty"`
	Status DataSyncStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataSyncList contains a list of DataSync
type DataSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataSync{}, &DataSyncList{})
}
