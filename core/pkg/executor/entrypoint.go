/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"
	"fmt"

	"drassi.run/core/pkg/model/records"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/otel"
	"go.uber.org/dig"
)

func Run(ctx context.Context, spec *JobSpec, scope *dig.Scope) (job *records.JobResult, err error) {
	var decorator JobRunDecorator
	if err = xdig.Populate(scope, &decorator); err != nil {
		return nil, err
	}

	exec, err := spec.CreateExecutor(ctx, scope)
	if err != nil {
		return nil, err
	}

	task := exec.CreateTask()
	task.Run = decorator.DecorateJobRun(task)
	task.Run = telemetryJobRun(spec.Id, task.Run)

	return task.Run(ctx)
}

func telemetryJobRun(jobId string, run JobRun) JobRun {
	return func(ctx context.Context) (_ *records.JobResult, err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("JobRun(%s)", jobId),
		)
		defer done(&err)
		return run(ctx)
	}
}
