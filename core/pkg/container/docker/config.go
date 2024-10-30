package docker

import (
	"fmt"
	"time"

	"drassi.run/core/pkg/container/types"
	composetypes "github.com/compose-spec/compose-go/v2/types"
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
		//MacAddress:   spec.MacAddress,
		Entrypoint:  spec.Entrypoint,
		WorkingDir:  spec.WorkingDir,
		Labels:      spec.Labels,
		StopSignal:  spec.StopSignal,
		StopTimeout: cc.stopTimeoutFrom(spec.StopGracePeriod),
	}

	cc.HostConfig = &dockercontainer.HostConfig{
		//AutoRemove:      spec.AutoRemove,
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
		//RestartPolicy:  restartPolicy,
		SecurityOpt: spec.SecurityOpt,
		LogConfig:   cc.logConfigFrom(spec.Logging),
		Isolation:   dockercontainer.Isolation(spec.Isolation),
		Sysctls:     spec.Sysctls,
		Runtime:     spec.Runtime,
		//MaskedPaths:   maskedPaths,
		Annotations: spec.Annotations,
	}

	if hc := spec.HealthCheck; hc != nil {
		cc.setHealCheck(spec.HealthCheck)
	}

	if err := cc.setResources(&spec.ContainerResource); err != nil {
		return err
	}

	cc.setStorage(&spec.ContainerStorage)
	cc.setNetwork(&spec.ContainerNetwork)

	return nil
}

func (cc *containerConfig) logConfigFrom(lc *types.LoggingConfig) dockercontainer.LogConfig {
	if lc == nil {
		return dockercontainer.LogConfig{}
	}
	return dockercontainer.LogConfig{
		Type:   lc.Driver,
		Config: lc.Options,
	}
}

func (cc *containerConfig) setHealCheck(hc *types.HealthCheckConfig) {
	cc.Config.Healthcheck = &dockercontainer.HealthConfig{
		Test:          hc.Test,
		Timeout:       hc.Timeout,
		Interval:      hc.Interval,
		Retries:       hc.Retries,
		StartPeriod:   hc.StartPeriod,
		StartInterval: hc.StartInterval,
	}
}

func (cc *containerConfig) stopTimeoutFrom(d *composetypes.Duration) *int {
	if d == nil {
		return nil
	}
	s := int(time.Duration(*d).Seconds())
	return &s
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
