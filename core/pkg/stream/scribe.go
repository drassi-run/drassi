/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"context"
	"drassi.run/core/pkg/scribe"
)

type scribeOutput struct {
	debug   bool
	handler Handler
}

func NewScribeOutput(h Handler) scribe.Output {
	return &scribeOutput{handler: h}
}

func (h *scribeOutput) SetDebug(b bool) {
	h.debug = b
}

func (h *scribeOutput) EnableDebug() bool {
	return h.debug
}

func (h *scribeOutput) Inscribe(ctx context.Context, message string) error {
	return h.handler.Handle(ctx, message)
}
