/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"errors"
	"io"
)

type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

func (s *Streams) Close() error {
	errs := make([]error, 0, 3)
	if c, ok := s.In.(io.Closer); ok {
		errs = append(errs, c.Close())
	}
	if c, ok := s.Out.(io.Closer); ok {
		errs = append(errs, c.Close())
	}
	if c, ok := s.Err.(io.Closer); ok {
		errs = append(errs, c.Close())
	}
	return errors.Join(errs...)
}
