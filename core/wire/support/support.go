/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_support

import (
	exec "drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/wire"
	"go.uber.org/dig"
)

func Wire(scope *dig.Scope) error {
	if err := provideTelemetry(scope); err != nil {
		return err
	}
	if err := provideEnv(scope); err != nil {
		return err
	}

	if err := scope.Provide(scrapeEnvProvider); err != nil {
		return err
	}
	if err := scope.Provide(scrapePostStartHook[exec.JobExecutor], dig.Name(wire.PostStart)); err != nil {
		return err
	}
	if err := scope.Provide(scrapePreStopHook[exec.JobExecutor], dig.Name(wire.PreStop)); err != nil {
		return err
	}

	return scope.Provide(NewTracker)
}

func provideTelemetry(scope *dig.Scope) error {
	return scope.Provide(support.NewTelemetry,
		dig.Name("telemetry"),
		dig.As(
			new(exec.JobRunDecorator),
			new(exec.StepRunDecorator),
			new(exec.ActionRunDecorator),
		),
	)
}

func provideEnv(scope *dig.Scope) error {
	return scope.Provide(exec.CIEnv, dig.Group(wire.EnvProvider))
}

type envProviderParams struct {
	dig.In

	EnvProvider []exec.EnvProvider `group:"env-provider"` // = [wire.EnvProvider]
}

func scrapeEnvProvider(p envProviderParams) exec.EnvProvider {
	return exec.MultiEnvProvider(p.EnvProvider)
}

type postStartHookParams[R any] struct {
	dig.In

	Hooks []exec.Hook[R] `group:"post-start"` // = [wire.PostStart]
}

func scrapePostStartHook[R any](p postStartHookParams[R]) exec.Hook[R] {
	return exec.Hooks[R](p.Hooks)
}

type preStopHookParams[R any] struct {
	dig.In

	Hooks []exec.Hook[R] `group:"pre-stop"` // = [wire.PreStop]
}

func scrapePreStopHook[R any](p preStopHookParams[R]) exec.Hook[R] {
	return exec.Hooks[R](p.Hooks)
}
