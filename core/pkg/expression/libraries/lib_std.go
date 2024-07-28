package libraries

import (
	expr "drassi.run/core/pkg/expression"
	op "drassi.run/core/pkg/expression/ast/operators"
)

func StdLib() expr.Library {
	return stdLib{}
}

type stdLib struct{}

func (lib stdLib) EnvOptions() []expr.EnvOption {
	opts := []expr.EnvOption{
		expr.WithFunction(op.LogicalNot, unaryFn(LogicalNot)),
		expr.WithFunction(op.LogicalAnd, variadicLazyFn(LogicalAnd)),
		expr.WithFunction(op.LogicalOr, variadicLazyFn(LogicalOr)),
		expr.WithFunction(op.Equals, binaryFn(Equals)),
		expr.WithFunction(op.NotEquals, binaryFn(NotEquals)),
		expr.WithFunction(op.Less, binaryFn(Less)),
		expr.WithFunction(op.LessEquals, binaryFn(LessEquals)),
		expr.WithFunction(op.Greater, binaryFn(Greater)),
		expr.WithFunction(op.GreaterEquals, binaryFn(GreaterEquals)),
		expr.WithFunction(op.PropertyAccess, oneRestLazyFn(Index)),
		expr.WithFunction(op.IndexAccess, oneRestLazyFn(Index)),

		expr.WithFunction("contains", binaryFn(Contains)),
		expr.WithFunction("startsWith", binaryFn(StartsWith)),
		expr.WithFunction("endsWith", binaryFn(EndsWith)),
		expr.WithFunction("format", oneRestLazyFn(Format)),
		expr.WithFunction("join", oneOptionFn(Join)),
		expr.WithFunction("toJSON", unaryFn(ToJSON)),
		expr.WithFunction("fromJSON", unaryFn(FromJson)),
	}

	return opts
}
