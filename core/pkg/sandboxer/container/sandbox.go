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
	"drassi.run/core/util/path"
	"drassi.run/core/util/tar"
)

const folderPerm fs.FileMode = 0o755

type sandbox struct {
	engine      container.Engine
	containerId string
	layout      sandboxer.Layout
	path        string
}

func newSandbox(ctx context.Context, engine container.Engine, containerId string) (*sandbox, error) {
	dir := "/opt/drassi/"
	sb := &sandbox{
		engine:      engine,
		containerId: containerId,
		layout: sandboxer.Layout{
			Workspace: filepath.Join(dir, "workspace"),
			Temp:      filepath.Join(dir, "temp"),
			Actions:   filepath.Join(dir, "actions"),
			Tools:     filepath.Join(dir, "tools"),
		},
	}

	layout := &sb.layout
	r, err := xtar.FileEntryReader(
		&xtar.FileEntry{Name: layout.Workspace, Mode: fs.ModeDir | folderPerm},
		&xtar.FileEntry{Name: layout.Temp, Mode: fs.ModeDir | 0o777},
		&xtar.FileEntry{Name: layout.Actions, Mode: fs.ModeDir | folderPerm},
		&xtar.FileEntry{Name: layout.Tools, Mode: fs.ModeDir | folderPerm},
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

func (sb *sandbox) ContainerInfo(ctx context.Context) (*sandboxer.ContainerInfo, error) {
	cinfo, err := sb.engine.ContainerInspect(ctx, sb.containerId)
	if err != nil {
		return nil, err
	}

	info := &sandboxer.ContainerInfo{
		Id:     sb.containerId,
		Mounts: cinfo.Mounts,
	}
	return info, nil
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

func (sb *sandbox) Execute(ctx context.Context, cmd, path []string, env map[string]string, workdir string, streams sandboxer.Streams) error {
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
