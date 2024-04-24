package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type Not struct {
	base.Node
	base.Container
}

func (n *Not) Value() any {
	panic("not implemented")
}

func (n *Not) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitNot(eCtx, n)
}

func (n *Not) TraceFullyRealized() bool {
	return false
}

func (n *Not) Expr() string {
	return fmt.Sprintf("!%s", n.Params[0].Expr())
}

func (n *Not) RealizedExpr(eCtx interfaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	return fmt.Sprintf("!%s", n.Params[0].RealizedExpr(eCtx))
}

func (n *Not) GetCtn() interfaces.Container {
	return n.Ctn
}

func (n *Not) SetCtn(c interfaces.Container) {
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
