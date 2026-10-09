/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"sync"

	"drassi.run/core/config"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/error"
	"golang.org/x/sync/errgroup"
)

// Launcher is a function that creates a sandbox from a provider-specific request.
type Launcher[Req any] func(ctx context.Context, req Req) (sandboxer.Sandbox, error)

// Provisioner orchestrates multi-runtime provisioning as a Launcher decorator.
// It is completely stateless and safe for concurrent use across multiple launches.
type Provisioner[Req any] struct {
	runtimes map[string]*config.Runtime
	factory  PipelineFactory[Req]
}

// New creates a new Provisioner with explicitly provided runtimes and factory.
func New[Req any](
	runtimes map[string]*config.Runtime,
	factory PipelineFactory[Req],
) *Provisioner[Req] {
	return &Provisioner[Req]{
		runtimes: runtimes,
		factory:  factory,
	}
}

// Launch decorates an inner Launcher with parallel preparation, sequential request transformation,
// inner execution, sequential post-launch actions, and automatic cleanup registration/rollback.
func (p *Provisioner[Req]) Launch(runtimeDir string, l Launcher[Req]) Launcher[Req] {
	if p.Empty() {
		return l
	}

	return func(ctx context.Context, req Req) (sb sandboxer.Sandbox, err error) {
		// 1. Deterministic ordering of runtimes by name
		names := make([]string, 0, len(p.runtimes))
		for name := range p.runtimes {
			names = append(names, name)
		}
		slices.Sort(names)

		pipelines := make([]Pipeline[Req], len(names))
		for i, name := range names {
			pipelines[i] = p.factory(runtimeDir, name, *p.runtimes[name])
		}

		var mu sync.Mutex
		var rb xerror.Rollbacker
		defer rb.Run(ctx)

		// 2. Parallel Prepare across all runtimes: parallel(pull -> mount)
		g, groupCtx := errgroup.WithContext(ctx)
		for i, name := range names {
			pipeline := pipelines[i]
			g.Go(func() error {
				for _, op := range pipeline {
					if c, err := op.Prepare(groupCtx); err != nil {
						return fmt.Errorf("operation %q prepare failed for runtime %q: %w", op.Name(), name, err)
					} else if c != nil {
						mu.Lock()
						rb.Add(c)
						mu.Unlock()
					}
				}
				return nil
			})
		}
		if err = g.Wait(); err != nil {
			return
		}

		// 3. Sequential PreLaunch request mutation
		for i, name := range names {
			for _, op := range pipelines[i] {
				if req, err = op.PreLaunch(ctx, req); err != nil {
					return nil, fmt.Errorf("operation %q pre-launch failed for runtime %q: %w", op.Name(), name, err)
				}
			}
		}

		// 4. Invoke inner launcher with mutated request
		if sb, err = l(ctx, req); err != nil {
			return
		}

		// 5. Sequential PostLaunch actions
		for i, name := range names {
			for _, op := range pipelines[i] {
				next, err := op.PostLaunch(ctx, sb)
				if err != nil {
					sb = cmp.Or(next, sb)
					_ = sb.Terminate(context.WithoutCancel(ctx))
					err = fmt.Errorf("operation %q post-launch failed for runtime %q: %w", op.Name(), name, err)
					return nil, err
				}
				sb = next
			}
		}

		// 6. Success: dismiss rollback and attach all cleanups to sandbox in LIFO order
		rb.Dismiss()
		sb = sandboxer.AddAfterCleanup(sb, rb.Cleanups()...)
		return
	}
}

func (p *Provisioner[Req]) Empty() bool {
	return p == nil || len(p.runtimes) == 0 || p.factory == nil
}
