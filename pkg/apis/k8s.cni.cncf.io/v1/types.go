package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const GroupName = "k8s.cni.cncf.io"
const GroupVersion = "v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type InterfaceMap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec []InterfaceMapSpec `json:"spec"`
}

type InterfaceMapSpec struct {
	Interface string `json:"interface"`
	Network   string `json:"network"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type InterfaceMapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []InterfaceMap `json:"items"`
}
