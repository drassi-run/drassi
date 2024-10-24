package host

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/fs"
	"drassi.run/core/util/path"
	"github.com/go-git/go-billy/v5/osfs"
)

type sandbox struct {
	layout sandboxer.Layout
}

func newSandbox(dir string) (*sandbox, error) {
	// - if dir is not absolute it will be joined with the cwd
	// - clean the result
	if d, err := filepath.Abs(dir); err != nil {
		return nil, err
	} else {
		dir = d
	}

	layout := sandboxer.Layout{
		Workspace: filepath.Join(dir, "workspace"),
		Temp:      filepath.Join(dir, "temp"),
		Actions:   filepath.Join(dir, "actions"),
		Tools:     filepath.Join(dir, "tools"),
	}

	dirs := []string{
		layout.Workspace,
		layout.Actions,
		layout.Tools,
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, xfs.DirPerm); err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(layout.Temp, xfs.AllPerm); err != nil {
		return nil, err
	}

	return &sandbox{layout: layout}, nil
}

func (sb *sandbox) Layout() *sandboxer.Layout {
	return &sb.layout
}

func (sb *sandbox) ContainerInfo(context.Context) (*sandboxer.ContainerInfo, error) {
	return nil, nil
}

func (sb *sandbox) Stat(_ context.Context, path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

func (sb *sandbox) CopyIn(ctx context.Context, reader io.Reader, dst string) error {
	fsys := osfs.New("/")
	return xfs.Write(ctx, fsys, reader, dst)
}

func (sb *sandbox) CopyOut(ctx context.Context, src string) (io.ReadCloser, error) {
	fsys := osfs.New("/")
	r := xfs.Read(ctx, fsys, src)
	return r, nil
}

func (sb *sandbox) Execute(ctx context.Context, cmd, path []string, env map[string]string, workdir string, streams sandboxer.Streams) error {
	// TODO lookup entrypoint under custom PATH
	c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)

	// env
	c.Env = make([]string, 0, len(env))
	for k, v := range env {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// path
	if p := os.Getenv("PATH"); p != "" {
		path = append(path, p)
	}
	if len(path) > 0 {
		p := strings.Join(path, string(os.PathListSeparator))
		c.Env = append(c.Env, "PATH="+p)
	}

	// workdir
	if workdir == "" {
		c.Dir = sb.layout.Workspace
	} else {
		c.Dir = xpath.Abs(workdir, sb.layout.Workspace)
	}

	// streams
	c.Stdin = streams.In()
	c.Stdout = streams.Out()
	c.Stderr = streams.Err()

	return c.Run()
}

func (sb *sandbox) Terminate(context.Context) error {
	jobDir := filepath.Dir(sb.layout.Workspace)
	return os.RemoveAll(jobDir)
}
