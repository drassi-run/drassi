/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_secret

import (
	"fmt"

	exec "drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/secret"
	"drassi.run/core/pkg/secret/redact"
	"drassi.run/core/wire"
	"go.uber.org/dig"
)

func Module() *wire.Module {
	fn := func(scope *dig.Scope) error {
		if err := scope.Provide(secret.NewMasker); err != nil {
			return fmt.Errorf("provide secret.NewMasker: %w", err)
		}

		if err := scope.Decorate(redact.Attacher[exec.Milieu]); err != nil {
			return fmt.Errorf("decorate cmdtypes.Attacher with secret.Masker: %w", err)
		}
		if err := scope.Provide(redact.Reporter[exec.Milieu], dig.Group("decorator")); err != nil {
			return fmt.Errorf("decorate cmdtypes.Reporter with secret.Masker: %w", err)
		}
		if err := scope.Decorate(redact.ScribeHandler); err != nil {
			return fmt.Errorf("decorate scribe.Handler with secret.Masker: %w", err)
		}
		if err := scope.Provide(redact.JobOutputs, dig.Name("masker")); err != nil {
			return fmt.Errorf("provide 'maskJobOutputs' JobRunDecorator: %w", err)
		}
		return nil
	}
	return wire.NewModule("core/secret", fn)
}
