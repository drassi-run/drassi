/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_secret

import (
	"context"

	"drassi.run/core/pkg/command/cmdtypes"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/secret"
	"drassi.run/core/util/dig"
)

func maskAttachmentUploader[R any](a cmdtypes.Attacher[R], sm secret.Masker) cmdtypes.Attacher[R] {
	return &maskAttacher[R]{
		Attacher: a,
		sm:       sm,
	}
}

type maskAttacher[R any] struct {
	cmdtypes.Attacher[R]
	sm secret.Masker
}

func (a *maskAttacher[R]) Upload(ctx context.Context, res R, att *cmdtypes.Attachment) error {
	att.Reader = secret.MaskReader(a.sm, att.Reader)
	return a.Attacher.Upload(ctx, res, att)
}

func maskIssueReporter[R any](sm secret.Masker) xdig.Decorator[cmdtypes.Reporter[R]] {
	return func(r cmdtypes.Reporter[R]) cmdtypes.Reporter[R] {
		return &maskReporter[R]{
			Reporter: r,
			sm:       sm,
		}
	}
}

type maskReporter[R any] struct {
	cmdtypes.Reporter[R]
	sm secret.Masker
}

const issueMessageMaxLength = 4096

func (r *maskReporter[R]) AddIssue(ctx context.Context, res R, iss *cmdtypes.Issue) error {
	iss.Message = r.sm.Mask(iss.Message)
	if len(iss.Message) > issueMessageMaxLength {
		iss.Message = iss.Message[:issueMessageMaxLength]
	}

	return r.Reporter.AddIssue(ctx, res, iss)
}

func maskScribeHandler(h scribe.Handler, sm secret.Masker) scribe.Handler {
	return func(ctx context.Context, line string) error {
		line = sm.Mask(line)
		return h(ctx, line)
	}
}

func maskJobRunDecorator(sm secret.Masker) executor.JobRunDecorator {
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
