/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"drassi.run/core/pkg/feature"
	"drassi.run/gha-runner/pkg/messages"
)

func NewSysVarFlags(m map[string]messages.Variable) feature.Flags {
	return sysVarFlags(m)
}

type sysVarFlags map[string]messages.Variable

func (f sysVarFlags) Flag(key string) (string, bool) {
	if v, ok := f[key]; !ok {
		return "", false
	} else {
		return v.Value, true
	}
}
