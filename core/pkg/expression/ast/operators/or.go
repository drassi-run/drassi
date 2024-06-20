package operators

import (
	"fmt"

	"drassi.run/core/pkg/expression/ast/ast_ifaces"
	"drassi.run/core/pkg/expression/ast/base"
)

type Or struct {
	base.Node
	base.Container
}

func (o *Or) Value() any {
	panic("not implemented")
}

func (o *Or) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitOr(eCtx, o)
}

func (o *Or) TraceFullyRealized() bool {
	return false
}

func (o *Or) Expr() string {
	return fmt.Sprintf("%s || %s", o.Params[0].Expr(), o.Params[1].Expr())
}

func (o *Or) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(o)
	if exist {
		return result
	}
	return fmt.Sprintf("%s || %s", o.Params[0].RealizedExpr(eCtx), o.Params[1].RealizedExpr(eCtx))
}

func (o *Or) GetCtn() ast_ifaces.Container {
	return o.Ctn
}

func (o *Or) SetCtn(c ast_ifaces.Container) {
	o.Ctn = c
}

func (o *Or) GetLevel() (level int) {
	return o.Level
}

func (o *Or) GetName() string {
	return o.Name
}

func (o *Or) SetName(name string) {
	o.Name = name
}
