/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_cmdhandler

import (
	"context"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/util/dig"
	"go.uber.org/dig"
)

func NewListener() executor.Listener {
	return new(cmdListener)
}

type cmdListener struct {
	executor.NoopJobListener
	executor.NoopStepListener

	cmdMgr command.FileManager
}

func (l *cmdListener) OnInitializeJob(_ executor.JobExecutor, scope *dig.Scope) executor.EventHandler {
	return &jobInitEventHandler{
		scope:    scope,
		listener: l,
	}
}

func (l *cmdListener) OnRunTask(_ executor.StepExecutor, task *executor.ActionRun) executor.EventHandler {
	return &taskRunEventHandler{
		task:   task,
		cmdMgr: l.cmdMgr,
	}
}

type jobInitEventHandler struct {
	scope    *dig.Scope
	ctx      context.Context
	listener *cmdListener
}

func (eh *jobInitEventHandler) Begin(ctx context.Context) error {
	eh.ctx = ctx
	return nil
}

func (eh *jobInitEventHandler) End(err error) error {
	if err != nil {
		return nil // skip
	}

	// NOTE: some handlers are depended on sandbox
	if err = eh.scope.Invoke(eh.registerConsoleCommands); err != nil {
		return err
	}

	if err = eh.scope.Invoke(eh.registerFileCommands); err != nil {
		return err
	}

	if err = eh.scope.Invoke(eh.provideEnv); err != nil {
		return err
	}

	var runner records.Runner
	if err = xdig.Populate(eh.scope, &runner); err != nil {
		return err
	}

	if runner.Debug == "1" {
		if err = eh.scope.Invoke(eh.setDiaryDebug); err != nil {
			return err
		}

		if err = eh.scope.Invoke(eh.setConsoleManagerDebug); err != nil {
			return err
		}
	}

	return xdig.Populate(eh.scope, &eh.listener.cmdMgr)
}

type consoleCommandParams struct {
	dig.In

	CmdMgr   command.ConsoleManager
	Handlers []*command.ConsoleHandler `group:"console-handlers"`
}

func (eh *jobInitEventHandler) registerConsoleCommands(p consoleCommandParams) error {
	for _, h := range p.Handlers {
		if err := p.CmdMgr.Register(h); err != nil {
			return err
		}
	}
	return nil
}

type fileCommandParams struct {
	dig.In

	CmdMgr   command.FileManager
	Handlers []*command.FileHandler `group:"file-handlers"`
}

func (eh *jobInitEventHandler) registerFileCommands(p fileCommandParams) error {
	for _, h := range p.Handlers {
		if err := p.CmdMgr.Register(h); err != nil {
			return err
		}
	}
	return nil
}

func (eh *jobInitEventHandler) provideEnv(env support.EnvProvider, stack executor.Stack, cmdMgr command.FileManager) {
	ep := func() map[string]string {
		exec := stack.Leaf()
		if exec == nil {
			return nil
		}

		suffix := exec.StepSpec().Uid
		return cmdMgr.Env(suffix)
	}
	env.ProvideEnv(ep)
}

func (eh *jobInitEventHandler) setDiaryDebug(diary scribe.Diary) {
	diary.SetDebug(true)
}

func (eh *jobInitEventHandler) setConsoleManagerDebug(cmdMgr command.ConsoleManager) error {
	cmd := &command.Command{Name: "echo", Value: "ON"}
	return cmdMgr.Process(eh.ctx, "", cmd)
}

type taskRunEventHandler struct {
	task   *executor.ActionRun
	ctx    context.Context
	cmdMgr command.FileManager
}

func (h *taskRunEventHandler) Begin(ctx context.Context) error {
	h.ctx = ctx
	return h.cmdMgr.Initialize(ctx, h.task.StepId)
}

func (h *taskRunEventHandler) End(err error) error {
	if err != nil {
		return nil // skip
	}

	return h.cmdMgr.Process(h.ctx, h.task.StepId)
}
