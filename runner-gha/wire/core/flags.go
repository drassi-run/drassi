/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_core

import (
	"drassi.run/core/pkg/feature"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/wire"
	"drassi.run/gha-runner/pkg/messages"
)

type sysVarFlags map[string]messages.Variable

func (f sysVarFlags) Flag(key string) (string, bool) {
	if v, ok := f[key]; !ok {
		return "", false
	} else {
		return v.Value, true
	}
}

// https://github.com/actions/runner/blob/v2.335.1/src/Runner.Worker/ExecutionContext.cs#L1394-L1417
func newSysVarFlags(req *messages.PipelineAgentJobRequest, dossier *records.Dossier) feature.Flags {
	sysVar := req.Variables
	dVar := dossier.Variables
	if _, ok := sysVar[wire.RunnerDebug]; !ok {
		if v, ok := dVar[wire.RunnerDebug]; ok {
			sysVar[wire.RunnerDebug] = messages.Variable{Value: v}
		}
	}
	if _, ok := sysVar[wire.StepDebug]; !ok {
		if v, ok := dVar[wire.StepDebug]; ok {
			sysVar[wire.StepDebug] = messages.Variable{Value: v}
		}
	}
	return sysVarFlags(sysVar)
}
