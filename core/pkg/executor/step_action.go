package executor

import (
	"context"
	"path/filepath"

	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/actions"
	"drassi.run/core/pkg/model/dossiers"
	"drassi.run/core/pkg/store/repository"
	"gopkg.in/yaml.v3"
)

// Example:
// + Using a public action: `uses: actions/aws@v2.0.1`
// + Using a public action in a subdirectory: `uses: actions/aws/ec2@main`
// + Using a local action: `uses: ./.github/actions/hello-world-action`
type ActionStepRun struct {
	BaseStepRun
	Repo *model.Repository

	Store repository.Store

	rev       string
	actionRun StepRun
}

func (sr *ActionStepRun) SetContextInfo(dossier *dossiers.Dossier) {
	gh := dossier.Github

	gh.Action = sr.Id
	gh.ActionRepository = sr.Repo.Repo
	gh.ActionRef = sr.Repo.Ref
}

func (sr *ActionStepRun) Initialize(ctx context.Context, exec StepExecutor) error {
	if rev, err := sr.Store.Fetch(ctx, sr.Repo, ""); err != nil {
		return err
	} else {
		sr.rev = rev
	}

	path := filepath.Join(sr.Repo.Path, "action.yml")
	r, err := sr.Store.File(ctx, sr.Repo, sr.rev, path)
	if err != nil {
		return err
	}
	defer r.Close()

	m := make(map[string]any)
	if err = yaml.NewDecoder(r).Decode(m); err != nil {
		return err
	}
	if err = model.Decode(m, sr.action); err != nil {
		return err
	}

	switch runs := sr.action.Runs.(type) {
	case *actions.JavaScriptRuns:
		sr.actionRun = &javaScriptActionRun{
			action: runs,
			repo:   sr.Repo,
			rev:    sr.rev,
		}
	case *actions.DockerRuns:
		sr.actionRun = &dockerActionRun{
			action: runs,
		}
	case *actions.CompositeRuns:
		sr.actionRun = &CompositeStepRun{
			action: runs,
		}
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
