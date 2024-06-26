package executor

import (
	"context"

	"drassi.run/core/pkg/model/dossiers"
)

// Example:
// + `uses: docker://alpine:3.8`
// + `uses: docker://gcr.io/cloud-builders/gradle`
type DockerStepRun struct {
	BaseStepRun
	Image string
}

func (sr *DockerStepRun) SetContextInfo(dossier *dossiers.Dossier) {
	gh := dossier.Github

	gh.Action = sr.Id
	gh.ActionRepository = ""
	gh.ActionRef = ""
}

func (sr *DockerStepRun) Initialize(ctx context.Context, exec StepExecutor) error {
	return exec.Sandbox().PullImage(ctx, sr.Image)
}

func (sr *DockerStepRun) PreTask() *Task {
	return nil
}

func (sr *DockerStepRun) MainTask() *Task {
	return &Task{
		StepId:    sr.Id,
		Stage:     StageMain,
		Condition: sr.Condition,
		Run:       sr.executeMain,
	}
}

func (sr *DockerStepRun) executeMain(ctx context.Context, exec StepExecutor) error {
	evalSupplier := &evaluationSupplier{dossier: exec.Dossier()}
	inputs, err := sr.Inputs.Evaluate("job.step", evalSupplier)
	if err != nil {
		return err
	}

	var entrypoint []string
	if ep, ok := inputs["entrypoint"]; ok {
		entrypoint = append(entrypoint, ep)
	}

	var cmd []string
	if c, ok := inputs["cmd"]; ok {
		cmd = append(cmd, c)
	}

	return exec.Sandbox().RunContainer(ctx, sr.Image, entrypoint, cmd, nil, "")
}

func (sr *DockerStepRun) PostTask() *Task {
	return nil
}
