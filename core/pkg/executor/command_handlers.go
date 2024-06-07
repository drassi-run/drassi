package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/executor/command"
	"github.com/dungdm93/drassi/core/pkg/executor/problem"
	"github.com/dungdm93/drassi/core/pkg/executor/secret"
)

type consoleCommandHandlers struct {
	cmdMgr *command.ConsoleCommandManager
}

func (h *consoleCommandHandlers) RegisterForJobExecutor(ctx context.Context, executor *JobExecutor) {
	_ = h.cmdMgr.RegisterCommand("add-mask", false, h.addSecretMask(executor))
	_ = h.cmdMgr.RegisterCommand("add-matcher", true, h.addProblemMatcher(ctx, executor))
	_ = h.cmdMgr.RegisterCommand("remove-matcher", true, h.removeProblemMatcher(ctx, executor))

	_ = h.cmdMgr.RegisterCommand("group", true, h.groupingLog(executor))
	_ = h.cmdMgr.RegisterCommand("endgroup", true, h.endGroupingLog(executor))
	_ = h.cmdMgr.RegisterCommand("debug", false, h.logMessage(executor))
	_ = h.cmdMgr.RegisterCommand("notice", false, h.logMessage(executor))
	_ = h.cmdMgr.RegisterCommand("warning", false, h.logMessage(executor))
	_ = h.cmdMgr.RegisterCommand("error", false, h.logMessage(executor))

	_ = h.cmdMgr.RegisterCommand("add-path", true, h.addPath(executor))
	_ = h.cmdMgr.RegisterCommand("set-env", true, h.setEnv(executor))
}

func (h *consoleCommandHandlers) RegisterForStepExecutor(ctx context.Context, executor *StepExecutor) {
	_ = h.cmdMgr.RegisterCommand("set-output", true, h.setOutput(executor))
	_ = h.cmdMgr.RegisterCommand("save-state", true, h.saveState(executor))
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L384
func (h *consoleCommandHandlers) addSecretMask(e *JobExecutor) command.ConsoleCommandHandler {
	return func(cmd *command.Command) error {
		if cmd.Value == "" {
			return errors.New("can't add secret mask for empty string in ##[add-mask] command")
		}

		s := secret.NewValueSecret(cmd.Value)
		e.AddSecretMask(s)
		for _, mask := range strings.FieldsFunc(cmd.Value, func(c rune) bool { return c == '\n' || c == '\r' }) {
			if mask != cmd.Value {
				s = secret.NewValueSecret(mask)
				e.AddSecretMask(s)
			}
		}

		return nil
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L451
func (h *consoleCommandHandlers) addProblemMatcher(ctx context.Context, e *JobExecutor) command.ConsoleCommandHandler {
	return func(cmd *command.Command) error {
		file := cmd.Value
		if file == "" {
			return errors.New("file path must be specified in ##[add-matcher] command")
		}

		configs, err := h.readProblemMatcherFile(ctx, e, file)
		if err != nil {
			return err
		}

		for _, config := range configs.ProblemMatcher {
			e.AddProblemMatcher(config)
		}
		return nil
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L498
func (h *consoleCommandHandlers) removeProblemMatcher(ctx context.Context, e *JobExecutor) command.ConsoleCommandHandler {
	return func(cmd *command.Command) error {
		file := cmd.Value
		owner := cmd.Params["owner"]
		if (file == "") == (owner == "") {
			return errors.New("either an owner name or a file path must be specified in ##[remove-matcher] command")
		}

		var owners []string
		if owner != "" {
			owners = []string{owner}
		} else {
			configs, err := h.readProblemMatcherFile(ctx, e, file)
			if err != nil {
				return err
			}
			owners = make([]string, len(configs.ProblemMatcher))
			for i, config := range configs.ProblemMatcher {
				owners[i] = config.Owner
			}
		}

		for _, o := range owners {
			e.RemoveProblemMatcher(o)
		}

		return nil
	}
}

func (h *consoleCommandHandlers) readProblemMatcherFile(ctx context.Context, e *JobExecutor, file string) (*problem.Configs, error) {
	if filepath.IsLocal(file) {
		ws := e.Sandbox().GetWorkspaceDir()
		file = filepath.Join(ws, file)
	}
	reader, err := e.Sandbox().CopyOut(ctx, file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	conf := new(problem.Configs)
	if err = json.NewDecoder(reader).Decode(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L751
func (h *consoleCommandHandlers) groupingLog(e *JobExecutor) command.ConsoleCommandHandler {
	return func(cmd *command.Command) error {
		e.Log("", "##[group]"+cmd.Value)
		return nil
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L751
func (h *consoleCommandHandlers) endGroupingLog(e *JobExecutor) command.ConsoleCommandHandler {
	return func(cmd *command.Command) error {
		e.Log("", "##[endgroup]")
		return nil
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L566
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L600
func (h *consoleCommandHandlers) logMessage(e *JobExecutor) command.ConsoleCommandHandler {
	return func(cmd *command.Command) error {
		// TODO
		return nil
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L417
func (h *consoleCommandHandlers) addPath(e *JobExecutor) command.ConsoleCommandHandler {
	return func(cmd *command.Command) error {
		paths := []string{cmd.Value}
		return e.AddPath(paths)
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L234
func (h *consoleCommandHandlers) setEnv(e *JobExecutor) command.ConsoleCommandHandler {
	return func(cmd *command.Command) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return errors.New("required field 'name' is missing in ##[set-output] command")
		}

		env := map[string]string{
			name: cmd.Value,
		}
		return e.SetEnv(env)
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L301
func (h *consoleCommandHandlers) setOutput(e *StepExecutor) command.ConsoleCommandHandler {
	return func(cmd *command.Command) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return fmt.Errorf("required field 'name' in ##[set-output] command")
		}

		output := map[string]string{
			name: cmd.Value,
		}
		return e.SetOutput(output)
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L336
func (h *consoleCommandHandlers) saveState(e *StepExecutor) command.ConsoleCommandHandler {
	return func(cmd *command.Command) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return fmt.Errorf("required field 'name' in ##[save-state] command")
		}

		state := map[string]string{
			name: cmd.Value,
		}
		return e.SaveState(state)
	}
}
