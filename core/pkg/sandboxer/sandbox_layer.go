/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package sandboxer

import (
	"context"
	"errors"
)

type layeredSandbox struct {
	Sandbox
	Base Sandbox
}

func NewLayeredSandbox(main Sandbox, base Sandbox) Sandbox {
	return &layeredSandbox{
		Sandbox: main,
		Base:    base,
	}
}

func (sb *layeredSandbox) Terminate(ctx context.Context) error {
	errs := make([]error, 2)
	errs[0] = sb.Sandbox.Terminate(ctx)
	errs[1] = sb.Base.Terminate(ctx)

	return errors.Join(errs...)
}

func (sb *layeredSandbox) Underlay() Sandbox {
	return sb.Base
}
