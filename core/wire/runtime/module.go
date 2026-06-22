/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_runtime

import (
	"fmt"

	"drassi.run/core/wire"
	"go.uber.org/dig"
)

func Module() *wire.Module {
	fn := func(scope *dig.Scope) error {
		if err := scope.Provide(NewContainerRuntime); err != nil {
			return fmt.Errorf("provide runtime.Container: %w", err)
		}
		return nil
	}
	return wire.NewModule("core/runtime", fn)
}
