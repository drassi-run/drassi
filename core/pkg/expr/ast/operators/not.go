package operators

import (
	"fmt"

	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expr/ast/base"
)

type Not struct {
	base.Node
	base.Container
}

func (n *Not) Value() any {
	panic("not implemented")
}

func (n *Not) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitNot(eCtx, n)
}

func (n *Not) TraceFullyRealized() bool {
	return false
}

func (n *Not) Expr() string {
	return fmt.Sprintf("!%s", n.Params[0].Expr())
}

func (n *Not) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	return fmt.Sprintf("!%s", n.Params[0].RealizedExpr(eCtx))
}

func (n *Not) GetCtn() ast_ifaces.Container {
	return n.Ctn
}

func (n *Not) SetCtn(c ast_ifaces.Container) {
	n.Ctn = c
}

func (n *Not) GetLevel() (level int) {
	return n.Level
}

func (n *Not) GetName() string {
	return n.Name
}

func (n *Not) SetName(name string) {
	n.Name = name
}
