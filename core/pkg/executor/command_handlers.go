package executor

import (
	"archive/tar"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/model/dossiers"
	utilreader "drassi.run/core/pkg/util/reader"
)

type handlerInfo[H any] struct {
	ctx     context.Context
	handler H
}

type commandHandlers struct {
	consoleMgr command.ConsoleCommandManager
	fileMgr    command.FileCommandManager

	job   handlerInfo[JobCommandHandler]
	steps []handlerInfo[StepCommandHandler]
}

func (h *commandHandlers) Register() {
	_ = h.consoleMgr.RegisterCommand("add-mask", false, h.addSecretMask)
	_ = h.consoleMgr.RegisterCommand("add-matcher", true, h.addProblemMatcher)
	_ = h.consoleMgr.RegisterCommand("remove-matcher", true, h.removeProblemMatcher)

	_ = h.consoleMgr.RegisterCommand("group", true, h.groupingLog)
	_ = h.consoleMgr.RegisterCommand("endgroup", true, h.endGroupingLog)
	_ = h.consoleMgr.RegisterCommand("debug", false, h.logMessage)
	_ = h.consoleMgr.RegisterCommand("notice", false, h.logMessage)
	_ = h.consoleMgr.RegisterCommand("warning", false, h.logMessage)
	_ = h.consoleMgr.RegisterCommand("error", false, h.logMessage)

	_ = h.consoleMgr.RegisterCommand("add-path", true, h.consoleAddPath)
	_ = h.consoleMgr.RegisterCommand("set-env", true, h.consoleSetEnv)
	_ = h.consoleMgr.RegisterCommand("set-output", true, h.consoleSetOutput)
	_ = h.consoleMgr.RegisterCommand("save-state", true, h.consoleSaveState)

	_ = h.fileMgr.RegisterCommand("GITHUB_PATH", h.fileAddPath)
	_ = h.fileMgr.RegisterCommand("GITHUB_ENV", h.fileSetEnv)
	_ = h.fileMgr.RegisterCommand("GITHUB_STATE", h.fileSaveState)
	_ = h.fileMgr.RegisterCommand("GITHUB_OUTPUT", h.fileSetOutput)
	_ = h.fileMgr.RegisterCommand("GITHUB_STEP_SUMMARY", h.createStepSummary)
}

func (h *commandHandlers) StartStep(ctx context.Context, stepHandler StepCommandHandler) (map[string]string, error) {
	env, err := h.fileMgr.Initialize(ctx, stepHandler.Sandbox())
	if err != nil {
		return nil, err
	}
	h.steps = append(h.steps, handlerInfo[StepCommandHandler]{
		ctx:     ctx,
		handler: stepHandler,
	})
	return env, nil
}

func (h *commandHandlers) EndStep(outcome dossiers.Result) (err error) {
	if outcome == dossiers.ResultSuccess {
		err = h.stepHandle(func(ctx context.Context, handler StepCommandHandler) error {
			if ctx.Err() != nil { // ctx is DONE
				return nil
			}
			return h.fileMgr.Process(ctx, handler.Sandbox())
		})
	}
	h.steps = h.steps[:len(h.steps)-1]
	return
}

func (h *commandHandlers) lineHandler(w io.Writer, handlers ...reporter.LineHandler) reporter.LineHandler {
	return func(line string) error {
		jh := h.job.handler
		if cmd := h.consoleMgr.ParseCommand(line); cmd != nil {
			if err := h.consoleMgr.Process(line, cmd); err != nil {
				jh.Log(TagError, err.Error())

				// set step outcome = failure
				_ = h.stepHandle(func(ctx context.Context, handler StepCommandHandler) error {
					if exec, ok := handler.(StepExecutor); ok {
						exec.SetOutcome(dossiers.ResultFailure)
					}
					return nil
				})
			}
			return nil
		}

		// run additional handlers, any errors will be ignored
		for _, hdl := range handlers {
			if err := hdl(line); err != nil {
				jh.Log(TagError, err.Error())
			}
		}

		_, err := io.WriteString(w, line)
		return err
	}
}

func (h *commandHandlers) stepHandle(fn func(context.Context, StepCommandHandler) error) error {
	if len(h.steps) == 0 {
		return errors.New("no step found")
	}
	currentStep := h.steps[len(h.steps)-1]
	handler := currentStep.handler
	ctx := currentStep.ctx

	return fn(ctx, handler)
}

func (h *commandHandlers) jobHandle(fn func(context.Context, JobCommandHandler) error) error {
	handler := h.job.handler
	ctx := h.job.ctx
	if len(h.steps) > 0 {
		currentStep := h.steps[len(h.steps)-1]
		ctx = currentStep.ctx
	}
	return fn(ctx, handler)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L384
func (h *commandHandlers) addSecretMask(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		if cmd.Value == "" {
			return errors.New("can't add secret mask for empty string in ##[add-mask] command")
		}

		s := secret.NewValueSecret(cmd.Value)
		jh.AddSecretMask(s)
		for _, mask := range strings.FieldsFunc(cmd.Value, func(c rune) bool { return c == '\n' || c == '\r' }) {
			if mask != cmd.Value {
				s = secret.NewValueSecret(mask)
				jh.AddSecretMask(s)
			}
		}

		return nil
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L451
func (h *commandHandlers) addProblemMatcher(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		file := cmd.Value
		if file == "" {
			return errors.New("file path must be specified in ##[add-matcher] command")
		}

		conf, err := h.readProblemMatcherFile(ctx, jh, file)
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
				jh.AddProblemMatcher(config.Owner, matcher)
			}
		}
		return nil
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L498
func (h *commandHandlers) removeProblemMatcher(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		file := cmd.Value
		owner := cmd.Params["owner"]
		if (file == "") == (owner == "") {
			return errors.New("either an owner name or a file path must be specified in ##[remove-matcher] command")
		}

		var owners []string
		if owner != "" {
			owners = []string{owner}
		} else {
			conf, err := h.readProblemMatcherFile(ctx, jh, file)
			if err != nil {
				return err
			}
			owners = make([]string, len(conf.Configs))
			for i, config := range conf.Configs {
				owners[i] = config.Owner
			}
		}

		for _, o := range owners {
			jh.RemoveProblemMatcher(o)
		}

		return nil
	})
}

func (h *commandHandlers) readProblemMatcherFile(ctx context.Context, jh JobCommandHandler, file string) (*problem.MatcherConfigs, error) {
	if filepath.IsLocal(file) {
		ws := jh.Sandbox().GetWorkspaceDir()
		file = filepath.Join(ws, file)
	}
	reader, err := jh.Sandbox().CopyOut(ctx, file)
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
func (h *commandHandlers) groupingLog(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		jh.Log(TagGroup, cmd.Value)
		return nil
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L751
func (h *commandHandlers) endGroupingLog(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		jh.Log(TagEndGroup, "")
		return nil
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L566
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L600
func (h *commandHandlers) logMessage(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		// TODO
		return nil
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L417
func (h *commandHandlers) consoleAddPath(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		paths := []string{cmd.Value}
		return jh.AddPath(paths)
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L234
func (h *commandHandlers) consoleSetEnv(cmd *command.Command) error {
	return h.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return errors.New("required field 'name' is missing in ##[set-output] command")
		}

		env := map[string]string{
			name: cmd.Value,
		}
		return jh.SetEnv(env)
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L301
func (h *commandHandlers) consoleSetOutput(cmd *command.Command) error {
	return h.stepHandle(func(ctx context.Context, sh StepCommandHandler) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return fmt.Errorf("required field 'name' in ##[set-output] command")
		}

		output := map[string]string{
			name: cmd.Value,
		}
		return sh.SetOutput(output)
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L336
func (h *commandHandlers) consoleSaveState(cmd *command.Command) error {
	return h.stepHandle(func(ctx context.Context, sh StepCommandHandler) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return fmt.Errorf("required field 'name' in ##[save-state] command")
		}

		state := map[string]string{
			name: cmd.Value,
		}
		return sh.SaveState(state)
	})
}

func (h *commandHandlers) fileAddPath(r io.Reader) error {
	return h.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		return utilreader.Untar(r, func(hdr *tar.Header, reader io.Reader) error {
			if hdr.Name != "" {
				return fmt.Errorf("expected read single file with empty name, got %s", hdr.Name)
			}

			if paths, err := utilreader.ReadLine(reader); err != nil {
				return err
			} else {
				return jh.AddPath(paths)
			}
		})
	})
}

func (h *commandHandlers) fileSetEnv(r io.Reader) error {
	return h.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		return utilreader.Untar(r, func(hdr *tar.Header, reader io.Reader) error {
			if hdr.Name != "" {
				return fmt.Errorf("expected read single file with empty name, got %s", hdr.Name)
			}

			if env, err := utilreader.ParseEnvVars(reader); err != nil {
				return err
			} else {
				return jh.SetEnv(env)
			}
		})
	})
}

func (h *commandHandlers) fileSetOutput(r io.Reader) error {
	return h.stepHandle(func(ctx context.Context, jh StepCommandHandler) error {
		return utilreader.Untar(r, func(hdr *tar.Header, reader io.Reader) error {
			if hdr.Name != "" {
				return fmt.Errorf("expected read single file with empty name, got %s", hdr.Name)
			}

			if output, err := utilreader.ParseEnvVars(reader); err != nil {
				return err
			} else {
				return jh.SetOutput(output)
			}
		})
	})
}

func (h *commandHandlers) fileSaveState(r io.Reader) error {
	return h.stepHandle(func(ctx context.Context, jh StepCommandHandler) error {
		return utilreader.Untar(r, func(hdr *tar.Header, reader io.Reader) error {
			if hdr.Name != "" {
				return fmt.Errorf("expected read single file with empty name, got %s", hdr.Name)
			}

			if state, err := utilreader.ParseEnvVars(reader); err != nil {
				return err
			} else {
				return jh.SaveState(state)
			}
		})
	})
}

func (h *commandHandlers) createStepSummary(r io.Reader) error {
	// TODO: implement GITHUB_STEP_SUMMARY file command
	return nil
}
