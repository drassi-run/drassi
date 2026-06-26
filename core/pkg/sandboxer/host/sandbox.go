/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

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
	"syscall"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/stream"
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

func (sb *sandbox) Execute(ctx context.Context, cmd, path []string, env map[string]string, workdir string, streams *stream.Streams) error {
	// TODO lookup entrypoint under custom PATH
	c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)

	// By default, CommandContext kills the direct process. If the command spawns its own child processes,
	// Go will kill the parent, but the sub-processes might become orphaned and keep running.
	// By assign the command to a new Process Group ID (SysProcAttr), entire process tree can be killed
	// by via manual invoke os SIGKILL.
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Override the default context cancellation behavior.
	// By default, Go calls `cmd.Process.Kill()`. Replacing it with a function
	// that targets the process group instead.
	c.Cancel = func() error {
		pgid := c.Process.Pid
		// The negative sign (-) before the PID targets the entire process group
		// rather than just the single parent process.
		return syscall.Kill(-pgid, syscall.SIGKILL)
	}

	// env
	c.Env = os.Environ()
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
	c.Stdin, c.Stdout, c.Stderr = streams.In, streams.Out, streams.Err

	err := c.Run()
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err() // process failed by ctx cancel/timeout
		}
	}
	return err
}

func (sb *sandbox) Terminate(context.Context) error {
	jobDir := filepath.Dir(sb.layout.Workspace)
	return os.RemoveAll(jobDir)
}
