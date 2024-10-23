package host

import (
	"context"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/path"
	"drassi.run/core/util/string"
)

const folderPerm fs.FileMode = 0o755

type engine struct {
	spec *HostSpec
}

func New(spec *HostSpec) (sandboxer.Engine, error) {
	if d, err := xpath.ResolveDir(spec.RootDir); err != nil {
		return nil, err
	} else {
		spec.RootDir = d
	}

	if err := os.MkdirAll(spec.RootDir, folderPerm); err != nil {
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

	if sb, err := newSandbox(sandboxDir); err != nil {
		return nil, err
	} else {
		sb.layout.Runtimes = e.spec.RuntimeDir
		res := &sandboxer.LaunchResponse{
			Sandbox: sb,
		}
		return res, nil
	}
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
