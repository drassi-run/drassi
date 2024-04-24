package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type LessThanOrEqual struct {
	base.Node
	base.Container
}

func (l *LessThanOrEqual) Value() any {
	panic("not implemented")
}

func (l *LessThanOrEqual) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitLessThanOrEqual(eCtx, l)
}

func (l *LessThanOrEqual) TraceFullyRealized() bool {
	return false
}

func (l *LessThanOrEqual) Expr() string {
	return fmt.Sprintf("(%s <= %s)", l.Params[0].Expr(), l.Params[1].Expr())
}

func (l *LessThanOrEqual) RealizedExpr(eCtx interfaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(l)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s <= %s)", l.Params[0].RealizedExpr(eCtx), l.Params[1].RealizedExpr(eCtx))
}

func (l *LessThanOrEqual) GetCtn() interfaces.Container {
	return l.Ctn
}

func (l *LessThanOrEqual) SetCtn(c interfaces.Container) {
	l.Ctn = c
}

func (l *LessThanOrEqual) GetLevel() (level int) {
	return l.Level
}

func (l *LessThanOrEqual) GetName() string {
	return l.Name
}

func (l *LessThanOrEqual) SetName(name string) {
	l.Name = name
}
