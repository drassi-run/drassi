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
	xdig "drassi.run/core/util/dig"
	"drassi.run/core/wire"
	"go.uber.org/dig"
)

const (
	ConsoleCommandHandlers = "console-handlers"
	FileCommandHandlers    = "file-handlers"
)

func ProvideTo(scope *dig.Scope) error {
	if err := scope.Provide(cmd.NewFileManager[exec.Milieu]); err != nil {
		return err
	}
	if err := scope.Provide(cmd.NewConsoleManager[exec.Milieu]); err != nil {
		return err
	}

	if err := provideConsoleHandlers(scope); err != nil {
		return err
	}

	if err := provideFileHandlers(scope); err != nil {
		return err
	}

	if err := scope.Provide(NewCommandDecorator, dig.Name("command")); err != nil {
		return err
	}

	if err := scope.Provide(NewCommandEnvProvider, dig.Group(wire.EnvProvider)); err != nil {
		return err
	}

	return xdig.Supply(scope, NewCommandInitHook[exec.JobExecutor](scope), dig.Group(wire.PostStart))
}

func provideConsoleHandlers(scope *dig.Scope) error {
	group := dig.Group(ConsoleCommandHandlers)
	if err := scope.Provide(ch.AddSecretMask[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.AddProblemMatcher[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.RemoveProblemMatcher[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.GroupingLog[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.EndGroupingLog[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.DebugMessage[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.LogMessage[exec.Milieu], dig.Group(ConsoleCommandHandlers+",flatten")); err != nil {
		return err
	}
	if err := scope.Provide(ch.ConsoleAddPath[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.ConsoleSetEnv[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.ConsoleSetOutput[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.ConsoleSaveState[exec.Milieu], group); err != nil {
		return err
	}
	return nil
}

func provideFileHandlers(scope *dig.Scope) error {
	group := dig.Group(FileCommandHandlers)
	if err := scope.Provide(ch.FileAddPath[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.FileSetEnv[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.FileSaveState[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.FileSetOutput[exec.Milieu], group); err != nil {
		return err
	}
	if err := scope.Provide(ch.CreateStepSummary[exec.Milieu], group); err != nil {
		return err
	}
	return nil
}
