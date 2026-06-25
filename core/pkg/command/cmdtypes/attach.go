/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmdtypes

import (
	"context"
	"io"
)

const STEP_SUMMARY = "STEP_SUMMARY"

type Attachment struct {
	Type   string
	Reader io.Reader
}

type Attacher[R any] interface {
	Upload(ctx context.Context, res R, attachment *Attachment) error
}

func BlackHole[R any]() Attacher[R] {
	return blackhole[R]{}
}

type blackhole[R any] struct{}

func (blackhole[R]) Upload(context.Context, R, *Attachment) error {
	return nil
}
