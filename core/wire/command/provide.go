/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_command

import (
	exec "drassi.run/core/pkg/executor"
	cmd "drassi.run/core/pkg/executor/command"
	ch "drassi.run/core/pkg/executor/command/cmdhandler"
	"go.uber.org/dig"
)

const (
	ConsoleCommandHandlers = "console-handlers"
	FileCommandHandlers    = "file-handlers"
)

func ProvideTo(scope *dig.Scope) error {
	if err := scope.Provide(cmd.NewFileManager[exec.SupportCommands]); err != nil {
		return err
	}
	if err := scope.Provide(cmd.NewConsoleManager[exec.SupportCommands]); err != nil {
		return err
	}

	if err := provideConsoleHandlers(scope); err != nil {
		return err
	}

	if err := provideFileHandlers(scope); err != nil {
		return err
	}

	return scope.Provide(NewCommandDecorator,
		dig.As(new(exec.ActionRunDecorator), new(exec.JobRunDecorator)),
		dig.Name("command"),
	)
}

func provideConsoleHandlers(scope *dig.Scope) error {
	group := dig.Group(ConsoleCommandHandlers)
	if err := scope.Provide(ch.AddSecretMask[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.AddProblemMatcher[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.RemoveProblemMatcher[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.GroupingLog[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.EndGroupingLog[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.DebugMessage[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.LogMessage[exec.SupportCommands], dig.Group(ConsoleCommandHandlers+",flatten")); err != nil {
		return err
	}
	if err := scope.Provide(ch.ConsoleAddPath[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.ConsoleSetEnv[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.ConsoleSetOutput[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.ConsoleSaveState[exec.SupportCommands], group); err != nil {
		return err
	}
	return nil
}

func provideFileHandlers(scope *dig.Scope) error {
	group := dig.Group(FileCommandHandlers)
	if err := scope.Provide(ch.FileAddPath[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.FileSetEnv[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.FileSaveState[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.FileSetOutput[exec.SupportCommands], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.CreateStepSummary[exec.SupportCommands], group); err != nil {
		return err
	}
	return nil
}
