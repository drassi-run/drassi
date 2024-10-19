package docker

import (
	"drassi.run/core/pkg/container"
	dockertypes "github.com/docker/docker/api/types"
	dockermount "github.com/docker/docker/api/types/mount"
)

type containerSpec struct {
	Spec container.ContainerSpec
}

func (cs *containerSpec) From(info dockertypes.ContainerJSON) error {
	cs.Spec.Name = info.Name
	cs.Spec.Entrypoint = []string{info.Path}
	cs.Spec.Command = info.Args
	cs.Spec.Image = info.Image
	cs.Spec.Hostname = info.Config.Hostname

	cs.setMounts(info)

	return nil
}

func (cs *containerSpec) setMounts(info dockertypes.ContainerJSON) {
	for _, m := range info.Mounts {
		mount := &container.Mount{
			Type:     string(m.Type),
			Target:   m.Destination,
			ReadOnly: !m.RW,
		}
		if m.Type == dockermount.TypeVolume {
			mount.Source = m.Name
			if driver := m.Driver; driver != "" {
				mount.VolumeOptions = &container.VolumeOptions{
					Driver: driver,
				}
			}
		} else {
			mount.Source = m.Source
		}

		cs.Spec.Mounts = append(cs.Spec.Mounts, mount)
	}
}
