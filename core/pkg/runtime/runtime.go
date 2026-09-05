/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package runtime

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"drassi.run/core/config"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/stream"
)

type Runtime interface {
	Run(ctx context.Context, scriptPath string, paths []string, env map[string]string, workdir string, streams *stream.Streams) error
}

type runtime struct {
	sandbox sandboxer.Sandbox
	cmd     []string
}

func NewRuntime(name string, sandbox sandboxer.Sandbox, cfg *config.Runtime) (Runtime, error) {
	layout := sandbox.Layout()
	binPath := filepath.Join(layout.Runtimes, name, cfg.Executable)

	cmd := make([]string, 0, len(cfg.Cmd)+1)
	cmd = append(cmd, binPath)
	if len(cfg.Cmd) > 0 {
		var hasPlaceholder bool
		for _, arg := range cfg.Cmd {
			if strings.Contains(arg, "{0}") {
				hasPlaceholder = true
				break
			}
		}
		if !hasPlaceholder {
			return nil, fmt.Errorf("runtime %q cmd must contain \"{0}\" placeholder", name)
		}
		cmd = append(cmd, cfg.Cmd...)
	} else {
		cmd = append(cmd, "{0}")
	}

	rt := &runtime{
		sandbox: sandbox,
		cmd:     cmd,
	}
	return rt, nil
}

func (r *runtime) Run(
	ctx context.Context,
	scriptPath string,
	paths []string,
	env map[string]string,
	workdir string,
	streams *stream.Streams,
) error {
	cmd := make([]string, len(r.cmd))
	for i, arg := range r.cmd {
		cmd[i] = strings.ReplaceAll(arg, "{0}", scriptPath)
	}

	return r.sandbox.Execute(ctx, cmd, paths, env, workdir, streams)
}
