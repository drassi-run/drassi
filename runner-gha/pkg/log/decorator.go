/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package log

import (
	"context"
	"fmt"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
)

type RecordStore interface {
	RecordUid(stage executor.Stage, uid string) string
}

type Decorator struct {
	mgr   *Manager
	store RecordStore
}

func NewDecorator(mgr *Manager, store RecordStore) *Decorator {
	return &Decorator{mgr: mgr, store: store}
}

func (d *Decorator) DecorateJobRun(task *executor.JobTask) executor.JobRun {
	if task.Stage == executor.StageMain {
		return task.Run
	}

	run := task.Run
	uid := d.store.RecordUid(task.Stage, task.JobSpec().Uid)
	return func(ctx context.Context) (*records.Job, error) {
		if err := d.mgr.Start(uid); err != nil {
			return nil, fmt.Errorf("start record log: %w", err)
		}

		rec, err := run(ctx)

		if ex := d.mgr.Stop(); ex != nil {
			return nil, fmt.Errorf("stop record log: %w", ex)
		}
		return rec, err
	}
}

func (d *Decorator) DecorateStepRun(task *executor.StepTask) executor.StepRun {
	run := task.Run
	uid := d.store.RecordUid(task.Stage, task.StepSpec().Uid)
	return func(ctx context.Context) (*records.Step, error) {
		if err := d.mgr.Start(uid); err != nil {
			return nil, fmt.Errorf("start record log: %w", err)
		}

		rec, err := run(ctx)

		if ex := d.mgr.Stop(); ex != nil {
			return nil, fmt.Errorf("stop record log: %w", ex)
		}
		return rec, err
	}
}
