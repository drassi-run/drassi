/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package worker

import (
	"context"
	"iter"
	"maps"
	"sync"
	"time"

	"drassi.run/core/util/context"
	"drassi.run/core/wire"
	"drassi.run/gha-runner/pkg/messages"
	"github.com/chainguard-dev/clog"
)

type flight struct {
	JobId   string
	Worker  *Worker
	Context context.Context
	Cancel  context.CancelCauseFunc
	DoneCh  chan struct{}
}

func NewManager() *Manager {
	return &Manager{
		inflight: make(map[string]*flight),
	}
}

type Manager struct {
	mu sync.Mutex
	wg sync.WaitGroup

	inflight map[string]*flight
}

func (m *Manager) Submit(req *messages.PipelineAgentJobRequest, modules ...*wire.Module) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx := xcontext.WithExpander(context.Background())
	ctx, cancel := context.WithCancelCause(ctx)
	f := &flight{
		JobId:   req.JobId,
		Worker:  NewWorker(req),
		Cancel:  cancel,
		Context: ctx,
		DoneCh:  make(chan struct{}),
	}
	m.inflight[req.JobId] = f

	m.wg.Add(1)
	go m.takeoff(f, modules...)
}

func (m *Manager) takeoff(f *flight, modules ...*wire.Module) {
	var err error
	defer m.landing(f, &err)
	err = f.Worker.Run(f.Context, modules...)
}

func (m *Manager) landing(f *flight, e *error) {
	var err error
	ctx := f.Context
	if ex := ctx.Err(); ex != nil {
		// external error: server canceled, process interrupt,...
		err = ex
	} else if e != nil && *e != nil {
		// internal error while running job
		err = *e
	}
	if err != nil {
		clog.ErrorContextf(ctx, "job=%q fail: %v", f.JobId, err)
	}
	if r := recover(); r != nil {
		clog.ErrorContextf(ctx, "job=%q panic: %v", f.JobId, r)
		if ex, ok := r.(error); ok {
			err = ex
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	f.Cancel(err)
	delete(m.inflight, f.JobId)
	close(f.DoneCh)
	m.wg.Done()
}

func (m *Manager) Cancel(reqId string, timeout time.Duration, err error) {
	if f, ok := m.inflight[reqId]; ok {
		ctx := f.Context
		if ex := xcontext.Expand(ctx, timeout); ex != nil {
			clog.ErrorContextf(ctx, "job=%q: %v", reqId, ex)
		}
		f.Cancel(err)
		<-f.DoneCh
		clog.WarnContextf(ctx, "job=%q canceled: %v", reqId, err)
	}
}

func (m *Manager) Inflight() iter.Seq[string] {
	return maps.Keys(m.inflight)
}

func (m *Manager) Wait() {
	m.wg.Wait()
}
