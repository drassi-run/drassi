package tools

import (
	"math"

	"github.com/dungdm93/drassi/core/pkg/expression/ast"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/functions"
)

func GetFunctionNode(names []string) []functions.IFnInfo[ast_ifaces.Fn] {
	var ret []functions.IFnInfo[ast_ifaces.Fn]
	for _, f := range names {
		switch f {
		case "always":
			ret = append(ret, functions.NewFunctionInfo[functions.Always]("always", 0, math.MaxInt32))
		case "cancelled":
			ret = append(ret, functions.NewFunctionInfo[functions.Cancelled]("cancelled", 0, 0))
		case "success":
			ret = append(ret, functions.NewFunctionInfo[functions.Success]("success", 0, 0))
		case "failure":
			ret = append(ret, functions.NewFunctionInfo[functions.Failure]("failure", 0, 0))
		case "hashfiles":
			ret = append(ret, functions.NewFunctionInfo[functions.HashFile]("hashfiles", 1, math.MaxUint8))
		default:
		}
	}
	return ret
}


func GetNamedValueNode(names []string) []ast.INamedValueInfo[ast.INamedValue]{
	var ret []ast.INamedValueInfo[ast.INamedValue]
	for _, f := range names {
		switch f {
		case "github":
			ret = append(ret, ast.NewNamedValueInfo[ast.ContextValueNode]("github"))
		case "strategy":
			ret = append(ret,ast.NewNamedValueInfo[ast.ContextValueNode]("strategy"))
		case "env":
			ret = append(ret, ast.NewNamedValueInfo[ast.ContextValueNode]("env"))
		case "steps":
			ret = append(ret, ast.NewNamedValueInfo[ast.ContextValueNode]("steps"))
		case "runner":
			ret = append(ret, ast.NewNamedValueInfo[ast.ContextValueNode]("runner"))
		case "needs":
			ret = append(ret, ast.NewNamedValueInfo[ast.ContextValueNode]("needs"))
		case "inputs":
			ret = append(ret, ast.NewNamedValueInfo[ast.ContextValueNode]("inputs"))
		default:
		}
	}
	return ret
}
