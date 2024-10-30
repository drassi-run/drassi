package docker

import (
	"drassi.run/core/pkg/container/types"
	"github.com/docker/cli/opts"
	dockertypes "github.com/docker/docker/api/types"
	dockermount "github.com/docker/docker/api/types/mount"
)

type containerSpec struct {
	Spec types.ContainerSpec
}

func (cs *containerSpec) From(info dockertypes.ContainerJSON) error {
	cs.Spec.Name = info.Name
	cs.Spec.Entrypoint = []string{info.Path}
	cs.Spec.Command = info.Args
	cs.Spec.Image = info.Image
	cs.Spec.DNS.HostName = info.Config.Hostname
	cs.Spec.DNS.DomainName = info.Config.Domainname
	cs.Spec.Environment = opts.ConvertKVStringsToMap(info.Config.Env)

	cs.setMounts(info)

	return nil
}

func (cs *containerSpec) setMounts(info dockertypes.ContainerJSON) {
	for _, m := range info.Mounts {
		mount := &types.Mount{
			Type:     string(m.Type),
			Target:   m.Destination,
			ReadOnly: !m.RW,
		}
		if m.Type == dockermount.TypeVolume {
			mount.Source = m.Name
			if driver := m.Driver; driver != "" {
				mount.VolumeOptions = &types.VolumeOptions{
					Driver: driver,
				}
			}
		} else {
			mount.Source = m.Source
		}

		cs.Spec.Mounts = append(cs.Spec.Mounts, mount)
	}
}
