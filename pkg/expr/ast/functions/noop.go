package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type NoOp struct {
	base.Node
	Fn
}

func (n *NoOp) Value() any {
	panic("not implemented")
}

func (n *NoOp) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitNoopFn(eCtx, n)
}

func (n *NoOp) TraceFullyRealized() bool {
	return false
}

func (n *NoOp) GetCtn() interfaces.Container {
	return n.Ctn
}

func (n *NoOp) SetCtn(c interfaces.Container) {
	n.Ctn = c
}

func (n *NoOp) GetLevel() (level int) {
	return n.Level
}

func (n *NoOp) GetName() string {
	return n.Name
}
func (n *NoOp) SetName(name string) {
	n.Name = name
}

func (n *NoOp) Expr() string {
	params := make([]string, len(n.Parameters()))
	for i, param := range n.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", n.GetName(), strings.Join(params, ", "))
}

func (n *NoOp) RealizedExpr(eCtx interfaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	params := make([]string, len(n.Parameters()))
	for i, param := range n.Parameters() {
		params[i] = param.RealizedExpr(eCtx)
	}
	return fmt.Sprintf("%s(%s)", n.GetName(), strings.Join(params, ", "))
}
