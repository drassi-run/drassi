package docker

import (
	"drassi.run/core/pkg/container/types"
	"github.com/docker/cli/opts"
	dockertypes "github.com/docker/docker/api/types"
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

	cs.setStorage(info)

	return nil
}
