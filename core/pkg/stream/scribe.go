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

type scribeDiary struct {
	debug   bool
	handler Handler
}

func NewScribeDiary(h Handler) scribe.Diary {
	return &scribeDiary{handler: h}
}

func (d *scribeDiary) SetDebug(b bool) {
	d.debug = b
}

func (d *scribeDiary) EnableDebug() bool {
	return d.debug
}

func (d *scribeDiary) Write(ctx context.Context, message string) error {
	return d.handler.Handle(ctx, message)
}
