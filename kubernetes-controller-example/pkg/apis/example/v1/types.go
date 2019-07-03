package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Dog describes a Dog resource.
type Dog struct {
	metav1.TypeMeta	`json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behaviour of the Dog
	Spec DogSpec	`json:"spec"`
}

type DogSpec struct {
	Name string `json:"name"`
}

// DogList is a list of Dog
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DogList is a list of Dog Resources
type DogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Dog `json:"items"`
}
