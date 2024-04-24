package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type Or struct {
	base.Node
	base.Container
}

func (o *Or) Value() any {
	panic("not implemented")
}

func (o *Or) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitOr(eCtx, o)
}

func (o *Or) TraceFullyRealized() bool {
	return false
}

func (o *Or) Expr() string {
	return fmt.Sprintf("%s || %s", o.Params[0].Expr(), o.Params[1].Expr())
}

func (o *Or) RealizedExpr(eCtx interfaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(o)
	if exist {
		return result
	}
	return fmt.Sprintf("%s || %s", o.Params[0].RealizedExpr(eCtx), o.Params[1].RealizedExpr(eCtx))
}

func (o *Or) GetCtn() interfaces.Container {
	return o.Ctn
}

func (o *Or) SetCtn(c interfaces.Container) {
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
