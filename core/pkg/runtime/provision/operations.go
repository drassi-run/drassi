/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision

import (
	"fmt"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/store/oci"
)

const (
	KeyHostMountDir = StateKey[string]("host_mount_dir")
	KeyMountID      = StateKey[string]("mount_id")
	KeyImage        = StateKey[*ocistore.Image]("image")
)

type Operation interface {
	Name() string
	PreLaunch(pctx *Context) error
	PostLaunch(pctx *Context, sb sandboxer.Sandbox) error
}

type Noop struct{}

func (op Noop) Name() string                                     { return "noop" }
func (op Noop) PreLaunch(_ *Context) error                       { return nil }
func (op Noop) PostLaunch(_ *Context, _ sandboxer.Sandbox) error { return nil }

type OpFunc struct {
	PreFunc  func(ctx *Context) error
	PostFunc func(ctx *Context, sb sandboxer.Sandbox) error
}

func (op *OpFunc) Name() string { return "func" }

func (op *OpFunc) PreLaunch(pctx *Context) error {
	if fn := op.PreFunc; fn != nil {
		return fn(pctx)
	}
	return nil
}

func (op *OpFunc) PostLaunch(pctx *Context, sb sandboxer.Sandbox) error {
	if fn := op.PostFunc; fn != nil {
		return fn(pctx, sb)
	}
	return nil
}

type pullOp struct {
	Noop
	store ocistore.Manager
}

// Pull returns an Operation that checks if the configured image is locally available,
// and pulls it using the provided ocistore.Manager if missing.
func Pull(store ocistore.Manager) Operation {
	return &pullOp{store: store}
}

func (op *pullOp) Name() string { return "pull" }

func (op *pullOp) PreLaunch(pctx *Context) error {
	img, err := op.store.Image(pctx, pctx.Config.Image)
	if err != nil {
		return fmt.Errorf("check image %q: %w", pctx.Config.Image, err)
	}
	if img == nil {
		if img, err = op.store.Pull(pctx, pctx.Config.Image); err != nil {
			return fmt.Errorf("pull image %q: %w", pctx.Config.Image, err)
		}
	}
	pctx.Set(KeyImage, img)
	return nil
}

type mountOp struct {
	Noop
	store ocistore.Manager
	opts  []ocistore.MountOption
}

// Mount returns an Operation that mounts the configured runtime image with the given MountOptions,
// and records KeyHostMountDir and KeyMountID in the Context.
func Mount(store ocistore.Manager, opts ...ocistore.MountOption) Operation {
	return &mountOp{store: store, opts: opts}
}

func (op *mountOp) Name() string { return "mount" }

func (op *mountOp) PreLaunch(pctx *Context) error {
	img, ok := pctx.Get(KeyImage)
	if !ok || img == nil {
		return fmt.Errorf("image %q not found in context: pull operation must be used first", pctx.Config.Image)
	}
	mountDir, id, err := op.store.Mount(pctx, img, op.opts...)
	if err != nil {
		return fmt.Errorf("mount image %q: %w", pctx.Config.Image, err)
	}
	pctx.Set(KeyHostMountDir, mountDir)
	pctx.Set(KeyMountID, id)
	return nil
}
