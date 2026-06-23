/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire

import (
	"fmt"

	"go.uber.org/dig"
)

const (
	PostStart   = "post-start"
	PreStop     = "pre-stop"
	EnvProvider = "env-provider"
)

const (
	RunnerDebug = "ACTIONS_RUNNER_DEBUG"
	StepDebug   = "ACTIONS_STEP_DEBUG"
)

type Module struct {
	name  string
	apply func(*dig.Scope) error
}

func (m *Module) Name() string {
	return m.name
}

func NewModule(name string, fn func(*dig.Scope) error) *Module {
	return &Module{name: name, apply: fn}
}

func Apply(scope *dig.Scope, modules ...*Module) error {
	for _, module := range modules {
		if err := module.apply(scope); err != nil {
			return fmt.Errorf("apply module %q: %w", module.name, err)
		}
	}
	return nil
}
