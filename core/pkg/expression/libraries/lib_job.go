package libraries

import (
	expr "drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/dossiers"
)

func JobLib(job *dossiers.Job) expr.Library {
	return &jobLib{job: job}
}

type jobLib struct {
	job *dossiers.Job
}

func (lib *jobLib) EnvOptions() []expr.EnvOption {
	opts := []expr.EnvOption{
		expr.WithFunction("success", nullaryFn(lib.Success)),
		expr.WithFunction("always", nullaryFn(lib.Always)),
		expr.WithFunction("cancelled", nullaryFn(lib.Cancelled)),
		expr.WithFunction("failure", nullaryFn(lib.Failure)),
	}

	return opts
}
