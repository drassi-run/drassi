/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"
	"errors"
	"fmt"

	"drassi.run/core/pkg/model/records"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/otel"
	"go.uber.org/dig"
)

func Run(ctx context.Context, spec *JobSpec, scope *dig.Scope) (job *records.Job, err error) {
	var decorator JobRunDecorator
	if err = xdig.Populate(scope, &decorator); err != nil {
		return nil, err
	}

	exec := spec.CreateExecutor(scope)
	initTask := &JobTask{
		Run:      exec.Initialize,
		Stage:    StagePre,
		Executor: exec,
	}
	initTask.Run = decorator.DecorateJobRun(initTask)
	initTask.Run = telemetryJobRun(spec.Id, StagePre, initTask.Run)
	job, err = initTask.Run(ctx)
	if err != nil {
		err = fmt.Errorf("initialize job: %w", err)
	}

	completeTask := &JobTask{
		Run:      exec.Finalize,
		Stage:    StagePost,
		Executor: exec,
	}
	// GitHub runner require plan completeTask before run mainTask
	completeTask.Run = decorator.DecorateJobRun(completeTask)
	completeTask.Run = telemetryJobRun(spec.Id, StagePost, completeTask.Run)
	// register completeTask to ensure it always be run
	defer func() {
		j, ex := completeTask.Run(ctx)
		if ex != nil {
			ex = fmt.Errorf("complete job: %w", err)
			if err != nil {
				err = errors.Join(err, ex)
			} else {
				err = ex
			}
		}
		job = j
	}()

	// run main only when initTask success
	if err == nil {
		mainTask := &JobTask{
			Run:      exec.RunJob,
			Stage:    StageMain,
			Executor: exec,
		}
		mainTask.Run = decorator.DecorateJobRun(mainTask)
		mainTask.Run = telemetryJobRun(spec.Id, StageMain, mainTask.Run)
		job, err = mainTask.Run(ctx)
		if err != nil {
			err = fmt.Errorf("running job: %w", err)
		}
	}

	return
}

func telemetryJobRun(jobId string, stage Stage, run JobRun) JobRun {
	return func(ctx context.Context) (_ *records.Job, err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("JobRun(%s, %s)", jobId, stage),
			xotel.DrassiStep(jobId), xotel.DrassiStage(stage),
		)
		defer done(&err)
		return run(ctx)
	}
}
