package sandboxer

import (
	"context"
	"io"

	"github.com/dungdm93/drasi/pkg/container"
)

type SandboxEnvironment struct {
	SandboxId string
	Sandboxer Sandboxer

	ContainerRuntime    container.ContainerRuntime
	JobContainerId      string
	ServiceContainerIds map[string]string
	ContainerNetwork    string
	ContainerVolumes    map[string]string

	WorkPath     string
	WorkflowPath string
	ActionsPath  string
}

// used to execute script or action
func (c *SandboxEnvironment) Execute(ctx context.Context, cmd []string, env map[string]string, workdir string) error {
	//if containerId != nill {
	//	TODO("execute in container")
	//} else {
	//	TODO("execute in sandbox")
	//}
	panic("implement me")
}

func (c *SandboxEnvironment) CopyIn(ctx context.Context, reader io.Reader, dst string) error {
	panic("implement me")
}

func (c *SandboxEnvironment) CopyOut(ctx context.Context, src string) (io.Reader, error) {
	panic("implement me")
}

// used to run docker action
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsusing-for-docker-container-actions
func (c *SandboxEnvironment) RunContainer(ctx context.Context, image string, entrypoint string, cmd []string, env map[string]string, workdir string) error {
	panic("implement me")
}

func (c *SandboxEnvironment) PullImage(ctx context.Context, image string) error {
	panic("implement me")
}

func (c *SandboxEnvironment) BuildImage(ctx context.Context, image string, dockerfile string, contextPath string) error {
	panic("implement me")
}

func (c *SandboxEnvironment) GetWorkPath() string {
	return c.WorkPath
}

func (c *SandboxEnvironment) GetWorkflowPath() string {
	return c.WorkflowPath
}

func (c *SandboxEnvironment) GetActionsPath() string {
	return c.ActionsPath
}
