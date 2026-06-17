/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package timeline

import (
	"context"
	"maps"
	"slices"
	"sync"
	"time"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command/cmdtypes"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"github.com/chainguard-dev/clog"
	"github.com/google/uuid"
)

func NewManager(recorder Recorder) *Manager {
	return NewManagerInterval(time.Second, recorder)
}

func NewManagerInterval(interval time.Duration, recorder Recorder) *Manager {
	ticker := time.NewTicker(interval)

	return &Manager{
		ticker:     ticker,
		recorder:   recorder,
		recordUids: make(map[string]map[executor.Stage]string),
		records:    make(map[string]*Record),
	}
}

type Manager struct {
	mu       sync.Mutex
	wg       sync.WaitGroup
	ticker   *time.Ticker
	recorder Recorder

	order      int
	recordUids map[string]map[executor.Stage]string

	records   map[string]*Record
	jobRecord *Record
}

func (m *Manager) RecordUid(stage executor.Stage, uid string) string {
	s, ok := m.recordUids[uid]
	if !ok {
		s = make(map[executor.Stage]string)
		m.recordUids[uid] = s
	}
	r, ok := s[stage]
	if !ok {
		r = uuid.New().String()
		s[stage] = r
	}
	return r
}

func (m *Manager) JobRecord() *Record {
	return m.jobRecord
}

func (m *Manager) InitJob(spec *executor.JobSpec) {
	o := &JobObject{JobSpec: spec}
	r := m.newRecord(spec.Uid, o)

	r.StartedAt = new(time.Now())
	r.State = StateInProgress

	m.jobRecord = r
}

func (m *Manager) FinishJob(rec *records.Job) {
	r := m.jobRecord
	o := r.Object.(*JobObject)

	r.State = StateCompleted
	r.CompletedAt = new(time.Now())
	if rec != nil {
		r.Result = ToResult(rec.Result)
		o.Outputs = maps.Clone(rec.Outputs)
	} else {
		r.Result = ResultFailed
	}
}

func (m *Manager) DecorateJobRun(task *executor.JobTask) executor.JobRun {
	if task.Stage == executor.StageMain {
		return task.Run
	}

	// planning for Setup job & Complete job step
	uid := m.RecordUid(task.Stage, task.JobSpec().Uid)
	var o *StepObject
	switch task.Stage {
	case executor.StagePre:
		o = setupJobObject(uid)
	case executor.StagePost:
		o = completeJobObject(uid)
	}
	r := m.newRecord(uid, o)
	switch task.Stage {
	case executor.StagePre:
		r.Name = "Set up job"
	case executor.StagePost:
		r.Name = "Complete job"
	}
	m.push(r)
	m.addToJob(r)

	run := task.Run
	return func(ctx context.Context) (*records.Job, error) {
		r.StartedAt = new(time.Now())
		r.State = StateInProgress
		m.push(r)

		rec, err := run(ctx)

		r.State = StateCompleted
		r.CompletedAt = new(time.Now())
		if err != nil {
			r.Result = ResultFailed
		} else {
			r.Result = ToResult(rec.Result)
			o.Outputs = maps.Clone(rec.Outputs)
		}
		m.push(r)
		return rec, err
	}
}

func (m *Manager) DecorateStepRun(task *executor.StepTask) executor.StepRun {
	// not record embedded step (inside composite action)
	if executor.Depth(task.Executor) > 1 {
		return task.Run
	}

	uid := m.RecordUid(task.Stage, task.StepSpec().Uid)
	o := &StepObject{StepSpec: task.StepSpec()}
	r := m.newRecord(uid, o)
	r.Name = task.Executor.Name(task.Stage)
	m.push(r)
	m.addToJob(r)

	run := task.Run
	return func(ctx context.Context) (*records.Step, error) {
		r.StartedAt = new(time.Now())
		r.State = StateInProgress
		r.Name = task.Executor.Name(task.Stage)
		m.push(r)

		rec, err := run(ctx)

		r.State = StateCompleted
		r.CompletedAt = new(time.Now())
		r.Name = task.Executor.Name(task.Stage)
		if err != nil {
			r.Result = ResultFailed
		} else {
			r.Result = ToResult(rec.Conclusion)
			o.Outputs = maps.Clone(rec.Outputs)
		}
		m.push(r)
		return rec, err
	}
}

func (m *Manager) AddIssue(stage executor.Stage, stepUid string, iss *cmdtypes.Issue) {
	uid := m.RecordUid(stage, stepUid)
	r := m.records[uid]
	r.Issues = append(r.Issues, iss)
}

func (m *Manager) Run(ctx context.Context) {
	l := clog.FromContext(ctx)
	m.wg.Add(1)
	defer m.wg.Done()

	for {
		select {
		case <-m.ticker.C:
			m.flush(ctx, l)
		case <-ctx.Done():
			m.ticker.Stop()

			// New 30s timeout context for flush remaining records
			var cancel context.CancelFunc
			ctx = context.WithoutCancel(ctx)
			ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
			//goland:noinspection GoDeferInLoop
			defer cancel()

			m.flush(ctx, l)
			return
		}
	}
}

func (m *Manager) Wait() {
	m.wg.Wait()
}

func (m *Manager) newRecord(uid string, obj any) *Record {
	r := &Record{
		Uid:    uid,
		Order:  m.order,
		Object: obj,
		State:  StatePending,
	}
	m.order++
	return r
}

func (m *Manager) push(r *Record) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[r.Uid] = r
}

func (m *Manager) addToJob(r *Record) {
	m.jobRecord.Children = append(m.jobRecord.Children, r)
}

func (m *Manager) flush(ctx context.Context, l *clog.Logger) {
	m.mu.Lock()
	r := slices.Collect(maps.Values(m.records))
	clear(m.records)
	m.mu.Unlock()

	if len(r) <= 0 {
		return
	}

	l.Info("flushing timeline.Record...")
	if err := m.recorder.Update(ctx, r...); err != nil {
		l.Errorf("failed to update record: %v", err)
	}
}

func setupJobObject(uid string) *StepObject {
	return metaStepObject("__init", uid, "Set up job", executor.StagePre)
}

func completeJobObject(uid string) *StepObject {
	return metaStepObject("__finish", uid, "Complete job", executor.StagePost)
}

func metaStepObject(id, uid, name string, stage executor.Stage) *StepObject {
	spec := &executor.StepSpec{
		Id:   id,
		Uid:  uid,
		Name: workflows.NewLiteralToken(name),
	}
	return &StepObject{
		StepSpec: spec,
		Stage:    stage,
	}
}
