package wire_reporter

import (
	"io"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/executor/secret"
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
	if err := scope.Decorate(maskStdout); err != nil {
		return err
	}
	if err := scope.Decorate(maskStderr); err != nil {
		return err
	}

	return nil
}

func Wire(scope *dig.Scope) error {
	return scope.Invoke(registerCallbacks)
}

type maskStdoutParams struct {
	dig.In

	StdOut       io.Writer `name:"stdout"`
	SecretMasker secret.Masker
}

type maskStdoutResult struct {
	dig.Out

	StdOut io.Writer `name:"stdout"`
}

func maskStdout(p maskStdoutParams) maskStdoutResult {
	w := secret.NewWriter(p.StdOut, p.SecretMasker)

	return maskStdoutResult{
		StdOut: w,
	}
}

type maskStderrParams struct {
	dig.In

	StdErr       io.Writer `name:"stderr"`
	SecretMasker secret.Masker
}

type maskStderrResult struct {
	dig.Out

	StdErr io.Writer `name:"stderr"`
}

func maskStderr(p maskStderrParams) maskStderrResult {
	w := secret.NewWriter(p.StdErr, p.SecretMasker)

	return maskStderrResult{
		StdErr: w,
	}
}

func registerCallbacks(rep reporter.Reporter, sup executor.Supervisor) error {
	sup.Register(executor.BeforeRunJobCallback(rep.StartJob))
	sup.Register(executor.AfterRunJobCallback(rep.EndJob))
	sup.Register(executor.BeforeRunStepCallback(rep.StartStep))
	sup.Register(executor.AfterRunStepCallback(rep.EndStep))

	return nil
}
