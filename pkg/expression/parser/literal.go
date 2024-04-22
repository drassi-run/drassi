package parser

import (
	"github.com/dungdm93/drasi/pkg/expression/evaluator"
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type Literal struct {
	base.ExpressionNodeBase
	value any
	kind  interfaces.ValueKind
	name  string
}

func NewLiteral(val any) *Literal {
	value, kind, _ := evaluator.ConvertToCanonicalValue(val)
	return &Literal{
		value: value,
		kind:  kind,
		name:  kind.ToString(),
	}
}

func (a *Literal) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitLiteral(eCtx, a)
}

func (l *Literal) Value() any {
	return l.value
}

func (l *Literal) TraceFullyRealized() bool {
	return false
}

func (l *Literal) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
	return l.value
}

func (l *Literal) ConvertToExpression() string {
	return evaluator.FormatValue(nil, l.value, l.kind)
}

func (l *Literal) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	return evaluator.FormatValue(nil, l.value, l.kind)
}

func (l *Literal) GetContainer() interfaces.IContainer {
	return l.Container
}

func (l *Literal) setContainer(c interfaces.IContainer) {
	l.Container = c
}

func (l *Literal) GetLevel() (level int) {
	return l.Level
}

func (l *Literal) GetName() string {
	return l.name
}
func (l *Literal) SetName(name string) {
	l.name = name
}
