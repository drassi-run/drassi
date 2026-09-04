/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package worker

import (
	"context"
	"fmt"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/util/context"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/error"
	"drassi.run/core/util/otel"
	"drassi.run/core/wire"
	"drassi.run/gitea-runner/pkg/reporter"
	gitea_wire "drassi.run/gitea-runner/wire"
	runnerv1 "gitea.dev/actionslib/runner/v1"
	"github.com/chainguard-dev/clog"
	"go.uber.org/dig"
)

type Worker struct {
	task   *runnerv1.Task
	ctx    context.Context
	cancel context.CancelCauseFunc
}

func New(task *runnerv1.Task) *Worker {
	return &Worker{task: task}
}

func (w *Worker) Context() context.Context {
	return w.ctx
}

func (w *Worker) Run(ctx context.Context, modules ...*wire.Module) (err error) {
	scope := dig.New().Scope("worker")
	if err = gitea_wire.Synthetic(scope, w.task, modules...); err != nil {
		return
	}

	if fn, ex := w.initOtel(scope); ex != nil {
		return ex
	} else {
		var done func(*error)
		ctx, done = fn(ctx)
		defer done(&err)
	}
	w.ctx, w.cancel = context.WithCancelCause(ctx)
	defer w.cancel(nil)

	if err = w.inject(scope); err != nil {
		return
	}

	if err = scope.Invoke(w.startServices); err != nil {
		return
	}
	defer scope.Invoke(w.stopServices)

	return w.run(scope)
}

func (w *Worker) run(scope *dig.Scope) (err error) {
	defer xerror.Recover(&err)
	l := clog.FromContext(w.ctx)

	l.Debug("convert task.WorkflowPayload into Workflow")
	workflow := new(workflows.Workflow)
	if err = decodeWorkflow(w.task.WorkflowPayload, workflow); err != nil {
		return fmt.Errorf("convert task.WorkflowPayload into Workflow: %w", err)
	}

	l.Debug("convert Workflow into JobSpec")
	spec, err := convertJobSpec(workflow)
	if err != nil {
		return fmt.Errorf("convert Workflow into JobSpec: %w", err)
	}

	scope = scope.Scope(fmt.Sprintf("job(%s)", spec.Id))

	l.Infof("running job %s", spec.Id)
	if job, err := executor.Run(w.ctx, spec, scope); err != nil {
		return err
	} else if job.Result != records.ResultSuccess {
		return fmt.Errorf("job.Result failed")
	}
	return nil
}

func (w *Worker) initOtel(scope *dig.Scope) (func(context.Context) (context.Context, func(*error)), error) {
	var forge *records.Forge
	if err := xdig.Populate(scope, &forge); err != nil {
		return nil, fmt.Errorf("inject records.Forge: %w", err)
	}

	fn := func(ctx context.Context) (context.Context, func(*error)) {
		// TODO set LogLevel=Debug if RunnerDebug=true
		return xotel.SetupTelemetry(ctx, "worker",
			xotel.Repo(forge.Repository),
			xotel.Workflow(forge.Workflow),
			xotel.Run(forge.RunId+"#"+forge.RunAttempt),
			xotel.Job(forge.Job),
		)
	}
	return fn, nil
}

func (w *Worker) inject(scope *dig.Scope) error {
	if err := xdig.Supply[xcontext.Provider](scope, w); err != nil {
		return err
	}
	if err := xdig.Supply(scope, w.cancel); err != nil {
		return err
	}
	return scope.Invoke(w.initDiary)
}

func (w *Worker) initDiary(diary scribe.Diary) {
	w.ctx = scribe.ContextWithScribe(w.ctx, diary)
}

func (w *Worker) startServices(log *reporter.LogStreamer, rep *reporter.Reporter) error {
	if err := log.Start(); err != nil {
		return fmt.Errorf("log start: %w", err)
	}
	if err := rep.Start(); err != nil {
		return fmt.Errorf("reporter start: %w", err)
	}
	return nil
}

func (w *Worker) stopServices(log *reporter.LogStreamer, rep *reporter.Reporter) {
	l := clog.FromContext(w.ctx)

	if err := log.Close(); err != nil {
		l.Errorf("log close: %v", err)
	}
	if err := rep.Close(); err != nil {
		l.Errorf("reporter close: %v", err)
	}
}
