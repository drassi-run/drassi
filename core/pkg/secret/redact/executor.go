/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package redact

import (
	"context"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/secret"
)

func JobOutputs(sm secret.Masker) executor.JobRunDecorator {
	return &maskJobOutputs{sm: sm}
}

type maskJobOutputs struct {
	sm secret.Masker
}

func (m *maskJobOutputs) DecorateJobRun(task *executor.JobTask) executor.JobRun {
	run := task.Run
	return func(ctx context.Context) (*records.JobResult, error) {
		s := scribe.FromContext(ctx)
		res, err := run(ctx)

		for k, v := range res.Outputs {
			if !m.sm.IsClean(v) {
				s.Warningf("Skip output %q since it may contain secret.", k)
				delete(res.Outputs, k)
			}
		}

		return res, err
	}
}
