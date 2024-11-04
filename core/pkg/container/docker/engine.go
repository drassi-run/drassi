package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"
	"time"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/cli"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/util/io"
	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerfilters "github.com/docker/docker/api/types/filters"
	dockerimage "github.com/docker/docker/api/types/image"
	dockernetwork "github.com/docker/docker/api/types/network"
	dockervolume "github.com/docker/docker/api/types/volume"
	dockerclient "github.com/docker/docker/client"
	dockererr "github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
)

// ProxyCommand
//   - [github.com/docker/cli/cli/connhelper.GetConnectionHelper]
func ProxyCommand(host string) []string {
	cmd := []string{"docker"}
	if host != "" {
		cmd = append(cmd, "--host", "unix://"+host)
	}
	cmd = append(cmd, "system", "dial-stdio")
	return cmd
}

type engine struct {
	client dockerclient.APIClient
}

var defaultOpts = []dockerclient.Opt{
	dockerclient.WithAPIVersionNegotiation(),
	dockerclient.WithUserAgent("drassi"),
}

func New(opts ...dockerclient.Opt) (container.Engine, error) {
	opts = append(defaultOpts, opts...)

	if client, err := dockerclient.NewClientWithOpts(opts...); err != nil {
		return nil, err
	} else {
		return &engine{client: client}, nil
	}
}

func (e *engine) Close() error {
	return e.client.Close()
}

func (e *engine) Address() string {
	return e.client.DaemonHost()
}

func (e *engine) ImagePull(ctx context.Context, ref string, opts *container.PullOptions) error {
	switch opts.PullPolicy {
	case "never":
		return nil
	case "", "missing":
		_, _, err := e.client.ImageInspectWithRaw(ctx, ref)
		if err == nil { // image existed
			return nil
		}
		if !dockererr.IsNotFound(err) {
			return err
		}
	case "always":
	default:
		return fmt.Errorf("unknown image pull policy: %s", opts.PullPolicy)
	}

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

func (e *engine) ContainerRun(ctx context.Context, spec *types.ContainerSpec, opts *container.RunOptions) (string, error) {
	stdio := opts.Stdio
	cc := new(containerConfig)
	if err := cc.From(spec, stdio); err != nil {
		return "", err
	}

	createResp, err := e.client.ContainerCreate(ctx, cc.Config, cc.HostConfig, cc.NetworkingConfig, cc.Platform, cc.Name)
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
		Env:          cli.ConvertMapToKVString(opts.Env),
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

func (e *engine) ContainerRemove(ctx context.Context, opts *container.RemoveOptions) error {
	if id := opts.Id; id != "" {
		if err := e.containerRemove(ctx, id); err != nil {
			return err
		}
	}

	if labels := opts.Labels; len(labels) > 0 {
		filter := e.filterOf(labels)
		if _, err := e.client.ContainersPrune(ctx, filter); err != nil {
			return err
		}

		containers, err := e.client.ContainerList(ctx, dockercontainer.ListOptions{
			All:     true,
			Filters: filter,
		})
		if err != nil {
			return err
		}

		return parallel(containers, func(ctn *dockertypes.Container) error {
			return e.containerRemove(ctx, ctn.ID)
		})
	}

	return nil
}

func (e *engine) containerRemove(ctx context.Context, id string) error {
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

func (e *engine) ContainerInspect(ctx context.Context, id string) (*types.ContainerSpec, error) {
	res, err := e.client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}
	cs := new(containerSpec)
	if err = cs.From(res); err != nil {
		return nil, err
	}
	return cs.Spec, nil
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

func (e *engine) NetworkCreate(ctx context.Context, spec *types.NetworkSpec) (string, error) {
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

func (e *engine) NetworkRemove(ctx context.Context, opts *container.RemoveOptions) error {
	if id := opts.Id; id != "" {
		if err := e.client.NetworkRemove(ctx, id); err != nil {
			return err
		}
	}

	if labels := opts.Labels; len(labels) > 0 {
		filter := e.filterOf(labels)
		if _, err := e.client.NetworksPrune(ctx, filter); err != nil {
			return err
		}

		networks, err := e.client.NetworkList(ctx, dockernetwork.ListOptions{
			Filters: filter,
		})
		if err != nil {
			return err
		}

		return parallel(networks, func(net *dockernetwork.Summary) error {
			return e.client.NetworkRemove(ctx, net.ID)
		})
	}

	return nil
}

func (e *engine) VolumeCreate(ctx context.Context, spec *types.VolumeSpec) (string, error) {
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

func (e *engine) VolumeRemove(ctx context.Context, opts *container.RemoveOptions) error {
	if id := opts.Id; id != "" {
		if err := e.client.VolumeRemove(ctx, id, true); err != nil {
			return err
		}
	}

	if labels := opts.Labels; len(labels) > 0 {
		filter := e.filterOf(labels)
		if _, err := e.client.VolumesPrune(ctx, filter); err != nil {
			return err
		}

		resp, err := e.client.VolumeList(ctx, dockervolume.ListOptions{
			Filters: filter,
		})
		if err != nil {
			return err
		}

		return parallel(resp.Volumes, func(vol **dockervolume.Volume) error {
			return e.client.VolumeRemove(ctx, (*vol).Name, true)
		})
	}

	return nil
}

type run = func() error

func (e *engine) streamingStdio(ctx context.Context, id string, exec bool, stdio *types.Stdio, streams container.Streams, fn run) run {
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

			reader := xio.NewContextReader(ctx, hijackedResp.Reader)
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

// Docker currently supported filters are:
//
//   - until (<timestamp>) - only remove images created before given timestamp
//   - label (label=<key>, label=<key>=<value>, label!=<key>, or label!=<key>=<value>) -
//     only remove images with (or without, in case label!=... is used) the specified labels.
//
// See: https://docs.docker.com/reference/cli/docker/image/prune/#filter
func (e *engine) filterOf(labels map[string]string) dockerfilters.Args {
	args := dockerfilters.NewArgs()
	for k, v := range labels {
		args.Add("label", fmt.Sprintf("%s=%s", k, v))
	}
	return args
}

func parallel[E any](list []E, op func(*E) error) error {
	if len(list) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errs := make([]error, len(list))

	for i := 0; i < len(list); i++ {
		wg.Add(1)
		go func(i int, e *E) {
			defer wg.Done()
			errs[i] = op(e)
		}(i, &list[i]) // using pointer to avoid copy data
	}

	wg.Wait()
	return errors.Join(errs...)
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
