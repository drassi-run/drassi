/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_log

import (
	"context"
	"fmt"

	exec "drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/stream"
	"drassi.run/core/wire"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/types"
	"go.uber.org/dig"
)

type Option func(o *options)
type options struct {
	dir        string
	maxLogSize int64
}

func WithMaxLogSize(n int64) Option {
	return func(o *options) {
		o.maxLogSize = n
	}
}

func WithDir(dir string) Option {
	return func(o *options) {
		o.dir = dir
	}
}

func Module(opts ...Option) *wire.Module {
	o := &options{
		dir:        "/tmp/gha-runner/",
		maxLogSize: int64(100) * 1024 * 1024, // 100MiB
	}
	for _, opt := range opts {
		opt(o)
	}

	fn := func(scope *dig.Scope) error {
		if err := scope.Provide(o.newLogManager); err != nil {
			return fmt.Errorf("provide log.Manager: %w", err)
		}
		if err := scope.Provide(streamSinkFactory); err != nil {
			return fmt.Errorf("provide stream.SinkFactory from log.Manager: %w", err)
		}
		if err := scope.Provide(scribeHandler); err != nil {
			return fmt.Errorf("provide scribe.Handler from log.Manager: %w", err)
		}
		if err := scope.Provide(log.NewDecorator, dig.Name("log")); err != nil {
			return fmt.Errorf("provide log.Decorator as %q StepRunDecorator: %w", "log", err)
		}
		return nil
	}
	return wire.NewModule("gha/log", fn)
}

func (o *options) newLogManager() (*log.Manager, error) {
	return log.NewManager(o.dir, o.maxLogSize)
}

type logStreamSink struct {
	logMgr *log.Manager
	store  types.RecordStore
}

func (s *logStreamSink) Emit(_ context.Context, res exec.Milieu, line string) error {
	stepUid := res.StepSpec().Uid
	uid := s.store.RecordUid(res.Stage(), stepUid)
	return s.logMgr.Handle(uid, line)
}

func streamSinkFactory(logMgr *log.Manager, store types.RecordStore) stream.SinkFactory[exec.Milieu] {
	sink := &logStreamSink{logMgr: logMgr, store: store}
	return func(_ string) stream.Sink[exec.Milieu] {
		return sink
	}
}

func scribeHandler(logMgr *log.Manager) scribe.Handler {
	return func(ctx context.Context, msg string) error {
		if uid, ok := log.StepUidFromContext(ctx); ok {
			return logMgr.Handle(uid, msg)
		}
		return nil
	}
}
