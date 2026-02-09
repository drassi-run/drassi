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
	"github.com/chainguard-dev/clog"
	"go.uber.org/dig"
)

func Wire(scope *dig.Scope) error {
	if err := provideStack(scope); err != nil {
		return err
	}

	if err := provideTelemetry(scope); err != nil {
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
	return scope.Provide(support.NewStack,
		dig.As(new(support.Stack), new(executor.JobRunDecorator), new(executor.StepRunDecorator)),
		dig.Name("stack"),
	)
}

func provideTelemetry(scope *dig.Scope) error {
	return scope.Provide(support.NewTelemetry,
		dig.As(new(executor.JobRunDecorator), new(executor.StepRunDecorator), new(executor.ActionRunDecorator)),
		dig.Name("telemetry"),
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
