package ast

import (
	"math"

	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/functions"
)

type FnStore struct {
	WellKnownFns map[string]functions.IFnInfo[ast_ifaces.Fn]
}

var e *FnStore

// Store functions that are available everywhere, see https://github.com/actions/runner/blob/dfaf6e06ee862f94e0daf11ac5e5d197b913080c/src/Sdk/DTExpressions2/Expressions2/ExpressionConstants.cs#L10

func NewFnStore() *FnStore {
	if e != nil {
		return e
	}
	e = &FnStore{WellKnownFns: make(map[string]functions.IFnInfo[ast_ifaces.Fn])}
	// WellKnownFns:
	// [x] contains
	// [x] endsWith
	// [x] startsWith
	// [x] join
	// [x] format
	// [x] toJson
	// [x] fromJSON
	e.WellKnownFns["contains"] = functions.NewFunctionInfo[functions.Contains]("contains", 2, 2)
	e.WellKnownFns["endswith"] = functions.NewFunctionInfo[functions.EndsWith]("endsWith", 2, 2)
	e.WellKnownFns["startswith"] = functions.NewFunctionInfo[functions.StartsWith]("startsWith", 2, 2)
	e.WellKnownFns["format"] = functions.NewFunctionInfo[functions.Format]("format", 1, math.MaxUint8)
	e.WellKnownFns["join"] = functions.NewFunctionInfo[functions.Join]("join", 1, 2)
	e.WellKnownFns["fromjson"] = functions.NewFunctionInfo[functions.FromJson]("fromJSON", 1, 1)
	e.WellKnownFns["tojson"] = functions.NewFunctionInfo[functions.ToJson]("toJSON", 1, 1)
	return e
}
