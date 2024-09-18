package libraries

import (
	expr "drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/dossiers"
)

func JobLib(jobInfo *dossiers.JobInfo) expr.Library {
	lib := &jobLib{info: jobInfo}

	return lib
}

type jobLib struct {
	info *dossiers.JobInfo
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
