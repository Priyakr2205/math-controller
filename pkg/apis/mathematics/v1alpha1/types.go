package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Math struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MathSpec   `json:"spec"`
	Status MathStatus `json:"status"`
}

type MathSpec struct {
	Number1   *int32 `json:"number1"`
	Number2   *int32 `json:"number2"`
	Operation string `json:"operation"`
}

type MathStatus struct {
	LastUpdateTime string `json:"lastUpdateTime"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FooList is a list of Foo resources
type MathList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Math `json:"items"`
}
