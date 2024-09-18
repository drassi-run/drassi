package wire_cmdhandler

import (
	"drassi.run/core/pkg/wire"
	"go.uber.org/dig"
)

func ProvideTo(scope *dig.Scope) error {
	if err := provideConsoleHandlers(scope); err != nil {
		return err
	}

	if err := provideFileHandlers(scope); err != nil {
		return err
	}

	return nil
}

func provideConsoleHandlers(scope *dig.Scope) error {
	group := dig.Group(wire.ConsoleCommandHandlers)
	if err := scope.Provide(addSecretMask, group); err != nil {
		return err
	}
	if err := scope.Provide(addProblemMatcher, group); err != nil {
		return err
	}
	if err := scope.Provide(removeProblemMatcher, group); err != nil {
		return err
	}
	if err := scope.Provide(groupingLog, group); err != nil {
		return err
	}
	if err := scope.Provide(endGroupingLog, group); err != nil {
		return err
	}
	if err := scope.Provide(logMessage, dig.Group(wire.ConsoleCommandHandlers+",flatten")); err != nil {
		return err
	}
	if err := scope.Provide(consoleAddPath, group); err != nil {
		return err
	}
	if err := scope.Provide(consoleSetEnv, group); err != nil {
		return err
	}
	if err := scope.Provide(consoleSetOutput, group); err != nil {
		return err
	}
	if err := scope.Provide(consoleSaveState, group); err != nil {
		return err
	}
	return nil
}

func provideFileHandlers(scope *dig.Scope) error {
	group := dig.Group(wire.FileCommandHandlers)
	if err := scope.Provide(fileAddPath, group); err != nil {
		return err
	}
	if err := scope.Provide(fileSetEnv, group); err != nil {
		return err
	}
	if err := scope.Provide(fileSaveState, group); err != nil {
		return err
	}
	if err := scope.Provide(fileSetOutput, group); err != nil {
		return err
	}
	if err := scope.Provide(createStepSummary, group); err != nil {
		return err
	}
	return nil
}
