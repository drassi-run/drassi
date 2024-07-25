package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func LogicalAnd(operands ...ref.LazyVal) ref.Val {
	var res ref.Val = types.NULL

	for _, o := range operands {
		res = o()
		if res.Type() == ref.TypeInvalid {
			return res
		}
		if isFalsy(res) {
			return res
		}
	}
	return res
}

func LogicalOr(operands ...ref.LazyVal) ref.Val {
	var res ref.Val = types.NULL

	for _, o := range operands {
		res = o()
		if res.Type() == ref.TypeInvalid {
			return res
		}
		if isTruthy(res) {
			return res
		}
	}
	return res
}

func LogicalNot(operand ref.Val) ref.Val {
	b := isFalsy(operand)
	return types.Boolean(b)
}

func isTruthy(v ref.Val) bool {
	b, ok := v.(traits.Logical)
	// non-Logical value e.g. List or Map, it's treated as `true`
	return !ok || b.ToBoolean()
}

func isFalsy(v ref.Val) bool {
	b, ok := v.(traits.Logical)
	return ok && !b.ToBoolean()
}
