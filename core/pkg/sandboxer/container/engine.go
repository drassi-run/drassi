/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"context"
	"errors"

	"drassi.run/core/pkg/container/specdef"
	"drassi.run/core/pkg/sandboxer"
)

func WithContainers(e sandboxer.Engine) sandboxer.Engine {
	if _, ok := e.(*containerizedEngine); ok {
		return e
	}
	return &containerizedEngine{Engine: e}
}

type containerizedEngine struct {
	sandboxer.Engine
}

func (e *containerizedEngine) Unwrap() sandboxer.Engine {
	return e.Engine
}

func (e *containerizedEngine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (resp *sandboxer.LaunchResponse, err error) {
	if resp, err = e.Engine.Launch(ctx, req); err != nil {
		return
	}

	if req.JobContainer == nil && len(req.ServiceContainers) == 0 {
		return
	}

	sb := resp.Sandbox
	if resp.ContainerEngine == nil {
		_ = sb.Terminate(ctx)
		return nil, errors.New("sandbox does not support container engine")
	}

	client, err := resp.ContainerEngine(ctx)
	if err != nil {
		_ = sb.Terminate(ctx)
		return nil, err
	}

	b := NewBootstrapper(client, req.Forge)
	defer func(ctx context.Context) {
		if err == nil { // success
			return
		}
		ctx = context.WithoutCancel(ctx)
		_ = b.Rollback(ctx)
		_ = sb.Terminate(ctx)
		resp = nil
	}(ctx)

	var jobSb sandboxer.Sandbox
	if req.JobContainer != nil {
		target := DefaultLayout
		opt := specdef.AddMount(resp.Mounter.OverlayMounts(target, true)...)

		if resp.JobContainer, err = b.RunJobContainer(ctx, req.JobContainer, opt); err != nil {
			return
		}
		if jobSb, err = NewSandbox(ctx, client, resp.JobContainer.Id, target); err != nil {
			return
		}
	}

	if len(req.ServiceContainers) > 0 {
		if resp.ServiceContainers, err = b.RunServiceContainers(ctx, req.ServiceContainers); err != nil {
			return
		}
	}

	resp.Sandbox = b.WrapSandbox(sb, jobSb)
	return
}
