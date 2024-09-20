package wire_reporter

import (
	"io"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/model/dossiers"
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
	h1 := func(je executor.JobExecutor) error {
		rep.StartJob()
		return nil
	}
	sup.Register(executor.BeforeRunJobCallback(h1))

	h2 := func(je executor.JobExecutor, result *dossiers.Job) error {
		rep.EndJob(result.Result, result.Outputs)
		return nil
	}
	sup.Register(executor.AfterRunJobCallback(h2))

	h3 := func(se executor.StepExecutor) error {
		rep.StartStep(se.StepId())
		return nil
	}
	sup.Register(executor.BeforeRunStepCallback(h3))

	h4 := func(se executor.StepExecutor, result *dossiers.Step) error {
		rep.EndStep(se.StepId(), result.Outcome)
		return nil
	}
	sup.Register(executor.AfterRunStepCallback(h4))

	return nil
}
