/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xsync

import (
	"context"
	"sync"
)

// Singleton ensure f invoke once, the same as sync.OnceValues does except it take context.Context as param
func Singleton[R any](f func(context.Context) (R, error)) func(context.Context) (R, error) {
	// Use a struct so that there's a single heap allocation.
	d := struct {
		f     func(context.Context) (R, error)
		once  sync.Once
		valid bool
		r     R // result
		e     error
		p     any // panic
	}{
		f: f,
	}
	return func(ctx context.Context) (R, error) {
		d.once.Do(func() {
			defer func() {
				d.f = nil
				d.p = recover()
				if !d.valid {
					panic(d.p)
				}
			}()
			d.r, d.e = d.f(ctx)
			d.valid = true
		})
		if !d.valid {
			panic(d.p)
		}
		return d.r, d.e
	}
}
