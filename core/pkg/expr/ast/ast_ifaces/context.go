package ast_ifaces

import (
	"github.com/dungdm93/drassi/core/pkg/expr"
	"github.com/dungdm93/drassi/core/pkg/secret_masker"
)

type Context interface {
	State() any
	Masker() secret_masker.Interface
	Trace() TraceWriter
	SetTraceResult(node ExprNode, result expr.Result)
	TryGetTraceResult(node ExprNode) (exist bool, result string)
	Visitor() Visitor
}
