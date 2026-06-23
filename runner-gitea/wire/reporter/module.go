/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_reporter

import (
	"context"
	"fmt"

	runnerv1 "code.gitea.io/actions-proto-go/runner/v1"
	exec "drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/stream"
	xcontext "drassi.run/core/util/context"
	"drassi.run/core/wire"
	"drassi.run/gitea-runner/pkg/gitea"
	"drassi.run/gitea-runner/pkg/reporter"
	"go.uber.org/dig"
)

type options struct {
	task *runnerv1.Task
}

func Module(task *runnerv1.Task) *wire.Module {
	o := options{task: task}

	fn := func(scope *dig.Scope) error {
		if err := scope.Provide(o.newLogStreamer); err != nil {
			return fmt.Errorf("provide reporter.LogStreamer: %w", err)
		}
		if err := scope.Provide(streamHandler); err != nil {
			return fmt.Errorf("provide stream.Handler from reporter.LogStreamer: %w", err)
		}
		if err := scope.Provide(scribeHandler); err != nil {
			return fmt.Errorf("provide scribe.Handler from reporter.LogStreamer.ContextHandle(...): %w", err)
		}

		if err := scope.Provide(o.newReporter); err != nil {
			return fmt.Errorf("provide reporter.LogStreamer: %w", err)
		}
		if err := scope.Provide(ident[*reporter.Reporter],
			dig.As(new(exec.JobRunDecorator), new(exec.StepRunDecorator)),
			dig.Name("reporter"),
		); err != nil {
			return fmt.Errorf("provide reporter.LogStreamer: %w", err)
		}
		return nil
	}
	return wire.NewModule("gitea/log", fn)
}

func (o *options) newLogStreamer(ctxProv xcontext.Provider, client gitea.Client) *reporter.LogStreamer {
	return reporter.NewLogStreamer(o.task.Id, ctxProv, client)
}

func streamHandler(ls *reporter.LogStreamer) stream.Handler {
	return ls
}

func scribeHandler(ls *reporter.LogStreamer) scribe.Handler {
	return ls.ContextHandle
}

func (o *options) newReporter(
	client gitea.Client,
	ctxProv xcontext.Provider,
	ls *reporter.LogStreamer,
	cancel context.CancelCauseFunc,
) *reporter.Reporter {
	return reporter.New(o.task.Id, client, ctxProv, ls, cancel)
}

func ident[T any](t T) T {
	return t
}
