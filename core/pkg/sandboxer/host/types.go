package host

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Host struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec HostSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type HostSpec struct {
	RootDir string `json:"rootDir" yaml:"rootDir"`
	Path    string `json:"path,omitempty" yaml:"path,omitempty"`
}
