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
	"path/filepath"
	"slices"
	"sync"

	"drassi.run/core/config"
	"drassi.run/core/pkg/sandboxer"
	"golang.org/x/sync/errgroup"
)

// Launcher is a function that creates a sandbox from a provider-specific request.
type Launcher[Req any] func(ctx context.Context, req Req) (sandboxer.Sandbox, error)

// Provisioner orchestrates multi-runtime provisioning as a Launcher decorator.
// It is completely stateless and safe for concurrent use across multiple launches.
type Provisioner[Req any] struct {
	runtimes map[string]*config.Runtime
	ops      []Operation[Req]
}

// New creates a new Provisioner with explicitly provided operations. Zero default ops are added.
func New[Req any](
	runtimes map[string]*config.Runtime,
	ops ...Operation[Req],
) *Provisioner[Req] {
	return &Provisioner[Req]{
		runtimes: runtimes,
		ops:      ops,
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

		contexts := make([]*Context, len(names))
		for i, name := range names {
			targetDir := filepath.Join(runtimeDir, name)
			contexts[i] = NewContext(ctx, name, p.runtimes[name], targetDir)
		}

		var (
			mu       sync.Mutex
			cleanups []sandboxer.Cleanup
		)

		rollback := func(ctx context.Context) {
			ctx = context.WithoutCancel(ctx)
			for _, cleanup := range slices.Backward(cleanups) {
				_ = cleanup(ctx)
			}
		}

		// 2. Parallel Prepare across all runtimes: parallel(pull -> mount)
		g, groupCtx := errgroup.WithContext(ctx)
		for _, pctx := range contexts {
			pctx.Context = groupCtx
			g.Go(func() error {
				for _, op := range p.ops {
					if c, err := op.Prepare(pctx); err != nil {
						return fmt.Errorf("operation %q prepare failed for runtime %q: %w", op.Name(), pctx.RuntimeName, err)
					} else if c != nil {
						mu.Lock()
						cleanups = append(cleanups, c)
						mu.Unlock()
					}
				}
				return nil
			})
		}
		if err = g.Wait(); err != nil {
			rollback(ctx)
			return
		}
		for _, pctx := range contexts {
			pctx.Context = ctx
		}

		// 3. Sequential PreLaunch request mutation
		for _, pctx := range contexts {
			for _, op := range p.ops {
				if req, err = op.PreLaunch(pctx, req); err != nil {
					rollback(ctx)
					return nil, fmt.Errorf("operation %q pre-launch failed for runtime %q: %w", op.Name(), pctx.RuntimeName, err)
				}
			}
		}

		// 4. Invoke inner launcher with mutated request
		if sb, err = l(ctx, req); err != nil {
			rollback(ctx)
			return
		}

		// 5. Sequential PostLaunch actions
		for _, pctx := range contexts {
			for _, op := range p.ops {
				next, err := op.PostLaunch(pctx, sb)
				if err != nil {
					sb = cmp.Or(next, sb)
					_ = sb.Terminate(context.WithoutCancel(ctx))
					rollback(ctx)
					return nil, fmt.Errorf("operation %q post-launch failed for runtime %q: %w", op.Name(), pctx.RuntimeName, err)
				}
				sb = next
			}
		}

		// 6. Success: attach all cleanups to sandbox in LIFO order
		if len(cleanups) > 0 {
			slices.Reverse(cleanups)
			sb = sandboxer.AddAfterCleanup(sb, cleanups...)
		}

		return
	}
}

func (p *Provisioner[Req]) Empty() bool {
	return p == nil || len(p.runtimes) == 0 || len(p.ops) == 0
}
