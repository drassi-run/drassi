package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expr/ast/base"
)

type NoOp struct {
	base.Node
	Fn
}

func (n *NoOp) Value() any {
	panic("not implemented")
}

func (n *NoOp) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitNoopFn(eCtx, n)
}

func (n *NoOp) TraceFullyRealized() bool {
	return false
}

func (n *NoOp) GetCtn() ast_ifaces.Container {
	return n.Ctn
}

func (n *NoOp) SetCtn(c ast_ifaces.Container) {
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

func (n *NoOp) RealizedExpr(eCtx ast_ifaces.Context) string {
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
