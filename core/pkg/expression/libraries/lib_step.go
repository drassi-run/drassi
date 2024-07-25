package libraries

import (
	expr "drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/dossiers"
)

func StepLib(stepInfo *dossiers.Github) expr.Library {
	lib := &stepLib{info: stepInfo}

	return lib
}

type stepLib struct {
	info *dossiers.Github
}

func (lib *stepLib) EnvOptions() []expr.EnvOption {
	opts := []expr.EnvOption{
		expr.WithFunction("success", nullaryFn(lib.Success)),
		expr.WithFunction("always", nullaryFn(lib.Always)),
		expr.WithFunction("cancelled", nullaryFn(lib.Cancelled)),
		expr.WithFunction("failure", nullaryFn(lib.Failure)),
	}

	return opts
}
