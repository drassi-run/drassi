package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/model/dossiers"
)

func (lib *jobLib) Always() ref.Val {
	return types.TRUE
}

func (lib *jobLib) Success() ref.Val {
	r := lib.info.Status == dossiers.ResultSuccess
	return types.Boolean(r)
}

func (lib *jobLib) Failure() ref.Val {
	r := lib.info.Status == dossiers.ResultFailure
	return types.Boolean(r)
}

func (lib *jobLib) Cancelled() ref.Val {
	r := lib.info.Status == dossiers.ResultCancelled
	return types.Boolean(r)
}

func (lib *stepLib) Always() ref.Val {
	return types.TRUE
}

func (lib *stepLib) Success() ref.Val {
	r := lib.info.ActionStatus == dossiers.ResultSuccess
	return types.Boolean(r)
}

func (lib *stepLib) Failure() ref.Val {
	r := lib.info.ActionStatus == dossiers.ResultFailure
	return types.Boolean(r)
}

func (lib *stepLib) Cancelled() ref.Val {
	r := lib.info.ActionStatus == dossiers.ResultCancelled
	return types.Boolean(r)
}
