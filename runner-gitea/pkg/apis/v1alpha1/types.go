package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GiteaRunner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GiteaRunnerSpec `json:"spec,omitempty"`
}

type GiteaRunnerSpec struct {
	UUID                  string   `json:"uuid,omitempty"`
	Token                 string   `json:"token,omitempty"`
	Address               string   `json:"address,omitempty"`
	InsecureSkipTLSVerify bool     `json:"insecureSkipTLSVerify,omitempty"`
	RunnerLabels          []string `json:"runnerLabels,omitempty"`
	// +default=5
	Concurrency int `json:"concurrency,omitempty"`

	SandboxerRef corev1.TypedLocalObjectReference `json:"sandboxerRef,omitempty"`
}
