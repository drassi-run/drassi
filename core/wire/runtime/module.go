/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_runtime

import (
	"fmt"

	"drassi.run/core/config"
	"drassi.run/core/pkg/runtime"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/dig"
	"drassi.run/core/wire"
	"go.uber.org/dig"
)

type Option func(o *options)

type options struct {
	runtimeConfigs map[string]*config.Runtime
}

func WithRuntimes(cfg map[string]*config.Runtime) Option {
	return func(o *options) {
		o.runtimeConfigs = cfg
	}
}

func Module(opts ...Option) *wire.Module {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	fn := func(scope *dig.Scope) error {
		if o.runtimeConfigs != nil {
			if err := xdig.Supply(scope, o.runtimeConfigs); err != nil {
				return fmt.Errorf("supply runtimes: %w", err)
			}
		}
		if err := scope.Provide(NewContainerRuntime); err != nil {
			return fmt.Errorf("provide runtime.Container: %w", err)
		}
		if err := scope.Provide(o.newProvider); err != nil {
			return fmt.Errorf("provide runtime.Provider: %w", err)
		}
		return nil
	}
	return wire.NewModule("core/runtime", fn)
}

func (o *options) newProvider(sb sandboxer.Sandbox) (runtime.Provider, error) {
	return runtime.NewProvider(sb, o.runtimeConfigs)
}
