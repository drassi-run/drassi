/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"context"
	"io"
	"io/fs"
	"net"
	"path/filepath"

	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/fs"
	"drassi.run/core/util/net"
	"drassi.run/core/util/path"
	"drassi.run/core/util/tar"
	dockerclient "github.com/docker/docker/client"
)

const jobDir = "/opt/drassi/"

type sandbox struct {
	client *client
	vm     *vm
	layout sandboxer.Layout
	path   string
}

func newSandbox(ctx context.Context, c *client, machine *vm) (*sandbox, error) {
	sb := &sandbox{
		client: c,
		vm:     machine,
		layout: sandboxer.Layout{
			Workspace: filepath.Join(jobDir, "workspace"),
			Temp:      filepath.Join(jobDir, "temp"),
			Actions:   filepath.Join(jobDir, "actions"),
			Tools:     filepath.Join(jobDir, "tools"),
		},
	}

	layout := &sb.layout
	r, err := xtar.FileEntryReader(
		&xtar.FileEntry{Name: layout.Workspace, Mode: fs.ModeDir | xfs.DirPerm},
		&xtar.FileEntry{Name: layout.Temp, Mode: fs.ModeDir | xfs.AllPerm},
		&xtar.FileEntry{Name: layout.Actions, Mode: fs.ModeDir | xfs.DirPerm},
		&xtar.FileEntry{Name: layout.Tools, Mode: fs.ModeDir | xfs.DirPerm},
	)
	if err != nil {
		return nil, err
	}
	if err = sb.CopyIn(ctx, r, "/"); err != nil {
		return nil, err
	}

	if p, err := c.Info(ctx); err != nil {
		return nil, err
	} else {
		sb.path = p
	}
	return sb, nil
}

func (sb *sandbox) Layout() *sandboxer.Layout {
	return &sb.layout
}

func (sb *sandbox) Stat(ctx context.Context, path string) (fs.FileInfo, error) {
	return sb.client.Stat(ctx, path)
}

func (sb *sandbox) CopyIn(ctx context.Context, reader io.Reader, dst string) error {
	return sb.client.CopyIn(ctx, reader, dst)
}

func (sb *sandbox) CopyOut(ctx context.Context, src string) (io.ReadCloser, error) {
	return sb.client.CopyOut(ctx, src)
}

func (sb *sandbox) Execute(ctx context.Context, cmd, path []string, env map[string]string, workdir string, streams *stream.Streams) error {
	if workdir == "" {
		workdir = sb.layout.Workspace
	} else {
		workdir = xpath.Abs(workdir, sb.layout.Workspace)
	}
	if sb.path != "" {
		path = append(path, sb.path)
	}
	return sb.client.Execute(ctx, cmd, path, env, workdir, streams)
}

func (sb *sandbox) Terminate(ctx context.Context) error {
	if sb.vm == nil {
		return nil
	}
	return sb.vm.stop(ctx)
}

func (sb *sandbox) Dialer(cmd []string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		inRead, inWrite := io.Pipe()
		outRead, outWrite := io.Pipe()

		streams := &stream.Streams{
			In:  inRead,
			Out: outWrite,
		}
		go func() {
			_ = sb.Execute(ctx, cmd, nil, nil, "", streams)
			_ = inWrite.Close()
			_ = outWrite.Close()
		}()
		return xnet.NewStdioConn(inWrite, outRead), nil
	}
}

func (sb *sandbox) dockerEngine() (c.Engine, error) {
	dialer := sb.Dialer(docker.ProxyCommand(""))
	client, err := docker.New(dockerclient.WithDialContext(dialer))
	if err != nil {
		return nil, err
	}
	return c.WithTelemetry(client), nil
}
