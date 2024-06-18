package evaluator

import (
	"github.com/dungdm93/drassi/core/pkg/expression"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expression/common"
	"github.com/dungdm93/drassi/core/pkg/expression/secret_masker"
)

type (
	context struct {
		state        any
		opt          *Option
		traceWriter  ast_ifaces.TraceWriter
		secretMasker secret_masker.Interface
		traceResults map[ast_ifaces.ExprNode]string
		visitor      ast_ifaces.Visitor
	}
)

func newContext(t ast_ifaces.TraceWriter, sm secret_masker.Interface, s any, o *Option, v ast_ifaces.Visitor) *context {
	if t == nil {
		panic(ErrorEmptyTrace)
	}
	if sm == nil {
		panic(ErrorEmptySecretMasker)
	}
	return &context{
		state:        s,
		traceWriter:  t,
		secretMasker: sm,
		visitor:      v,
		opt:          o,
		traceResults: map[ast_ifaces.ExprNode]string{},
	}
}

func (c *context) State() any {
	return c.state
}

func (c *context) Masker() secret_masker.Interface {
	return c.secretMasker
}

func (c *context) Trace() ast_ifaces.TraceWriter {
	return c.traceWriter
}

func (c *context) SetTraceResult(node ast_ifaces.ExprNode, result expression.Result) {
	if _, exist := c.traceResults[node]; exist {
		delete(c.traceResults, node)
	}
	value := formatValueFromResult(c.secretMasker, result)
	c.traceResults[node] = value
}

func (c *context) TryGetTraceResult(node ast_ifaces.ExprNode) (exist bool, result string) {
	result, exist = c.traceResults[node]
	return
}

func (c *context) Visitor() ast_ifaces.Visitor {
	return c.visitor
}

func formatValueFromResult(masker secret_masker.Interface, result expression.Result) string {
	return common.FormatValue(masker, result.Value(), result.Kind())
}
