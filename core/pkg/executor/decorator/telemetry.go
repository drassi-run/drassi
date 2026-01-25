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

func (t Telemetry) DecorateActionRun(stage executor.Stage, action *executor.ActionRun) *executor.ActionRun {
	base := action.Run
	dec := func(ctx context.Context, e executor.StepExecutor) (err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("ActionRun(%s, %s)", e.StepSpec().Id, stage),
		)
		defer done(&err)
		return base(ctx, e)
	}
	return &executor.ActionRun{
		Condition: action.Condition,
		Run:       dec,
	}
}

func (t Telemetry) DecorateStepRun(stage executor.Stage, e executor.StepExecutor, step executor.StepRun) executor.StepRun {
	return func(ctx context.Context) *records.Step {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("StepRun(%s, %s)", e.StepSpec().Id, stage),
			xotel.DrassiStep(e.StepSpec().Id), xotel.DrassiStage(stage),
		)
		defer done(nil)
		return step(ctx)
	}
}
