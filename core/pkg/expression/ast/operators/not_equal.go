package operators

import (
	"fmt"

	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/base"
)

type NotEqual struct {
	base.Node
	base.Container
}

func (n *NotEqual) Value() any {
	panic("not implemented")
}

func (n *NotEqual) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitNotEqual(eCtx, n)
}

func (n *NotEqual) TraceFullyRealized() bool {
	return false
}

func (n *NotEqual) Expr() string {
	return fmt.Sprintf("%s != %s", n.Params[0].Expr(), n.Params[1].Expr())
}

func (n *NotEqual) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	return fmt.Sprintf("%s != %s", n.Params[0].RealizedExpr(eCtx), n.Params[1].RealizedExpr(eCtx))
}

func (n *NotEqual) GetCtn() ast_ifaces.Container {
	return n.Ctn
}

func (n *NotEqual) SetCtn(c ast_ifaces.Container) {
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
