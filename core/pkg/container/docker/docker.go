package docker

import (
	"context"
	"errors"
	"fmt"

	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	dockererr "github.com/docker/docker/errdefs"
	"github.com/dungdm93/drassi/core/pkg/container"
)

type docker struct {
	client dockerclient.APIClient
}

func (d *docker) Close() error {
	return d.client.Close()
}

func (d *docker) LaunchContainer(ctx context.Context, req container.LaunchContainerRequest) (container.LaunchContainerResponse, error) {
	return container.LaunchContainerResponse{}, nil
}

func (d *docker) TerminateContainer(ctx context.Context, req container.TerminateContainerRequest) (container.TerminateContainerResponse, error) {
	res := container.TerminateContainerResponse{}
	err := d.client.ContainerStop(ctx, req.ContainerId, dockercontainer.StopOptions{
		Timeout: req.Timeout,
	})
	if err != nil {
		if dockererr.IsNotFound(err) {
			return res, nil
		}
		return res, err
	}
	err = d.client.ContainerRemove(ctx, req.ContainerId, dockercontainer.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	// In the case of `docker run --rm ...`
	if err != nil && !dockererr.IsNotFound(err) && !dockererr.IsConflict(err) {
		return res, err
	}
	return res, nil
}

func (d *docker) ExecuteContainer(ctx context.Context, req container.ExecuteContainerRequest) (container.ExecuteContainerResponse, error) {
	const detach = false // run command in the foreground
	const tty = true

	res := container.ExecuteContainerResponse{}
	idResp, err := d.client.ContainerExecCreate(ctx, req.ContainerId, dockertypes.ExecConfig{
		User:         req.User,
		Cmd:          req.Cmd,
		WorkingDir:   req.Workdir,
		Env:          []string{},
		Tty:          tty,
		Detach:       detach,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return res, fmt.Errorf("failed to create exec: %w", err)
	}
	if idResp.ID == "" {
		return res, errors.New("exec ID empty")
	}

	hijackedResp, err := d.client.ContainerExecAttach(ctx, idResp.ID, dockertypes.ExecStartCheck{
		Detach: detach,
		Tty:    tty,
	})
	if err != nil {
		return res, fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer hijackedResp.Close()

	// TODO: improve handling hijackedIOStreamer (https://github.com/docker/cli/blob/26.0/cli/command/container/exec.go#L188-L198)
	if err := streamingResponse(ctx, hijackedResp, tty); err != nil {
		return res, err
	}
	// TODO: Monitor TTY (https://github.com/docker/cli/blob/26.0/cli/command/container/exec.go#L203)

	inspectResp, err := d.client.ContainerExecInspect(ctx, idResp.ID)
	if err != nil {
		return res, fmt.Errorf("failed to inspect exec: %w", err)
	}
	if inspectResp.ExitCode != 0 {
		return res, fmt.Errorf("exitcode '%d': failure", inspectResp.ExitCode)
	} else {
		return res, nil
	}
}

func (d *docker) CopyFromContainer(ctx context.Context, req container.CopyFromContainerRequest) (container.CopyFromContainerResponse, error) {
	r, _, err := d.client.CopyFromContainer(ctx, req.ContainerId, req.SourcePath)
	return container.CopyFromContainerResponse{Reader: r}, err
}

func (d *docker) CopyToContainer(ctx context.Context, req container.CopyToContainerRequest) (container.CopyToContainerResponse, error) {
	err := d.client.CopyToContainer(ctx, req.ContainerId, req.DestinationPath, req.Content, dockertypes.CopyToContainerOptions{})
	return container.CopyToContainerResponse{}, err
}
