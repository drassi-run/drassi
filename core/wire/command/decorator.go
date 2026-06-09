package wire_command

import (
	"context"

	exec "drassi.run/core/pkg/executor"
	cmd "drassi.run/core/pkg/executor/command"
)

type commandDecorator struct {
	fileMgr cmd.FileManager[exec.Milieu]
}

func NewCommandDecorator(fileMgr cmd.FileManager[exec.Milieu]) exec.ActionRunDecorator {
	return &commandDecorator{fileMgr}
}

func (c *commandDecorator) DecorateActionRun(task *exec.ActionTask) exec.ActionRun {
	run := task.Run
	res := exec.NewMilieu(task.Executor.StepExecutor())
	return func(ctx context.Context) error {
		if err := c.fileMgr.Initialize(ctx, res); err != nil {
			return err
		}
		if err := run(ctx); err != nil {
			return err
		}
		return c.fileMgr.Process(ctx, res)
	}
}
