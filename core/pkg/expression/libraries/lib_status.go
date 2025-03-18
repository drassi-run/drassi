/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package libraries

import (
	expr "drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/records"
)

type StatusProvider interface {
	Status() records.Result
}

func StatusLib(p StatusProvider) expr.Library {
	return &statusLib{StatusProvider: p}
}

type statusLib struct {
	StatusProvider
}

func (lib *statusLib) EnvOptions() []expr.Option {
	opts := []expr.Option{
		expr.WithFunction("success", nullaryFn(lib.Success)),
		expr.WithFunction("always", nullaryFn(lib.Always)),
		expr.WithFunction("cancelled", nullaryFn(lib.Cancelled)),
		expr.WithFunction("failure", nullaryFn(lib.Failure)),
	}

	return opts
}
