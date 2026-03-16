/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scribe

import "context"

type Handler func(context.Context, string) error

type fwDiary struct {
	debug   bool
	handler Handler
}

func NewForwardDiary(h Handler) Diary {
	return &fwDiary{handler: h}
}

func (d *fwDiary) SetDebug(b bool) {
	d.debug = b
}

func (d *fwDiary) EnableDebug() bool {
	return d.debug
}

func (d *fwDiary) Write(ctx context.Context, message string) error {
	return d.handler(ctx, message)
}
