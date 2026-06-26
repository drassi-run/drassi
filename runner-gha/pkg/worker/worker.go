/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package worker

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/util/context"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/error"
	"drassi.run/core/util/otel"
	"drassi.run/core/wire"
	"drassi.run/gha-runner/pkg/lease"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/log/logtypes"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/timeline"
	gha_wire "drassi.run/gha-runner/wire"
	"github.com/chainguard-dev/clog"
	"go.uber.org/dig"
)

func NewWorker(msg *messages.PipelineAgentJobRequest) *Worker {
	return &Worker{msg: msg}
}

type Worker struct {
	msg         *messages.PipelineAgentJobRequest
	lease       lease.Lease
	timelineMgr *timeline.Manager

	ctx   context.Context
	wgLog sync.WaitGroup
	wgSvc sync.WaitGroup
}

func (w *Worker) Context() context.Context {
	return w.ctx
}

func (w *Worker) Run(ctx context.Context, modules ...*wire.Module) (err error) {
	scope := dig.New().Scope("worker")
	if err = gha_wire.Synthetic(scope, w.msg, modules...); err != nil {
		return
	}

	if fn, ex := w.initOtel(scope); ex != nil {
		return ex
	} else {
		var done func(*error)
		ctx, done = fn(ctx)
		defer done(&err)
	}
	w.ctx = ctx

	if err = w.inject(scope); err != nil {
		return
	}

	defer w.complete()

	if cancel := w.runServices(); cancel != nil {
		defer cancel() // cancel timelineMgr.Run & lease.Renew
	}

	if err = scope.Invoke(w.runLogSubscribers); err != nil {
		return
	}
	defer scope.Invoke(w.closeLogSubscribers)

	return w.run(scope)
}

func (w *Worker) run(scope *dig.Scope) (err error) {
	defer xerror.Recover(&err)
	l := clog.FromContext(w.ctx)

	l.Debug("convert PipelineAgentJobRequest to JobSpec")
	spec, err := messages.ToJobSpec(w.msg)
	if err != nil {
		return fmt.Errorf("convert PipelineAgentJobRequest to JobSpec: %w", err)
	}
	scope = scope.Scope(fmt.Sprintf("job(%s)", spec.Id))

	l.Infof("running job %s", spec.Id)
	w.timelineMgr.InitJob(spec)
	job, err := executor.Run(w.ctx, spec, scope)
	w.timelineMgr.FinishJob(job)

	return err
}

func (w *Worker) inject(scope *dig.Scope) error {
	if err := xdig.Supply[xcontext.Provider](scope, w); err != nil {
		return fmt.Errorf("provide xcontext.Provider: %w", err)
	}
	if err := xdig.Populate(scope, &w.lease); err != nil {
		return fmt.Errorf("inject lease: %w", err)
	}
	if err := xdig.Populate(scope, &w.timelineMgr); err != nil {
		return fmt.Errorf("inject timeline.Manager: %w", err)
	}
	return scope.Invoke(w.initDiary)
}

func (w *Worker) runServices() context.CancelFunc {
	ctx, cancel := context.WithCancel(w.ctx)

	// Run lease.Renew
	w.wgSvc.Add(1)
	go func() {
		defer w.wgSvc.Done()
		w.lease.Renew(ctx)
	}()

	// Run timelineMgr
	w.wgSvc.Add(1)
	go func() {
		defer w.wgSvc.Done()
		w.timelineMgr.Run(ctx)
	}()

	return cancel
}

type logSubscriberParams struct {
	dig.In

	LogManager  *log.Manager
	Subscribers []logtypes.Subscriber `group:"log-subscribers"` // = [LogSubscribers]
}

func (w *Worker) runLogSubscribers(p logSubscriberParams) {
	for _, sub := range p.Subscribers {
		ch := p.LogManager.Subscribe()
		w.wgLog.Add(1)
		go func() {
			defer w.wgLog.Done()
			sub.Run(w.ctx, ch)
		}()
	}
}

func (w *Worker) closeLogSubscribers(p logSubscriberParams) {
	l := clog.FromContext(w.ctx)

	if lm := p.LogManager; lm != nil {
		// stop any current running session, e.g. because of panic
		if err := lm.Stop(); err != nil {
			l.Errorf("stop log.Manager failed: %v", err)
		}
		if err := lm.Close(); err != nil {
			l.Errorf("close log.Manager failed: %v", err)
		}
	}

	w.wgLog.Wait()

	for _, sub := range p.Subscribers {
		if closer, ok := sub.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				l.Errorf("close log.Subscriber failed: %v", err)
			}
		}
	}

	// remove job's log folder
	if lm := p.LogManager; lm != nil {
		if err := lm.Dispose(); err != nil {
			l.Errorf("remove job's log folder: %v", err)
		}
	}
}

func (w *Worker) complete() {
	l := clog.FromContext(w.ctx)
	w.wgSvc.Wait()

	var r *timeline.Record
	if tm := w.timelineMgr; tm != nil {
		r = tm.JobRecord()
	}

	ls := w.lease
	if ls == nil {
		return
	}

	ctx := context.WithoutCancel(w.ctx)
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := ls.Complete(ctx, r); err != nil {
		l.Errorf("completing job error: %v", err)
	} else {
		l.Info("complete job")
	}
}

func (w *Worker) initOtel(scope *dig.Scope) (func(context.Context) (context.Context, func(*error)), error) {
	var gh *records.Github
	if err := xdig.Populate(scope, &gh); err != nil {
		return nil, fmt.Errorf("inject records.Github: %w", err)
	}

	fn := func(ctx context.Context) (context.Context, func(*error)) {
		// TODO set LogLevel=Debug if RunnerDebug=true
		jobMatrix := ""
		if w.msg.JobName != "__default" {
			jobMatrix = "/" + w.msg.JobName
		}
		return xotel.SetupTelemetry(ctx, "worker",
			xotel.Repo(gh.Repository),
			xotel.Workflow(gh.Workflow),
			xotel.Run(gh.RunId+"#"+gh.RunAttempt),
			xotel.Job(gh.Job+jobMatrix),
		)
	}
	return fn, nil
}

func (w *Worker) initDiary(diary scribe.Diary) {
	w.ctx = scribe.ContextWithScribe(w.ctx, diary)
}
