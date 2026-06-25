/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_scribe

import (
	"fmt"

	"drassi.run/core/pkg/feature"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/wire"
	"go.uber.org/dig"
)

func Module() *wire.Module {
	fn := func(scope *dig.Scope) error {
		if err := scope.Provide(scribe.NewForwardDiary); err != nil {
			return fmt.Errorf("provide scribe.ForwardDiary: %w", err)
		}
		if err := scope.Decorate(configureDiary); err != nil {
			return fmt.Errorf("configure scribe.Diary: %w", err)
		}

		return nil
	}
	return wire.NewModule("core/scribe", fn)
}

func configureDiary(diary scribe.Diary, flags feature.Flags) scribe.Diary {
	debug := feature.Bool(flags, wire.StepDebug, false)
	diary.SetDebug(debug)

	return diary
}
