package wire_command

import (
	"context"

	exec "drassi.run/core/pkg/executor"
	cmd "drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/scribe"
	"go.uber.org/dig"
)

type CommandDecorator struct {
	fileMgr cmd.FileManager[exec.Milieu]
}

func NewCommandDecorator(fileMgr cmd.FileManager[exec.Milieu]) *CommandDecorator {
	return &CommandDecorator{fileMgr}
}

func (c *CommandDecorator) DecorateActionRun(task *exec.ActionTask) exec.ActionRun {
	run := task.Run
	sup := exec.NewMilieu(task.Executor.StepExecutor())
	return func(ctx context.Context) error {
		if err := c.fileMgr.Initialize(ctx, sup); err != nil {
			return err
		}
		if err := run(ctx); err != nil {
			return err
		}
		return c.fileMgr.Process(ctx, sup)
	}
}

type CommandInitDecorator struct {
	dig.In

	ConsMgr      cmd.ConsoleManager[exec.Milieu]
	ConsHandlers []*cmd.ConsoleHandler[exec.Milieu] `group:"console-handlers"`

	FileMgr      cmd.FileManager[exec.Milieu]
	FileHandlers []*cmd.FileHandler[exec.Milieu] `group:"file-handlers"`

	Runner records.Runner
	Diary  scribe.Diary
}

func NewCommandInitDecorator(p *CommandInitDecorator) *CommandInitDecorator {
	return p
}

func (c *CommandInitDecorator) DecorateJobRun(task *exec.JobTask) exec.JobRun {
	if task.Stage != exec.StagePre {
		return task.Run
	}

	// decorator for Initialize job
	run := task.Run
	return func(ctx context.Context) (res *records.Job, err error) {
		if res, err = run(ctx); err != nil {
			return
		}

		if err = c.registerConsoleCommands(); err != nil {
			return
		}
		if err = c.registerFileCommands(); err != nil {
			return
		}

		if c.Runner.Debug == "1" {
			c.setDiaryDebug()
			if err = c.setConsoleManagerDebug(ctx); err != nil {
				return
			}
		}

		return
	}
}

func (c *CommandInitDecorator) registerConsoleCommands() error {
	for _, h := range c.ConsHandlers {
		if err := c.ConsMgr.Register(h); err != nil {
			return err
		}
	}
	return nil
}

func (c *CommandInitDecorator) registerFileCommands() error {
	for _, h := range c.FileHandlers {
		if err := c.FileMgr.Register(h); err != nil {
			return err
		}
	}
	return nil
}

func (c *CommandInitDecorator) setDiaryDebug() {
	c.Diary.SetDebug(true)
}

func (c *CommandInitDecorator) setConsoleManagerDebug(ctx context.Context) error {
	com := &cmd.Command{Name: "echo", Value: "ON"}
	return c.ConsMgr.Process(ctx, nil, "", com)
}
