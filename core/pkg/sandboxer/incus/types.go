package incus

import (
	"github.com/lxc/incus/v6/shared/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Incus struct {
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
	// Creation source
	Source api.InstanceSource `json:"source,omitempty" yaml:"source,omitempty"`

	// Type (container or virtual-machine)
	// Example: container
	Type api.InstanceType `json:"type" yaml:"type"`

	// Cloud instance size (AWS, GCP, Azure, ...) to emulate with limits
	// Example: t1.micro
	InstanceSize string `json:"instance_size,omitempty" yaml:"instance_size,omitempty"`

	// Architecture name
	// Example: x86_64
	Architecture string `json:"architecture,omitempty" yaml:"architecture,omitempty"`

	// Instance configuration (see https://linuxcontainers.org/incus/docs/main/instances/)
	// Example: {"security.nesting": "true"}
	Config map[string]string `json:"config,omitempty" yaml:"config,omitempty"`

	// Instance devices (see https://linuxcontainers.org/incus/docs/main/instances/)
	// Example: {"root": {"type": "disk", "pool": "default", "path": "/"}}
	Devices map[string]map[string]string `json:"devices,omitempty" yaml:"devices,omitempty"`

	// Whether the instance is ephemeral (deleted on shutdown)
	// Example: false
	Ephemeral bool `json:"ephemeral,omitempty" yaml:"ephemeral,omitempty"`

	// List of profiles applied to the instance
	// Example: ["default"]
	Profiles []string `json:"profiles,omitempty" yaml:"profiles,omitempty"`

	// If set, instance will be restored to the provided snapshot name
	// Example: snap0
	Restore string `json:"restore,omitempty" yaml:"restore,omitempty"`

	// Whether the instance currently has saved state on disk
	// Example: false
	Stateful bool `json:"stateful,omitempty" yaml:"stateful,omitempty"`

	// Instance description
	// Example: My test instance
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}
