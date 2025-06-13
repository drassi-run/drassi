/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_cmdhandler

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/util/path"
	"drassi.run/core/util/tar"
	"k8s.io/apimachinery/pkg/util/sets"
)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L384
func AddSecretMask(secretMasker secret.Masker) *command.ConsoleHandler {
	run := func(ctx context.Context, cmd *command.Command) error {
		if cmd.Value == "" {
			return fmt.Errorf("%w %q: empty value", command.ErrInvalidCommand, "add-mask")
		}
		s := secret.NewValueSecret(cmd.Value)
		secretMasker.AddSecret(s)
		for mask := range splitLine(cmd.Value) {
			if mask != cmd.Value {
				s = secret.NewValueSecret(mask)
				secretMasker.AddSecret(s)
			}
		}
		scribe.Debugf(ctx, "Added secret mask")
		return nil
	}
	return command.NewConsoleHandler("add-mask", false, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L451
func AddProblemMatcher(m map[string]problem.Matcher, sb sandboxer.Sandbox, stack executor.Stack) *command.ConsoleHandler {
	run := func(ctx context.Context, cmd *command.Command) error {
		file := cmd.Value
		if file == "" {
			return fmt.Errorf("%w %q: empty file path (in cmd value)", command.ErrInvalidCommand, "add-matcher")
		}
		if pt := getPathTranslator(stack.Leaf()); pt != nil {
			if p, ok := pt.TranslatePath(file); ok {
				file = p
			}
		}

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

		var owners = make([]string, len(conf.Configs))
		for i, config := range conf.Configs {
			if matcher, err := problem.NewMatcher(config.Severity, config.Patterns); err != nil {
				return err
			} else {
				owners[i] = config.Owner
				m[config.Owner] = matcher
			}
		}
		scribe.Debugf(ctx,
			"Added matchers: %s. Problem matchers scan action output for known warning or error strings and report these inline.",
			strings.Join(owners, ", "))
		return nil
	}
	return command.NewConsoleHandler("add-matcher", true, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L498
func RemoveProblemMatcher(m map[string]problem.Matcher, sb sandboxer.Sandbox, stack executor.Stack) *command.ConsoleHandler {
	run := func(ctx context.Context, cmd *command.Command) error {
		file := cmd.Value
		owner := cmd.Params["owner"]
		if (file == "") == (owner == "") {
			return fmt.Errorf("%w %q: either owner or file must be specified, but not both", command.ErrInvalidCommand, "remove-matcher")
		}
		if file != "" {
			if pt := getPathTranslator(stack.Leaf()); pt != nil {
				if p, ok := pt.TranslatePath(file); ok {
					file = p
				}
			}
		}

		var owners []string
		if owner != "" {
			owners = []string{owner}
		} else {
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
		scribe.Debugf(ctx, "Removed matchers: %s", strings.Join(owners, ", "))
		return nil
	}
	return command.NewConsoleHandler("remove-matcher", true, run)
}

func readProblemMatcherFile(ctx context.Context, sb sandboxer.Sandbox, file string) (*problem.MatcherConfigs, error) {
	layout := sb.Layout()
	file = xpath.Abs(file, layout.Workspace)

	reader, err := sb.CopyOut(ctx, file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	found := false
	conf := new(problem.MatcherConfigs)
	err = xtar.Untar(ctx, reader, func(hdr *tar.Header, tr io.Reader) error {
		if !xtar.IsRegular(hdr) {
			return fmt.Errorf("%w: un-expected %s file", ErrInvalidFile, xtar.FileType(hdr.Typeflag))
		}
		if found {
			return fmt.Errorf("%w: un-expected multiple files", ErrInvalidFile)
		}
		found = true
		return json.NewDecoder(tr).Decode(conf)
	})
	return conf, err
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L751
func GroupingLog() *command.ConsoleHandler {
	run := func(ctx context.Context, cmd *command.Command) error {
		scribe.Log(ctx, scribe.TagGroup, cmd.Value)
		return nil
	}
	return command.NewConsoleHandler("group", true, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L751
func EndGroupingLog() *command.ConsoleHandler {
	run := func(ctx context.Context, cmd *command.Command) error {
		scribe.Log(ctx, scribe.TagEndGroup, cmd.Value)
		return nil
	}
	return command.NewConsoleHandler("endgroup", true, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L566
func DebugMessage() *command.ConsoleHandler {
	run := func(ctx context.Context, cmd *command.Command) error {
		scribe.Log(ctx, scribe.TagDebug, cmd.Value)
		return nil
	}
	return command.NewConsoleHandler("debug", false, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L600
func LogMessage() []*command.ConsoleHandler {
	run := func(ctx context.Context, cmd *command.Command) error {
		scribe.Log(ctx, cmd.Name, cmd.Value)
		return nil
	}
	return []*command.ConsoleHandler{
		command.NewConsoleHandler("notice", false, run),
		command.NewConsoleHandler("warning", false, run),
		command.NewConsoleHandler("error", false, run),
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L417
func ConsoleAddPath(stack executor.Stack) *command.ConsoleHandler {
	run := func(ctx context.Context, cmd *command.Command) error {
		if cmd.Value == "" {
			return fmt.Errorf("%w %q: missing value", command.ErrInvalidCommand, "add-path")
		}

		job := stack.Job()
		if job == nil {
			return ErrNoJobRunning
		}
		paths := []string{cmd.Value}
		scribe.Debugf(ctx, "Add path: %q", cmd.Value)
		job.AddPath(paths)
		return nil
	}
	return command.NewConsoleHandler("add-path", true, run)
}

var setEnvBlockList = sets.New("NODE_OPTIONS")

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L234
func ConsoleSetEnv(stack executor.Stack, tracker executor.Tracker) *command.ConsoleHandler {
	run := func(ctx context.Context, cmd *command.Command) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return fmt.Errorf("%w %q: required field %q is missing", command.ErrInvalidCommand, "set-env", "name")
		}

		if setEnvBlockList.Has(name) {
			iss := &executor.Issue{
				Type:    executor.IssueTypeError,
				Message: fmt.Sprintf("Can't update %q environment variable using ::%s:: command.", name, cmd.Name),
			}
			if err := tracker.AddIssue(iss); err != nil {
				return err
			}
		}

		env := map[string]string{name: cmd.Value}
		scribe.Debugf(ctx, "Set env: %s = %s", name, cmd.Value)

		for _, step := range stack.Stack() {
			step.SetEnv(env)
		}
		stack.Job().SetEnv(env)
		return nil
	}
	return command.NewConsoleHandler("set-env", true, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L301
func ConsoleSetOutput(stack executor.Stack) *command.ConsoleHandler {
	run := func(ctx context.Context, cmd *command.Command) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return fmt.Errorf("%w %q: required field %q is missing", command.ErrInvalidCommand, "set-output", "name")
		}

		step := stack.Leaf()
		if step == nil {
			return ErrNoStepRunning
		}

		output := map[string]string{
			name: cmd.Value,
		}
		scribe.Debugf(ctx, "Set output: %s = %s", name, cmd.Value)
		step.SetOutput(output)
		return nil
	}
	return command.NewConsoleHandler("set-output", true, run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L336
func ConsoleSaveState(stack executor.Stack) *command.ConsoleHandler {
	run := func(ctx context.Context, cmd *command.Command) error {
		name, ok := cmd.Params["name"]
		if !ok || name == "" {
			return fmt.Errorf("%w %q: required field %q is missing", command.ErrInvalidCommand, "save-state", "name")
		}

		step := stack.Root()
		if step == nil {
			return ErrNoStepRunning
		}

		state := map[string]string{
			name: cmd.Value,
		}
		scribe.Debugf(ctx, "Save intra-action state: %s = %s", name, cmd.Value)
		step.SaveState(state)
		return nil
	}
	return command.NewConsoleHandler("save-state", true, run)
}
