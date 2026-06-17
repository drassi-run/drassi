/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package worker

import (
	"drassi.run/core/pkg/flag"
	"drassi.run/core/wire"
	"drassi.run/gha-runner/pkg/messages"
)

type variableFlags map[string]messages.Variable

func (f variableFlags) Flag(key string) (string, bool) {
	if v, ok := f[key]; !ok {
		return "", false
	} else {
		return v.Value, true
	}
}

// https://github.com/actions/runner/blob/v2.335.1/src/Runner.Worker/ExecutionContext.cs#L1394-L1417
func initFlag(sysVar map[string]messages.Variable, vars map[string]string) flag.Flags {
	if _, ok := sysVar[wire.RunnerDebug]; !ok {
		if v, ok := vars[wire.RunnerDebug]; ok {
			sysVar[wire.RunnerDebug] = messages.Variable{Value: v}
		}
	}
	if _, ok := sysVar[wire.StepDebug]; !ok {
		if v, ok := vars[wire.StepDebug]; ok {
			sysVar[wire.StepDebug] = messages.Variable{Value: v}
		}
	}
	return variableFlags(sysVar)
}
