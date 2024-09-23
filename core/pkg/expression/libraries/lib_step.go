package libraries

import (
	expr "drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/records"
)

func StepLib(step *records.Step) expr.Library {
	return &stepLib{step: step}
}

type stepLib struct {
	step *records.Step
}

func (lib *stepLib) EnvOptions() []expr.Option {
	opts := []expr.Option{
		expr.WithFunction("success", nullaryFn(lib.Success)),
		expr.WithFunction("always", nullaryFn(lib.Always)),
		expr.WithFunction("cancelled", nullaryFn(lib.Cancelled)),
		expr.WithFunction("failure", nullaryFn(lib.Failure)),
	}

	return opts
}
