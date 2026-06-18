/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package worker

import "drassi.run/gha-runner/pkg/messages"

type variableFlags map[string]messages.Variable

func (f variableFlags) Flag(key string) (string, bool) {
	if v, ok := f[key]; !ok {
		return "", false
	} else {
		return v.Value, true
	}
}
