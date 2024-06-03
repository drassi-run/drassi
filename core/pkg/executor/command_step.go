package executor

import (
	"fmt"

	"github.com/dungdm93/drassi/core/pkg/executor/command"
)

func (e *StepExecutor) setOutput(cmd *command.Command) error {
	name, ok := cmd.Params["name"]
	if !ok || name == "" {
		return fmt.Errorf("required field 'name' in ##[set-output] command")
	}

	output := map[string]string{
		name: cmd.Value,
	}
	return e.SetOutput(output)
}

func (e *StepExecutor) saveState(cmd *command.Command) error {
	name, ok := cmd.Params["name"]
	if !ok || name == "" {
		return fmt.Errorf("required field 'name' in ##[save-state] command")
	}

	state := map[string]string{
		name: cmd.Value,
	}
	return e.SaveState(state)
}
