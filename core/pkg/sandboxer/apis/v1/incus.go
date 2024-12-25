package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type IncusSandboxer struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec IncusSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type IncusSpec struct {
	Endpoint string        `json:"endpoint" yaml:"endpoint"`
	Template IncusTemplate `json:"template,omitempty" yaml:"template,omitempty"`
}

// IncusTemplate
// [github.com/lxc/incus/v6/shared/api.InstancesPost]
type IncusTemplate struct {
	// OCI image name, e.g: ghcr.io/drassi-run/ubuntu:22.04
	Image string `json:"source,omitempty" yaml:"source,omitempty"`

	// Instance architecture, e.g: x86_64
	Architecture string `json:"architecture,omitempty" yaml:"architecture,omitempty"`

	// Cloud instance size (AWS, GCP, Azure, ...) to emulate with limits
	// Example: t1.micro
	InstanceSize string `json:"instance_size,omitempty" yaml:"instance_size,omitempty"`

	// List of profiles applied to the instance
	// Example: ["default"]
	Profiles []string `json:"profiles,omitempty" yaml:"profiles,omitempty"`

	// Instance configuration (see https://linuxcontainers.org/incus/docs/main/instances/)
	// Example: {"security.nesting": "true"}
	Config map[string]string `json:"config,omitempty" yaml:"config,omitempty"`

	// Instance devices (see https://linuxcontainers.org/incus/docs/main/instances/)
	// Example: {"root": {"type": "disk", "pool": "default", "path": "/"}}
	Devices map[string]map[string]string `json:"devices,omitempty" yaml:"devices,omitempty"`

	// Whether the instance is ephemeral (deleted on shutdown)
	// Example: false
	Ephemeral bool `json:"ephemeral,omitempty" yaml:"ephemeral,omitempty"`
}
