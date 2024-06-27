package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/secret"
)

type handlerInfo[E any] struct {
	ctx  context.Context
	exec E
}

type consoleCommandHandlers struct {
	cmdMgr command.ConsoleCommandManager

	job   handlerInfo[JobExecutor]
	steps []handlerInfo[StepExecutor]
}

func (h *consoleCommandHandlers) StartJob(ctx context.Context, jobExec JobExecutor) error {
	if h.job.exec != nil {
		return fmt.Errorf("job %s still running, it need to be end first", h.job.exec.JobId())
	}

	h.job.ctx = ctx
	h.job.exec = jobExec

	_ = h.cmdMgr.RegisterCommand("add-mask", false, h.addSecretMask)
	_ = h.cmdMgr.RegisterCommand("add-matcher", true, h.addProblemMatcher)
	_ = h.cmdMgr.RegisterCommand("remove-matcher", true, h.removeProblemMatcher)

	_ = h.cmdMgr.RegisterCommand("group", true, h.groupingLog)
	_ = h.cmdMgr.RegisterCommand("endgroup", true, h.endGroupingLog)
	_ = h.cmdMgr.RegisterCommand("debug", false, h.logMessage)
	_ = h.cmdMgr.RegisterCommand("notice", false, h.logMessage)
	_ = h.cmdMgr.RegisterCommand("warning", false, h.logMessage)
	_ = h.cmdMgr.RegisterCommand("error", false, h.logMessage)

	_ = h.cmdMgr.RegisterCommand("add-path", true, h.addPath)
	_ = h.cmdMgr.RegisterCommand("set-env", true, h.setEnv)

	h.clearSteps()
	return nil
}

func (h *consoleCommandHandlers) EndJob() {
	h.clearSteps()

	// un-register commands
	_ = h.cmdMgr.RegisterCommand("add-mask", false, nil)
	_ = h.cmdMgr.RegisterCommand("add-matcher", true, nil)
	_ = h.cmdMgr.RegisterCommand("remove-matcher", true, nil)

	_ = h.cmdMgr.RegisterCommand("group", true, nil)
	_ = h.cmdMgr.RegisterCommand("endgroup", true, nil)
	_ = h.cmdMgr.RegisterCommand("debug", false, nil)
	_ = h.cmdMgr.RegisterCommand("notice", false, nil)
	_ = h.cmdMgr.RegisterCommand("warning", false, nil)
	_ = h.cmdMgr.RegisterCommand("error", false, nil)

	_ = h.cmdMgr.RegisterCommand("add-path", true, nil)
	_ = h.cmdMgr.RegisterCommand("set-env", true, nil)

	h.job.ctx = nil
	h.job.exec = nil
}

func (h *consoleCommandHandlers) clearSteps() {
	// un-register commands
	_ = h.cmdMgr.RegisterCommand("set-output", true, nil)
	_ = h.cmdMgr.RegisterCommand("save-state", true, nil)

	h.steps = nil
}

func (h *consoleCommandHandlers) StartStep(ctx context.Context, stepExec StepExecutor) error {
	if h.job.exec == nil {
		return errors.New("job need to be started before starting a step")
	}

	h.steps = append(h.steps, handlerInfo[StepExecutor]{
		ctx:  ctx,
		exec: stepExec,
	})
	if len(h.steps) == 1 {
		_ = h.cmdMgr.RegisterCommand("set-output", true, h.setOutput)
		_ = h.cmdMgr.RegisterCommand("save-state", true, h.saveState)
	}
	return nil
}

func (h *consoleCommandHandlers) EndStep() {
	if len(h.steps) == 1 {
		// un-register commands
		_ = h.cmdMgr.RegisterCommand("set-output", true, nil)
		_ = h.cmdMgr.RegisterCommand("save-state", true, nil)
	}
	if len(h.steps) > 0 {
		h.steps = h.steps[:len(h.steps)-1]
	}
}

func (h *consoleCommandHandlers) stepHandle(fn func(ctx context.Context, exec StepExecutor) error) error {
	if len(h.steps) == 0 {
		return errors.New("no step found")
	}
	currentStep := h.steps[len(h.steps)-1]
	exec := currentStep.exec
	ctx := currentStep.ctx

	return fn(ctx, exec)
}

func (h *consoleCommandHandlers) jobHandle(fn func(ctx context.Context, exec JobExecutor) error) error {
	exec := h.job.exec
	if exec == nil {
		return errors.New("no job found")
	}
	ctx := h.job.ctx
	if len(h.steps) > 0 {
		currentStep := h.steps[len(h.steps)-1]
		ctx = currentStep.ctx
	}

	return fn(ctx, exec)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L384
func (h *consoleCommandHandlers) addSecretMask(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, e JobExecutor) error {
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
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L451
func (h *consoleCommandHandlers) addProblemMatcher(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, e JobExecutor) error {
		file := cmd.Value
		if file == "" {
			return errors.New("file path must be specified in ##[add-matcher] command")
		}

		conf, err := h.readProblemMatcherFile(ctx, e, file)
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
				e.AddProblemMatcher(config.Owner, matcher)
			}
		}
		return nil
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L498
func (h *consoleCommandHandlers) removeProblemMatcher(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, e JobExecutor) error {
		file := cmd.Value
		owner := cmd.Params["owner"]
		if (file == "") == (owner == "") {
			return errors.New("either an owner name or a file path must be specified in ##[remove-matcher] command")
		}

		var owners []string
		if owner != "" {
			owners = []string{owner}
		} else {
			conf, err := h.readProblemMatcherFile(ctx, e, file)
			if err != nil {
				return err
			}
			owners = make([]string, len(conf.Configs))
			for i, config := range conf.Configs {
				owners[i] = config.Owner
			}
		}

		for _, o := range owners {
			e.RemoveProblemMatcher(o)
		}

		return nil
	})
}

func (h *consoleCommandHandlers) readProblemMatcherFile(ctx context.Context, e JobExecutor, file string) (*problem.MatcherConfigs, error) {
	if filepath.IsLocal(file) {
		ws := e.Sandbox().GetWorkspaceDir()
		file = filepath.Join(ws, file)
	}
	reader, err := e.Sandbox().CopyOut(ctx, file)
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
func (h *consoleCommandHandlers) groupingLog(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, e JobExecutor) error {
		e.Log(TagGroup, cmd.Value)
		return nil
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L751
func (h *consoleCommandHandlers) endGroupingLog(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, e JobExecutor) error {
		e.Log(TagEndGroup, "")
		return nil
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L566
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L600
func (h *consoleCommandHandlers) logMessage(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, e JobExecutor) error {
		// TODO
		return nil
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L417
func (h *consoleCommandHandlers) addPath(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, e JobExecutor) error {
		paths := []string{cmd.Value}
		return e.AddPath(paths)
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L234
func (h *consoleCommandHandlers) setEnv(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, e JobExecutor) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return errors.New("required field 'name' is missing in ##[set-output] command")
		}

		env := map[string]string{
			name: cmd.Value,
		}
		return e.SetEnv(env)
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L301
func (h *consoleCommandHandlers) setOutput(cmd *command.Command) error {
	return h.stepHandle(func(ctx context.Context, e StepExecutor) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return fmt.Errorf("required field 'name' in ##[set-output] command")
		}

		output := map[string]string{
			name: cmd.Value,
		}
		return e.SetOutput(output)
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L336
func (h *consoleCommandHandlers) saveState(cmd *command.Command) error {
	return h.stepHandle(func(ctx context.Context, e StepExecutor) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return fmt.Errorf("required field 'name' in ##[save-state] command")
		}

		state := map[string]string{
			name: cmd.Value,
		}
		return e.SaveState(state)
	})
}
