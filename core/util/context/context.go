package xcontext

import (
	"context"
	"time"
)

const elapse = 3 * time.Second

func ExpandTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if dead, ok := ctx.Deadline(); ok {
		if time.Until(dead) > elapse {
			return context.WithCancel(ctx)
		}
	}
	ctx = context.WithoutCancel(ctx)
	return context.WithTimeout(ctx, elapse)
}
