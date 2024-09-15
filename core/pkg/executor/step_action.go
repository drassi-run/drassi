package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/actions"
	"drassi.run/core/pkg/store/repository"
	"drassi.run/core/pkg/store/repository/gitstore"
	"github.com/go-git/go-git/v5/plumbing/object"
	"gopkg.in/yaml.v3"
)

// Example:
// + Using a public action: `uses: actions/aws@v2.0.1`
// + Using a public action in a subdirectory: `uses: actions/aws/ec2@main`
// + Using a local action: `uses: ./.github/actions/hello-world-action`
type ActionStepRun struct {
	BaseStepRun
	Repo *repository.Repository

	Store gitstore.Store

	rev       string
	actionRun StepRun
}

func (sr *ActionStepRun) Repository() *repository.Repository {
	return sr.Repo
}

func (sr *ActionStepRun) Initialize(ctx context.Context, exec StepExecutor) error {
	gh := exec.Dossier().Github
	if rev, err := sr.Store.Fetch(ctx, sr.Repo, gh.Token); err != nil {
		return err
	} else {
		sr.rev = rev
	}

	if err := sr.loadAction(ctx); err != nil {
		return err
	}

	return sr.actionRun.Initialize(ctx, exec)
}

func (sr *ActionStepRun) PreTask() *Task {
	return sr.actionRun.PreTask()
}

func (sr *ActionStepRun) MainTask() *Task {
	return sr.actionRun.MainTask()
}

func (sr *ActionStepRun) PostTask() *Task {
	return sr.actionRun.PostTask()
}

func (sr *ActionStepRun) loadAction(ctx context.Context) error {
	// 1. First, try reading "action.yml" or "action.yaml" file
	for _, f := range []string{"action.yml", "action.yaml"} {
		path := filepath.Join(sr.Repo.Path, f)
		if r, err := sr.Store.File(ctx, sr.Repo, sr.rev, path); err == nil {
			return sr.loadActionManifest(r)
		} else if !errors.Is(err, object.ErrFileNotFound) {
			return err
		}
	}

	// 2. Second, try reading "Dockerfile" or "dockerfile"
	for _, f := range []string{"Dockerfile", "dockerfile"} {
		path := filepath.Join(sr.Repo.Path, f)
		if r, err := sr.Store.File(ctx, sr.Repo, sr.rev, path); err == nil {
			r.Close()
			return sr.createDockerfileAction(path)
		} else if !errors.Is(err, object.ErrFileNotFound) {
			return err
		}
	}

	return fmt.Errorf(`file "action.yml", "action.yaml" and "Dockerfile" not found in your given path`)
}

func (sr *ActionStepRun) loadActionManifest(r io.ReadCloser) error {
	defer r.Close()

	m := make(map[string]any)
	if err := yaml.NewDecoder(r).Decode(m); err != nil {
		return err
	}
	action := new(actions.Action)
	if err := model.Decode(m, action); err != nil {
		return err
	}

	if actionRun, err := FromAction(action); err != nil {
		return err
	} else {
		sr.actionRun = actionRun
	}

	return nil
}

func (sr *ActionStepRun) createDockerfileAction(dockerfile string) error {
	sr.actionRun = &DockerStepRun{
		Image: dockerfile,
	}
	return nil
}
