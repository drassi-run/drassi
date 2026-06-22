/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_scribe

import (
	"context"
	"fmt"

	"drassi.run/core/pkg/feature"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/secret"
	"drassi.run/core/wire"
	"go.uber.org/dig"
)

func Module() *wire.Module {
	fn := func(scope *dig.Scope) error {
		if err := scope.Decorate(maskSecretHandler); err != nil {
			return fmt.Errorf("decorate scribe.Handler with secret.Masker: %w", err)
		}
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

func maskSecretHandler(h scribe.Handler, sm secret.Masker) scribe.Handler {
	return func(ctx context.Context, line string) error {
		line = sm.Mask(line)
		return h(ctx, line)
	}
}

func configureDiary(diary scribe.Diary, flags feature.Flags) scribe.Diary {
	debug := feature.Bool(flags, wire.StepDebug, false)
	diary.SetDebug(debug)

	return diary
}
