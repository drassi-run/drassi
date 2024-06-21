package executor

import (
	"context"
	"strings"

	"drassi.run/core/pkg/model/actions"
)

type dockerActionRun struct {
	action *actions.DockerRuns
	image  string
}

func (ar *dockerActionRun) Initialize(ctx context.Context, exec StepExecutor) error {
	if i, ok := strings.CutPrefix(ar.action.Image, "docker://"); ok {
		// TODO pull image
		ar.image = i
	} else {
		ar.image = "" // random string
		// TODO build image from ./path/to/Dockerfile
	}
	return nil
}

func (ar *dockerActionRun) PreTask() *Task {
	if ar.action.PreEntrypoint == "" {
		return nil
	}
	return &Task{
		Stage:     StagePre,
		Condition: ar.action.PreIf,
		Run:       ar.executePre,
	}
}

func (ar *dockerActionRun) executePre(ctx context.Context, exec StepExecutor) error {
	entrypoint := []string{ar.action.PreEntrypoint}
	env := ar.evaluateEnv()
	return exec.Sandbox().RunContainer(ctx, ar.image, entrypoint, nil, env, "")
}

func (ar *dockerActionRun) MainTask() *Task {
	return &Task{
		Stage: StageMain,
		Run:   ar.executeMain,
	}
}

func (ar *dockerActionRun) executeMain(ctx context.Context, exec StepExecutor) error {
	entrypoint := []string{ar.action.Entrypoint}
	env := ar.evaluateEnv()
	return exec.Sandbox().RunContainer(ctx, ar.image, entrypoint, nil, env, "")
}

func (ar *dockerActionRun) PostTask() *Task {
	if ar.action.PostEntrypoint == "" {
		return nil
	}
	return &Task{
		Stage:     StagePost,
		Condition: ar.action.PreIf,
		Run:       ar.executePre,
	}
}

func (ar *dockerActionRun) executePost(ctx context.Context, exec StepExecutor) error {
	entrypoint := []string{ar.action.PostEntrypoint}
	env := ar.evaluateEnv()
	return exec.Sandbox().RunContainer(ctx, ar.image, entrypoint, nil, env, "")
}

func (ar *dockerActionRun) Action() actions.Runs {
	return ar.action
}

func (ar *dockerActionRun) evaluateEnv() map[string]string {
	// TODO compute env from  ar.action.Env
	return nil
}
