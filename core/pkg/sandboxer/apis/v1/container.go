package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ContainerSandboxer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ContainerSandboxerSpec `json:"spec,omitempty"`
}

type ContainerSandboxerSpec struct {
	Image string `json:"image,omitempty"`
}
