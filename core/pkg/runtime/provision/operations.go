/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision

import (
	"context"
	"fmt"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/store/oci"
)

const (
	KeyHostMountDir = StateKey[string]("host_mount_dir")
	KeyMountID      = StateKey[string]("mount_id")
	KeyImage        = StateKey[*ocistore.Image]("image")
)

type Operation[Req any] interface {
	Name() string
	Prepare(pctx *Context) (sandboxer.Cleanup, error)
	PreLaunch(pctx *Context, req Req) (Req, error)
	PostLaunch(pctx *Context, sb sandboxer.Sandbox) (sandboxer.Sandbox, error)
}

type Noop[Req any] struct{}

func (Noop[Req]) Name() string                                  { return "noop" }
func (Noop[Req]) Prepare(_ *Context) (sandboxer.Cleanup, error) { return nil, nil }
func (Noop[Req]) PreLaunch(_ *Context, req Req) (Req, error)    { return req, nil }
func (Noop[Req]) PostLaunch(_ *Context, sb sandboxer.Sandbox) (sandboxer.Sandbox, error) {
	return sb, nil
}

type pullOp[Req any] struct {
	Noop[Req]
	store ocistore.Manager
}

// Pull returns an Operation that checks if the configured image is locally available,
// and pulls it using the provided ocistore.Manager if missing.
func Pull[Req any](store ocistore.Manager) Operation[Req] {
	return &pullOp[Req]{store: store}
}

func (op *pullOp[Req]) Name() string { return "pull" }

func (op *pullOp[Req]) Prepare(pctx *Context) (sandboxer.Cleanup, error) {
	img, err := op.store.Image(pctx, pctx.Config.Image)
	if err != nil {
		return nil, fmt.Errorf("check image %q: %w", pctx.Config.Image, err)
	}
	if img == nil {
		if img, err = op.store.Pull(pctx, pctx.Config.Image); err != nil {
			return nil, fmt.Errorf("pull image %q: %w", pctx.Config.Image, err)
		}
	}
	pctx.Set(KeyImage, img)
	return nil, nil
}

type mountOp[Req any] struct {
	Noop[Req]
	store ocistore.Manager
	opts  []ocistore.MountOption
}

// Mount returns an Operation that mounts the configured runtime image with the given MountOptions,
// records KeyHostMountDir and KeyMountID in Context, and returns an unmount Cleanup closure.
func Mount[Req any](store ocistore.Manager, opts ...ocistore.MountOption) Operation[Req] {
	return &mountOp[Req]{store: store, opts: opts}
}

func (op *mountOp[Req]) Name() string { return "mount" }

func (op *mountOp[Req]) Prepare(pctx *Context) (sandboxer.Cleanup, error) {
	img, ok := pctx.Get(KeyImage)
	if !ok || img == nil {
		return nil, fmt.Errorf("image %q not found in context: pull operation must be used first", pctx.Config.Image)
	}
	mountDir, id, err := op.store.Mount(pctx, img, op.opts...)
	if err != nil {
		return nil, fmt.Errorf("mount image %q: %w", pctx.Config.Image, err)
	}
	pctx.Set(KeyHostMountDir, mountDir)
	pctx.Set(KeyMountID, id)

	cleanup := func(ctx context.Context) error {
		return op.store.Unmount(ctx, id)
	}
	return cleanup, nil
}
