package sandboxer

import (
	"context"
	"io"

	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/model/records"
)

type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

type Sandbox interface {
	Execute(ctx context.Context, cmd []string, env map[string]string, workdir string, streams *Streams) error
	CopyIn(ctx context.Context, reader io.Reader, dst string) error
	CopyOut(ctx context.Context, src string) (io.ReadCloser, error)

	RunContainer(ctx context.Context, image string, entrypoint []string, cmd []string, env map[string]string, workdir string) error
	PullImage(ctx context.Context, image string) error
	BuildImage(ctx context.Context, image string, dockerfile string, contextPath string) error

	Paths() []string

	// The full path the repository is cloned to, and where the job runs from
	GetWorkspaceDir() string

	// The full path to the directory where actions are downloaded into
	GetActionsDir() string

	// The full path to the directory containing preinstalled tools for GitHub-hosted runners
	GetToolsDir() string

	// The full path to the directory where file commands, workflow/event.json and run scripts are located
	GetTempDir() string
}

type SandboxRuntime interface {
	Close() error
	Connect(context.Context) error

	LaunchSandbox(context.Context, LaunchSandboxRequest) (LaunchSandboxResponse, error)
	TerminateSandbox(context.Context, TerminateSandboxRequest) (TerminateSandboxResponse, error)
	ExecuteSandbox(context.Context, ExecuteSandboxRequest) (ExecuteSandboxResponse, error)

	CopyFromSandbox(context.Context, CopyFromSandboxRequest) (CopyFromSandboxResponse, error)
	CopyToSandbox(context.Context, CopyToSandboxRequest) (CopyToSandboxResponse, error)
}

type LaunchSandboxRequest struct {
	JobId             string
	JobEnv            map[string]string
	JobContainer      *types.ContainerSpec
	ServiceContainers map[string]*types.ContainerSpec
}

type LaunchSandboxResponse struct {
	Sandbox   Sandbox
	Container *records.Container
	Services  map[string]*records.Container

	Env map[string]string
}

type TerminateSandboxRequest struct {
	Sandbox Sandbox
	Timeout *int
}

type TerminateSandboxResponse struct {
}

type ExecuteSandboxRequest struct {
	SandboxId string
	Cmd       []string
	Env       map[string]string
	User      string
	Group     string
	Workdir   string
	Streams   *Streams
}

type ExecuteSandboxResponse struct {
	ExitCode int
}

type CopyFromSandboxRequest struct {
	SandboxId  string
	SourcePath string
}

type CopyFromSandboxResponse struct {
	Reader io.ReadCloser
}

type CopyToSandboxRequest struct {
	SandboxId       string
	DestinationPath string
	Content         io.Reader
}

type CopyToSandboxResponse struct {
}
