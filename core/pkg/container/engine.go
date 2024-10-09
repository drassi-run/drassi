package container

import (
	"context"
	"io"
)

type Engine interface {
	io.Closer

	ImagePull(ctx context.Context, ref string) error
	ImageBuild(ctx context.Context, context io.Reader, opts *BuildOptions) error

	ContainerRun(ctx context.Context, spec *ContainerSpec, opts *RunOptions) (string, error)
	ContainerExec(ctx context.Context, id string, opts *ExecOptions) (string, error)
	ContainerRemove(ctx context.Context, id string) error
	CopyIn(ctx context.Context, id string, opts *CopyInOptions) error
	CopyOut(ctx context.Context, id string, opts *CopyOutOptions) (io.ReadCloser, error)

	NetworkCreate(ctx context.Context, spec *NetworkSpec) (string, error)
	NetworkRemove(ctx context.Context, id string) error

	VolumeCreate(ctx context.Context, spec *VolumeSpec) (string, error)
	VolumeRemove(ctx context.Context, id string) error
}

type BuildOptions struct {
	Tags []string
}

type RunOptions struct {
	Stdio   *Stdio
	Streams Streams
}

type ExecOptions struct {
	Cmd     []string
	Env     map[string]string
	Path    []string
	Workdir string
	Stdio   *Stdio
	Streams Streams
}

type CopyInOptions struct {
	Reader          io.Reader
	DestinationPath string
}

type CopyOutOptions struct {
	SourcePath string
}

type Streams interface {
	In() io.Reader
	Out() io.Writer
	Err() io.Writer
}
