package executor

import (
	"context"
	xdig "drassi.run/core/util/dig"
	xotel "drassi.run/core/util/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type stepRunDecorator struct {
	StepRun
	supervisor Supervisor
}

func DecorateStepRun(sr StepRun) StepRun {
	if _, ok := sr.(*stepRunDecorator); ok {
		return sr
	}

	if csr, ok := sr.(*CompositeStepRun); ok {
		for i, s := range csr.StepRuns {
			csr.StepRuns[i] = DecorateStepRun(s)
		}
	}
	return &stepRunDecorator{StepRun: sr}
}

func (sr *stepRunDecorator) Initialize(ctx context.Context, scope *dig.Scope) (err error) {
	if err = xdig.Populate(scope, &sr.supervisor); err != nil {
		return err
	}

	ctx, span := xotel.StartSpan(ctx, "StepRun.Initialize")
	defer xotel.EndSpan(span, &err)

	stop := sr.supervisor.StartContext(ctx)
	defer stop()

	return sr.StepRun.Initialize(ctx, scope)
}

func (sr *stepRunDecorator) PreTask() *Task {
	return sr.decorateTask(StepRun.PreTask)
}

func (sr *stepRunDecorator) MainTask() *Task {
	return sr.decorateTask(StepRun.MainTask)
}

func (sr *stepRunDecorator) PostTask() *Task {
	return sr.decorateTask(StepRun.PostTask)
}

func (sr *stepRunDecorator) decorateTask(fn func(StepRun) *Task) *Task {
	task := fn(sr.StepRun)
	if task == nil {
		return nil
	}

	delegate := task.Run
	task.Run = func(ctx context.Context, se StepExecutor) (err error) {
		ctx, span := xotel.StartSpan(ctx, "Task.Run", trace.WithAttributes(
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
