package types

type ContainerSpec struct {
	Name       string
	Image      string
	PullPolicy string

	Command     []string
	Entrypoint  []string
	WorkingDir  string
	Environment map[string]string
	Labels      map[string]string
	Annotations map[string]string

	ContainerNetwork
	ContainerStorage
	Devices           []string
	DeviceCgroupRules []string

	ContainerRuntime
	ContainerResource
	ContainerSecurity
}
