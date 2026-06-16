/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	"drassi.run/gha-runner/pkg/lease"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/log/logtypes"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/timeline"
	"drassi.run/gha-runner/pkg/types"
	"go.uber.org/dig"
)

func New(msg *messages.PipelineAgentJobRequest) *Worker {
	return &Worker{msg: msg}
}

type Worker struct {
	msg         *messages.PipelineAgentJobRequest
	lease       lease.Lease
	timelineMgr *timeline.Manager

	ctx     context.Context
	cancel  context.CancelCauseFunc
	waiters []types.Waiter
	closers []io.Closer
}

func (w *Worker) Context() context.Context {
	return w.ctx
}

func (w *Worker) Wait() {
	for _, waiter := range w.waiters {
		waiter.Wait()
	}
}

func (w *Worker) Cancel(cause error) {
	if cancel := w.cancel; cancel != nil {
		w.cancel(cause)
	}
}

func (w *Worker) Run(ctx context.Context, scope *dig.Scope) (err error) {
	w.ctx, w.cancel = context.WithCancelCause(ctx)
	defer w.cancel(nil)

	defer w.close()
	if err = w.setup(scope); err != nil {
		return err
	}
	if err = w.runServices(ctx, scope); err != nil {
		return err
	}

	var job *records.Job
	defer func() {
		if job != nil {
			w.timelineMgr.FinishJob(job)
		}
		ex := w.lease.Complete(w.ctx, w.timelineMgr.JobRecord())
		err = errors.Join(err, ex)
	}()

	req := w.lease.GetMessage()
	spec, err := messages.ToJobSpec(req)
	if err != nil {
		return fmt.Errorf("convert PipelineAgentJobRequest to JobSpec: %w", err)
	}
	scope = scope.Scope(fmt.Sprintf("job(%s)", spec.Id))

	w.timelineMgr.InitJob(spec)
	job, err = executor.Run(w.ctx, spec, scope)
	return err
}

func (w *Worker) runServices(ctx context.Context, scope *dig.Scope) error {
	go w.lease.Renew(ctx)

	go w.timelineMgr.Run(w.ctx)
	w.waiters = append(w.waiters, w.timelineMgr)

	if err := scope.Invoke(w.runLogSubscribers); err != nil {
		return err
	}

	return nil
}

func (w *Worker) close() error {
	errs := make([]error, 0)
	for _, c := range slices.Backward(w.closers) {
		errs = append(errs, c.Close())
	}
	return errors.Join(errs...)
}

type logSubscriberParams struct {
	dig.In

	LogManager  *log.Manager
	Subscribers []logtypes.Subscriber `group:"log-subscribers"` // = [LogSubscribers]
}

func (w *Worker) runLogSubscribers(p logSubscriberParams) {
	for _, sub := range p.Subscribers {
		ch := p.LogManager.Subscribe()
		go sub.Run(ch)
		w.waiters = append(w.waiters, sub)
	}
}
