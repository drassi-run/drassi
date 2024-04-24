package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type LessThan struct {
	base.Node
	base.Container
}

func (l *LessThan) Value() any {
	panic("not implemented")
}

func (l *LessThan) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitLessThan(eCtx, l)
}

func (l *LessThan) TraceFullyRealized() bool {
	return false
}

func (l *LessThan) Expr() string {
	return fmt.Sprintf("(%s < %s)", l.Params[0].Expr(), l.Params[1].Expr())
}

func (l *LessThan) RealizedExpr(eCtx interfaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(l)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s < %s)", l.Params[0].RealizedExpr(eCtx), l.Params[1].RealizedExpr(eCtx))
}

func (l *LessThan) GetCtn() interfaces.Container {
	return l.Ctn
}

func (l *LessThan) SetCtn(c interfaces.Container) {
	l.Ctn = c
}

func (l *LessThan) GetLevel() (level int) {
	return l.Level
}

func (l *LessThan) GetName() string {
	return l.Name
}

func (l *LessThan) SetName(name string) {
	l.Name = name
}
