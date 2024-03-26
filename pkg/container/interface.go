package container

import (
	"context"
	"io"

	"github.com/compose-spec/compose-go/v2/types"
)

type ContainerRuntime interface {
	io.Closer
	LaunchContainer(context.Context, LaunchContainerRequest) (LaunchContainerResponse, error)
	TerminateContainer(context.Context, TerminateContainerRequest) (TerminateContainerResponse, error)
	ExecuteContainer(context.Context, ExecuteContainerRequest) (ExecuteContainerResponse, error)

	// See https://docs.docker.com/engine/reference/commandline/cp/ for the specification.
	CopyFromContainer(context.Context, CopyFromContainerRequest) (CopyFromContainerResponse, error)
	CopyToContainer(context.Context, CopyToContainerRequest) (CopyToContainerResponse, error)
}

type LaunchContainerRequest struct {
	Config types.ServiceConfig
}

type LaunchContainerResponse struct {
	ContainerId string
}

type TerminateContainerRequest struct {
	ContainerId string
	Timeout     *int
}

type TerminateContainerResponse struct {
}

type ExecuteContainerRequest struct {
	ContainerId string
	Cmd         []string
	Env         map[string]string
	User        string
	Workdir     string
}

type ExecuteContainerResponse struct {
}

type CopyFromContainerRequest struct {
	ContainerId string
	SourcePath  string
}

type CopyFromContainerResponse struct {
	Reader io.ReadCloser
}

type CopyToContainerRequest struct {
	ContainerId     string
	DestinationPath string
	Content         io.Reader
}

type CopyToContainerResponse struct {
}
