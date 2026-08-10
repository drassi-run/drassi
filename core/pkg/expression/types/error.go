/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"errors"
	"fmt"

	"drassi.run/core/pkg/expression/types/ref"
)

type Error struct {
	error
}

func NewError(format string, a ...any) *Error {
	e := fmt.Errorf(format, a...)
	return &Error{e}
}

func WrapError(e error) *Error {
	return &Error{e}
}

func (e *Error) Type() ref.Type {
	return ref.TypeInvalid
}

func (e *Error) Value() any {
	return e.error
}

func (e *Error) Equal(_ ref.Val) bool {
	return false
}

func (e *Error) Unwrap() error {
	return e.error
}

func (e *Error) MarshalText() ([]byte, error) {
	return nil, errors.New("error is non-marshalable")
}
