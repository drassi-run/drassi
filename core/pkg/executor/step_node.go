package executor

import (
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/store/repository"
	"drassi.run/core/util/dig"
	"go.uber.org/dig"
)

type NodeStepRun struct {
	BaseStepRun

	Runtime string
	Main    string

	Pre   string
	PreIf workflows.Conditional

	Post   string
	PostIf workflows.Conditional

	// injected values
	exprEnv expression.Env
	sandbox sandboxer.Sandbox
	streams sandboxer.Streams
	repo    *repository.Repository
}

func (sr *NodeStepRun) Initialize(exec StepExecutor, scope *dig.Scope) error {
	if err := xdig.Populate(scope, &sr.exprEnv); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &sr.sandbox); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &sr.streams); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &sr.repo); err != nil {
		return err
	}

	return nil
}

func (sr *NodeStepRun) PreTask() *Task {
	if sr.Pre == "" {
		return nil
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L451-L471
	condition := sr.PreIf
	if condition == "" {
		condition = "always()"
	}
	return &Task{
		StepId:    sr.Id,
		Stage:     StagePre,
		Condition: condition,
		Run:       sr.execute(StagePre),
	}
}

func (sr *NodeStepRun) MainTask() *Task {
	return &Task{
		StepId:    sr.Id,
		Stage:     StageMain,
		Condition: sr.Condition,
		Run:       sr.execute(StageMain),
	}
}

func (sr *NodeStepRun) PostTask() *Task {
	if sr.Post == "" {
		return nil
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L451-L471
	condition := sr.PostIf
	if condition == "" {
		condition = "always()"
	}
	return &Task{
		StepId:    sr.Id,
		Stage:     StagePost,
		Condition: condition,
		Run:       sr.execute(StagePost),
	}
}

func (sr *NodeStepRun) execute(stage Stage) TaskRun {
	return func(exec StepExecutor) error {
		ctx := exec.Context()

		scriptPath := sr.computeScriptPath(stage)
		cmd := []string{"node", scriptPath}

		inputs := make(map[string]string)
		if err := evaluator.Evaluate(sr.exprEnv, sr.Inputs, &inputs); err != nil {
			return err
		}

		env := exec.ComposeEnv()
		for k, v := range inputs {
			k = strings.ToUpper(k)
			env["INPUT_"+k] = v
		}

		paths := exec.JobExecutor().SystemPaths()
		return sr.sandbox.Execute(ctx, cmd, paths, env, "", sr.streams)
	}
}

func (sr *NodeStepRun) computeScriptPath(stage Stage) string {
	var script string
	switch stage {
	case StagePre:
		script = sr.Pre
	case StagePost:
		script = sr.Post
	case StageMain:
		script = sr.Main
	}

	layout := sr.sandbox.Layout()
	scriptPath := filepath.Join(layout.Actions, repository.Location(sr.repo), script)
	return scriptPath
}
