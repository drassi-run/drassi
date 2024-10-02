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
	if err := scope.Provide(AddSecretMask, group); err != nil {
		return err
	}
	if err := scope.Provide(AddProblemMatcher, group); err != nil {
		return err
	}
	if err := scope.Provide(RemoveProblemMatcher, group); err != nil {
		return err
	}
	if err := scope.Provide(GroupingLog, group); err != nil {
		return err
	}
	if err := scope.Provide(EndGroupingLog, group); err != nil {
		return err
	}
	if err := scope.Provide(DebugMessage, group); err != nil {
		return err
	}
	if err := scope.Provide(LogMessage, dig.Group(wire.ConsoleCommandHandlers+",flatten")); err != nil {
		return err
	}
	if err := scope.Provide(ConsoleAddPath, group); err != nil {
		return err
	}
	if err := scope.Provide(ConsoleSetEnv, group); err != nil {
		return err
	}
	if err := scope.Provide(ConsoleSetOutput, group); err != nil {
		return err
	}
	if err := scope.Provide(ConsoleSaveState, group); err != nil {
		return err
	}
	return nil
}

func provideFileHandlers(scope *dig.Scope) error {
	group := dig.Group(wire.FileCommandHandlers)
	if err := scope.Provide(FileAddPath, group); err != nil {
		return err
	}
	if err := scope.Provide(FileSetEnv, group); err != nil {
		return err
	}
	if err := scope.Provide(FileSaveState, group); err != nil {
		return err
	}
	if err := scope.Provide(FileSetOutput, group); err != nil {
		return err
	}
	if err := scope.Provide(CreateStepSummary, group); err != nil {
		return err
	}
	return nil
}
