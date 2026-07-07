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
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/types"
)

type decorator struct {
	mgr   *Manager
	store types.RecordStore
}

func NewDecorator(mgr *Manager, store types.RecordStore) executor.StepRunDecorator {
	return &decorator{mgr: mgr, store: store}
}

func (d *decorator) DecorateStepRun(task *executor.StepTask) executor.StepRun {
	var id, uid string
	if task.Kind == workflows.StepKindJob {
		id, uid = task.JobId(), task.JobSpec().Uid
	} else {
		id, uid = task.StepId(), task.StepSpec().Uid
	}
	uid = d.store.RecordUid(task.Stage, uid)
	attr := xotel.Step(string(task.Stage) + "/" + id)
	run := task.Run

	return func(ctx context.Context) (rec *records.StepResult, err error) {
		if err = d.mgr.Start(uid, attr); err != nil {
			return nil, fmt.Errorf("start record log: %w", err)
		}

		defer func() {
			if ex := d.mgr.Stop(); ex == nil {
				return
			} else if err == nil {
				err = fmt.Errorf("stop record log: %w", ex)
			}
		}()

		return run(ctx)
	}
}
