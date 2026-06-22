/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_common

import (
	"fmt"

	exec "drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/feature"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/problem"
	"drassi.run/core/pkg/secret"
	xdig "drassi.run/core/util/dig"
	"drassi.run/core/wire"
	"go.uber.org/dig"
)

type Option func(o *options)
type options struct {
	defaultFeatureFlags  bool // use feature.Empty as feature.Flags
	defaultExpressionEnv bool // create empty expression.Env
	defaultDossier       bool // create default records.Dossier
}

func UseEmptyFeatureFlags(b bool) Option {
	return func(o *options) {
		o.defaultFeatureFlags = b
	}
}

func ProvideDefaultExpressionEnv(b bool) Option {
	return func(o *options) {
		o.defaultExpressionEnv = b
	}
}

func ProvideDefaultDossier(b bool) Option {
	return func(o *options) {
		o.defaultDossier = b
	}
}

func Module(opts ...Option) *wire.Module {
	o := &options{
		defaultFeatureFlags:  true,
		defaultExpressionEnv: true,
		defaultDossier:       true,
	}
	for _, opt := range opts {
		opt(o)
	}

	fn := func(scope *dig.Scope) error {
		if err := xdig.Supply(scope, exec.CIEnv, dig.Group(wire.EnvProvider)); err != nil {
			return fmt.Errorf("provide CIEnv: %w", err)
		}

		if err := scope.Provide(secret.NewMasker); err != nil {
			return fmt.Errorf("provide secret.NewMasker: %w", err)
		}

		if err := scope.Provide(problem.NewMatchers); err != nil {
			return fmt.Errorf("provide problem.Matchers: %w", err)
		}

		if o.defaultDossier {
			if err := scope.Provide(newDossier); err != nil {
				return fmt.Errorf("provide default records.Dossier: %w", err)
			}
		}

		if err := scope.Provide(getGitHub); err != nil {
			return fmt.Errorf("provide records.GitHub: %w", err)
		}

		if err := scope.Provide(getEnv); err != nil {
			return fmt.Errorf("provide Env: %w", err)
		}

		if o.defaultExpressionEnv {
			if err := scope.Provide(expression.NewEnv); err != nil {
				return fmt.Errorf("provide default expression.Env: %w", err)
			}
		}

		if o.defaultFeatureFlags {
			if err := xdig.Supply(scope, feature.Empty); err != nil {
				return fmt.Errorf("provide default (empty) feature.Flags: %w", err)
			}
		}

		if err := scope.Provide(collectEnvProvider); err != nil {
			return fmt.Errorf("collect EnvProviders: %w", err)
		}
		if err := scope.Provide(collectPostStartHook[exec.JobExecutor], dig.Name(wire.PostStart)); err != nil {
			return fmt.Errorf("collect 'post-start' Hooks: %w", err)
		}
		if err := scope.Provide(collectPreStopHook[exec.JobExecutor], dig.Name(wire.PreStop)); err != nil {
			return fmt.Errorf("collect 'pre-stop' Hooks: %w", err)
		}

		return nil
	}
	return wire.NewModule("core/common", fn)
}

func newDossier() *records.Dossier {
	d := new(records.Dossier)
	d.Github = new(records.Github)
	d.Env = make(map[string]string)
	return d
}

func getGitHub(d *records.Dossier) *records.Github {
	return d.Github
}

func getEnv(d *records.Dossier) map[string]string {
	return d.Env
}
