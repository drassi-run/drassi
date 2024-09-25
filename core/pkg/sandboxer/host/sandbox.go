package host

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/sandboxer"
)

type hostSandbox struct {
	sandboxRuntime sandboxer.SandboxRuntime
	sandboxId      string

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

var _ sandboxer.Sandbox = (*hostSandbox)(nil)

func (s *hostSandbox) Execute(ctx context.Context, cmd []string, env map[string]string, workdir string, streams *sandboxer.Streams) error {
	if s.jobContainerId == "" {
		if workdir == "" || strings.HasPrefix(workdir, "./") {
			workdir = filepath.Join(s.GetWorkspaceDir(), workdir)
		} else if !strings.HasPrefix(workdir, "/") {
			return fmt.Errorf("unexpected workdir %s", workdir)
		}
		req := sandboxer.ExecuteSandboxRequest{
			SandboxId: s.sandboxId,
			Cmd:       cmd,
			Env:       env,
			Workdir:   workdir,
			Streams:   streams,
		}
		res, err := s.sandboxRuntime.ExecuteSandbox(ctx, req)
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

func (s *hostSandbox) CopyIn(ctx context.Context, reader io.Reader, dst string) error {
	if s.jobContainerId == "" {
		req := sandboxer.CopyToSandboxRequest{
			SandboxId:       s.sandboxId,
			DestinationPath: dst,
			Content:         reader,
		}
		_, err := s.sandboxRuntime.CopyToSandbox(ctx, req)
		if err != nil {
			return err
		}
	} else {
		//	TODO("copy to container")
		panic("copy to container")
	}
	return nil
}

func (s *hostSandbox) CopyOut(ctx context.Context, src string) (io.ReadCloser, error) {
	if s.jobContainerId == "" {
		req := sandboxer.CopyFromSandboxRequest{
			SandboxId:  s.sandboxId,
			SourcePath: src,
		}
		res, err := s.sandboxRuntime.CopyFromSandbox(ctx, req)
		if err != nil {
			return nil, err
		}
		return res.Reader, nil
	} else {
		//	TODO("copy to container")
		panic("copy to container")
	}
}

func (s *hostSandbox) RunContainer(ctx context.Context, image string, entrypoint []string, cmd []string, env map[string]string, workdir string) error {
	//TODO implement me
	panic("implement me")
}

func (s *hostSandbox) PullImage(ctx context.Context, image string) error {
	//TODO implement me
	panic("implement me")
}

func (s *hostSandbox) BuildImage(ctx context.Context, image string, dockerfile string, contextPath string) error {
	//TODO implement me
	panic("implement me")
}

func (s *hostSandbox) Paths() []string {
	path := os.Getenv("PATH")
	// Split path by os.PathListSeparator
	return strings.FieldsFunc(path, func(r rune) bool {
		return r == os.PathListSeparator
	})
}

func (s *hostSandbox) GetWorkspaceDir() string {
	return s.workspaceDir
}
func (s *hostSandbox) GetActionsDir() string {
	return s.actionsDir
}
func (s *hostSandbox) GetToolsDir() string {
	return s.toolsDir
}
func (s *hostSandbox) GetTempDir() string {
	return s.tempDir
}
