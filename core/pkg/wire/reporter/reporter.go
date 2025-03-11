package wire_reporter

import (
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/reporter"
	"go.uber.org/dig"
)

func ProvideTo(scope *dig.Scope) error {
	if err := scope.Provide(reporter.Reporter.Stdin, dig.Name("stdin")); err != nil {
		return err
	}
	if err := scope.Provide(reporter.Reporter.Stdout, dig.Name("stdout")); err != nil {
		return err
	}
	if err := scope.Provide(reporter.Reporter.Stderr, dig.Name("stderr")); err != nil {
		return err
	}

	return nil
}

func Wire(scope *dig.Scope) error {
	return scope.Invoke(registerCallbacks)
}

func registerCallbacks(rep reporter.Reporter, sup executor.Supervisor) error {
	sup.Register(executor.BeforeRunJobCallback(rep.StartJob))
	sup.Register(executor.AfterRunJobCallback(rep.EndJob))
	sup.Register(executor.BeforeRunStepCallback(rep.StartStep))
	sup.Register(executor.AfterRunStepCallback(rep.EndStep))

	return nil
}
