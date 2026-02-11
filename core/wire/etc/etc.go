/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package etc

import (
	"context"
	"io"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/util/context"
	"drassi.run/core/util/dig"
	"github.com/chainguard-dev/clog"
	"go.uber.org/dig"
)

func Wire(scope *dig.Scope) error {
	if err := provideStack(scope); err != nil {
		return err
	}

	if err := scope.Provide(NewTracker); err != nil {
		return err
	}

	if err := scope.Provide(support.NewEnvProvider); err != nil {
		return err
	}

	return scope.Invoke(provideEnv)
}

func provideEnv(envProv support.EnvProvider) {
	envProv.ProvideEnv(support.CIEnv())
}

func provideStack(scope *dig.Scope) error {
	s := new(stack)
	err := xdig.Supply(scope, s,
		dig.As(new(executor.Stack), new(xcontext.Provider)),
	)
	if err != nil {
		return err
	}

	l := &stackListener{stack: s}
	return xdig.Supply(scope, l,
		dig.As(new(executor.JobListener), new(executor.StepListener)),
		dig.Name("stack"),
	)
}

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
