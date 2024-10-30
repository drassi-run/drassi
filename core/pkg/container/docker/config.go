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
		User:         spec.User,
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
		Privileged:   spec.Privileged,
		IpcMode:      dockercontainer.IpcMode(spec.IpcMode),
		NetworkMode:  dockercontainer.NetworkMode(spec.NetworkMode),
		PidMode:      dockercontainer.PidMode(spec.PidMode),
		UTSMode:      dockercontainer.UTSMode(spec.UTSMode),
		UsernsMode:   dockercontainer.UsernsMode(spec.UserMode),
		CgroupnsMode: dockercontainer.CgroupnsMode(spec.CgroupMode),
		CapAdd:       spec.CapAdd,
		CapDrop:      spec.CapDrop,
		GroupAdd:     spec.GroupAdd,
		SecurityOpt:  spec.SecurityOpt,
		Sysctls:      spec.Sysctls,
		//MaskedPaths:   maskedPaths,
		Annotations: spec.Annotations,
	}

	if err := cc.setResources(&spec.ContainerResource); err != nil {
		return err
	}
	cc.setRuntime(&spec.ContainerRuntime)
	cc.setStorage(&spec.ContainerStorage)
	cc.setNetwork(&spec.ContainerNetwork)

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
