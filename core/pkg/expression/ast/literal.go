package ast

import (
	"github.com/dungdm93/drassi/core/pkg/expression"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/base"
	"github.com/dungdm93/drassi/core/pkg/expression/common"
)

type Lit struct {
	base.Node
	value any
	kind  expression.ResultKind
	name  string
}

func newLiteral(val any) *Lit {
	value, kind := common.ToCanonicalValue(val)
	return &Lit{
		value: value,
		kind:  kind,
		name:  kind.String(),
	}
}

func (l *Lit) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitLiteral(eCtx, l)
}

func (l *Lit) Value() any {
	return l.value
}

func (l *Lit) TraceFullyRealized() bool {
	return false
}

func (l *Lit) Expr() string {
	return common.FormatValue(nil, l.value, l.kind)
}

func (l *Lit) RealizedExpr(eCtx ast_ifaces.Context) string {
	return common.FormatValue(nil, l.value, l.kind)
}

func (l *Lit) GetCtn() ast_ifaces.Container {
	return l.Ctn
}

func (l *Lit) SetCtn(c ast_ifaces.Container) {
	l.Ctn = c
}

func (l *Lit) GetLevel() (level int) {
	return l.Level
}

func (l *Lit) GetName() string {
	return l.name
}

func (l *Lit) SetName(name string) {
	l.name = name
}
