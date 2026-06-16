/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package worker

import (
	"context"
	"fmt"
	"time"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/util/error"
	"drassi.run/gha-runner/pkg/lease"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/log/logtypes"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/timeline"
	"drassi.run/gha-runner/pkg/types"
	"github.com/chainguard-dev/clog"
	"go.uber.org/dig"
)

func New(msg *messages.PipelineAgentJobRequest) *Worker {
	return &Worker{msg: msg}
}

type Worker struct {
	msg         *messages.PipelineAgentJobRequest
	lease       lease.Lease
	logMgr      *log.Manager
	timelineMgr *timeline.Manager

	ctx     context.Context
	waiters []types.Waiter
}

func (w *Worker) Context() context.Context {
	return w.ctx
}

func (w *Worker) Run(ctx context.Context, scope *dig.Scope) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	w.ctx = ctx

	defer w.complete()

	if err := w.setup(scope); err != nil {
		return err
	}
	if cancel = w.runServices(); cancel != nil {
		defer cancel() // cancel timelineMgr.Run & lease.Renew
	}
	if err := scope.Invoke(w.runLogSubscribers); err != nil {
		return err
	}

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

func (w *Worker) runServices() context.CancelFunc {
	ctx, cancel := context.WithCancel(w.ctx)

	go w.lease.Renew(ctx)

	go w.timelineMgr.Run(ctx)
	w.waiters = append(w.waiters, w.timelineMgr)

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
		go sub.Run(w.ctx, ch)
		w.waiters = append(w.waiters, sub)
	}
}

func (w *Worker) complete() {
	l := clog.FromContext(w.ctx)

	// close log.Manager and its subscriber channels
	if lm := w.logMgr; lm != nil {
		if err := lm.Close(); err != nil {
			l.Errorf("close log.Manager failed: %v", err)
		}
	}
	// waiting for logtypes.Subscriber + timeline.Manager
	w.Wait()
	// remove job's log folder
	if lm := w.logMgr; lm != nil {
		if err := lm.Dispose(); err != nil {
			l.Errorf("dispose log.Manager failed: %v", err)
		}
	}

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

func (w *Worker) Wait() {
	for _, waiter := range w.waiters {
		waiter.Wait()
	}
}
