package sandboxer

import (
	"context"
	"io"

	"github.com/dungdm93/drasi/pkg/container"
)

type Sandbox interface {
	Execute(ctx context.Context, cmd []string, env map[string]string, workdir string) error
	CopyIn(ctx context.Context, reader io.Reader, dst string) error
	CopyOut(ctx context.Context, src string) (io.ReadCloser, error)

	RunContainer(ctx context.Context, image string, entrypoint []string, cmd []string, env map[string]string, workdir string) error
	PullImage(ctx context.Context, image string) error
	BuildImage(ctx context.Context, image string, dockerfile string, contextPath string) error

	GetWorkPath() string
	GetWorkflowPath() string
	GetActionsPath() string
}

type Sandboxer interface {
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
	JobContainer      *container.ContainerConfig
	ServiceContainers map[string]*container.ContainerConfig
}

type LaunchSandboxResponse struct {
	Sandbox Sandbox
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
