/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package etc

import (
	"drassi.run/core/pkg/executor"
	"drassi.run/core/util/context"
	"drassi.run/core/util/dig"
	"go.uber.org/dig"
)

func Wire(scope *dig.Scope) error {
	return scope.Invoke(provideEnv)
}

func provideEnv(sup executor.Supervisor) {
	m := map[string]string{
		"CI":             "true",
		"GITHUB_ACTIONS": "true",
	}
	sup.Register(executor.Env(m))
}

func x(scope *dig.Scope) error {
	s := new(stack)
	err := xdig.Supply(scope, s,
		dig.As(new(executor.Stack), new(xcontext.Provider)),
	)
	if err != nil {
		return err
	}

	l := &listener{stack: s}
	return xdig.Supply(scope, l,
		dig.As(new(executor.JobListener), new(executor.StepListener)),
		dig.Name("stack"),
	)
}
