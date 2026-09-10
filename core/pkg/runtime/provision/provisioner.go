/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision

import (
	"context"
	"fmt"
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
	runtimes    map[string]*config.Runtime
	targetDirFn func(name string) string
	ops         []Operation[Req]
}

// New creates a new Provisioner with explicitly provided operations. Zero default ops are added.
func New[Req any](
	runtimes map[string]*config.Runtime,
	targetDirFn func(name string) string,
	ops ...Operation[Req],
) *Provisioner[Req] {
	return &Provisioner[Req]{
		runtimes:    runtimes,
		targetDirFn: targetDirFn,
		ops:         ops,
	}
}

// Launch decorates an inner Launcher with parallel preparation, sequential request transformation,
// inner execution, sequential post-launch actions, and automatic cleanup registration/rollback.
func (p *Provisioner[Req]) Launch(ctx context.Context, l Launcher[Req]) Launcher[Req] {
	return func(launchCtx context.Context, req Req) (sandboxer.Sandbox, error) {
		if p == nil || len(p.runtimes) == 0 || len(p.ops) == 0 {
			return l(launchCtx, req)
		}

		// 1. Deterministic ordering of runtimes by name
		names := make([]string, 0, len(p.runtimes))
		for name := range p.runtimes {
			names = append(names, name)
		}
		slices.Sort(names)

		contexts := make([]*Context, len(names))
		for i, name := range names {
			targetDir := ""
			if p.targetDirFn != nil {
				targetDir = p.targetDirFn(name)
			}
			contexts[i] = NewContext(launchCtx, name, p.runtimes[name], targetDir)
		}

		var (
			mu       sync.Mutex
			cleanups []sandboxer.Cleanup
		)

		rollback := func() {
			for _, cleanup := range slices.Backward(cleanups) {
				_ = cleanup(launchCtx)
			}
		}

		// 2. Parallel Prepare across all runtimes: parallel(pull -> mount)
		g, _ := errgroup.WithContext(launchCtx)
		for _, pctx := range contexts {
			pctx := pctx
			g.Go(func() error {
				for _, op := range p.ops {
					cleanup, err := op.Prepare(pctx)
					if err != nil {
						return fmt.Errorf("operation %q prepare failed for runtime %q: %w", op.Name(), pctx.RuntimeName, err)
					}
					if cleanup != nil {
						mu.Lock()
						cleanups = append(cleanups, cleanup)
						mu.Unlock()
					}
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			rollback()
			return nil, err
		}

		// 3. Sequential PreLaunch request mutation
		var err error
		for _, pctx := range contexts {
			for _, op := range p.ops {
				req, err = op.PreLaunch(pctx, req)
				if err != nil {
					rollback()
					return nil, fmt.Errorf("operation %q pre-launch failed for runtime %q: %w", op.Name(), pctx.RuntimeName, err)
				}
			}
		}

		// 4. Invoke inner launcher with mutated request
		sb, err := l(launchCtx, req)
		if err != nil {
			rollback()
			return nil, err
		}

		// 5. Sequential PostLaunch actions
		for _, pctx := range contexts {
			for _, op := range p.ops {
				nextSb, err := op.PostLaunch(pctx, sb)
				if err != nil {
					if nextSb != nil {
						_ = nextSb.Terminate(launchCtx)
					} else if sb != nil {
						_ = sb.Terminate(launchCtx)
					}
					rollback()
					return nil, fmt.Errorf("operation %q post-launch failed for runtime %q: %w", op.Name(), pctx.RuntimeName, err)
				}
				sb = nextSb
			}
		}

		// 6. Success: attach all cleanups to sandbox in LIFO order
		if len(cleanups) > 0 {
			sb = sandboxer.AddAfterCleanup(sb, cleanups...)
		}

		return sb, nil
	}
}
