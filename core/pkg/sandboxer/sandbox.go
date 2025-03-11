package sandboxer

import (
	"context"
	"io"
	"io/fs"

	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/stream"
)

type Sandbox interface {
	Layout() *Layout

	Stat(ctx context.Context, path string) (fs.FileInfo, error)
	CopyIn(ctx context.Context, reader io.Reader, dst string) error
	CopyOut(ctx context.Context, src string) (io.ReadCloser, error)
	Execute(ctx context.Context, cmd, path []string, env map[string]string, workdir string, streams stream.Streams) error

	Terminate(ctx context.Context) error
}

type ContainerInfo struct {
	Id      string
	Network string
	Mounts  []*types.Mount
	Labels  map[string]string
}

type Layout struct {
	// Workspace is location repository is cloned to, and is job's default workdir
	Workspace string

	// Temp is where file commands, workflow/event.json and scripts are located
	Temp string

	// Actions is location where actions are downloaded into
	// It's job-scoped configuration
	Actions string

	// Tools directory contains preinstalled tools for GitHub-hosted runner
	// It's repo-scoped configuration
	Tools string

	// Runtimes directory contains node.js (and others) runtimes
	// It's runner-scoped configuration
	Runtimes string
}
