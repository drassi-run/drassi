package executor

import (
	"context"
	"fmt"

	"drassi.run/core/util/dig"
	"drassi.run/core/util/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type telemetryStepRun struct {
	StepRun
	supervisor Supervisor
}

func WithTelemetryStepRun(sr StepRun) StepRun {
	if _, ok := sr.(*telemetryStepRun); ok {
		return sr
	}

	if csr, ok := sr.(*CompositeStepRun); ok {
		for i, s := range csr.StepRuns {
			csr.StepRuns[i] = WithTelemetryStepRun(s)
		}
	}
	return &telemetryStepRun{StepRun: sr}
}

func (sr *telemetryStepRun) Initialize(ctx context.Context, scope *dig.Scope) (err error) {
	if err = xdig.Populate(scope, &sr.supervisor); err != nil {
		return err
	}

	ctx, span := xotel.StartSpan(ctx, fmt.Sprintf("Task.Initialize(%s)", sr.StepId()))
	defer xotel.EndSpan(span, &err)

	stop := sr.supervisor.StartContext(ctx)
	defer stop()

	return sr.StepRun.Initialize(ctx, scope)
}

func (sr *telemetryStepRun) PreTask() *Task {
	return sr.decorateTask(StepRun.PreTask)
}

func (sr *telemetryStepRun) MainTask() *Task {
	return sr.decorateTask(StepRun.MainTask)
}

func (sr *telemetryStepRun) PostTask() *Task {
	return sr.decorateTask(StepRun.PostTask)
}

func (sr *telemetryStepRun) decorateTask(fn func(StepRun) *Task) *Task {
	task := fn(sr.StepRun)
	if task == nil {
		return nil
	}

	delegate := task.Run
	task.Run = func(ctx context.Context, se StepExecutor) (err error) {
		spanName := fmt.Sprintf("Task.Run(%s)", sr.StepId())
		ctx, span := xotel.StartSpan(ctx, spanName, trace.WithAttributes(
			xotel.DrassiStage(string(task.Stage)),
			xotel.DrassiStep(task.StepId),
		))
		defer xotel.EndSpan(span, &err)

		stop := sr.supervisor.StartContext(ctx)
		defer stop()

		return delegate(ctx, se)
	}
	return task
}
