package host

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/apis/v1alpha1"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/util/fs"
	"drassi.run/core/util/path"
	"drassi.run/core/util/string"
)

type engine struct {
	spec *v1alpha1.HostSandboxerSpec
}

func New(spec *v1alpha1.HostSandboxerSpec) (sandboxer.Engine, error) {
	if d, err := xpath.ResolveDir(spec.RootDir); err != nil {
		return nil, err
	} else {
		spec.RootDir = d
	}

	if err := os.MkdirAll(spec.RootDir, xfs.DirPerm); err != nil {
		return nil, err
	}

	return &engine{spec: spec}, nil
}

func (e *engine) Close() error {
	return nil
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	sandboxDir := e.sandboxDir(req)
	sandboxDir = filepath.Join(e.spec.RootDir, sandboxDir)

	sb, err := newSandbox(sandboxDir)
	if err != nil {
		return nil, err
	}
	sb.layout.Runtimes = e.spec.RuntimeDir

	client, err := docker.New()
	if err != nil {
		return nil, err
	}
	client = c.WithTelemetry(client)

	b := container.NewBootstrapper(client)
	return b.Bootstrap(ctx, sb, req)
}

func (e *engine) sandboxDir(req *sandboxer.LaunchRequest) string {
	var server string
	if u, err := url.Parse(req.Github.ServerUrl); err == nil {
		server = u.Host
	}
	server = strings.ToLower(server)
	repo := strings.ToLower(req.Github.Repository)

	workflow := strings.TrimSuffix(req.Github.Workflow, ".yml")
	workflow = strings.TrimSuffix(workflow, ".yaml")
	workflow = xstring.Normalize(workflow)

	job := xstring.Normalize(req.Github.Job)
	run := xstring.Normalize(req.Github.RunId)
	attempt := xstring.Normalize(req.Github.RunAttempt)

	path := filepath.Join(server, repo, workflow, job, run+"_"+attempt)
	return path
}
