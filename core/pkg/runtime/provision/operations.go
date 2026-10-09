/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision

import (
	"context"
	"fmt"

	"drassi.run/core/config"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/store/oci"
)

type PipelineFactory[Req any] func(runtimeDir string, name string, rt config.Runtime) Pipeline[Req]
type Pipeline[Req any] []Operation[Req]

type Operation[Req any] interface {
	Name() string
	Prepare(ctx context.Context) (sandboxer.Cleanup, error)
	PreLaunch(ctx context.Context, req Req) (Req, error)
	PostLaunch(ctx context.Context, sb sandboxer.Sandbox) (sandboxer.Sandbox, error)
}

type Noop[Req any] struct{}

func (Noop[Req]) Name() string                                         { return "noop" }
func (Noop[Req]) Prepare(_ context.Context) (sandboxer.Cleanup, error) { return nil, nil }
func (Noop[Req]) PreLaunch(_ context.Context, req Req) (Req, error)    { return req, nil }
func (Noop[Req]) PostLaunch(_ context.Context, sb sandboxer.Sandbox) (sandboxer.Sandbox, error) {
	return sb, nil
}

type pullOp[Req any] struct {
	Noop[Req]
	store ocistore.Manager
	state *State
	image string
}

// Pull returns an Operation that pulls the specified image using the provided ocistore.Manager.
func Pull[Req any](store ocistore.Manager, state *State, image string) Operation[Req] {
	return &pullOp[Req]{
		store: store,
		state: state,
		image: image,
	}
}

func (op *pullOp[Req]) Name() string { return "pull" }

func (op *pullOp[Req]) Prepare(ctx context.Context) (sandboxer.Cleanup, error) {
	img, err := op.store.Pull(ctx, op.image)
	if err != nil {
		return nil, fmt.Errorf("pull image %q: %w", op.image, err)
	}
	op.state.Set(KeyImage, img)
	return nil, nil
}

type mountOp[Req any] struct {
	Noop[Req]
	store ocistore.Manager
	state *State
	opts  []ocistore.MountOption
}

// Mount returns an Operation that mounts the configured runtime image with the given options
func Mount[Req any](store ocistore.Manager, state *State, opts ...ocistore.MountOption) Operation[Req] {
	return &mountOp[Req]{
		store: store,
		state: state,
		opts:  opts,
	}
}

func (op *mountOp[Req]) Name() string { return "mount" }

func (op *mountOp[Req]) Prepare(ctx context.Context) (sandboxer.Cleanup, error) {
	img, ok := op.state.Get(KeyImage)
	if !ok || img == nil {
		return nil, fmt.Errorf("image not found in state: pull operation must be used first")
	}
	mnt, err := op.store.Mount(ctx, img, op.opts...)
	if err != nil {
		return nil, fmt.Errorf("mount image: %w", err)
	}
	return mnt.Release, nil
}
