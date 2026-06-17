/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_command

import (
	"context"

	cmd "drassi.run/core/pkg/command"
	exec "drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/feature"
	"drassi.run/core/wire"
	"go.uber.org/dig"
)

type commandInitHook[R any] struct {
	scope *dig.Scope
}

func NewCommandInitHook[R any](scope *dig.Scope) exec.Hook[R] {
	return &commandInitHook[R]{scope}
}

func (c *commandInitHook[R]) Hook(ctx context.Context, _ R) error {
	if err := c.scope.Invoke(c.registerConsoleCommands); err != nil {
		return err
	}
	if err := c.scope.Invoke(c.registerFileCommands); err != nil {
		return err
	}

	return c.scope.Invoke(c.setConsoleManagerDebug(ctx))
}

type consoleCommandParams struct {
	dig.In

	ConsMgr  cmd.ConsoleManager[exec.Milieu]
	Handlers []*cmd.ConsoleHandler[exec.Milieu] `group:"console-handlers"`
}

func (c *commandInitHook[R]) registerConsoleCommands(p consoleCommandParams) error {
	for _, h := range p.Handlers {
		if err := p.ConsMgr.Register(h); err != nil {
			return err
		}
	}
	return nil
}

type fileCommandParams struct {
	dig.In

	FileMgr  cmd.FileManager[exec.Milieu]
	Handlers []*cmd.FileHandler[exec.Milieu] `group:"file-handlers"`
}

func (c *commandInitHook[R]) registerFileCommands(p fileCommandParams) error {
	for _, h := range p.Handlers {
		if err := p.FileMgr.Register(h); err != nil {
			return err
		}
	}
	return nil
}

func (c *commandInitHook[R]) setConsoleManagerDebug(ctx context.Context) func(feature.Flags, cmd.ConsoleManager[exec.Milieu]) error {
	return func(flags feature.Flags, consMgr cmd.ConsoleManager[exec.Milieu]) error {
		if !feature.Bool(flags, wire.StepDebug, false) {
			return nil
		}

		com := &cmd.Command{Name: "echo", Value: "ON"}
		return consMgr.Process(ctx, nil, "", com)
	}
}
