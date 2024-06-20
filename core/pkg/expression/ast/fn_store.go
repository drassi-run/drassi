package ast

import (
	"math"

	"drassi.run/core/pkg/expression/ast/ast_ifaces"
	"drassi.run/core/pkg/expression/ast/functions"
)

type FnStore struct {
	WellKnownFns map[string]functions.IFnInfo[ast_ifaces.Fn]
}

var e *FnStore

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
	// [] toJson
	// [] fromJSON
	e.WellKnownFns["contains"] = functions.NewFunctionInfo[functions.Contains]("contains", 2, 2)
	e.WellKnownFns["endswith"] = functions.NewFunctionInfo[functions.EndsWith]("endsWith", 2, 2)
	e.WellKnownFns["startswith"] = functions.NewFunctionInfo[functions.StartsWith]("startsWith", 2, 2)
	e.WellKnownFns["join"] = functions.NewFunctionInfo[functions.Join]("join", 1, 2)
	e.WellKnownFns["format"] = functions.NewFunctionInfo[functions.Format]("format", 1, math.MaxUint8)
	e.WellKnownFns["fromjson"] = functions.NewFunctionInfo[functions.FromJson]("fromJSON", 1, 1)
	e.WellKnownFns["tojson"] = functions.NewFunctionInfo[functions.ToJson]("toJSON", 1, 1)
	e.WellKnownFns["always"] = functions.NewFunctionInfo[functions.Always]("always", 0, 2147483647)
	e.WellKnownFns["cancelled"] = functions.NewFunctionInfo[functions.Cancelled]("cancelled", 0, 0)
	e.WellKnownFns["success"] = functions.NewFunctionInfo[functions.Success]("success", 0, 0)
	e.WellKnownFns["failure"] = functions.NewFunctionInfo[functions.Failure]("failure", 0, 0)
	return e
}
