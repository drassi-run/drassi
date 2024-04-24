package interfaces

import (
	"github.com/dungdm93/drasi/pkg/expr"
	"github.com/dungdm93/drasi/pkg/secret_masker"
)

type Context interface {
	State() any
	Masker() secret_masker.SecretMasker
	Trace() TraceWriter
	SetTraceResult(node Node, result expr.Result)
	TryGetTraceResult(node Node) (exist bool, result string)
	Visitor() Visitor
}
