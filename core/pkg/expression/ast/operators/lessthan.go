package operators

import (
	"fmt"

	"drassi.run/core/pkg/expression/ast/ast_ifaces"
	"drassi.run/core/pkg/expression/ast/base"
)

type LessThan struct {
	base.Node
	base.Container
}

func (l *LessThan) Value() any {
	panic("not implemented")
}

func (l *LessThan) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitLessThan(eCtx, l)
}

func (l *LessThan) TraceFullyRealized() bool {
	return false
}

func (l *LessThan) Expr() string {
	return fmt.Sprintf("(%s < %s)", l.Params[0].Expr(), l.Params[1].Expr())
}

func (l *LessThan) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(l)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s < %s)", l.Params[0].RealizedExpr(eCtx), l.Params[1].RealizedExpr(eCtx))
}

func (l *LessThan) GetCtn() ast_ifaces.Container {
	return l.Ctn
}

func (l *LessThan) SetCtn(c ast_ifaces.Container) {
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
