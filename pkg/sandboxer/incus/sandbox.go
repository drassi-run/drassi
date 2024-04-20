package incus

import (
	"context"
	"github.com/dungdm93/drasi/pkg/sandboxer"
	"io"

	"github.com/dungdm93/drasi/pkg/container"
)

type incusSandbox struct {
	sandboxId string
	sandboxer sandboxer.Sandboxer

	containerRuntime    container.ContainerRuntime
	jobContainerId      string
	serviceContainerIds map[string]string
	containerNetwork    string
	containerVolumes    map[string]string

	workPath     string
	workflowPath string
	actionsPath  string
}

// used to execute script or action
func (c *incusSandbox) Execute(ctx context.Context, cmd []string, env map[string]string, workdir string) error {
	//if containerId != nill {
	//	TODO("execute in container")
	//} else {
	//	TODO("execute in sandbox")
	//}
	panic("implement me")
}

func (c *incusSandbox) CopyIn(ctx context.Context, reader io.Reader, dst string) error {
	panic("implement me")
}

func (c *incusSandbox) CopyOut(ctx context.Context, src string) (io.Reader, error) {
	panic("implement me")
}

// used to run docker action
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsusing-for-docker-container-actions
func (c *incusSandbox) RunContainer(ctx context.Context, image string, entrypoint string, cmd []string, env map[string]string, workdir string) error {
	panic("implement me")
}

func (c *incusSandbox) PullImage(ctx context.Context, image string) error {
	panic("implement me")
}

func (c *incusSandbox) BuildImage(ctx context.Context, image string, dockerfile string, contextPath string) error {
	panic("implement me")
}

func (c *incusSandbox) GetWorkPath() string {
	return c.workPath
}

func (c *incusSandbox) GetWorkflowPath() string {
	return c.workflowPath
}

func (c *incusSandbox) GetActionsPath() string {
	return c.actionsPath
}
