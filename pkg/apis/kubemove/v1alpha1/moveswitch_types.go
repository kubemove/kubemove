package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MoveSwitchSpec defines the desired state of MoveSwitch
// +k8s:openapi-gen=true
type MoveSwitchSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	MoveEngine string `json:"moveEngine"`
	Active     bool   `json:"active"`
}

// MoveSwitchStatus defines the observed state of MoveSwitch
// +k8s:openapi-gen=true
type MoveSwitchStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Stage  string
	Status string
	Reason string
	Volume []*VolumeStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MoveSwitch is the Schema for the moveswitches API
// +k8s:openapi-gen=true
type MoveSwitch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MoveSwitchSpec   `json:"spec,omitempty"`
	Status MoveSwitchStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MoveSwitchList contains a list of MoveSwitch
type MoveSwitchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MoveSwitch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MoveSwitch{}, &MoveSwitchList{})
}
