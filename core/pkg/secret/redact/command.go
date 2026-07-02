/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package redact

import (
	"context"

	"drassi.run/core/pkg/command/cmdtypes"
	"drassi.run/core/pkg/secret"
	"drassi.run/core/util/dig"
)

func Attacher[R any](a cmdtypes.Attacher[R], sm secret.Masker) cmdtypes.Attacher[R] {
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

func Reporter[R any](sm secret.Masker) xdig.Decorator[cmdtypes.Reporter[R]] {
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
