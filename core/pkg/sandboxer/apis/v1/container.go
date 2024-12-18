package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ContainerSandboxer struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec ContainerSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type ContainerSpec struct {
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
}
