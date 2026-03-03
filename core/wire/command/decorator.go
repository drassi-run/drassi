package wire_command

import (
	"context"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/scribe"
	xdig "drassi.run/core/util/dig"
	"go.uber.org/dig"
)

type CommandDecorator struct {
	cmdMgr command.FileManager[executor.SupportCommands]
}

func NewCommandDecorator(cmdMgr command.FileManager[executor.SupportCommands]) *CommandDecorator {
	return &CommandDecorator{cmdMgr}
}

func (c *CommandDecorator) DecorateActionRun(task *executor.ActionTask) executor.ActionRun {
	run := task.Run
	sup := executor.NewSupportCommands(task.Executor.StepExecutor())
	return func(ctx context.Context) error {
		if err := c.cmdMgr.Initialize(ctx, sup); err != nil {
			return err
		}
		if err := run(ctx); err != nil {
			return err
		}
		return c.cmdMgr.Process(ctx, sup)
	}
}

func (c *CommandDecorator) DecorateJobRun(task *executor.JobTask) executor.JobRun {
	if task.Stage != executor.StagePre {
		return task.Run
	}

	// decorator for Initialize job
	run := task.Run
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
		//if err = scope.Invoke(c.provideEnv); err != nil {
		//	return
		//}

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

	ConsMgr  command.ConsoleManager[executor.SupportCommands]
	Handlers []*command.ConsoleHandler[executor.SupportCommands] `group:"console-handlers"`
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

	FileMgr  command.FileManager[executor.SupportCommands]
	Handlers []*command.FileHandler[executor.SupportCommands] `group:"file-handlers"`
}

func (c *CommandDecorator) registerFileCommands(p fileCommandParams) error {
	for _, h := range p.Handlers {
		if err := p.FileMgr.Register(h); err != nil {
			return err
		}
	}
	return nil
}

//func (c *CommandDecorator) provideEnv(env support.EnvProvider, stack *support.Stack, cmdMgr command.FileManager) {
//	ep := func() map[string]string {
//		_, exec := stack.CurrentStep()
//		if exec == nil {
//			return nil
//		}
//
//		suffix := exec.StepSpec().Uid
//		return cmdMgr.Env(suffix)
//	}
//	env.ProvideEnv(ep)
//}

func (c *CommandDecorator) setDiaryDebug(diary scribe.Diary) {
	diary.SetDebug(true)
}

func (c *CommandDecorator) setConsoleManagerDebug(ctx context.Context) func(command.ConsoleManager[executor.SupportCommands]) error {
	return func(consMgr command.ConsoleManager[executor.SupportCommands]) error {
		cmd := &command.Command{Name: "echo", Value: "ON"}
		return consMgr.Process(ctx, nil, "", cmd)
	}
}
