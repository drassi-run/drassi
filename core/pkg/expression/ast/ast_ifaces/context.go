package ast_ifaces

import (
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/secret_masker"
)

type Context interface {
	State() any
	Masker() secret_masker.Interface
	Trace() TraceWriter
	SetTraceResult(node ExprNode, result expression.Result)
	TryGetTraceResult(node ExprNode) (exist bool, result string)
	Visitor() Visitor
}
