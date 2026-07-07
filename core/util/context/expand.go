/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xcontext

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

const defaultExpands = 15 * time.Second

type expanderKey struct{}

type expander = atomic.Pointer[time.Time]

// WithExpander add expander into ctx, which can be expanded with duration later
func WithExpander(ctx context.Context) context.Context {
	return context.WithValue(ctx, expanderKey{}, new(expander))
}

func Expand(ctx context.Context, d time.Duration) error {
	if e, ok := ctx.Value(expanderKey{}).(*expander); !ok {
		return fmt.Errorf("context does not have expander")
	} else if !e.CompareAndSwap(nil, new(time.Now().Add(d))) {
		return fmt.Errorf("expander.Time already set")
	}
	return nil
}

func ExpandContext(ctx context.Context, err error) (context.Context, context.CancelFunc) {
	if e, ok := ctx.Value(expanderKey{}).(*expander); !ok {
		return ctx, nil
	} else if t := e.Load(); t != nil {
		return context.WithDeadlineCause(ctx, *t, err)
	}

	return context.WithTimeout(ctx, defaultExpands)
}
