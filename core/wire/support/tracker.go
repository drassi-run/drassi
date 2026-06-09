/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_support

import (
	"context"
	"io"

	"drassi.run/core/pkg/executor/support"
	"github.com/chainguard-dev/clog"
)

func NewTracker() support.Tracker {
	return new(tracker)
}

type tracker struct{}

func (t *tracker) AddIssue(ctx context.Context, issue *support.Issue) error {
	l := clog.FromContext(ctx)
	l.Warnf("Issue: Type=%d, Category=%s, Message=%s, Data=%v", issue.Type, issue.Category, issue.Message, issue.Data)
	return nil
}

func (t *tracker) AttachFile(ctx context.Context, kind, name string, reader io.Reader) error {
	l := clog.FromContext(ctx)
	l.Warnf("AttachFile: Kind=%s, Name=%s", kind, name)
	return nil
}
