package ast_ifaces

import (
	"github.com/dungdm93/drassi/core/pkg/expression"
	"github.com/dungdm93/drassi/core/pkg/secret_masker"
)

type Context interface {
	State() any
	Masker() secret_masker.Interface
	Trace() TraceWriter
	SetTraceResult(node ExprNode, result expression.Result)
	TryGetTraceResult(node ExprNode) (exist bool, result string)
	Visitor() Visitor
}
