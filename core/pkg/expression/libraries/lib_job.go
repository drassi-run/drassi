package libraries

import (
	expr "drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/records"
)

func JobLib(job *records.Job) expr.Library {
	return &jobLib{job: job}
}

type jobLib struct {
	job *records.Job
}

func (lib *jobLib) EnvOptions() []expr.Option {
	opts := []expr.Option{
		expr.WithFunction("success", nullaryFn(lib.Success)),
		expr.WithFunction("always", nullaryFn(lib.Always)),
		expr.WithFunction("cancelled", nullaryFn(lib.Cancelled)),
		expr.WithFunction("failure", nullaryFn(lib.Failure)),
	}

	return opts
}
