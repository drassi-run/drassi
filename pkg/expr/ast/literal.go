package ast

import (
	"github.com/dungdm93/drasi/pkg/expr"
	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/common"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type literal struct {
	base.Node
	value any
	kind  expr.ResultKind
	name  string
}

func newLiteral(val any) *literal {
	value, kind := common.ToCanonicalValue(val)
	return &literal{
		value: value,
		kind:  kind,
		name:  kind.ToString(),
	}
}

func (l *literal) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitLiteral(eCtx, l)
}

func (l *literal) Value() any {
	return l.value
}

func (l *literal) TraceFullyRealized() bool {
	return false
}

func (l *literal) Expr() string {
	return common.FormatValue(nil, l.value, l.kind)
}

func (l *literal) RealizedExpr(eCtx interfaces.Context) string {
	return common.FormatValue(nil, l.value, l.kind)
}

func (l *literal) GetCtn() interfaces.Container {
	return l.Ctn
}

func (l *literal) SetCtn(c interfaces.Container) {
	l.Ctn = c
}

func (l *literal) GetLevel() (level int) {
	return l.Level
}

func (l *literal) GetName() string {
	return l.name
}

func (l *literal) SetName(name string) {
	l.name = name
}
