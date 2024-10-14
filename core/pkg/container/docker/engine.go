package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"drassi.run/core/pkg/container"
	utilio "drassi.run/core/pkg/util/io"
	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerimage "github.com/docker/docker/api/types/image"
	dockernetwork "github.com/docker/docker/api/types/network"
	dockervolume "github.com/docker/docker/api/types/volume"
	dockerclient "github.com/docker/docker/client"
	dockererr "github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
)

type empty = struct{}
type run = func() error

type engine struct {
	client dockerclient.APIClient
}

var defaultOpts = []dockerclient.Opt{
	dockerclient.WithAPIVersionNegotiation(),
	dockerclient.WithUserAgent("drassi"),
}

func New(opts ...dockerclient.Opt) (container.Engine, error) {
	opts = append(defaultOpts, opts...)

	if cli, err := dockerclient.NewClientWithOpts(opts...); err != nil {
		return nil, err
	} else {
		return &engine{client: cli}, nil
	}
}

func (e *engine) Close() error {
	return e.client.Close()
}

func (e *engine) Address() string {
	return e.client.DaemonHost()
}

func (e *engine) ImagePull(ctx context.Context, ref string, opts *container.PullOptions) error {
	var cred string
	if opts.RegistryAuth != nil {
		cred = opts.RegistryAuth.Credential()
	}
	reader, err := e.client.ImagePull(ctx, ref, dockerimage.PullOptions{
		RegistryAuth: cred,
	})
	if err != nil {
		return err
	}

	defer reader.Close()
	// cli.ImagePull is asynchronous.
	// The reader needs to be read completely for the pull operation to complete.
	// If stdout is not required, consider using io.Discard instead of os.Stdout.
	// TODO: show download progress
	_, err = io.Copy(os.Stdout, reader)
	return err
}

func (e *engine) ImageBuild(ctx context.Context, context io.Reader, opts *container.BuildOptions) error {
	//TODO implement me
	panic("implement me")
}

func (e *engine) ContainerRun(ctx context.Context, spec *container.ContainerSpec, opts *container.RunOptions) (string, error) {
	stdio := opts.Stdio
	cc := new(containerConfig)
	if err := cc.From(spec, stdio); err != nil {
		return "", err
	}

	createResp, err := e.client.ContainerCreate(ctx, cc.Config, cc.HostConfig, cc.NetworkingConfig, nil, spec.Name)
	if err != nil {
		return "", err
	}
	ctnID := createResp.ID

	fn := func() error {
		return e.client.ContainerStart(ctx, ctnID, dockercontainer.StartOptions{})
	}
	if stdio.AttachStdin() || stdio.AttachStdout() || stdio.AttachStderr() {
		fn = e.streamingStdio(ctx, ctnID, false, stdio, opts.Streams, fn)
	}
	if stdio.AttachStdout() || stdio.AttachStderr() {
		fn = e.waitFinish(ctx, ctnID, cc.HostConfig.AutoRemove, fn)
	}

	err = fn()
	return ctnID, err
}

func (e *engine) ContainerExec(ctx context.Context, id string, opts *container.ExecOptions) (string, error) {
	stdio := opts.Stdio

	idResp, err := e.client.ContainerExecCreate(ctx, id, dockercontainer.ExecOptions{
		Cmd:          opts.Cmd,
		WorkingDir:   opts.Workdir,
		Env:          convertMapToKVString(opts.Env),
		Tty:          stdio.Tty,
		Detach:       stdio.Detach(),
		AttachStdin:  stdio.AttachStdin(),
		AttachStdout: stdio.AttachStdout(),
		AttachStderr: stdio.AttachStderr(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create exec: %w", err)
	}
	if idResp.ID == "" {
		return "", errors.New("exec ID empty")
	}
	execID := idResp.ID

	fn := func() error { return nil }
	if stdio.AttachStdin() || stdio.AttachStdout() || stdio.AttachStderr() {
		fn = e.streamingStdio(ctx, execID, true, stdio, opts.Streams, fn)
	}
	if stdio.AttachStdout() || stdio.AttachStderr() {
		fn = e.exitCode(ctx, execID, true, fn)
	}

	err = fn()
	return execID, err
}

func (e *engine) ContainerRemove(ctx context.Context, id string) error {
	err := e.client.ContainerStop(ctx, id, dockercontainer.StopOptions{})
	if err != nil {
		if dockererr.IsNotFound(err) {
			return nil
		}
		return err
	}
	err = e.client.ContainerRemove(ctx, id, dockercontainer.RemoveOptions{
		RemoveVolumes: true, // Remove anonymous volumes associated with the container
		RemoveLinks:   true, // Remove the specified link
		Force:         true, // Force the removal of a running container (uses SIGKILL)
	})
	// In the case of `docker run --rm ...`
	if err != nil && !dockererr.IsNotFound(err) && !dockererr.IsConflict(err) {
		return err
	}
	return nil
}

func (e *engine) Stat(ctx context.Context, id string, path string) (fs.FileInfo, error) {
	if ps, err := e.client.ContainerStatPath(ctx, id, path); err != nil {
		return nil, err
	} else {
		return &fileInfo{ps}, nil
	}
}

func (e *engine) CopyIn(ctx context.Context, id string, opts *container.CopyInOptions) error {
	return e.client.CopyToContainer(ctx, id, opts.DestinationPath, opts.Reader, dockercontainer.CopyToContainerOptions{})
}

func (e *engine) CopyOut(ctx context.Context, id string, opts *container.CopyOutOptions) (io.ReadCloser, error) {
	r, _, err := e.client.CopyFromContainer(ctx, id, opts.SourcePath)
	return r, err
}

func (e *engine) NetworkCreate(ctx context.Context, spec *container.NetworkSpec) (string, error) {
	config := dockernetwork.CreateOptions{
		Labels:  spec.Labels,
		Driver:  spec.Driver,
		Options: spec.Options,
	}
	if spec.IPAMDriver != "" || len(spec.IPAMOptions) > 0 {
		config.IPAM = &dockernetwork.IPAM{
			Driver:  spec.IPAMDriver,
			Options: spec.IPAMOptions,
		}
	}

	if res, err := e.client.NetworkCreate(ctx, spec.Name, config); err != nil {
		return "", err
	} else {
		return res.ID, nil
	}
}

func (e *engine) NetworkRemove(ctx context.Context, id string) error {
	return e.client.NetworkRemove(ctx, id)
}

func (e *engine) VolumeCreate(ctx context.Context, spec *container.VolumeSpec) (string, error) {
	config := dockervolume.CreateOptions{
		Name:       spec.Name,
		Labels:     spec.Labels,
		Driver:     spec.Driver,
		DriverOpts: spec.Options,
	}

	if res, err := e.client.VolumeCreate(ctx, config); err != nil {
		return "", err
	} else {
		return res.Name, nil
	}
}

func (e *engine) VolumeRemove(ctx context.Context, id string) error {
	return e.client.VolumeRemove(ctx, id, true)
}

func (e *engine) streamingStdio(ctx context.Context, id string, exec bool, stdio *container.Stdio, streams container.Streams, fn run) run {
	ctx, cancel := context.WithCancel(ctx)

	return func() (err error) {
		defer cancel()

		var hijackedResp dockertypes.HijackedResponse
		if exec {
			hijackedResp, err = e.client.ContainerExecAttach(ctx, id, dockercontainer.ExecAttachOptions{
				Detach: stdio.Detach(),
				Tty:    stdio.Tty,
			})
		} else {
			hijackedResp, err = e.client.ContainerAttach(ctx, id, dockercontainer.AttachOptions{
				Stream: true,
				Stdin:  stdio.AttachStdin(),
				Stdout: stdio.AttachStdout(),
				Stderr: stdio.AttachStderr(),
			})
		}
		if err != nil {
			return
		}

		errC := make(chan error, 1)
		go func() {
			defer close(errC)
			defer hijackedResp.Close()

			reader := utilio.NewContextReader(ctx, hijackedResp.Reader)
			outWriter := streams.Out()
			errWriter := streams.Err()

			var err error
			if stdio.Tty {
				_, err = io.Copy(outWriter, reader)
			} else {
				// https://github.com/moby/moby/blob/v27.3.1/client/container_attach.go#L36
				_, err = stdcopy.StdCopy(outWriter, errWriter, reader)
			}
			errC <- err
		}()

		if err = fn(); err != nil {
			return
		}

		return <-errC
	}
}

func (e *engine) waitFinish(ctx context.Context, id string, autoRemove bool, fn run) run {
	ctx, cancel := context.WithCancel(ctx)
	condition := dockercontainer.WaitConditionNextExit
	if autoRemove {
		condition = dockercontainer.WaitConditionRemoved
	}

	return func() error {
		defer cancel()

		resultC, waitErrC := e.client.ContainerWait(ctx, id, condition)

		errC := make(chan error, 1)
		go func() {
			defer close(errC)

			select {
			case result := <-resultC:
				if result.Error != nil {
					errC <- errors.New(result.Error.Message)
				} else if result.StatusCode != 0 {
					errC <- fmt.Errorf("container exited with status code: %d", result.StatusCode)
				}
			case err := <-waitErrC:
				errC <- err
			}
		}()

		if err := fn(); err != nil {
			return err
		}

		return <-errC
	}
}

func (e *engine) exitCode(ctx context.Context, id string, exec bool, fn run) run {
	// ContainerExec
	if exec {
		return func() error {
			if err := fn(); err != nil {
				return err
			}

			inspectResp, err := e.client.ContainerExecInspect(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to inspect exec: %w", err)
			}
			if ec := inspectResp.ExitCode; ec != 0 {
				err = fmt.Errorf("exitcode '%d': failure", ec)
			}
			return err
		}
	}

	// ContainerRun
	return func() error {
		if err := fn(); err != nil {
			return err
		}

		inspectResp, err := e.client.ContainerInspect(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to inspect container: %w", err)
		}
		if ec := inspectResp.State.ExitCode; ec != 0 {
			err = fmt.Errorf("exitcode '%d': failure", ec)
		}
		return err
	}
}

type fileInfo struct {
	dockercontainer.PathStat
}

func (fi *fileInfo) Name() string {
	return fi.PathStat.Name
}

func (fi *fileInfo) Size() int64 {
	return fi.PathStat.Size
}

func (fi *fileInfo) Mode() fs.FileMode {
	return fi.PathStat.Mode
}

func (fi *fileInfo) ModTime() time.Time {
	return fi.PathStat.Mtime
}

func (fi *fileInfo) IsDir() bool {
	return fi.Mode().IsDir()
}

func (fi *fileInfo) Sys() any {
	return nil
}
