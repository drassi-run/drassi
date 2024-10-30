package docker

import (
	"fmt"

	"drassi.run/core/pkg/container/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockernetwork "github.com/docker/docker/api/types/network"
)

type containerConfig struct {
	Config           *dockercontainer.Config
	HostConfig       *dockercontainer.HostConfig
	NetworkingConfig *dockernetwork.NetworkingConfig
}

func (cc *containerConfig) From(spec *types.ContainerSpec, stdio *types.Stdio) error {
	cc.Config = &dockercontainer.Config{
		Tty:          stdio.Tty,
		OpenStdin:    stdio.Interactive,
		AttachStdin:  stdio.AttachStdin(),
		AttachStdout: stdio.AttachStdout(),
		AttachStderr: stdio.AttachStderr(),
		Env:          convertMapToKVString(spec.Environment),
		Cmd:          spec.Command,
		Image:        spec.Image,
		Volumes:      nil,
		Entrypoint:   spec.Entrypoint,
		WorkingDir:   spec.WorkingDir,
		Labels:       spec.Labels,
	}

	cc.HostConfig = &dockercontainer.HostConfig{
		Annotations: spec.Annotations,
	}

	if err := cc.setResources(&spec.ContainerResource); err != nil {
		return err
	}
	cc.setStorage(&spec.ContainerStorage)
	cc.setNetwork(&spec.ContainerNetwork)
	cc.setRuntime(&spec.ContainerRuntime)
	cc.setSecurity(&spec.ContainerSecurity)

	return nil
}

func convertMapToKVString(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}

	r := make([]string, 0, len(m))
	for k, v := range m {
		r = append(r, fmt.Sprintf("%s=%s", k, v))
	}
	return r
}
