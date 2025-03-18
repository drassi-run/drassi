/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/fs"
	"drassi.run/core/util/path"
	"drassi.run/core/util/tar"
)

const jobDir = "/opt/drassi/"

var defaultLayout = sandboxer.Layout{
	Workspace: filepath.Join(jobDir, "workspace"),
	Temp:      filepath.Join(jobDir, "temp"),
	Actions:   filepath.Join(jobDir, "actions"),
	Tools:     filepath.Join(jobDir, "tools"),
}

type sandbox struct {
	engine      container.Engine
	containerId string
	layout      sandboxer.Layout
	path        string
}

func newSandbox(ctx context.Context, engine container.Engine, containerId string) (*sandbox, error) {
	sb := &sandbox{
		engine:      engine,
		containerId: containerId,
		layout:      defaultLayout,
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

	if info, err := engine.ContainerInspect(ctx, containerId); err != nil {
		return nil, err
	} else {
		sb.path = info.Environment["PATH"]
	}

	return sb, nil
}

func (sb *sandbox) Layout() *sandboxer.Layout {
	return &sb.layout
}

func (sb *sandbox) Stat(ctx context.Context, path string) (fs.FileInfo, error) {
	return sb.engine.Stat(ctx, sb.containerId, path)
}

func (sb *sandbox) CopyIn(ctx context.Context, reader io.Reader, dst string) error {
	return sb.engine.CopyIn(ctx, sb.containerId, &container.CopyInOptions{
		Reader:          reader,
		DestinationPath: dst,
	})
}

func (sb *sandbox) CopyOut(ctx context.Context, src string) (io.ReadCloser, error) {
	return sb.engine.CopyOut(ctx, sb.containerId, &container.CopyOutOptions{
		SourcePath: src,
	})
}

func (sb *sandbox) Execute(ctx context.Context, cmd, path []string, env map[string]string, workdir string, streams stream.Streams) error {
	opts := &container.ExecOptions{
		Cmd:     cmd,
		Env:     env,
		Workdir: workdir,
		Stdio: &types.Stdio{
			Tty:         false,
			Interactive: false,
			Attach:      types.Stdout | types.Stderr,
		},
		Streams: streams,
	}

	// path
	if sb.path != "" {
		path = append(path, sb.path)
	}
	if len(path) > 0 {
		p := strings.Join(path, string(os.PathListSeparator))
		opts.Env["PATH"] = p
	}

	// workdir
	if workdir == "" {
		opts.Workdir = sb.layout.Workspace
	} else {
		opts.Workdir = xpath.Abs(workdir, sb.layout.Workspace)
	}

	_, err := sb.engine.ContainerExec(ctx, sb.containerId, opts)
	return err
}

func (sb *sandbox) Terminate(ctx context.Context) error {
	return sb.engine.ContainerRemove(ctx, &container.RemoveOptions{Id: sb.containerId})
}
