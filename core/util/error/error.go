/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xerror

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

func Recover(err *error) {
	switch r := recover().(type) {
	case nil:
		return
	case error:
		*err = r
	default:
		*err = fmt.Errorf("panic: %v", r)
	}
}

// Refine wraps err with fmt.Errorf if err is non nil.
// Intended for use with defer and a named error return.
// Inspired by https://github.com/golang/go/issues/32676.
func Refine(err *error, f string, v ...any) {
	if *err != nil {
		*err = fmt.Errorf(f+": %w", append(v, *err)...)
	}
}

type Rollbacker struct {
	b  bool
	cu []func(ctx context.Context) error
}

func (r *Rollbacker) Add(fn func(ctx context.Context) error) {
	if fn == nil {
		return
	}
	r.cu = append(r.cu, fn)
}

func (r *Rollbacker) Run(ctx context.Context) error {
	if r.b {
		return nil // success
	}

	ctx = context.WithoutCancel(ctx)
	return r.Do(ctx)
}

func (r *Rollbacker) Do(ctx context.Context) error {
	var errs []error
	for _, fn := range slices.Backward(r.cu) {
		err := fn(ctx)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (r *Rollbacker) Dismiss() {
	r.b = true
}
