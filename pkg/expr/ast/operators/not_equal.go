package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type NotEqual struct {
	base.Node
	base.Container
}

func (n *NotEqual) Value() any {
	panic("not implemented")
}

func (n *NotEqual) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitNotEqual(eCtx, n)
}

func (n *NotEqual) TraceFullyRealized() bool {
	return false
}

func (n *NotEqual) Expr() string {
	return fmt.Sprintf("%s != %s", n.Params[0].Expr(), n.Params[1].Expr())
}

func (n *NotEqual) RealizedExpr(eCtx interfaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	return fmt.Sprintf("%s != %s", n.Params[0].RealizedExpr(eCtx), n.Params[1].RealizedExpr(eCtx))
}

func (n *NotEqual) GetCtn() interfaces.Container {
	return n.Ctn
}

func (n *NotEqual) SetCtn(c interfaces.Container) {
	n.Ctn = c
}

func (n *NotEqual) GetLevel() (level int) {
	return n.Level
}

func (n *NotEqual) GetName() string {
	return n.Name
}

func (n *NotEqual) SetName(name string) {
	n.Name = name
}
