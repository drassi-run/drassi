package evaluator

import (
	"github.com/dungdm93/drasi/pkg/expr"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
	"github.com/dungdm93/drasi/pkg/secret_masker"
)

type (
	context struct {
		state        any
		opt          *Option
		traceWriter  interfaces.TraceWriter
		secretMasker secret_masker.Interface
		traceResults map[interfaces.Node]string
		visitor      interfaces.Visitor
	}
)

func newContext(t interfaces.TraceWriter, sm secret_masker.Interface, s any, o *Option, v interfaces.Visitor) *context {
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
		traceResults: map[interfaces.Node]string{},
	}
}

func (c *context) State() any {
	return c.state
}

func (c *context) Masker() secret_masker.Interface {
	return c.secretMasker
}

func (c *context) Trace() interfaces.TraceWriter {
	return c.traceWriter
}

func (c *context) SetTraceResult(node interfaces.Node, result expr.Result) {
	if _, exist := c.traceResults[node]; exist {
		delete(c.traceResults, node)
	}
	value := formatValueFromResult(c.secretMasker, result)
	c.traceResults[node] = value
}

func (c *context) TryGetTraceResult(node interfaces.Node) (exist bool, result string) {
	result, exist = c.traceResults[node]
	return
}

func (c *context) Visitor() interfaces.Visitor {
	return c.visitor
}
