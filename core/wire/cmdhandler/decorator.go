package wire_cmdhandler

import (
	"context"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/scribe"
	xdig "drassi.run/core/util/dig"
	"go.uber.org/dig"
)

type CommandDecorator struct {
	cmdMgr command.FileManager
}

func NewCommandDecorator(cmdMgr command.FileManager) *CommandDecorator {
	return &CommandDecorator{cmdMgr}
}

func (c *CommandDecorator) DecorateActionRun(action *executor.ActionRun) executor.ActionTask {
	stepId := action.StepId()
	run := action.Run
	return func(ctx context.Context) error {
		if err := c.cmdMgr.Initialize(ctx, stepId); err != nil {
			return err
		}
		if err := run(ctx); err != nil {
			return err
		}
		return c.cmdMgr.Process(ctx, stepId)
	}
}

func (c *CommandDecorator) DecorateJobRun(job *executor.JobRun) executor.JobTask {
	if job.Stage != executor.StagePre {
		return job.Run
	}

	// decorator for Initialize job
	run := job.Run
	var scope *dig.Scope //TODO
	return func(ctx context.Context) (res *records.Job, err error) {
		if res, err = run(ctx); err != nil {
			return
		}

		if err = scope.Invoke(c.registerConsoleCommands); err != nil {
			return
		}
		if err = scope.Invoke(c.registerFileCommands); err != nil {
			return
		}
		if err = scope.Invoke(c.provideEnv); err != nil {
			return
		}

		var runner records.Runner
		if err = xdig.Populate(scope, &runner); err != nil {
			return
		}
		if runner.Debug == "1" {
			if err = scope.Invoke(c.setDiaryDebug); err != nil {
				return
			}
			if err = scope.Invoke(c.setConsoleManagerDebug(ctx)); err != nil {
				return
			}
		}

		err = xdig.Populate(scope, &c.cmdMgr)
		return
	}
}

type consoleCommandParams struct {
	dig.In

	ConsMgr  command.ConsoleManager
	Handlers []*command.ConsoleHandler `group:"console-handlers"`
}

func (c *CommandDecorator) registerConsoleCommands(p consoleCommandParams) error {
	for _, h := range p.Handlers {
		if err := p.ConsMgr.Register(h); err != nil {
			return err
		}
	}
	return nil
}

type fileCommandParams struct {
	dig.In

	FileMgr  command.FileManager
	Handlers []*command.FileHandler `group:"file-handlers"`
}

func (c *CommandDecorator) registerFileCommands(p fileCommandParams) error {
	for _, h := range p.Handlers {
		if err := p.FileMgr.Register(h); err != nil {
			return err
		}
	}
	return nil
}

func (c *CommandDecorator) provideEnv(env support.EnvProvider, stack executor.Stack, cmdMgr command.FileManager) {
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

func (c *CommandDecorator) setDiaryDebug(diary scribe.Diary) {
	diary.SetDebug(true)
}

func (c *CommandDecorator) setConsoleManagerDebug(ctx context.Context) func(command.ConsoleManager) error {
	return func(consMgr command.ConsoleManager) error {
		cmd := &command.Command{Name: "echo", Value: "ON"}
		return consMgr.Process(ctx, "", cmd)
	}
}
