package executor

import (
	"context"
)

// Example:
// + `uses: docker://alpine:3.8`
// + `uses: docker://gcr.io/cloud-builders/gradle`
type DockerStepRun struct {
	BaseStepRun
	Image string
}

func (sr *DockerStepRun) Initialize(ctx context.Context, rCtx *StepRunContext) error {
	return rCtx.Sandbox().PullImage(ctx, sr.Image)
}

func (sr *DockerStepRun) PreTask() *Task {
	return nil
}

func (sr *DockerStepRun) MainTask() *Task {
	return &Task{
		StepID:    sr.UUID,
		Stage:     StageMain,
		Condition: sr.Condition,
		Run:       sr.executeMain,
	}
}

func (sr *DockerStepRun) executeMain(ctx context.Context, rCtx *StepRunContext) error {
	inputs, err := sr.Inputs.Evaluate("job.step", rCtx)
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

	return rCtx.Sandbox().RunContainer(ctx, sr.Image, entrypoint, cmd, nil, "")
}

func (sr *DockerStepRun) PostTask() *Task {
	return nil
}
