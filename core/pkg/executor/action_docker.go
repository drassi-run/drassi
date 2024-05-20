package executor

import (
	"context"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/model/actions"
)

type dockerActionExecutor struct {
	action *actions.DockerRuns
	image  string
}

func (e *dockerActionExecutor) Initialize(ctx context.Context, rCtx *StepRunContext) error {
	if i, ok := strings.CutPrefix(e.action.Image, "docker://"); ok {
		// TODO pull image
		e.image = i
	} else {
		e.image = "" // random string
		// TODO build image from ./path/to/Dockerfile
	}
	return nil
}

func (e *dockerActionExecutor) PreTask() *Task {
	if e.action.PreEntrypoint == "" {
		return nil
	}
	return &Task{
		Stage:     StagePre,
		Condition: e.action.PreIf,
		Run:       e.executePre,
	}
}

func (e *dockerActionExecutor) executePre(ctx context.Context, rCtx *StepRunContext) error {
	entrypoint := []string{e.action.PreEntrypoint}
	env := e.evaluateEnv()
	return rCtx.Sandbox().RunContainer(ctx, e.image, entrypoint, nil, env, "")
}

func (e *dockerActionExecutor) MainTask() *Task {
	return &Task{
		Stage: StageMain,
		Run:   e.executeMain,
	}
}

func (e *dockerActionExecutor) executeMain(ctx context.Context, rCtx *StepRunContext) error {
	entrypoint := []string{e.action.Entrypoint}
	env := e.evaluateEnv()
	return rCtx.Sandbox().RunContainer(ctx, e.image, entrypoint, nil, env, "")
}

func (e *dockerActionExecutor) PostTask() *Task {
	if e.action.PostEntrypoint == "" {
		return nil
	}
	return &Task{
		Stage:     StagePost,
		Condition: e.action.PreIf,
		Run:       e.executePre,
	}
}

func (e *dockerActionExecutor) executePost(ctx context.Context, rCtx *StepRunContext) error {
	entrypoint := []string{e.action.PostEntrypoint}
	env := e.evaluateEnv()
	return rCtx.Sandbox().RunContainer(ctx, e.image, entrypoint, nil, env, "")
}

func (e *dockerActionExecutor) Action() actions.Runs {
	return e.action
}

func (e *dockerActionExecutor) evaluateEnv() map[string]string {
	// TODO compute env from  e.action.Env
	return nil
}
