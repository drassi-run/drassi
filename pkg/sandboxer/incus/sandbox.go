package incus

import (
	"context"
	"fmt"
	"io"

	"github.com/dungdm93/drasi/pkg/container"
	"github.com/dungdm93/drasi/pkg/sandboxer"
)

type incusSandbox struct {
	sandboxer sandboxer.Sandboxer
	sandboxId string

	containerRuntime    container.ContainerRuntime
	jobContainerId      string
	serviceContainerIds map[string]string
	containerNetwork    string
	containerVolumes    map[string]string

	workspaceDir string
	actionsDir   string
	toolsDir     string
	tempDir      string
}

var _ sandboxer.Sandbox = (*incusSandbox)(nil)

func (s *incusSandbox) Execute(ctx context.Context, cmd []string, env map[string]string, workdir string) error {
	if s.jobContainerId == "" {
		req := sandboxer.ExecuteSandboxRequest{
			SandboxId: s.sandboxId,
			Cmd:       cmd,
			Env:       env,
			Workdir:   workdir,
		}
		res, err := s.sandboxer.ExecuteSandbox(ctx, req)
		if err != nil {
			return err
		}
		if res.ExitCode != 0 {
			return fmt.Errorf("execute command failed with exit code %d", res.ExitCode)
		}
	} else {
		//	TODO("execute in container")
		panic("execute in container")
	}
	return nil
}

func (s *incusSandbox) CopyIn(ctx context.Context, reader io.Reader, dst string) error {
	if s.jobContainerId == "" {
		req := sandboxer.CopyToSandboxRequest{
			SandboxId:       s.sandboxId,
			DestinationPath: dst,
			Content:         reader,
		}
		_, err := s.sandboxer.CopyToSandbox(ctx, req)
		if err != nil {
			return err
		}
	} else {
		//	TODO("copy to container")
		panic("copy to container")
	}
	return nil
}

func (s *incusSandbox) CopyOut(ctx context.Context, src string) (io.ReadCloser, error) {
	if s.jobContainerId == "" {
		req := sandboxer.CopyFromSandboxRequest{
			SandboxId:  s.sandboxId,
			SourcePath: src,
		}
		res, err := s.sandboxer.CopyFromSandbox(ctx, req)
		if err != nil {
			return nil, err
		}
		return res.Reader, nil
	} else {
		//	TODO("copy to container")
		panic("copy to container")
	}
}

// used to run docker action
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsusing-for-docker-container-actions
func (s *incusSandbox) RunContainer(ctx context.Context, image string, entrypoint []string, cmd []string, env map[string]string, workdir string) error {
	//TODO implement me
	panic("implement me")
}

func (s *incusSandbox) PullImage(ctx context.Context, image string) error {
	panic("implement me")
}

func (s *incusSandbox) BuildImage(ctx context.Context, image string, dockerfile string, contextPath string) error {
	panic("implement me")
}

func (s *incusSandbox) GetWorkspaceDir() string {
	return s.workspaceDir
}
func (s *incusSandbox) GetActionsDir() string {
	return s.actionsDir
}
func (s *incusSandbox) GetToolsDir() string {
	return s.toolsDir
}
func (s *incusSandbox) GetTempDir() string {
	return s.tempDir
}
