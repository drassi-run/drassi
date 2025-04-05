/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scribe

import "context"

type Diary interface {
	SetDebug(b bool)
	EnableDebug() bool
	Write(ctx context.Context, message string) error
}

type discard struct{}

func (d discard) SetDebug(bool)                       {}
func (d discard) EnableDebug() bool                   { return false }
func (d discard) Write(context.Context, string) error { return nil }
