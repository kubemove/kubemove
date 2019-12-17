package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DataVolume defines volume space for dataSync
type DataVolume struct {
	Namespace       string            `json:"namespace"`
	Name            string            `json:"name"`
	PVC             string            `json:"pvc"`
	RemoteName      string            `json:"remoteName"`
	RemoteNamespace string            `json:"remoteNamespace"`
	Param           map[string]string `json:"param"`
}

// DataSyncSpec defines the desired state of DataSync
// +k8s:openapi-gen=true
type DataSyncSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Volume         []*DataVolume     `json:"volume"`
	Namespace      string            `json:"namespace"`
	PluginProvider string            `json:"plugin"`
	MoveEngine     string            `json:"moveEngine"`
	Backup         bool              `json:"backup"`
	Restore        bool              `json:"restore"`
	Config         map[string]string `json:"config"`
}

// DataSyncStatus defines the observed state of DataSync
// +k8s:openapi-gen=true
type DataSyncStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Stage          string          `json:"stage"`
	Status         string          `json:"status"`
	CompletionTime string          `json:"completionTime"`
	Volumes        []*VolumeStatus `json:"volume"`
	Reason         string          `json:"reason"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataSync is the Schema for the datasyncs API
// +k8s:openapi-gen=true
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
