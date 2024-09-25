package wire_cmdhandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/logging"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/sandboxer"
)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L384
func addSecretMask(secretMasker secret.Masker) *command.ConsoleHandler {
	run := func(cmd *command.Command) error {
		if cmd.Value == "" {
			return errors.New("can't add secret mask for empty string in ##[add-mask] command")
		}
		s := secret.NewValueSecret(cmd.Value)
		secretMasker.AddSecret(s)
		for mask := range splitLine(cmd.Value) {
			if mask != cmd.Value {
				s = secret.NewValueSecret(mask)
				secretMasker.AddSecret(s)
			}
		}
		return nil
	}
	return command.NewConsoleHandler("add-mask", false, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L451
func addProblemMatcher(m map[string]problem.Matcher, sb sandboxer.Sandbox, sup executor.Supervisor) *command.ConsoleHandler {
	run := func(cmd *command.Command) error {
		file := cmd.Value
		if file == "" {
			return errors.New("file path must be specified in ##[add-matcher] command")
		}

		ctx := sup.Context()
		conf, err := readProblemMatcherFile(ctx, sb, file)
		if err != nil {
			return err
		}
		if len(conf.Configs) == 0 {
			return nil
		}
		if err = conf.Validate(); err != nil {
			return err
		}

		for _, config := range conf.Configs {
			if matcher, err := problem.NewMatcher(config.Severity, config.Patterns); err != nil {
				return err
			} else {
				m[config.Owner] = matcher
			}
		}
		return nil
	}
	return command.NewConsoleHandler("add-matcher", true, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L498
func removeProblemMatcher(m map[string]problem.Matcher, sb sandboxer.Sandbox, sup executor.Supervisor) *command.ConsoleHandler {
	run := func(cmd *command.Command) error {
		file := cmd.Value
		owner := cmd.Params["owner"]
		if (file == "") == (owner == "") {
			return errors.New("either an owner name or a file path must be specified in ##[remove-matcher] command")
		}

		var owners []string
		if owner != "" {
			owners = []string{owner}
		} else {
			ctx := sup.Context()
			conf, err := readProblemMatcherFile(ctx, sb, file)
			if err != nil {
				return err
			}
			owners = make([]string, len(conf.Configs))
			for i, config := range conf.Configs {
				owners[i] = config.Owner
			}
		}

		for _, o := range owners {
			delete(m, o)
		}

		return nil
	}
	return command.NewConsoleHandler("remove-matcher", true, run)
}

func readProblemMatcherFile(ctx context.Context, sb sandboxer.Sandbox, file string) (*problem.MatcherConfigs, error) {
	if filepath.IsLocal(file) {
		ws := sb.GetWorkspaceDir()
		file = filepath.Join(ws, file)
	}
	reader, err := sb.CopyOut(ctx, file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	conf := new(problem.MatcherConfigs)
	if err = json.NewDecoder(reader).Decode(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L751
func groupingLog(l logging.Logger) *command.ConsoleHandler {
	run := func(cmd *command.Command) error {
		l.Log(logging.TagGroup, cmd.Value)
		return nil
	}
	return command.NewConsoleHandler("group", true, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L751
func endGroupingLog(l logging.Logger) *command.ConsoleHandler {
	run := func(cmd *command.Command) error {
		l.Log(logging.TagEndGroup, "")
		return nil
	}
	return command.NewConsoleHandler("endgroup", true, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L566
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L600
func logMessage(l logging.Logger) []*command.ConsoleHandler {
	run := func(cmd *command.Command) error {
		tag := fmt.Sprintf("##[%s]", cmd.Name)
		l.Log(tag, cmd.Value)
		return nil
	}
	return []*command.ConsoleHandler{
		command.NewConsoleHandler("debug", false, run),
		command.NewConsoleHandler("notice", false, run),
		command.NewConsoleHandler("warning", false, run),
		command.NewConsoleHandler("error", false, run),
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L417
func consoleAddPath(sup executor.Supervisor) *command.ConsoleHandler {
	run := func(cmd *command.Command) error {
		job := sup.Job()
		if job == nil {
			return errors.New("no job found")
		}
		paths := []string{cmd.Value}
		return job.AddPath(paths)
	}
	return command.NewConsoleHandler("add-path", true, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L234
func consoleSetEnv(sup executor.Supervisor) *command.ConsoleHandler {
	run := func(cmd *command.Command) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return errors.New("required field 'name' is missing in ##[set-output] command")
		}

		step := sup.CurrentStep()
		if step == nil {
			return errors.New("no step found")
		}

		env := map[string]string{
			name: cmd.Value,
		}
		return step.SetEnv(env)
	}
	return command.NewConsoleHandler("set-env", true, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L301
func consoleSetOutput(sup executor.Supervisor) *command.ConsoleHandler {
	run := func(cmd *command.Command) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return fmt.Errorf("required field 'name' in ##[set-output] command")
		}

		step := sup.CurrentStep()
		if step == nil {
			return errors.New("no step found")
		}

		output := map[string]string{
			name: cmd.Value,
		}
		return step.SetOutput(output)
	}
	return command.NewConsoleHandler("set-output", true, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L336
func consoleSaveState(sup executor.Supervisor) *command.ConsoleHandler {
	run := func(cmd *command.Command) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return fmt.Errorf("required field 'name' in ##[save-state] command")
		}

		step := sup.CurrentStep()
		if step == nil {
			return errors.New("no step found")
		}

		state := map[string]string{
			name: cmd.Value,
		}
		return step.SaveState(state)
	}
	return command.NewConsoleHandler("save-state", true, run)
}
