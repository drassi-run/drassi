package support

import (
	"context"
	"fmt"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	xotel "drassi.run/core/util/otel"
)

type Telemetry struct{}

func NewTelemetry() *Telemetry {
	return new(Telemetry)
}

func (t *Telemetry) DecorateActionRun(task *executor.ActionTask) executor.ActionRun {
	stepId := task.StepId()
	stage := task.Stage
	run := task.Run
	return func(ctx context.Context) (err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("ActionRun(%s, %s)", stepId, stage),
		)
		defer done(&err)
		return run(ctx)
	}
}

func (t *Telemetry) DecorateStepRun(task *executor.StepTask) executor.StepRun {
	stepId := task.StepId()
	stage := task.Stage
	run := task.Run
	return func(ctx context.Context) (_ *records.Step, err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("StepRun(%s, %s)", stepId, stage),
			xotel.DrassiStep(stepId), xotel.DrassiStage(stage),
		)
		defer done(&err)
		return run(ctx)
	}
}

func (t *Telemetry) DecorateJobRun(task *executor.JobTask) executor.JobRun {
	jobId := task.JobId()
	stage := task.Stage
	run := task.Run
	return func(ctx context.Context) (_ *records.Job, err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("JobRun(%s, %s)", jobId, stage),
			xotel.DrassiStep(jobId), xotel.DrassiStage(stage),
		)
		defer done(&err)
		return run(ctx)
	}
}
