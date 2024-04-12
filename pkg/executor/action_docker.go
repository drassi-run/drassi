package executor

import (
	"context"
	"strings"

	"github.com/dungdm93/drasi/pkg/model/actions"
)

type dockerActionRunner struct {
	rCtx   *RunContext
	action *actions.DockerRuns
	image  string
}

func (e *dockerActionRunner) Initialize(ctx context.Context) error {
	if i, ok := strings.CutPrefix(e.action.Image, "docker://"); ok {
		// TODO pull image
		e.image = i
	} else {
		e.image = "" // random string
		// TODO build image from ./path/to/Dockerfile
	}
	return nil
}

func (e *dockerActionRunner) PreTask() *Task {
	if e.action.PreEntrypoint == "" {
		return nil
	}
	return &Task{
		Stage:     StagePre,
		Condition: e.action.PreIf,
		Run:       e.executePre,
	}
}

func (e *dockerActionRunner) executePre(ctx context.Context) error {
	entrypoint := []string{e.action.PreEntrypoint}
	env := e.evaluateEnv()
	return e.rCtx.RunContainer(ctx, e.image, entrypoint, nil, env, "")
}

func (e *dockerActionRunner) MainTask() *Task {
	return &Task{
		Stage: StageMain,
		Run:   e.executeMain,
	}
}

func (e *dockerActionRunner) executeMain(ctx context.Context) error {
	entrypoint := []string{e.action.Entrypoint}
	env := e.evaluateEnv()
	return e.rCtx.RunContainer(ctx, e.image, entrypoint, nil, env, "")
}

func (e *dockerActionRunner) PostTask() *Task {
	if e.action.PostEntrypoint == "" {
		return nil
	}
	return &Task{
		Stage:     StagePost,
		Condition: e.action.PreIf,
		Run:       e.executePre,
	}
}

func (e *dockerActionRunner) executePost(ctx context.Context) error {
	entrypoint := []string{e.action.PostEntrypoint}
	env := e.evaluateEnv()
	return e.rCtx.RunContainer(ctx, e.image, entrypoint, nil, env, "")
}

func (e *dockerActionRunner) evaluateEnv() map[string]string {
	// TODO compute env from  e.action.Env
	return nil
}
