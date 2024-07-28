package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
)

func Equals(left, right ref.Val) ref.Val {
	r := equalWeak(left, right)
	return types.Boolean(r)
}

func NotEquals(left, right ref.Val) ref.Val {
	r := !equalWeak(left, right)
	return types.Boolean(r)
}
