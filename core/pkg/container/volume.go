package container

// https://github.com/moby/moby/blob/v27.3.1/api/types/volume/create_options.go#L10-L29
// https://github.com/containers/podman/blob/v5.2.4/pkg/domain/entities/types/volumes.go#L8-L21
type VolumeSpec struct {
	Name   string
	Labels map[string]string

	Driver  string
	Options map[string]string
}
