package decorator

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

func (t Telemetry) DecorateActionRun(action *executor.ActionRun) executor.ActionTask {
	stepId := action.StepId()
	stage := action.Stage
	run := action.Run
	return func(ctx context.Context) (err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("ActionRun(%s, %s)", stepId, stage),
		)
		defer done(&err)
		return run(ctx)
	}
}

func (t Telemetry) DecorateStepRun(step *executor.StepRun) executor.StepTask {
	stepId := step.StepId()
	stage := step.Stage
	run := step.Run
	return func(ctx context.Context) (_ *records.Step, err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("StepRun(%s, %s)", stepId, stage),
			xotel.DrassiStep(stepId), xotel.DrassiStage(stage),
		)
		defer done(&err)
		return run(ctx)
	}
}
