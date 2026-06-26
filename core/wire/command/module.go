/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_command

import (
	"fmt"

	cmd "drassi.run/core/pkg/command"
	ch "drassi.run/core/pkg/command/cmdhandler"
	"drassi.run/core/pkg/command/cmdtypes"
	exec "drassi.run/core/pkg/executor"
	xdig "drassi.run/core/util/dig"
	"drassi.run/core/wire"
	"go.uber.org/dig"
)

const (
	ConsoleCommandHandlers = "console-handlers"
	FileCommandHandlers    = "file-handlers"
)

type Option func(o *options)
type options struct {
	defaultIssueReporter      bool // use cmdtypes.Discard as cmdtypes.Reporter
	defaultAttachmentUploader bool // use cmdtypes.BlackHole as cmdtypes.Attacher
}

func UseDiscardIssueReporter(b bool) Option {
	return func(o *options) {
		o.defaultIssueReporter = b
	}
}

func UseBlackHoleAttachmentUploader(b bool) Option {
	return func(o *options) {
		o.defaultAttachmentUploader = b
	}
}

func Module(opts ...Option) *wire.Module {
	o := &options{
		defaultIssueReporter: true,
	}
	for _, opt := range opts {
		opt(o)
	}

	fn := func(scope *dig.Scope) error {
		if err := scope.Provide(cmd.NewFileManager[exec.Milieu]); err != nil {
			return fmt.Errorf("provide command.FileManager: %w", err)
		}
		if err := scope.Provide(cmd.NewConsoleManager[exec.Milieu]); err != nil {
			return fmt.Errorf("provide command.ConsoleManager: %w", err)
		}

		if err := provideConsoleHandlers(scope); err != nil {
			return err
		}

		if err := provideFileHandlers(scope); err != nil {
			return err
		}

		if err := scope.Provide(NewCommandDecorator, dig.Name("command")); err != nil {
			return fmt.Errorf("provide 'file-command' ActionRunDecorator: %w", err)
		}

		if err := scope.Provide(NewCommandEnvProvider, dig.Group(wire.EnvProvider)); err != nil {
			return fmt.Errorf("provide 'file-command' EnvProvider: %w", err)
		}

		if err := xdig.Supply(scope, NewCommandInitHook[exec.JobExecutor](scope), dig.Group(wire.PostStart)); err != nil {
			return fmt.Errorf("provide 'command' post-start Hook: %w", err)
		}

		if o.defaultIssueReporter {
			if err := scope.Provide(cmdtypes.Discard[exec.Milieu]); err != nil {
				return fmt.Errorf("provide 'discard' cmdtypes.Reporter: %w", err)
			}
		}

		if err := scope.Provide(refineIssueFileProp, dig.Group("decorator")); err != nil {
			return fmt.Errorf("decorate cmdtypes.Reporter by refine 'file' prop: %w", err)
		}
		if err := xdig.Decorate[cmdtypes.Reporter[exec.Milieu]](scope); err != nil {
			return fmt.Errorf("decorate cmdtypes.Reporter: %w", err)
		}

		if o.defaultAttachmentUploader {
			if err := scope.Provide(cmdtypes.BlackHole[exec.Milieu]); err != nil {
				return fmt.Errorf("provide 'blackhole' cmdtypes.Attacher: %w", err)
			}
		}

		return nil
	}
	return wire.NewModule("core/command", fn)
}

func provideConsoleHandlers(scope *dig.Scope) error {
	group := dig.Group(ConsoleCommandHandlers)
	if err := scope.Provide(ch.AddSecretMask[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'add-secret' command.ConsoleHandler: %w", err)
	}
	if err := scope.Provide(ch.AddProblemMatcher[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'add-matcher' command.ConsoleHandler: %w", err)
	}
	if err := scope.Provide(ch.RemoveProblemMatcher[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'remove-matcher' command.ConsoleHandler: %w", err)
	}
	if err := scope.Provide(ch.GroupingLog[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'group' command.ConsoleHandler: %w", err)
	}
	if err := scope.Provide(ch.EndGroupingLog[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'endgroup' command.ConsoleHandler: %w", err)
	}
	if err := scope.Provide(ch.DebugMessage[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'debug' command.ConsoleHandler: %w", err)
	}
	if err := scope.Provide(ch.LogMessage[exec.Milieu], dig.Group(ConsoleCommandHandlers+",flatten")); err != nil {
		return fmt.Errorf("provide log command.ConsoleHandler: %w", err)
	}
	if err := scope.Provide(ch.ConsoleAddPath[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'add-path' command.ConsoleHandler: %w", err)
	}
	if err := scope.Provide(ch.ConsoleSetEnv[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'set-env' command.ConsoleHandler: %w", err)
	}
	if err := scope.Provide(ch.ConsoleSetOutput[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'set-output' command.ConsoleHandler: %w", err)
	}
	if err := scope.Provide(ch.ConsoleSaveState[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'save-state' command.ConsoleHandler: %w", err)
	}
	return nil
}

func provideFileHandlers(scope *dig.Scope) error {
	group := dig.Group(FileCommandHandlers)
	if err := scope.Provide(ch.FileAddPath[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'GITHUB_PATH' command.FileHandler: %w", err)
	}
	if err := scope.Provide(ch.FileSetEnv[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'GITHUB_ENV' command.FileHandler: %w", err)
	}
	if err := scope.Provide(ch.FileSaveState[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'GITHUB_STATE' command.FileHandler: %w", err)
	}
	if err := scope.Provide(ch.FileSetOutput[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'GITHUB_OUTPUT' command.FileHandler: %w", err)
	}
	if err := scope.Provide(ch.CreateStepSummary[exec.Milieu], group); err != nil {
		return fmt.Errorf("provide 'GITHUB_STEP_SUMMARY' command.FileHandler: %w", err)
	}
	return nil
}
