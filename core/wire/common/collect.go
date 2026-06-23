/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_common

import (
	exec "drassi.run/core/pkg/executor"
	"go.uber.org/dig"
)

type envProviderParams struct {
	dig.In

	EnvProvider []exec.EnvProvider `group:"env-provider"` // = [wire.EnvProvider]
}

func collectEnvProvider(p envProviderParams) exec.EnvProvider {
	return exec.MultiEnvProvider(p.EnvProvider)
}

type postStartHookParams[R any] struct {
	dig.In

	Hooks []exec.Hook[R] `group:"post-start"` // = [wire.PostStart]
}

func collectPostStartHook[R any](p postStartHookParams[R]) exec.Hook[R] {
	if len(p.Hooks) == 1 {
		return p.Hooks[0]
	}
	return exec.Hooks[R](p.Hooks)
}

type preStopHookParams[R any] struct {
	dig.In

	Hooks []exec.Hook[R] `group:"pre-stop"` // = [wire.PreStop]
}

func collectPreStopHook[R any](p preStopHookParams[R]) exec.Hook[R] {
	if len(p.Hooks) == 1 {
		return p.Hooks[0]
	}
	return exec.Hooks[R](p.Hooks)
}
