/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

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
	"drassi.run/core/pkg/stream"
	xcontext "drassi.run/core/util/context"
	xio "drassi.run/core/util/io"
	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/pkg/stdcopy"
	dockercontainer "github.com/moby/moby/api/types/container"
	dockernetwork "github.com/moby/moby/api/types/network"
	dockervolume "github.com/moby/moby/api/types/volume"
	dockerclient "github.com/moby/moby/client"
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
	client *dockerclient.Client
}

var defaultOpts = []dockerclient.Opt{
	dockerclient.WithUserAgent("drassi"),
}

func New(opts ...dockerclient.Opt) (container.Engine, error) {
	opts = append(defaultOpts, opts...)

	if client, err := dockerclient.New(opts...); err != nil {
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
		_, err := e.client.ImageInspect(ctx, ref)
		if err == nil { // image existed
			return nil
		}
		if !errdefs.IsNotFound(err) {
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
	resp, err := e.client.ImagePull(ctx, ref, dockerclient.ImagePullOptions{
		RegistryAuth: cred,
	})
	if err != nil {
		return err
	}

	defer resp.Close()
	// cli.ImagePull is asynchronous.
	// The reader needs to be read completely for the pull operation to complete.
	// If stdout is not required, consider using io.Discard instead of os.Stdout.
	// TODO: show download progress
	_, err = io.Copy(os.Stdout, resp)
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

	createResp, err := e.client.ContainerCreate(ctx, dockerclient.ContainerCreateOptions{
		Config:           cc.Config,
		HostConfig:       cc.HostConfig,
		NetworkingConfig: cc.NetworkingConfig,
		Platform:         cc.Platform,
		Name:             cc.Name,
	})
	if err != nil {
		return "", err
	}
	ctnID := createResp.ID

	fn := e.start(ctx, ctnID, false, stdio)
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

	idResp, err := e.client.ExecCreate(ctx, id, dockerclient.ExecCreateOptions{
		Cmd:          opts.Cmd,
		WorkingDir:   opts.Workdir,
		Env:          cli.ConvertMapToKVString(opts.Env),
		TTY:          stdio.Tty,
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

	var fn run
	if stdio.AttachStdin() || stdio.AttachStdout() || stdio.AttachStderr() {
		fn = e.streamingStdio(ctx, execID, true, stdio, opts.Streams, fn)
	} else {
		fn = e.start(ctx, execID, true, stdio)
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
		if _, err := e.client.ContainerPrune(ctx, dockerclient.ContainerPruneOptions{Filters: filter}); err != nil {
			return err
		}

		containers, err := e.client.ContainerList(ctx, dockerclient.ContainerListOptions{
			All:     true,
			Filters: filter,
		})
		if err != nil {
			return err
		}

		return parallel(containers.Items, func(ctn *dockercontainer.Summary) error {
			return e.containerRemove(ctx, ctn.ID)
		})
	}

	return nil
}

func (e *engine) containerRemove(ctx context.Context, id string) error {
	_, err := e.client.ContainerStop(ctx, id, dockerclient.ContainerStopOptions{})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return err
	}
	_, err = e.client.ContainerRemove(ctx, id, dockerclient.ContainerRemoveOptions{
		RemoveVolumes: true, // Remove anonymous volumes associated with the container
		RemoveLinks:   true, // Remove the specified link
		Force:         true, // Force the removal of a running container (uses SIGKILL)
	})
	// In the case of `docker run --rm ...`
	if err != nil && !errdefs.IsNotFound(err) && !errdefs.IsConflict(err) {
		return err
	}
	return nil
}

func (e *engine) ContainerInspect(ctx context.Context, id string) (*types.ContainerSpec, error) {
	res, err := e.client.ContainerInspect(ctx, id, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return nil, err
	}
	cs := new(containerSpec)
	if err = cs.From(res.Container); err != nil {
		return nil, err
	}
	return cs.Spec, nil
}

func (e *engine) Stat(ctx context.Context, id string, path string) (fs.FileInfo, error) {
	if ps, err := e.client.ContainerStatPath(ctx, id, dockerclient.ContainerStatPathOptions{Path: path}); err != nil {
		return nil, err
	} else {
		return &fileInfo{ps.Stat}, nil
	}
}

func (e *engine) CopyIn(ctx context.Context, id string, opts *container.CopyInOptions) error {
	_, err := e.client.CopyToContainer(ctx, id, dockerclient.CopyToContainerOptions{
		DestinationPath: opts.DestinationPath,
		Content:         opts.Reader,
	})
	return err
}

func (e *engine) CopyOut(ctx context.Context, id string, opts *container.CopyOutOptions) (io.ReadCloser, error) {
	res, err := e.client.CopyFromContainer(ctx, id, dockerclient.CopyFromContainerOptions{
		SourcePath: opts.SourcePath,
	})
	if err != nil {
		return nil, err
	}
	return res.Content, nil
}

func (e *engine) NetworkCreate(ctx context.Context, spec *types.NetworkSpec) (string, error) {
	config := dockerclient.NetworkCreateOptions{
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
		if _, err := e.client.NetworkRemove(ctx, id, dockerclient.NetworkRemoveOptions{}); err != nil {
			return err
		}
	}

	if labels := opts.Labels; len(labels) > 0 {
		filter := e.filterOf(labels)
		if _, err := e.client.NetworkPrune(ctx, dockerclient.NetworkPruneOptions{Filters: filter}); err != nil {
			return err
		}

		networks, err := e.client.NetworkList(ctx, dockerclient.NetworkListOptions{
			Filters: filter,
		})
		if err != nil {
			return err
		}

		return parallel(networks.Items, func(net *dockernetwork.Summary) error {
			_, err := e.client.NetworkRemove(ctx, net.ID, dockerclient.NetworkRemoveOptions{})
			return err
		})
	}

	return nil
}

func (e *engine) VolumeCreate(ctx context.Context, spec *types.VolumeSpec) (string, error) {
	config := dockerclient.VolumeCreateOptions{
		Name:       spec.Name,
		Labels:     spec.Labels,
		Driver:     spec.Driver,
		DriverOpts: spec.Options,
	}

	if res, err := e.client.VolumeCreate(ctx, config); err != nil {
		return "", err
	} else {
		return res.Volume.Name, nil
	}
}

func (e *engine) VolumeRemove(ctx context.Context, opts *container.RemoveOptions) error {
	if id := opts.Id; id != "" {
		if _, err := e.client.VolumeRemove(ctx, id, dockerclient.VolumeRemoveOptions{Force: true}); err != nil {
			return err
		}
	}

	if labels := opts.Labels; len(labels) > 0 {
		filter := e.filterOf(labels)
		if _, err := e.client.VolumePrune(ctx, dockerclient.VolumePruneOptions{Filters: filter}); err != nil {
			return err
		}

		resp, err := e.client.VolumeList(ctx, dockerclient.VolumeListOptions{
			Filters: filter,
		})
		if err != nil {
			return err
		}

		return parallel(resp.Items, func(vol *dockervolume.Volume) error {
			_, err := e.client.VolumeRemove(ctx, vol.Name, dockerclient.VolumeRemoveOptions{Force: true})
			return err
		})
	}

	return nil
}

type run = func() error

func (e *engine) start(ctx context.Context, id string, exec bool, stdio *types.Stdio) run {
	// ContainerExec
	if exec {
		return func() error {
			_, err := e.client.ExecStart(ctx, id, dockerclient.ExecStartOptions{
				Detach: stdio.Detach(),
				TTY:    stdio.Tty,
			})
			return err
		}
	}

	// ContainerRun
	return func() error {
		_, err := e.client.ContainerStart(ctx, id, dockerclient.ContainerStartOptions{})
		return err
	}
}

func (e *engine) streamingStdio(ctx context.Context, id string, exec bool, stdio *types.Stdio, streams *stream.Streams, fn run) run {
	ctx, cancel := context.WithCancel(ctx)

	return func() (err error) {
		defer cancel()

		var hijackedResp dockerclient.HijackedResponse
		if exec {
			attachRes, err := e.client.ExecAttach(ctx, id, dockerclient.ExecAttachOptions{
				TTY: stdio.Tty,
			})
			if err != nil {
				return err
			}
			hijackedResp = attachRes.HijackedResponse
		} else {
			attachRes, err := e.client.ContainerAttach(ctx, id, dockerclient.ContainerAttachOptions{
				Stream: true,
				Stdin:  stdio.AttachStdin(),
				Stdout: stdio.AttachStdout(),
				Stderr: stdio.AttachStderr(),
			})
			if err != nil {
				return err
			}
			hijackedResp = attachRes.HijackedResponse
		}

		errC := make(chan error, 1)
		go func() {
			defer close(errC)
			defer hijackedResp.Close()

			reader := xio.NewContextReader(ctx, hijackedResp.Reader)
			outWriter, errWriter := streams.Out, streams.Err

			var err error
			if stdio.Tty {
				_, err = io.Copy(outWriter, reader)
			} else {
				// https://github.com/moby/moby/blob/docker-v29.7.2/client/container_attach.go#L52
				_, err = stdcopy.StdCopy(outWriter, errWriter, reader)
			}
			errC <- err
		}()

		if fn != nil {
			if err = fn(); err != nil {
				return
			}
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

		waitRes := e.client.ContainerWait(ctx, id, dockerclient.ContainerWaitOptions{
			Condition: condition,
		})

		errC := make(chan error, 1)
		go func() {
			defer close(errC)

			select {
			case result := <-waitRes.Result:
				if result.Error != nil {
					errC <- errors.New(result.Error.Message)
				} else if result.StatusCode != 0 {
					errC <- fmt.Errorf("container exited with status code: %d", result.StatusCode)
				}
			case err := <-waitRes.Error:
				ctx, cancel := xcontext.ExpandContext(ctx, err)
				defer cancel()
				_, _ = e.client.ContainerKill(ctx, id, dockerclient.ContainerKillOptions{Signal: "SIGKILL"})

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

			inspectResp, err := e.client.ExecInspect(ctx, id, dockerclient.ExecInspectOptions{})
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

		inspectResp, err := e.client.ContainerInspect(ctx, id, dockerclient.ContainerInspectOptions{})
		if err != nil {
			return fmt.Errorf("failed to inspect container: %w", err)
		}
		if ec := inspectResp.Container.State.ExitCode; ec != 0 {
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
func (e *engine) filterOf(labels map[string]string) dockerclient.Filters {
	args := make(dockerclient.Filters)
	for k, v := range labels {
		args.Add("label", fmt.Sprintf("%s=%s", k, v))
	}
	return args
}

func parallel[E any](list []E, op func(*E) error) error {
	if len(list) == 0 {
		return nil
	}

	errs := make([]error, len(list))
	var wg sync.WaitGroup

	for i := range list {
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

func (fi *fileInfo) Name() string       { return fi.PathStat.Name }
func (fi *fileInfo) Size() int64        { return fi.PathStat.Size }
func (fi *fileInfo) Mode() fs.FileMode  { return fi.PathStat.Mode }
func (fi *fileInfo) ModTime() time.Time { return fi.PathStat.Mtime }
func (fi *fileInfo) IsDir() bool        { return fi.Mode().IsDir() }
func (fi *fileInfo) Sys() any           { return nil }
