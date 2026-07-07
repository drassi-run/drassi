/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package redact

import (
	"context"

	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/secret"
)

func ScribeHandler(h scribe.Handler, sm secret.Masker) scribe.Handler {
	return func(ctx context.Context, line string) error {
		line = sm.Mask(line)
		return h(ctx, line)
	}
}
