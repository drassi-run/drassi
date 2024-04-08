package sandboxer

import (
	"context"
	"io"
)

type Sandboxer interface {
	io.Closer
	LaunchSandbox(context.Context, LaunchSandboxRequest) (LaunchSandboxResponse, error)
	TerminateSandbox(context.Context, TerminateSandboxRequest) (TerminateSandboxResponse, error)
	ExecuteSandbox(context.Context, ExecuteSandboxRequest) (ExecuteSandboxResponse, error)

	CopyFromSandbox(context.Context, CopyFromSandboxRequest) (CopyFromSandboxResponse, error)
	CopyToSandbox(context.Context, CopyToSandboxRequest) (CopyToSandboxResponse, error)
}

type LaunchSandboxRequest struct {
}

type LaunchSandboxResponse struct {
	SandboxId string
}

type TerminateSandboxRequest struct {
	SandboxId string
	Timeout   *int
}

type TerminateSandboxResponse struct {
}

type ExecuteSandboxRequest struct {
	SandboxId string
	Cmd       []string
	Env       map[string]string
	User      string
	Workdir   string
}

type ExecuteSandboxResponse struct {
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
