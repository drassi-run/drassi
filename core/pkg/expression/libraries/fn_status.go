package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/model/records"
)

func (lib *statusLib) Always() ref.Val {
	return types.TRUE
}

func (lib *statusLib) Success() ref.Val {
	r := lib.Status() == records.ResultSuccess
	return types.Boolean(r)
}

func (lib *statusLib) Failure() ref.Val {
	r := lib.Status() == records.ResultFailure
	return types.Boolean(r)
}

func (lib *statusLib) Cancelled() ref.Val {
	r := lib.Status() == records.ResultCancelled
	return types.Boolean(r)
}
