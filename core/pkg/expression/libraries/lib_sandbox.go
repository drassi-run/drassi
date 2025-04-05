/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package libraries

import (
	expr "drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/context"
)

func SandboxLib(c xcontext.Provider, sb sandboxer.Sandbox) expr.Library {
	return &sandboxLib{contextual: c, sandbox: sb}
}

type sandboxLib struct {
	contextual xcontext.Provider
	sandbox    sandboxer.Sandbox
}

func (s *sandboxLib) EnvOptions() []expr.Option {
	// TODO add hashFiles function
	return nil
}
