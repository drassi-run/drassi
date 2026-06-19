/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_timeline

import (
	"fmt"
	"time"

	exec "drassi.run/core/pkg/executor"
	"drassi.run/core/wire"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/timeline"
	"go.uber.org/dig"
)

type Option func(o *options)
type options struct {
	interval time.Duration
}

func WithInterval(d time.Duration) Option {
	return func(o *options) {
		o.interval = d
	}
}

func Module(opts ...Option) *wire.Module {
	o := &options{
		interval: time.Second,
	}
	for _, opt := range opts {
		opt(o)
	}

	fn := func(scope *dig.Scope) error {
		if err := scope.Provide(o.newTimelineManager); err != nil {
			return fmt.Errorf("provide timeline.Manager: %w", err)
		}

		if err := scope.Provide(iden, dig.As(new(log.RecordStore))); err != nil {
			return fmt.Errorf("provide timeline.Manager as log.RecordStore: %w", err)
		}

		if err := scope.Provide(iden,
			dig.As(new(exec.JobRunDecorator), new(exec.StepRunDecorator)),
			dig.Name("timeline"),
		); err != nil {
			return fmt.Errorf("provide timeline.Manager as %q JobRunDecorator & StepRunDecorator: %w", "timeline", err)
		}

		if err := scope.Provide(timeline.NewIssueReporter); err != nil {
			return fmt.Errorf("provide timeline.Manager as cmdtypes.Reporter: %w", err)
		}

		return nil
	}
	return wire.NewModule("log/timeline", fn)
}

func (o *options) newTimelineManager(rec timeline.Recorder) *timeline.Manager {
	return timeline.NewManager(o.interval, rec)
}

func iden(mgr *timeline.Manager) *timeline.Manager {
	return mgr
}
