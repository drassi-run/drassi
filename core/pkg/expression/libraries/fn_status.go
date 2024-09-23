package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/model/records"
)

func (lib *jobLib) Always() ref.Val {
	return types.TRUE
}

func (lib *jobLib) Success() ref.Val {
	r := lib.job.Result == records.ResultSuccess
	return types.Boolean(r)
}

func (lib *jobLib) Failure() ref.Val {
	r := lib.job.Result == records.ResultFailure
	return types.Boolean(r)
}

func (lib *jobLib) Cancelled() ref.Val {
	r := lib.job.Result == records.ResultCancelled
	return types.Boolean(r)
}

func (lib *stepLib) Always() ref.Val {
	return types.TRUE
}

func (lib *stepLib) Success() ref.Val {
	r := lib.step.Outcome == records.ResultSuccess
	return types.Boolean(r)
}

func (lib *stepLib) Failure() ref.Val {
	r := lib.step.Outcome == records.ResultFailure
	return types.Boolean(r)
}

func (lib *stepLib) Cancelled() ref.Val {
	r := lib.step.Outcome == records.ResultCancelled
	return types.Boolean(r)
}
