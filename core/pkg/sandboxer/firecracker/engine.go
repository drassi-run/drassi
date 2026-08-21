/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/util/fs"
	"drassi.run/core/util/path"
)

type engine struct {
	Config
}

func New(config *Config) (sandboxer.Engine, error) {
	if config.Bin == "" {
		config.Bin = "firecracker"
	}
	if config.AgentPort == 0 {
		config.AgentPort = defaultAgentPort
	}
	if config.VcpuCount == 0 {
		config.VcpuCount = 2
	}
	if config.MemSizeMiB == 0 {
		config.MemSizeMiB = 2048
	}
	if config.KernelArgs == "" {
		config.KernelArgs = DefaultConfig().KernelArgs
	}
	if config.AgentWait == 0 {
		config.AgentWait = 30
	}

	if d, err := xpath.ResolveDir(config.RootDir); err != nil {
		return nil, err
	} else {
		config.RootDir = d
	}
	if err := os.MkdirAll(config.RootDir, xfs.DirPerm); err != nil {
		return nil, err
	}
	if err := config.ensureKernel(); err != nil {
		return nil, err
	}
	if err := config.ensureRootfs(); err != nil {
		return nil, err
	}
	return &engine{Config: *config}, nil
}

func (e *engine) Close() error {
	return nil
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	if e.kernel == "" {
		return nil, fmt.Errorf("firecracker kernel is required")
	}
	if e.rootfs == "" {
		return nil, fmt.Errorf("firecracker rootfs is required")
	}
	if _, err := exec.LookPath(e.Bin); err != nil {
		return nil, fmt.Errorf("firecracker binary %q: %w", e.Bin, err)
	}

	dir := filepath.Join(e.RootDir, e.sandboxDir(req))
	if err := os.MkdirAll(dir, xfs.DirPerm); err != nil {
		return nil, err
	}

	machine, err := startVM(&e.Config, dir)
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}

	cl := &client{
		dial: func(ctx context.Context) (net.Conn, error) {
			return DialHybridVsock(ctx, machine.vsockPath(), e.AgentPort)
		},
	}
	if err = e.waitAgent(ctx, cl, machine); err != nil {
		_ = machine.stop(context.Background())
		return nil, err
	}

	sb, err := newSandbox(ctx, cl, machine)
	if err != nil {
		_ = machine.stop(context.Background())
		return nil, err
	}

	dockerEngine, err := sb.dockerEngine()
	if err != nil {
		_ = sb.Terminate(ctx)
		return nil, err
	}
	s := sandboxer.AddBeforeCleanup(sb, func(context.Context) error {
		return dockerEngine.Close()
	})
	b := container.NewBootstrapper(dockerEngine)
	resp, err := b.Bootstrap(ctx, s, req)
	if err != nil {
		_ = s.Terminate(ctx)
		return nil, err
	}
	return resp, nil
}

func (e *engine) waitAgent(ctx context.Context, cl *client, machine *vm) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(e.AgentWait)*time.Second)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var last error
	for {
		select {
		case err := <-machine.waitErr:
			if err == nil {
				err = fmt.Errorf("firecracker exited")
			}
			return fmt.Errorf("waiting for agent: %w", err)
		case <-ctx.Done():
			if last != nil {
				return fmt.Errorf("waiting for agent: %w", last)
			}
			return fmt.Errorf("waiting for agent: %w", ctx.Err())
		case <-ticker.C:
			if _, err := cl.Info(ctx); err == nil {
				return nil
			} else {
				last = err
			}
		}
	}
}

func (e *engine) sandboxDir(req *sandboxer.LaunchRequest) string {
	if req.Uid != "" {
		return req.Uid
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
