package executor

import (
	"archive/tar"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/logger"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/model/dossiers"
	utilreader "drassi.run/core/pkg/util/reader"
)

type CommandController interface {
	Register()
	LineHandler(w io.Writer, handlers ...logger.LineHandler) logger.LineHandler
	StartStep(ctx context.Context, stepHandler StepCommandHandler) error
	EndStep(outcome dossiers.Result) (err error)
}

type commandController struct {
	consoleMgr command.ConsoleManager
	fileMgr    command.FileManager

	job   handlerInfo[JobCommandHandler]
	steps []handlerInfo[StepCommandHandler]
}

type handlerInfo[H any] struct {
	ctx     context.Context
	handler H
}

func (cc *commandController) Register() {
	_ = cc.consoleMgr.Register(command.NewConsoleHandler("add-mask", false, cc.addSecretMask))
	_ = cc.consoleMgr.Register(command.NewConsoleHandler("add-matcher", true, cc.addProblemMatcher))
	_ = cc.consoleMgr.Register(command.NewConsoleHandler("remove-matcher", true, cc.removeProblemMatcher))

	_ = cc.consoleMgr.Register(command.NewConsoleHandler("group", true, cc.groupingLog))
	_ = cc.consoleMgr.Register(command.NewConsoleHandler("endgroup", true, cc.endGroupingLog))
	_ = cc.consoleMgr.Register(command.NewConsoleHandler("debug", false, cc.logMessage))
	_ = cc.consoleMgr.Register(command.NewConsoleHandler("notice", false, cc.logMessage))
	_ = cc.consoleMgr.Register(command.NewConsoleHandler("warning", false, cc.logMessage))
	_ = cc.consoleMgr.Register(command.NewConsoleHandler("error", false, cc.logMessage))

	_ = cc.consoleMgr.Register(command.NewConsoleHandler("add-path", true, cc.consoleAddPath))
	_ = cc.consoleMgr.Register(command.NewConsoleHandler("set-env", true, cc.consoleSetEnv))
	_ = cc.consoleMgr.Register(command.NewConsoleHandler("set-output", true, cc.consoleSetOutput))
	_ = cc.consoleMgr.Register(command.NewConsoleHandler("save-state", true, cc.consoleSaveState))

	_ = cc.fileMgr.Register(command.NewFileHandler("GITHUB_PATH", cc.fileAddPath))
	_ = cc.fileMgr.Register(command.NewFileHandler("GITHUB_ENV", cc.fileSetEnv))
	_ = cc.fileMgr.Register(command.NewFileHandler("GITHUB_STATE", cc.fileSaveState))
	_ = cc.fileMgr.Register(command.NewFileHandler("GITHUB_OUTPUT", cc.fileSetOutput))
	_ = cc.fileMgr.Register(command.NewFileHandler("GITHUB_STEP_SUMMARY", cc.createStepSummary))
}

func (cc *commandController) StartStep(ctx context.Context, stepHandler StepCommandHandler) error {
	if env, err := cc.fileMgr.Initialize(ctx, stepHandler.Sandbox()); err != nil {
		return err
	} else {
		if err = stepHandler.SetEnv(env, false); err != nil {
			return err
		}
	}

	cc.steps = append(cc.steps, handlerInfo[StepCommandHandler]{
		ctx:     ctx,
		handler: stepHandler,
	})
	return nil
}

func (cc *commandController) EndStep(outcome dossiers.Result) (err error) {
	if outcome == dossiers.ResultSuccess {
		err = cc.stepHandle(func(ctx context.Context, handler StepCommandHandler) error {
			if ctx.Err() != nil { // ctx is DONE
				return nil
			}
			return cc.fileMgr.Process(ctx, handler.Sandbox())
		})
	}
	cc.steps = cc.steps[:len(cc.steps)-1]
	return
}

func (cc *commandController) LineHandler(w io.Writer, handlers ...logger.LineHandler) logger.LineHandler {
	return func(line string) error {
		jh := cc.job.handler
		if cmd := cc.consoleMgr.ParseCommand(line); cmd != nil {
			if err := cc.consoleMgr.Process(line, cmd); err != nil {
				jh.Log(TagError, err.Error())

				// set step outcome = failure
				_ = cc.stepHandle(func(ctx context.Context, handler StepCommandHandler) error {
					if exec, ok := handler.(StepExecutor); ok {
						exec.SetResult(dossiers.ResultFailure)
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

func (cc *commandController) stepHandle(fn func(context.Context, StepCommandHandler) error) error {
	if len(cc.steps) == 0 {
		return errors.New("no step found")
	}
	currentStep := cc.steps[len(cc.steps)-1]
	handler := currentStep.handler
	ctx := currentStep.ctx

	return fn(ctx, handler)
}

func (cc *commandController) jobHandle(fn func(context.Context, JobCommandHandler) error) error {
	handler := cc.job.handler
	ctx := cc.job.ctx
	if len(cc.steps) > 0 {
		currentStep := cc.steps[len(cc.steps)-1]
		ctx = currentStep.ctx
	}
	return fn(ctx, handler)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L384
func (cc *commandController) addSecretMask(cmd *command.Command) error {
	return cc.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
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
func (cc *commandController) addProblemMatcher(cmd *command.Command) error {
	return cc.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		file := cmd.Value
		if file == "" {
			return errors.New("file path must be specified in ##[add-matcher] command")
		}

		conf, err := cc.readProblemMatcherFile(ctx, jh, file)
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
func (cc *commandController) removeProblemMatcher(cmd *command.Command) error {
	return cc.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		file := cmd.Value
		owner := cmd.Params["owner"]
		if (file == "") == (owner == "") {
			return errors.New("either an owner name or a file path must be specified in ##[remove-matcher] command")
		}

		var owners []string
		if owner != "" {
			owners = []string{owner}
		} else {
			conf, err := cc.readProblemMatcherFile(ctx, jh, file)
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

func (cc *commandController) readProblemMatcherFile(ctx context.Context, jh JobCommandHandler, file string) (*problem.MatcherConfigs, error) {
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
func (cc *commandController) groupingLog(cmd *command.Command) error {
	return cc.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		jh.Log(TagGroup, cmd.Value)
		return nil
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L751
func (cc *commandController) endGroupingLog(cmd *command.Command) error {
	return cc.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		jh.Log(TagEndGroup, "")
		return nil
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L566
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L600
func (cc *commandController) logMessage(cmd *command.Command) error {
	return cc.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		// TODO
		return nil
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L417
func (cc *commandController) consoleAddPath(cmd *command.Command) error {
	return cc.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		paths := []string{cmd.Value}
		return jh.AddPath(paths)
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L234
func (cc *commandController) consoleSetEnv(cmd *command.Command) error {
	return cc.stepHandle(func(ctx context.Context, sh StepCommandHandler) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return errors.New("required field 'name' is missing in ##[set-output] command")
		}

		env := map[string]string{
			name: cmd.Value,
		}
		return sh.SetEnv(env, true)
	})
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L301
func (cc *commandController) consoleSetOutput(cmd *command.Command) error {
	return cc.stepHandle(func(ctx context.Context, sh StepCommandHandler) error {
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
func (cc *commandController) consoleSaveState(cmd *command.Command) error {
	return cc.stepHandle(func(ctx context.Context, sh StepCommandHandler) error {
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

func (cc *commandController) fileAddPath(r io.Reader) error {
	return cc.jobHandle(func(ctx context.Context, jh JobCommandHandler) error {
		return utilreader.Untar(ctx, r, func(hdr *tar.Header, reader io.Reader) error {
			if hdr.Name != "" {
				return fmt.Errorf("expected read single file with empty name, got %s", hdr.Name)
			}

			if paths, err := cc.readLine(reader); err != nil {
				return err
			} else {
				return jh.AddPath(paths)
			}
		})
	})
}

func (cc *commandController) fileSetEnv(r io.Reader) error {
	return cc.stepHandle(func(ctx context.Context, sh StepCommandHandler) error {
		return utilreader.Untar(ctx, r, func(hdr *tar.Header, reader io.Reader) error {
			if hdr.Name != "" {
				return fmt.Errorf("expected read single file with empty name, got %s", hdr.Name)
			}

			if env, err := cc.parseEnvVars(reader); err != nil {
				return err
			} else {
				return sh.SetEnv(env, true)
			}
		})
	})
}

func (cc *commandController) fileSetOutput(r io.Reader) error {
	return cc.stepHandle(func(ctx context.Context, sh StepCommandHandler) error {
		return utilreader.Untar(ctx, r, func(hdr *tar.Header, reader io.Reader) error {
			if hdr.Name != "" {
				return fmt.Errorf("expected read single file with empty name, got %s", hdr.Name)
			}

			if output, err := cc.parseEnvVars(reader); err != nil {
				return err
			} else {
				return sh.SetOutput(output)
			}
		})
	})
}

func (cc *commandController) fileSaveState(r io.Reader) error {
	return cc.stepHandle(func(ctx context.Context, sh StepCommandHandler) error {
		return utilreader.Untar(ctx, r, func(hdr *tar.Header, reader io.Reader) error {
			if hdr.Name != "" {
				return fmt.Errorf("expected read single file with empty name, got %s", hdr.Name)
			}

			if state, err := cc.parseEnvVars(reader); err != nil {
				return err
			} else {
				return sh.SaveState(state)
			}
		})
	})
}

func (cc *commandController) createStepSummary(r io.Reader) error {
	// TODO: implement GITHUB_STEP_SUMMARY file command
	return nil
}

func (cc *commandController) readLine(reader io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		l := scanner.Text()
		if l != "" && !strings.HasPrefix(l, "#") {
			lines = append(lines, l)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L342-L403
func (cc *commandController) parseEnvVars(reader io.Reader) (map[string]string, error) {
	env := make(map[string]string)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		equalsIndex := strings.Index(line, "=")
		heredocIndex := strings.Index(line, "<<")

		// Normal style NAME=VALUE
		if 0 <= equalsIndex && (heredocIndex < 0 || equalsIndex < heredocIndex) {
			key, value := line[:equalsIndex], line[equalsIndex+1:]
			if key == "" {
				return nil, fmt.Errorf("invalid nil key in line: %s", line)
			}
			env[key] = value
			continue
		}

		// Heredoc style NAME<<EOF
		if 0 <= heredocIndex && (equalsIndex < 0 || heredocIndex < equalsIndex) {
			key, delimiter := line[:heredocIndex], line[heredocIndex+2:]
			if key == "" || delimiter == "" {
				return nil, fmt.Errorf("invalid format %q. key and delimiter MUST NOT be empty", line)
			}
			value, finish := make([]string, 0), false
			for scanner.Scan() {
				doc := scanner.Text()
				if doc == delimiter {
					finish = true
					break
				}
				value = append(value, doc)
			}
			if err := scanner.Err(); err != nil {
				return nil, err
			}
			if !finish {
				return nil, fmt.Errorf("invalid value. EOF marker missing new line")
			}

			env[key] = strings.Join(value, "\n")
			continue
		}

		return nil, fmt.Errorf("invalid format: %q", line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return env, nil
}
